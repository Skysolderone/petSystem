import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiRequest } from "@/services/api";
import type { CreateDevicePayload, Device, DeviceCommandResponse, DeviceDataPoint, DeviceStatus } from "@/types/device";

export function useDevicesList() {
  return useQuery({
    queryKey: ["devices"],
    queryFn: async () => {
      const response = await apiRequest<Device[]>("/devices?page=1&page_size=20");
      return response.data;
    },
    refetchInterval: 30_000,
  });
}

export function useCreateDevice() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: CreateDevicePayload) => {
      const response = await apiRequest<Device>("/devices", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      return response.data;
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["devices"] });
    },
  });
}

export function useDeviceStatus(deviceId?: string) {
  return useQuery({
    queryKey: ["device-status", deviceId],
    enabled: Boolean(deviceId),
    queryFn: async () => {
      const response = await apiRequest<DeviceStatus>(`/devices/${deviceId}/status`);
      return response.data;
    },
    refetchInterval: 20_000,
  });
}

export function useDeviceData(deviceId?: string, metric?: string) {
  return useQuery({
    queryKey: ["device-data", deviceId, metric ?? "all"],
    enabled: Boolean(deviceId),
    queryFn: async () => {
      const query = metric ? `?metric=${encodeURIComponent(metric)}&hours=24&limit=24` : "?hours=24&limit=24";
      const response = await apiRequest<DeviceDataPoint[]>(`/devices/${deviceId}/data${query}`);
      return response.data;
    },
    refetchInterval: 20_000,
  });
}

export function useSendDeviceCommand(deviceId?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (command: { command: string; params?: Record<string, unknown> }) => {
      if (!deviceId) {
        throw new Error("deviceId is required");
      }

      const response = await apiRequest<DeviceCommandResponse>(`/devices/${deviceId}/command`, {
        method: "POST",
        body: JSON.stringify(command),
      });
      return response.data;
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["devices"] }),
        queryClient.invalidateQueries({ queryKey: ["device-status", deviceId] }),
        queryClient.invalidateQueries({ queryKey: ["device-data", deviceId] }),
      ]);
    },
  });
}

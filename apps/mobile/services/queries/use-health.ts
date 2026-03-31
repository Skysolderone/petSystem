import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiRequest } from "@/services/api";
import type { CreateHealthRecordPayload, HealthAIAnswer, HealthAlert, HealthRecord, HealthSummary } from "@/types/health";

export function useHealthRecords(petId?: string) {
  return useQuery({
    queryKey: ["health-records", petId],
    enabled: Boolean(petId),
    queryFn: async () => {
      const response = await apiRequest<HealthRecord[]>(`/pets/${petId}/health?page=1&page_size=10`);
      return response.data;
    },
  });
}

export function useHealthSummary(petId?: string) {
  return useQuery({
    queryKey: ["health-summary", petId],
    enabled: Boolean(petId),
    queryFn: async () => {
      const response = await apiRequest<HealthSummary>(`/pets/${petId}/health/summary`);
      return response.data;
    },
  });
}

export function useHealthAlerts(petId?: string) {
  return useQuery({
    queryKey: ["health-alerts", petId],
    enabled: Boolean(petId),
    queryFn: async () => {
      const response = await apiRequest<HealthAlert[]>(`/pets/${petId}/health/alerts`);
      return response.data;
    },
  });
}

export function useCreateHealthRecord(petId?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: CreateHealthRecordPayload) => {
      if (!petId) {
        throw new Error("petId is required");
      }
      const response = await apiRequest<HealthRecord>(`/pets/${petId}/health`, {
        method: "POST",
        body: JSON.stringify(payload),
      });
      return response.data;
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["health-records", petId] }),
        queryClient.invalidateQueries({ queryKey: ["health-summary", petId] }),
        queryClient.invalidateQueries({ queryKey: ["health-alerts", petId] }),
      ]);
    },
  });
}

export function useAskHealthAI(petId?: string) {
  return useMutation({
    mutationFn: async (question: string) => {
      if (!petId) {
        throw new Error("petId is required");
      }
      const response = await apiRequest<HealthAIAnswer>(`/pets/${petId}/health/ask-ai`, {
        method: "POST",
        body: JSON.stringify({ question }),
      });
      return response.data;
    },
  });
}

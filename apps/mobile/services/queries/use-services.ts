import { useQuery } from "@tanstack/react-query";

import { apiRequest } from "@/services/api";
import type { Booking } from "@/types/booking";
import type { ServiceAvailability, ServiceProvider } from "@/types/service";

const defaultLocation = {
  lat: 31.2304,
  lng: 121.4737,
};

export function useServicesList(type?: string) {
  return useQuery({
    queryKey: ["services", type ?? "all"],
    queryFn: async () => {
      const query = new URLSearchParams({
        lat: String(defaultLocation.lat),
        lng: String(defaultLocation.lng),
      });
      if (type) {
        query.set("type", type);
      }
      const response = await apiRequest<ServiceProvider[]>(`/services?${query.toString()}`);
      return response.data;
    },
  });
}

export function useServiceDetail(serviceId?: string) {
  return useQuery({
    queryKey: ["service-detail", serviceId],
    enabled: Boolean(serviceId),
    queryFn: async () => {
      const response = await apiRequest<ServiceProvider>(`/services/${serviceId}`);
      return response.data;
    },
  });
}

export function useServiceAvailability(serviceId?: string) {
  return useQuery({
    queryKey: ["service-availability", serviceId],
    enabled: Boolean(serviceId),
    queryFn: async () => {
      const response = await apiRequest<ServiceAvailability>(`/services/${serviceId}/availability`);
      return response.data;
    },
  });
}

export function useServiceReviews(serviceId?: string) {
  return useQuery({
    queryKey: ["service-reviews", serviceId],
    enabled: Boolean(serviceId),
    queryFn: async () => {
      const response = await apiRequest<Booking[]>(`/services/${serviceId}/reviews`);
      return response.data;
    },
  });
}

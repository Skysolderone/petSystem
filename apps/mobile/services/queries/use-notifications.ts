import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiRequest } from "@/services/api";
import { useNotificationStore } from "@/stores/notification-store";
import type {
  NotificationItem,
  NotificationReadAllResponse,
  NotificationReadResponse,
  PushTokenPayload,
  PushTokenResponse,
} from "@/types/notification";

export function useNotificationsList() {
  return useQuery({
    queryKey: ["notifications"],
    queryFn: async () => {
      const response = await apiRequest<NotificationItem[]>("/notifications?page=1&page_size=20");
      return response.data;
    },
  });
}

export function useMarkNotificationRead(notificationId?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const response = await apiRequest<NotificationReadResponse>(`/notifications/${notificationId}/read`, {
        method: "PUT",
      });
      return response.data;
    },
    onSuccess: async (data) => {
      useNotificationStore.getState().setUnreadCount(data.unread_count);
      await queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });
}

export function useMarkAllNotificationsRead() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const response = await apiRequest<NotificationReadAllResponse>("/notifications/read-all", {
        method: "PUT",
      });
      return response.data;
    },
    onSuccess: async (data) => {
      useNotificationStore.getState().setUnreadCount(data.unread_count);
      await queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });
}

export async function registerPushToken(payload: PushTokenPayload) {
  const response = await apiRequest<PushTokenResponse>("/notifications/push-token", {
    method: "POST",
    body: JSON.stringify(payload),
  });
  return response.data;
}

export async function unregisterPushToken(token: string) {
  const response = await apiRequest<{ deleted: boolean }>("/notifications/push-token", {
    method: "DELETE",
    body: JSON.stringify({ token }),
  });
  return response.data;
}

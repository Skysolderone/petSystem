import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiRequest } from "@/services/api";
import { useAuthStore } from "@/stores/auth-store";
import type { User } from "@/types/auth";

interface UpdateUserPayload {
  nickname?: string;
  email?: string;
}

interface UpdateLocationPayload {
  latitude?: number;
  longitude?: number;
}

interface UploadAssetPayload {
  uri: string;
  name?: string;
  type?: string;
}

export function useUserMe() {
  const session = useAuthStore((state) => state.session);

  return useQuery({
    queryKey: ["me"],
    enabled: Boolean(session?.accessToken),
    queryFn: async () => {
      const response = await apiRequest<User>("/users/me");
      return response.data;
    },
  });
}

export function useUpdateUserProfile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: UpdateUserPayload) => {
      const response = await apiRequest<User>("/users/me", {
        method: "PUT",
        body: JSON.stringify(payload),
      });
      return response.data;
    },
    onSuccess: async (user) => {
      await useAuthStore.getState().updateUser(user);
      await queryClient.invalidateQueries({ queryKey: ["me"] });
    },
  });
}

export function useUpdateUserLocation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: UpdateLocationPayload) => {
      const response = await apiRequest<User>("/users/me/location", {
        method: "PUT",
        body: JSON.stringify(payload),
      });
      return response.data;
    },
    onSuccess: async (user) => {
      await useAuthStore.getState().updateUser(user);
      await queryClient.invalidateQueries({ queryKey: ["me"] });
    },
  });
}

export function useUploadUserAvatar() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ uri, name = "avatar.jpg", type = "image/jpeg" }: UploadAssetPayload) => {
      const formData = new FormData();
      formData.append("file", { uri, name, type } as never);

      const response = await apiRequest<User>("/users/me/avatar", {
        method: "PUT",
        body: formData,
      });
      return response.data;
    },
    onSuccess: async (user) => {
      await useAuthStore.getState().updateUser(user);
      await queryClient.invalidateQueries({ queryKey: ["me"] });
    },
  });
}

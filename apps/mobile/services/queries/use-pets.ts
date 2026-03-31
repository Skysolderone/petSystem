import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiRequest } from "@/services/api";
import type { CreatePetPayload, Pet } from "@/types/pet";

interface PetsPage {
  items: Pet[];
  nextPage?: number;
}

export function usePetsList() {
  return useInfiniteQuery({
    queryKey: ["pets"],
    initialPageParam: 1,
    queryFn: async ({ pageParam }) => {
      const response = await apiRequest<Pet[]>(`/pets?page=${pageParam}&page_size=20`);
      const nextPage =
        response.meta?.page && response.meta?.page_size && response.meta?.total
          ? response.meta.page * response.meta.page_size < response.meta.total
            ? response.meta.page + 1
            : undefined
          : undefined;

      return {
        items: response.data,
        nextPage,
      } satisfies PetsPage;
    },
    getNextPageParam: (lastPage) => lastPage.nextPage,
  });
}

export function useCreatePet() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: CreatePetPayload) => {
      const response = await apiRequest<Pet>("/pets", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      return response.data;
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["pets"] });
    },
  });
}

export function usePetDetail(petId?: string) {
  return useQuery({
    queryKey: ["pet-detail", petId],
    enabled: Boolean(petId),
    queryFn: async () => {
      const response = await apiRequest<Pet>(`/pets/${petId}`);
      return response.data;
    },
  });
}

export function useUpdatePet(petId?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: Partial<CreatePetPayload>) => {
      const response = await apiRequest<Pet>(`/pets/${petId}`, {
        method: "PUT",
        body: JSON.stringify(payload),
      });
      return response.data;
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["pets"] }),
        queryClient.invalidateQueries({ queryKey: ["pet-detail", petId] }),
      ]);
    },
  });
}

export function useUploadPetAvatar(petId?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ uri, name = "pet-avatar.jpg", type = "image/jpeg" }: { uri: string; name?: string; type?: string }) => {
      const formData = new FormData();
      formData.append("file", { uri, name, type } as never);

      const response = await apiRequest<Pet>(`/pets/${petId}/avatar`, {
        method: "PUT",
        body: formData,
      });
      return response.data;
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["pets"] }),
        queryClient.invalidateQueries({ queryKey: ["pet-detail", petId] }),
      ]);
    },
  });
}

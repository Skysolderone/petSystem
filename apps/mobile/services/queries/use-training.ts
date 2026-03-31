import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiRequest } from "@/services/api";
import type {
  CreateTrainingPlanPayload,
  GenerateTrainingPlanPayload,
  TrainingPlan,
  UpdateTrainingPlanPayload,
} from "@/types/training";

export function useTrainingPlans(petId?: string) {
  return useQuery({
    queryKey: ["training-plans", petId],
    enabled: Boolean(petId),
    queryFn: async () => {
      const response = await apiRequest<TrainingPlan[]>(`/pets/${petId}/training`);
      return response.data;
    },
  });
}

export function useTrainingPlanDetail(planId?: string) {
  return useQuery({
    queryKey: ["training-plan", planId],
    enabled: Boolean(planId),
    queryFn: async () => {
      const response = await apiRequest<TrainingPlan>(`/training/${planId}`);
      return response.data;
    },
  });
}

export function useCreateTrainingPlan(petId?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: CreateTrainingPlanPayload) => {
      const response = await apiRequest<TrainingPlan>(`/pets/${petId}/training`, {
        method: "POST",
        body: JSON.stringify(payload),
      });
      return response.data;
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["training-plans", petId] }),
        queryClient.invalidateQueries({ queryKey: ["notifications"] }),
      ]);
    },
  });
}

export function useGenerateTrainingPlan(petId?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: GenerateTrainingPlanPayload) => {
      const response = await apiRequest<TrainingPlan>(`/pets/${petId}/training/generate`, {
        method: "POST",
        body: JSON.stringify(payload),
      });
      return response.data;
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["training-plans", petId] }),
        queryClient.invalidateQueries({ queryKey: ["notifications"] }),
      ]);
    },
  });
}

export function useUpdateTrainingPlan(planId?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: UpdateTrainingPlanPayload) => {
      const response = await apiRequest<TrainingPlan>(`/training/${planId}`, {
        method: "PUT",
        body: JSON.stringify(payload),
      });
      return response.data;
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["training-plan", planId] }),
        queryClient.invalidateQueries({ queryKey: ["training-plans"] }),
      ]);
    },
  });
}

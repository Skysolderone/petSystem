import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiRequest } from "@/services/api";
import type { Comment, CommunityAIAnswer, CreatePostPayload, Post } from "@/types/community";

export function useCommunityFeed() {
  return useQuery({
    queryKey: ["community-feed"],
    queryFn: async () => {
      const response = await apiRequest<Post[]>("/posts?page=1&page_size=20");
      return response.data;
    },
  });
}

export function useCreatePost() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: CreatePostPayload) => {
      const response = await apiRequest<Post>("/posts", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      return response.data;
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["community-feed"] });
    },
  });
}

export function useTogglePostLike(postId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      await apiRequest(`/posts/${postId}/like`, {
        method: "POST",
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["community-feed"] });
    },
  });
}

export function usePostComments(postId: string) {
  return useQuery({
    queryKey: ["post-comments", postId],
    queryFn: async () => {
      const response = await apiRequest<Comment[]>(`/posts/${postId}/comments`);
      return response.data;
    },
  });
}

export function useCreateComment(postId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (content: string) => {
      const response = await apiRequest<Comment>(`/posts/${postId}/comments`, {
        method: "POST",
        body: JSON.stringify({ content }),
      });
      return response.data;
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["community-feed"] }),
        queryClient.invalidateQueries({ queryKey: ["post-comments", postId] }),
      ]);
    },
  });
}

export function useAskCommunityAI() {
  return useMutation({
    mutationFn: async (question: string) => {
      const response = await apiRequest<CommunityAIAnswer>("/community/ask-ai", {
        method: "POST",
        body: JSON.stringify({ question }),
      });
      return response.data;
    },
  });
}

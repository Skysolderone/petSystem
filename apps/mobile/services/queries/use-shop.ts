import { useQuery } from "@tanstack/react-query";

import { apiRequest } from "@/services/api";
import type { Product } from "@/types/shop";

export function useShopProducts(category?: string, query?: string) {
  return useQuery({
    queryKey: ["shop-products", category ?? "all", query ?? ""],
    queryFn: async () => {
      const search = new URLSearchParams();
      if (category) {
        search.set("category", category);
      }
      if (query) {
        search.set("q", query);
      }
      const suffix = search.toString() ? `?${search.toString()}` : "";
      const response = await apiRequest<Product[]>(`/shop/products${suffix}`);
      return response.data;
    },
  });
}

export function useShopRecommendations(petId?: string) {
  return useQuery({
    queryKey: ["shop-recommendations", petId],
    enabled: Boolean(petId),
    queryFn: async () => {
      const response = await apiRequest<Product[]>(`/shop/recommendations/${petId}`);
      return response.data;
    },
  });
}

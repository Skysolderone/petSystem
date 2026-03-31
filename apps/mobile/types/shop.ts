export interface Product {
  id: string;
  provider_id?: string;
  name: string;
  description: string;
  category: string;
  price: number;
  currency: string;
  images: string[];
  pet_species: string[];
  tags: string[];
  external_url?: string;
  rating: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  recommended_reason?: string;
}

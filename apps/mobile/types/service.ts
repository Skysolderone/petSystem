export interface ServiceProvider {
  id: string;
  user_id: string;
  name: string;
  type: string;
  description: string;
  address: string;
  latitude: number;
  longitude: number;
  phone: string;
  photos: string[];
  rating: number;
  review_count: number;
  is_verified: boolean;
  open_hours: Record<string, unknown>;
  services: Array<Record<string, unknown>>;
  tags: string[];
  created_at: string;
  updated_at: string;
  distance_km?: number;
}

export interface ServiceAvailability {
  provider_id: string;
  slots: string[];
}

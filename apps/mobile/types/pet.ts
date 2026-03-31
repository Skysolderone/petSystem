export interface Pet {
  id: string;
  owner_id: string;
  name: string;
  species: string;
  breed: string;
  gender: string;
  birth_date?: string;
  weight?: number;
  avatar_url: string;
  microchip?: string;
  is_neutered: boolean;
  allergies: string[];
  notes: string;
  health_score?: number;
  created_at: string;
  updated_at: string;
}

export interface CreatePetPayload {
  name: string;
  species: string;
  breed?: string;
  gender?: string;
  weight?: number;
  avatar_url?: string;
  is_neutered?: boolean;
  allergies?: string[];
  notes?: string;
}

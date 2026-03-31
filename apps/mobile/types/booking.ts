export interface Booking {
  id: string;
  user_id: string;
  pet_id: string;
  provider_id: string;
  service_name: string;
  status: string;
  start_time: string;
  end_time?: string;
  price: number;
  currency: string;
  notes: string;
  cancel_reason: string;
  rating?: number;
  review: string;
  created_at: string;
  updated_at: string;
}

export interface CreateBookingPayload {
  provider_id: string;
  pet_id: string;
  service_name: string;
  start_time: string;
  price: number;
  notes?: string;
}

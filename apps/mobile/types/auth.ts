export interface User {
  id: string;
  phone?: string;
  email?: string;
  nickname: string;
  avatar_url: string;
  latitude?: number;
  longitude?: number;
  role: string;
  plan_type: string;
  plan_expiry?: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface AuthSession {
  user: User;
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

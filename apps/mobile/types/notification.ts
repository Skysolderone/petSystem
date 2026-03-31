export interface NotificationItem {
  id: string;
  user_id: string;
  type: string;
  title: string;
  body: string;
  data: Record<string, unknown>;
  is_read: boolean;
  created_at: string;
  updated_at: string;
}

export interface NotificationStreamMessage {
  type: string;
  timestamp: string;
  notification?: NotificationItem;
  notifications?: NotificationItem[];
  unread_count?: number;
}

export interface NotificationReadResponse {
  id: string;
  read: boolean;
  unread_count: number;
}

export interface NotificationReadAllResponse {
  updated: number;
  unread_count: number;
}

export interface PushTokenPayload {
  token: string;
  provider?: string;
  platform?: string;
}

export interface PushTokenResponse {
  token: string;
  provider: string;
  platform: string;
  is_active: boolean;
  last_seen_at: string;
}

export interface HealthRecord {
  id: string;
  pet_id: string;
  type: string;
  title: string;
  description: string;
  data: Record<string, unknown>;
  due_date?: string;
  attachments: string[];
  recorded_at: string;
  created_at: string;
  updated_at: string;
}

export interface HealthAlert {
  id: string;
  pet_id: string;
  alert_type: string;
  severity: string;
  title: string;
  message: string;
  source: string;
  is_read: boolean;
  is_dismissed: boolean;
  created_at: string;
}

export interface HealthSummary {
  score: number;
  status: "excellent" | "stable" | "watch" | "critical" | string;
  insights: string[];
  recommended_actions: string[];
  data_points_analyzed: number;
  generated_at: string;
}

export interface HealthAIAnswer {
  question: string;
  answer: string;
  created_at: string;
}

export interface CreateHealthRecordPayload {
  type: string;
  title: string;
  description?: string;
  data?: Record<string, unknown>;
  due_date?: string;
  attachments?: string[];
  recorded_at?: string;
}

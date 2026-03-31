export interface TrainingStep {
  day?: number;
  title?: string;
  instruction?: string;
  duration_minutes?: number;
  category?: string;
  difficulty?: string;
  [key: string]: unknown;
}

export interface TrainingPlan {
  id: string;
  pet_id: string;
  title: string;
  description: string;
  difficulty: string;
  category: string;
  steps: TrainingStep[];
  ai_generated: boolean;
  progress: number;
  created_at: string;
  updated_at: string;
}

export interface CreateTrainingPlanPayload {
  title: string;
  description?: string;
  difficulty?: string;
  category?: string;
  steps?: TrainingStep[];
}

export interface GenerateTrainingPlanPayload {
  goal: string;
  difficulty?: string;
  category?: string;
}

export interface UpdateTrainingPlanPayload {
  title?: string;
  description?: string;
  difficulty?: string;
  category?: string;
  steps?: TrainingStep[];
  progress?: number;
}

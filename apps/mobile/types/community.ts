export interface Post {
  id: string;
  author_id: string;
  pet_id?: string;
  type: string;
  title: string;
  content: string;
  images: string[];
  tags: string[];
  latitude?: number;
  longitude?: number;
  like_count: number;
  comment_count: number;
  is_published: boolean;
  created_at: string;
  updated_at: string;
}

export interface Comment {
  id: string;
  post_id: string;
  author_id: string;
  parent_id?: string;
  content: string;
  like_count: number;
  created_at: string;
  updated_at: string;
}

export interface CreatePostPayload {
  pet_id?: string;
  type?: string;
  title?: string;
  content: string;
  tags?: string[];
}

export interface CommunityAIAnswer {
  question: string;
  answer: string;
  created_at: string;
}

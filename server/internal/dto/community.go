package dto

import (
	"time"

	"petverse/server/internal/model"
)

type CreatePostRequest struct {
	PetID     *string  `json:"pet_id"`
	Type      string   `json:"type"`
	Title     string   `json:"title"`
	Content   string   `json:"content" binding:"required,min=1,max=2000"`
	Images    []string `json:"images"`
	Tags      []string `json:"tags"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type UpdatePostRequest struct {
	Title     *string  `json:"title"`
	Content   *string  `json:"content"`
	Images    []string `json:"images"`
	Tags      []string `json:"tags"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type CreateCommentRequest struct {
	ParentID *string `json:"parent_id"`
	Content  string  `json:"content" binding:"required,min=1,max=1000"`
}

type AskCommunityAIRequest struct {
	Question string `json:"question" binding:"required,min=3,max=500"`
}

type CommunityAIResponse struct {
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	CreatedAt time.Time `json:"created_at"`
}

type PostResponse struct {
	ID           string    `json:"id"`
	AuthorID     string    `json:"author_id"`
	PetID        *string   `json:"pet_id,omitempty"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Images       []string  `json:"images"`
	Tags         []string  `json:"tags"`
	Latitude     *float64  `json:"latitude,omitempty"`
	Longitude    *float64  `json:"longitude,omitempty"`
	LikeCount    int       `json:"like_count"`
	CommentCount int       `json:"comment_count"`
	IsPublished  bool      `json:"is_published"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CommentResponse struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	AuthorID  string    `json:"author_id"`
	ParentID  *string   `json:"parent_id,omitempty"`
	Content   string    `json:"content"`
	LikeCount int       `json:"like_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ToggleLikeResponse struct {
	Liked     bool `json:"liked"`
	LikeCount int  `json:"like_count"`
}

func ToPostResponse(post *model.Post) PostResponse {
	var petID *string
	if post.PetID != nil {
		value := post.PetID.String()
		petID = &value
	}

	return PostResponse{
		ID:           post.ID.String(),
		AuthorID:     post.AuthorID.String(),
		PetID:        petID,
		Type:         post.Type,
		Title:        post.Title,
		Content:      post.Content,
		Images:       decodeStringArray(post.Images),
		Tags:         decodeStringArray(post.Tags),
		Latitude:     post.Latitude,
		Longitude:    post.Longitude,
		LikeCount:    post.LikeCount,
		CommentCount: post.CommentCount,
		IsPublished:  post.IsPublished,
		CreatedAt:    post.CreatedAt,
		UpdatedAt:    post.UpdatedAt,
	}
}

func ToCommentResponse(comment *model.Comment) CommentResponse {
	var parentID *string
	if comment.ParentID != nil {
		value := comment.ParentID.String()
		parentID = &value
	}
	return CommentResponse{
		ID:        comment.ID.String(),
		PostID:    comment.PostID.String(),
		AuthorID:  comment.AuthorID.String(),
		ParentID:  parentID,
		Content:   comment.Content,
		LikeCount: comment.LikeCount,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}
}

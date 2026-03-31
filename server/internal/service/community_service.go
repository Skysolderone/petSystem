package service

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"petverse/server/internal/dto"
	"petverse/server/internal/model"
	"petverse/server/internal/pkg/ai"
	"petverse/server/internal/pkg/apperror"
	"petverse/server/internal/pkg/pagination"
)

type CommunityService struct {
	posts communityRepository
	pets  communityPetRepository
	llm   *ai.Assistant
}

type communityRepository interface {
	CreatePost(ctx context.Context, post *model.Post) error
	ListPosts(ctx context.Context, page, pageSize int, tag string) ([]model.Post, int64, error)
	GetPostByID(ctx context.Context, id uuid.UUID) (*model.Post, error)
	UpdatePost(ctx context.Context, post *model.Post) error
	DeletePost(ctx context.Context, post *model.Post) error
	ToggleLike(ctx context.Context, userID, postID uuid.UUID) (bool, int, error)
	ListComments(ctx context.Context, postID uuid.UUID) ([]model.Comment, error)
	CreateComment(ctx context.Context, comment *model.Comment) error
	GetCommentByID(ctx context.Context, id uuid.UUID) (*model.Comment, error)
	DeleteComment(ctx context.Context, comment *model.Comment) error
}

type communityPetRepository interface {
	GetByIDAndOwner(ctx context.Context, id, ownerID uuid.UUID) (*model.Pet, error)
}

type CommunityServiceOption func(*CommunityService)

func WithCommunityAssistant(assistant *ai.Assistant) CommunityServiceOption {
	return func(service *CommunityService) {
		service.llm = assistant
	}
}

func NewCommunityService(posts communityRepository, pets communityPetRepository, options ...CommunityServiceOption) *CommunityService {
	service := &CommunityService{posts: posts, pets: pets}
	for _, option := range options {
		option(service)
	}
	return service
}

func (s *CommunityService) ListPosts(ctx context.Context, page, pageSize int, tag string) ([]model.Post, int64, int, int, error) {
	page, pageSize = pagination.Normalize(page, pageSize)
	posts, total, err := s.posts.ListPosts(ctx, page, pageSize, tag)
	if err != nil {
		return nil, 0, 0, 0, apperror.Wrap(http.StatusInternalServerError, "list_posts_failed", "failed to load posts", err)
	}
	return posts, total, page, pageSize, nil
}

func (s *CommunityService) CreatePost(ctx context.Context, userID uuid.UUID, req dto.CreatePostRequest) (*model.Post, error) {
	var petID *uuid.UUID
	if req.PetID != nil && *req.PetID != "" {
		parsedPetID, err := uuid.Parse(*req.PetID)
		if err != nil {
			return nil, apperror.New(http.StatusBadRequest, "invalid_pet_id", "pet id is invalid")
		}
		pet, err := s.pets.GetByIDAndOwner(ctx, parsedPetID, userID)
		if err != nil {
			return nil, apperror.Wrap(http.StatusInternalServerError, "pet_lookup_failed", "failed to load pet", err)
		}
		if pet == nil {
			return nil, apperror.New(http.StatusNotFound, "pet_not_found", "pet not found")
		}
		petID = &parsedPetID
	}

	postType := req.Type
	if postType == "" {
		postType = "post"
	}

	post := &model.Post{
		AuthorID:    userID,
		PetID:       petID,
		Type:        postType,
		Title:       req.Title,
		Content:     req.Content,
		Images:      datatypes.JSON(dto.EncodeStringArray(req.Images)),
		Tags:        datatypes.JSON(dto.EncodeStringArray(req.Tags)),
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		IsPublished: true,
	}
	if err := s.posts.CreatePost(ctx, post); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "create_post_failed", "failed to create post", err)
	}
	return post, nil
}

func (s *CommunityService) GetPost(ctx context.Context, postID uuid.UUID) (*model.Post, error) {
	post, err := s.posts.GetPostByID(ctx, postID)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "get_post_failed", "failed to load post", err)
	}
	if post == nil {
		return nil, apperror.New(http.StatusNotFound, "post_not_found", "post not found")
	}
	return post, nil
}

func (s *CommunityService) UpdatePost(ctx context.Context, userID, postID uuid.UUID, req dto.UpdatePostRequest) (*model.Post, error) {
	post, err := s.GetPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	if post.AuthorID != userID {
		return nil, apperror.New(http.StatusForbidden, "forbidden", "cannot edit another user's post")
	}

	if req.Title != nil {
		post.Title = *req.Title
	}
	if req.Content != nil {
		post.Content = *req.Content
	}
	if req.Images != nil {
		post.Images = datatypes.JSON(dto.EncodeStringArray(req.Images))
	}
	if req.Tags != nil {
		post.Tags = datatypes.JSON(dto.EncodeStringArray(req.Tags))
	}
	if req.Latitude != nil {
		post.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		post.Longitude = req.Longitude
	}

	if err := s.posts.UpdatePost(ctx, post); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "update_post_failed", "failed to update post", err)
	}
	return post, nil
}

func (s *CommunityService) DeletePost(ctx context.Context, userID, postID uuid.UUID) error {
	post, err := s.GetPost(ctx, postID)
	if err != nil {
		return err
	}
	if post.AuthorID != userID {
		return apperror.New(http.StatusForbidden, "forbidden", "cannot delete another user's post")
	}
	if err := s.posts.DeletePost(ctx, post); err != nil {
		return apperror.Wrap(http.StatusInternalServerError, "delete_post_failed", "failed to delete post", err)
	}
	return nil
}

func (s *CommunityService) ToggleLike(ctx context.Context, userID, postID uuid.UUID) (bool, int, error) {
	if _, err := s.GetPost(ctx, postID); err != nil {
		return false, 0, err
	}
	liked, likeCount, err := s.posts.ToggleLike(ctx, userID, postID)
	if err != nil {
		return false, 0, apperror.Wrap(http.StatusInternalServerError, "toggle_like_failed", "failed to toggle like", err)
	}
	return liked, likeCount, nil
}

func (s *CommunityService) ListComments(ctx context.Context, postID uuid.UUID) ([]model.Comment, error) {
	if _, err := s.GetPost(ctx, postID); err != nil {
		return nil, err
	}
	comments, err := s.posts.ListComments(ctx, postID)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "list_comments_failed", "failed to load comments", err)
	}
	return comments, nil
}

func (s *CommunityService) CreateComment(ctx context.Context, userID, postID uuid.UUID, req dto.CreateCommentRequest) (*model.Comment, error) {
	if _, err := s.GetPost(ctx, postID); err != nil {
		return nil, err
	}

	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		parsedParentID, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, apperror.New(http.StatusBadRequest, "invalid_parent_id", "parent id is invalid")
		}
		parentID = &parsedParentID
	}

	comment := &model.Comment{
		PostID:   postID,
		AuthorID: userID,
		ParentID: parentID,
		Content:  req.Content,
	}
	if err := s.posts.CreateComment(ctx, comment); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "create_comment_failed", "failed to create comment", err)
	}
	return comment, nil
}

func (s *CommunityService) DeleteComment(ctx context.Context, userID, commentID uuid.UUID) error {
	comment, err := s.posts.GetCommentByID(ctx, commentID)
	if err != nil {
		return apperror.Wrap(http.StatusInternalServerError, "get_comment_failed", "failed to load comment", err)
	}
	if comment == nil {
		return apperror.New(http.StatusNotFound, "comment_not_found", "comment not found")
	}
	if comment.AuthorID != userID {
		return apperror.New(http.StatusForbidden, "forbidden", "cannot delete another user's comment")
	}
	if err := s.posts.DeleteComment(ctx, comment); err != nil {
		return apperror.Wrap(http.StatusInternalServerError, "delete_comment_failed", "failed to delete comment", err)
	}
	return nil
}

func (s *CommunityService) AskAI(ctx context.Context, question string) string {
	if s.llm != nil {
		answer, err := s.llm.AnswerCommunity(ctx, question)
		if err == nil && answer != "" {
			return answer
		}
	}

	loweredQuestion := strings.ToLower(question)
	answerParts := []string{}

	switch {
	case strings.Contains(question, "寄养") || strings.Contains(loweredQuestion, "boarding"):
		answerParts = append(answerParts, "优先确认环境清洁、遛放频率、视频探视和应急医疗流程。")
	case strings.Contains(question, "疫苗") || strings.Contains(question, "医院"):
		answerParts = append(answerParts, "优先选择可提供病历、疫苗批次和复诊建议的正规医院。")
	case strings.Contains(question, "美容"):
		answerParts = append(answerParts, "先确认是否支持猫犬分区、烘干方式和敏感皮肤护理。")
	default:
		answerParts = append(answerParts, "优先看真实评价、预约时段和是否有清晰的服务边界。")
	}

	answerParts = append(answerParts, "这是一条规则生成的社区建议，最终仍要结合宠物个体情况判断。")
	return strings.Join(answerParts, "")
}

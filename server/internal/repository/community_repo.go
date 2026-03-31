package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"petverse/server/internal/model"
	"petverse/server/internal/pkg/pagination"
)

type CommunityRepository struct {
	db *gorm.DB
}

func NewCommunityRepository(db *gorm.DB) *CommunityRepository {
	return &CommunityRepository{db: db}
}

func (r *CommunityRepository) CreatePost(ctx context.Context, post *model.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *CommunityRepository) ListPosts(ctx context.Context, page, pageSize int, tag string) ([]model.Post, int64, error) {
	var (
		posts []model.Post
		total int64
	)

	query := r.db.WithContext(ctx).Model(&model.Post{}).Where("is_published = ?", true)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(pagination.Offset(page, pageSize)).
		Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (r *CommunityRepository) GetPostByID(ctx context.Context, id uuid.UUID) (*model.Post, error) {
	var post model.Post
	err := r.db.WithContext(ctx).First(&post, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *CommunityRepository) UpdatePost(ctx context.Context, post *model.Post) error {
	return r.db.WithContext(ctx).Save(post).Error
}

func (r *CommunityRepository) DeletePost(ctx context.Context, post *model.Post) error {
	return r.db.WithContext(ctx).Delete(post).Error
}

func (r *CommunityRepository) ToggleLike(ctx context.Context, userID, postID uuid.UUID) (bool, int, error) {
	var like model.PostLike
	err := r.db.WithContext(ctx).Where("user_id = ? AND post_id = ?", userID, postID).First(&like).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := r.db.WithContext(ctx).Create(&model.PostLike{UserID: userID, PostID: postID}).Error; err != nil {
			return false, 0, err
		}
		if err := r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", postID).UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
			return false, 0, err
		}
		count, countErr := r.likeCount(ctx, postID)
		return true, count, countErr
	}
	if err != nil {
		return false, 0, err
	}

	if err := r.db.WithContext(ctx).Delete(&like).Error; err != nil {
		return false, 0, err
	}
	if err := r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", postID).UpdateColumn("like_count", gorm.Expr("CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END")).Error; err != nil {
		return false, 0, err
	}
	count, countErr := r.likeCount(ctx, postID)
	return false, count, countErr
}

func (r *CommunityRepository) ListComments(ctx context.Context, postID uuid.UUID) ([]model.Comment, error) {
	var comments []model.Comment
	err := r.db.WithContext(ctx).Where("post_id = ?", postID).Order("created_at ASC").Find(&comments).Error
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *CommunityRepository) CreateComment(ctx context.Context, comment *model.Comment) error {
	if err := r.db.WithContext(ctx).Create(comment).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", comment.PostID).UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
}

func (r *CommunityRepository) GetCommentByID(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.WithContext(ctx).First(&comment, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *CommunityRepository) DeleteComment(ctx context.Context, comment *model.Comment) error {
	if err := r.db.WithContext(ctx).Delete(comment).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ? AND comment_count > 0", comment.PostID).UpdateColumn("comment_count", gorm.Expr("comment_count - 1")).Error
}

func (r *CommunityRepository) likeCount(ctx context.Context, postID uuid.UUID) (int, error) {
	var post model.Post
	if err := r.db.WithContext(ctx).Select("like_count").First(&post, "id = ?", postID).Error; err != nil {
		return 0, err
	}
	return post.LikeCount, nil
}

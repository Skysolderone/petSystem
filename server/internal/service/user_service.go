package service

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"petverse/server/internal/dto"
	"petverse/server/internal/model"
	"petverse/server/internal/pkg/apperror"
)

type UserService struct {
	users userRepository
}

type userRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
}

func NewUserService(users userRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) GetMe(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "user_lookup_failed", "failed to query user", err)
	}
	if user == nil {
		return nil, apperror.New(http.StatusNotFound, "user_not_found", "user not found")
	}
	return user, nil
}

func (s *UserService) UpdateMe(ctx context.Context, userID uuid.UUID, req dto.UpdateUserRequest) (*model.User, error) {
	user, err := s.GetMe(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}
	if req.Email != nil {
		if *req.Email == "" {
			user.Email = nil
		} else {
			user.Email = req.Email
		}
	}
	if req.AvatarURL != nil {
		user.AvatarURL = *req.AvatarURL
	}

	if err := s.users.Update(ctx, user); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "update_user_failed", "failed to update user", err)
	}
	return user, nil
}

func (s *UserService) UpdateLocation(ctx context.Context, userID uuid.UUID, req dto.UpdateUserLocationRequest) (*model.User, error) {
	user, err := s.GetMe(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.Latitude != nil {
		user.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		user.Longitude = req.Longitude
	}

	if err := s.users.Update(ctx, user); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "update_user_failed", "failed to update user location", err)
	}
	return user, nil
}

func (s *UserService) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) (*model.User, error) {
	return s.UpdateMe(ctx, userID, dto.UpdateUserRequest{
		AvatarURL: &avatarURL,
	})
}

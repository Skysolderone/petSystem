package service

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"petverse/server/internal/dto"
	"petverse/server/internal/model"
	"petverse/server/internal/pkg/apperror"
	petjwt "petverse/server/internal/pkg/jwt"
	"petverse/server/internal/pkg/password"
	"petverse/server/internal/pkg/socialauth"
)

type AuthService struct {
	users          authUserRepository
	tokens         *petjwt.Manager
	appleVerifier  socialauth.Verifier
	googleVerifier socialauth.Verifier
}

type authUserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByPhone(ctx context.Context, phone string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByWechatOpenID(ctx context.Context, openID string) (*model.User, error)
	GetByAppleID(ctx context.Context, appleID string) (*model.User, error)
	GetByGoogleID(ctx context.Context, googleID string) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
}

type AuthServiceOption func(*AuthService)

func WithAppleVerifier(verifier socialauth.Verifier) AuthServiceOption {
	return func(service *AuthService) {
		service.appleVerifier = verifier
	}
}

func WithGoogleVerifier(verifier socialauth.Verifier) AuthServiceOption {
	return func(service *AuthService) {
		service.googleVerifier = verifier
	}
}

func NewAuthService(users authUserRepository, tokens *petjwt.Manager, options ...AuthServiceOption) *AuthService {
	service := &AuthService{
		users:  users,
		tokens: tokens,
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*model.User, petjwt.TokenPair, error) {
	existing, err := s.users.GetByPhone(ctx, req.Phone)
	if err != nil {
		return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "user_lookup_failed", "failed to query user", err)
	}
	if existing != nil {
		return nil, petjwt.TokenPair{}, apperror.New(http.StatusConflict, "phone_already_exists", "phone already registered")
	}

	hashedPassword, err := password.Hash(req.Password)
	if err != nil {
		return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "password_hash_failed", "failed to process password", err)
	}

	phone := req.Phone
	user := &model.User{
		Phone:    &phone,
		Password: hashedPassword,
		Nickname: req.Nickname,
		Role:     "user",
		PlanType: "free",
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "create_user_failed", "failed to create user", err)
	}

	tokenPair, err := s.tokens.GenerateTokens(user.ID)
	if err != nil {
		return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "token_issue_failed", "failed to issue tokens", err)
	}

	return user, tokenPair, nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*model.User, petjwt.TokenPair, error) {
	user, err := s.users.GetByPhone(ctx, req.Phone)
	if err != nil {
		return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "user_lookup_failed", "failed to query user", err)
	}
	if user == nil {
		return nil, petjwt.TokenPair{}, apperror.New(http.StatusUnauthorized, "invalid_credentials", "invalid phone or password")
	}

	if err := password.Compare(user.Password, req.Password); err != nil {
		return nil, petjwt.TokenPair{}, apperror.New(http.StatusUnauthorized, "invalid_credentials", "invalid phone or password")
	}

	tokenPair, err := s.tokens.GenerateTokens(user.ID)
	if err != nil {
		return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "token_issue_failed", "failed to issue tokens", err)
	}

	return user, tokenPair, nil
}

func (s *AuthService) LoginWithWechat(ctx context.Context, req dto.WechatLoginRequest) (*model.User, petjwt.TokenPair, error) {
	user, err := s.users.GetByWechatOpenID(ctx, req.OpenID)
	if err != nil {
		return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "user_lookup_failed", "failed to query user", err)
	}
	if user == nil {
		user, err = s.createSocialUser(ctx, socialCreateInput{
			Nickname:      firstNonEmpty(req.Nickname, "微信用户"),
			AvatarURL:     req.AvatarURL,
			WechatOpenID:  &req.OpenID,
			DefaultSecret: "wechat",
		})
		if err != nil {
			return nil, petjwt.TokenPair{}, err
		}
	} else if profileNeedsRefresh(user, req.Nickname, req.AvatarURL, nil) {
		applySocialProfile(user, req.Nickname, req.AvatarURL, nil)
		if err := s.users.Update(ctx, user); err != nil {
			return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "update_user_failed", "failed to update social profile", err)
		}
	}

	return s.issueTokenPair(user)
}

func (s *AuthService) LoginWithApple(ctx context.Context, req dto.AppleLoginRequest) (*model.User, petjwt.TokenPair, error) {
	appleID := strings.TrimSpace(req.AppleID)
	email := req.Email
	nickname := strings.TrimSpace(req.Nickname)
	avatarURL := strings.TrimSpace(req.AvatarURL)

	if strings.TrimSpace(req.IdentityToken) != "" {
		identity, err := s.verifyIdentityToken(ctx, s.appleVerifier, req.IdentityToken, "apple")
		if err != nil {
			return nil, petjwt.TokenPair{}, err
		}
		if appleID != "" && appleID != identity.Subject {
			return nil, petjwt.TokenPair{}, apperror.New(http.StatusBadRequest, "apple_subject_mismatch", "apple account does not match identity token")
		}

		appleID = identity.Subject
		if email == nil && identity.Email != "" {
			resolvedEmail := identity.Email
			email = &resolvedEmail
		}
		if nickname == "" {
			nickname = identity.Name
		}
	}
	if appleID == "" {
		return nil, petjwt.TokenPair{}, apperror.New(http.StatusBadRequest, "apple_id_required", "apple id or identity token is required")
	}

	user, err := s.users.GetByAppleID(ctx, appleID)
	if err != nil {
		return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "user_lookup_failed", "failed to query user", err)
	}
	if user == nil && email != nil && *email != "" {
		existingByEmail, err := s.users.GetByEmail(ctx, *email)
		if err != nil {
			return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "user_lookup_failed", "failed to query user", err)
		}
		if existingByEmail != nil {
			user = existingByEmail
			user.AppleID = &appleID
			applySocialProfile(user, nickname, avatarURL, email)
			if err := s.users.Update(ctx, user); err != nil {
				return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "update_user_failed", "failed to bind apple account", err)
			}
		}
	}

	if user == nil {
		user, err = s.createSocialUser(ctx, socialCreateInput{
			Nickname:      firstNonEmpty(nickname, "Apple 用户"),
			AvatarURL:     avatarURL,
			AppleID:       &appleID,
			Email:         email,
			DefaultSecret: "apple",
		})
		if err != nil {
			return nil, petjwt.TokenPair{}, err
		}
	} else if profileNeedsRefresh(user, nickname, avatarURL, email) {
		applySocialProfile(user, nickname, avatarURL, email)
		if err := s.users.Update(ctx, user); err != nil {
			return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "update_user_failed", "failed to update social profile", err)
		}
	}

	return s.issueTokenPair(user)
}

func (s *AuthService) LoginWithGoogle(ctx context.Context, req dto.GoogleLoginRequest) (*model.User, petjwt.TokenPair, error) {
	googleID := strings.TrimSpace(req.GoogleID)
	email := req.Email
	nickname := strings.TrimSpace(req.Nickname)
	avatarURL := strings.TrimSpace(req.AvatarURL)

	if strings.TrimSpace(req.IdentityToken) != "" {
		identity, err := s.verifyIdentityToken(ctx, s.googleVerifier, req.IdentityToken, "google")
		if err != nil {
			return nil, petjwt.TokenPair{}, err
		}
		if googleID != "" && googleID != identity.Subject {
			return nil, petjwt.TokenPair{}, apperror.New(http.StatusBadRequest, "google_subject_mismatch", "google account does not match identity token")
		}

		googleID = identity.Subject
		if email == nil && identity.Email != "" {
			resolvedEmail := identity.Email
			email = &resolvedEmail
		}
		if nickname == "" {
			nickname = identity.Name
		}
		if avatarURL == "" {
			avatarURL = identity.Picture
		}
	}
	if googleID == "" {
		return nil, petjwt.TokenPair{}, apperror.New(http.StatusBadRequest, "google_id_required", "google id or identity token is required")
	}

	user, err := s.users.GetByGoogleID(ctx, googleID)
	if err != nil {
		return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "user_lookup_failed", "failed to query user", err)
	}
	if user == nil && email != nil && *email != "" {
		existingByEmail, err := s.users.GetByEmail(ctx, *email)
		if err != nil {
			return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "user_lookup_failed", "failed to query user", err)
		}
		if existingByEmail != nil {
			user = existingByEmail
			user.GoogleID = &googleID
			applySocialProfile(user, nickname, avatarURL, email)
			if err := s.users.Update(ctx, user); err != nil {
				return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "update_user_failed", "failed to bind google account", err)
			}
		}
	}

	if user == nil {
		user, err = s.createSocialUser(ctx, socialCreateInput{
			Nickname:      firstNonEmpty(nickname, "Google 用户"),
			AvatarURL:     avatarURL,
			GoogleID:      &googleID,
			Email:         email,
			DefaultSecret: "google",
		})
		if err != nil {
			return nil, petjwt.TokenPair{}, err
		}
	} else if profileNeedsRefresh(user, nickname, avatarURL, email) {
		applySocialProfile(user, nickname, avatarURL, email)
		if err := s.users.Update(ctx, user); err != nil {
			return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "update_user_failed", "failed to update social profile", err)
		}
	}

	return s.issueTokenPair(user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*model.User, petjwt.TokenPair, error) {
	claims, err := s.tokens.ParseToken(refreshToken)
	if err != nil {
		return nil, petjwt.TokenPair{}, apperror.New(http.StatusUnauthorized, "invalid_refresh_token", "refresh token is invalid")
	}
	if claims.TokenType != "refresh" {
		return nil, petjwt.TokenPair{}, apperror.New(http.StatusUnauthorized, "invalid_refresh_token", "refresh token is invalid")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, petjwt.TokenPair{}, apperror.New(http.StatusUnauthorized, "invalid_refresh_token", "refresh token is invalid")
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "user_lookup_failed", "failed to query user", err)
	}
	if user == nil {
		return nil, petjwt.TokenPair{}, apperror.New(http.StatusUnauthorized, "invalid_refresh_token", "refresh token is invalid")
	}

	tokenPair, err := s.tokens.GenerateTokens(user.ID)
	if err != nil {
		return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "token_issue_failed", "failed to issue tokens", err)
	}

	return user, tokenPair, nil
}

type socialCreateInput struct {
	Nickname      string
	AvatarURL     string
	Email         *string
	WechatOpenID  *string
	AppleID       *string
	GoogleID      *string
	DefaultSecret string
}

func (s *AuthService) createSocialUser(ctx context.Context, input socialCreateInput) (*model.User, error) {
	hashedPassword, err := password.Hash(input.DefaultSecret + "-" + uuid.NewString())
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "password_hash_failed", "failed to prepare social login", err)
	}

	user := &model.User{
		Email:        input.Email,
		Password:     hashedPassword,
		Nickname:     input.Nickname,
		AvatarURL:    input.AvatarURL,
		Role:         "user",
		PlanType:     "free",
		WechatOpenID: input.WechatOpenID,
		AppleID:      input.AppleID,
		GoogleID:     input.GoogleID,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "create_user_failed", "failed to create user", err)
	}
	return user, nil
}

func (s *AuthService) issueTokenPair(user *model.User) (*model.User, petjwt.TokenPair, error) {
	tokenPair, err := s.tokens.GenerateTokens(user.ID)
	if err != nil {
		return nil, petjwt.TokenPair{}, apperror.Wrap(http.StatusInternalServerError, "token_issue_failed", "failed to issue tokens", err)
	}
	return user, tokenPair, nil
}

func (s *AuthService) verifyIdentityToken(ctx context.Context, verifier socialauth.Verifier, identityToken, provider string) (*socialauth.Identity, error) {
	if verifier == nil {
		return nil, apperror.New(http.StatusNotImplemented, provider+"_identity_not_configured", provider+" identity token login is not configured")
	}

	identity, err := verifier.VerifyIDToken(ctx, identityToken)
	if err != nil {
		return nil, apperror.Wrap(http.StatusUnauthorized, provider+"_identity_invalid", "identity token is invalid", err)
	}
	return identity, nil
}

func applySocialProfile(user *model.User, nickname, avatarURL string, email *string) {
	if nickname != "" {
		user.Nickname = nickname
	}
	if avatarURL != "" {
		user.AvatarURL = avatarURL
	}
	if email != nil && *email != "" {
		user.Email = email
	}
}

func profileNeedsRefresh(user *model.User, nickname, avatarURL string, email *string) bool {
	switch {
	case nickname != "" && user.Nickname != nickname:
		return true
	case avatarURL != "" && user.AvatarURL != avatarURL:
		return true
	case email != nil && *email != "" && (user.Email == nil || *user.Email != *email):
		return true
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

var ErrUnauthenticated = errors.New("unauthenticated")

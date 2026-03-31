package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"petverse/server/internal/dto"
	"petverse/server/internal/model"
	petjwt "petverse/server/internal/pkg/jwt"
	"petverse/server/internal/pkg/password"
	"petverse/server/internal/pkg/socialauth"
)

type fakeUserRepo struct {
	usersByID     map[uuid.UUID]*model.User
	usersByPhone  map[string]*model.User
	usersByEmail  map[string]*model.User
	usersByWechat map[string]*model.User
	usersByApple  map[string]*model.User
	usersByGoogle map[string]*model.User
}

type fakeSocialVerifier struct {
	identity *socialauth.Identity
	err      error
}

func (f fakeSocialVerifier) VerifyIDToken(_ context.Context, _ string) (*socialauth.Identity, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.identity, nil
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		usersByID:     map[uuid.UUID]*model.User{},
		usersByPhone:  map[string]*model.User{},
		usersByEmail:  map[string]*model.User{},
		usersByWechat: map[string]*model.User{},
		usersByApple:  map[string]*model.User{},
		usersByGoogle: map[string]*model.User{},
	}
}

func (r *fakeUserRepo) Create(_ context.Context, user *model.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	r.usersByID[user.ID] = user
	if user.Phone != nil {
		r.usersByPhone[*user.Phone] = user
	}
	if user.Email != nil {
		r.usersByEmail[*user.Email] = user
	}
	if user.WechatOpenID != nil {
		r.usersByWechat[*user.WechatOpenID] = user
	}
	if user.AppleID != nil {
		r.usersByApple[*user.AppleID] = user
	}
	if user.GoogleID != nil {
		r.usersByGoogle[*user.GoogleID] = user
	}
	return nil
}

func (r *fakeUserRepo) GetByPhone(_ context.Context, phone string) (*model.User, error) {
	return r.usersByPhone[phone], nil
}

func (r *fakeUserRepo) GetByEmail(_ context.Context, email string) (*model.User, error) {
	return r.usersByEmail[email], nil
}

func (r *fakeUserRepo) GetByWechatOpenID(_ context.Context, openID string) (*model.User, error) {
	return r.usersByWechat[openID], nil
}

func (r *fakeUserRepo) GetByAppleID(_ context.Context, appleID string) (*model.User, error) {
	return r.usersByApple[appleID], nil
}

func (r *fakeUserRepo) GetByGoogleID(_ context.Context, googleID string) (*model.User, error) {
	return r.usersByGoogle[googleID], nil
}

func (r *fakeUserRepo) GetByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	return r.usersByID[id], nil
}

func (r *fakeUserRepo) Update(_ context.Context, user *model.User) error {
	r.usersByID[user.ID] = user
	if user.Phone != nil {
		r.usersByPhone[*user.Phone] = user
	}
	if user.Email != nil {
		r.usersByEmail[*user.Email] = user
	}
	if user.WechatOpenID != nil {
		r.usersByWechat[*user.WechatOpenID] = user
	}
	if user.AppleID != nil {
		r.usersByApple[*user.AppleID] = user
	}
	if user.GoogleID != nil {
		r.usersByGoogle[*user.GoogleID] = user
	}
	return nil
}

func TestAuthServiceRegisterAndLogin(t *testing.T) {
	t.Parallel()

	repo := newFakeUserRepo()
	tokens := petjwt.NewManager("secret", "test", 15*time.Minute, time.Hour)
	service := NewAuthService(repo, tokens)

	registerReq := dto.RegisterRequest{
		Phone:    "13800138000",
		Password: "strong-pass",
		Nickname: "Tester",
	}

	registeredUser, tokenPair, err := service.Register(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if registeredUser.ID == uuid.Nil {
		t.Fatal("Register() did not assign user ID")
	}
	if registeredUser.Password == registerReq.Password {
		t.Fatal("Register() stored raw password")
	}
	if tokenPair.AccessToken == "" || tokenPair.RefreshToken == "" {
		t.Fatal("Register() did not issue tokens")
	}

	loggedInUser, _, err := service.Login(context.Background(), dto.LoginRequest{
		Phone:    registerReq.Phone,
		Password: registerReq.Password,
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if loggedInUser.ID != registeredUser.ID {
		t.Fatalf("Login() returned wrong user: got %s want %s", loggedInUser.ID, registeredUser.ID)
	}
}

func TestAuthServiceRefresh(t *testing.T) {
	t.Parallel()

	repo := newFakeUserRepo()
	hashedPassword, err := password.Hash("strong-pass")
	if err != nil {
		t.Fatalf("password.Hash() error = %v", err)
	}

	phone := "13800138001"
	user := &model.User{
		ID:       uuid.New(),
		Phone:    &phone,
		Password: hashedPassword,
		Nickname: "Refresh User",
	}
	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("repo.Create() error = %v", err)
	}

	tokens := petjwt.NewManager("secret", "test", 15*time.Minute, time.Hour)
	service := NewAuthService(repo, tokens)

	issuedTokens, err := tokens.GenerateTokens(user.ID)
	if err != nil {
		t.Fatalf("GenerateTokens() error = %v", err)
	}

	refreshedUser, refreshedTokens, err := service.Refresh(context.Background(), issuedTokens.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if refreshedUser.ID != user.ID {
		t.Fatalf("Refresh() returned wrong user: got %s want %s", refreshedUser.ID, user.ID)
	}
	if refreshedTokens.AccessToken == "" || refreshedTokens.RefreshToken == "" {
		t.Fatal("Refresh() did not issue new tokens")
	}
}

func TestAuthServiceLoginWithGoogleIdentityToken(t *testing.T) {
	t.Parallel()

	repo := newFakeUserRepo()
	tokens := petjwt.NewManager("secret", "test", 15*time.Minute, time.Hour)
	service := NewAuthService(
		repo,
		tokens,
		WithGoogleVerifier(fakeSocialVerifier{
			identity: &socialauth.Identity{
				Subject: "google-real-user",
				Email:   "google.real@example.com",
				Name:    "Google Real User",
				Picture: "https://example.com/google.png",
			},
		}),
	)

	user, tokenPair, err := service.LoginWithGoogle(context.Background(), dto.GoogleLoginRequest{
		IdentityToken: "google.jwt.token",
	})
	if err != nil {
		t.Fatalf("LoginWithGoogle() error = %v", err)
	}
	if user.GoogleID == nil || *user.GoogleID != "google-real-user" {
		t.Fatalf("unexpected google id: %+v", user)
	}
	if user.Email == nil || *user.Email != "google.real@example.com" {
		t.Fatalf("unexpected email: %+v", user)
	}
	if user.AvatarURL != "https://example.com/google.png" {
		t.Fatalf("unexpected avatar: %+v", user)
	}
	if tokenPair.AccessToken == "" || tokenPair.RefreshToken == "" {
		t.Fatal("LoginWithGoogle() did not issue tokens")
	}
}

func TestAuthServiceLoginWithAppleIdentityTokenBindsExistingEmail(t *testing.T) {
	t.Parallel()

	repo := newFakeUserRepo()
	tokens := petjwt.NewManager("secret", "test", 15*time.Minute, time.Hour)

	hashedPassword, err := password.Hash("strong-pass")
	if err != nil {
		t.Fatalf("password.Hash() error = %v", err)
	}
	email := "apple.bind@example.com"
	existingUser := &model.User{
		ID:       uuid.New(),
		Email:    &email,
		Password: hashedPassword,
		Nickname: "Existing User",
	}
	if err := repo.Create(context.Background(), existingUser); err != nil {
		t.Fatalf("repo.Create() error = %v", err)
	}

	service := NewAuthService(
		repo,
		tokens,
		WithAppleVerifier(fakeSocialVerifier{
			identity: &socialauth.Identity{
				Subject: "apple-real-user",
				Email:   email,
				Name:    "Apple Real User",
			},
		}),
	)

	user, _, err := service.LoginWithApple(context.Background(), dto.AppleLoginRequest{
		IdentityToken: "apple.jwt.token",
	})
	if err != nil {
		t.Fatalf("LoginWithApple() error = %v", err)
	}
	if user.ID != existingUser.ID {
		t.Fatalf("LoginWithApple() returned wrong user: got %s want %s", user.ID, existingUser.ID)
	}
	if user.AppleID == nil || *user.AppleID != "apple-real-user" {
		t.Fatalf("expected bound apple id, got %+v", user)
	}
}

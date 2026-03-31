package dto

type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required,min=6,max=20"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Nickname string `json:"nickname" binding:"required,min=1,max=50"`
}

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required,min=6,max=20"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type WechatLoginRequest struct {
	OpenID    string `json:"open_id" binding:"required,min=3,max=100"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

type AppleLoginRequest struct {
	AppleID   string  `json:"apple_id" binding:"required,min=3,max=100"`
	Email     *string `json:"email"`
	Nickname  string  `json:"nickname"`
	AvatarURL string  `json:"avatar_url"`
}

type GoogleLoginRequest struct {
	GoogleID  string  `json:"google_id" binding:"required,min=3,max=100"`
	Email     *string `json:"email"`
	Nickname  string  `json:"nickname"`
	AvatarURL string  `json:"avatar_url"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
}

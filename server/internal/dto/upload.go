package dto

type UpdateUserRequest struct {
	Nickname  *string `json:"nickname"`
	Email     *string `json:"email"`
	AvatarURL *string `json:"avatar_url"`
}

type UpdateUserLocationRequest struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type UploadResponse struct {
	URL      string `json:"url"`
	Path     string `json:"path"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MIMEType string `json:"mime_type"`
}

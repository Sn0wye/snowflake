package dto

type OAuthResponse struct {
	AccessToken  string `json:"access_token" example:"123e4567-e89b-12d3-a456-426614174000"`
	RefreshToken string `json:"refresh_token" example:"a1b2c3d4e5f6..."`
} // @name OAuthResponse

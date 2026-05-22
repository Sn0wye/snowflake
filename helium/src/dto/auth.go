package dto

type ProfileResponse struct {
	ID           string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name         string `json:"name" example:"John Doe"`
	Username     string `json:"username" example:"johndoe"`
	Email        string `json:"email" example:"johndoe@example.com"`
	AnnualIncome int    `json:"annual_income" example:"100000"`
	Debt         int    `json:"debt" example:"50000"`
	AssetsValue  int    `json:"assets_value" example:"200000"`
} // @name ProfileResponse

type LoginRequest struct {
	Email    string `json:"email" validate:"required" example:"johndoe@example.com"`
	Password string `json:"password" validate:"required" example:"P@ssw0rd!"`
} // @name LoginRequest

type LoginResponse struct {
	AccessToken  string `json:"access_token" example:"123e4567-e89b-12d3-a456-426614174000"`
	RefreshToken string `json:"refresh_token" example:"a1b2c3d4e5f6..."`
} // @name LoginResponse

type RegisterRequest struct {
	Name         string `json:"name" validate:"required" example:"John Doe"`
	Username     string `json:"username" validate:"required" example:"johndoe"`
	Password     string `json:"password" validate:"required" example:"P@ssw0rd!"`
	Email        string `json:"email" validate:"required,email" example:"johndoe@example.com"`
	AnnualIncome int    `json:"annual_income" validate:"required,number,gte=0" example:"100000"`
	Debt         int    `json:"debt" validate:"required,number,gte=0" example:"50000"`
	AssetsValue  int    `json:"assets_value" validate:"required,number,gte=0" example:"200000"`
} // @name RegisterRequest

type RegisterResponse struct {
	AccessToken  string `json:"access_token" example:"123e4567-e89b-12d3-a456-426614174000"`
	RefreshToken string `json:"refresh_token" example:"a1b2c3d4e5f6..."`
} // @name RegisterResponse

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"a1b2c3d4e5f6..."`
} // @name RefreshRequest

type RefreshResponse struct {
	AccessToken  string `json:"access_token" example:"123e4567-e89b-12d3-a456-426614174000"`
	RefreshToken string `json:"refresh_token" example:"a1b2c3d4e5f6..."`
} // @name RefreshResponse

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"a1b2c3d4e5f6..."`
} // @name LogoutRequest

package service

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrEmailAlreadyTaken     = errors.New("email already taken")
	ErrInvalidCredentials    = errors.New("invalid email or password")
	ErrRefreshTokenNotFound  = errors.New("refresh token not found")
	ErrRefreshTokenExpired   = errors.New("refresh token expired")
	ErrOAuthAccountNotFound  = errors.New("oauth account not found")
	ErrUsernameGenerationFailed = errors.New("failed to generate unique username")
)

func mapNotFound(err error, domainErr error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domainErr
	}
	return err
}

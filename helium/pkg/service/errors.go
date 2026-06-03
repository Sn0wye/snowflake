package service

import (
	"errors"

	"github.com/getsnowflake/snowflake/helium/pkg/repository"
)

var (
	ErrUserNotFound             = errors.New("user not found")
	ErrEmailAlreadyTaken        = errors.New("email already taken")
	ErrInvalidCredentials       = errors.New("invalid email or password")
	ErrRefreshTokenNotFound     = errors.New("refresh token not found")
	ErrRefreshTokenExpired      = errors.New("refresh token expired")
)

func mapNotFound(err error, domainErr error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return domainErr
	}
	return err
}

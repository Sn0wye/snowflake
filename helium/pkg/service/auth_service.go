package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/getsnowflake/snowflake/helium/pkg/jwt"
	"github.com/getsnowflake/snowflake/helium/pkg/messaging"
	"github.com/getsnowflake/snowflake/helium/pkg/repository"
	"github.com/getsnowflake/snowflake/helium/src/dto"
	"github.com/getsnowflake/snowflake/helium/src/models"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	Profile(db *gorm.DB, userID string) (dto.ProfileResponse, error)
	Register(db *gorm.DB, req dto.RegisterRequest) (dto.RegisterResponse, error)
	Login(db *gorm.DB, req dto.LoginRequest) (dto.LoginResponse, error)
	Refresh(db *gorm.DB, refreshTokenString string) (dto.RefreshResponse, error)
	Logout(db *gorm.DB, refreshTokenString string) error
}

type authService struct {
	repos *repository.Factory
	jwt   *jwt.JWT
	token TokenService
	rmq   *messaging.MessagingService
}

func newAuthService(repos *repository.Factory, j *jwt.JWT, token TokenService, rmq *messaging.MessagingService) AuthService {
	return &authService{repos: repos, jwt: j, token: token, rmq: rmq}
}

func (s *authService) Profile(db *gorm.DB, userID string) (dto.ProfileResponse, error) {
	user, err := s.repos.User.FindByID(db, userID)
	if err != nil {
		return dto.ProfileResponse{}, mapNotFound(err, ErrUserNotFound)
	}

	return dto.ProfileResponse{
		ID:           user.ID.String(),
		Name:         user.Name,
		Username:     user.Username,
		Email:        user.Email,
		AnnualIncome: user.AnnualIncome,
		Debt:         user.Debt,
		AssetsValue:  user.AssetsValue,
	}, nil
}

func (s *authService) Register(db *gorm.DB, req dto.RegisterRequest) (dto.RegisterResponse, error) {
	_, err := s.repos.User.FindByEmail(db, req.Email)
	if err == nil {
		return dto.RegisterResponse{}, ErrEmailAlreadyTaken
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.RegisterResponse{}, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return dto.RegisterResponse{}, errors.Wrap(err, "failed to hash password")
	}

	user := models.User{
		Username:     req.Username,
		Password:     string(hashedPassword),
		Email:        req.Email,
		Name:         req.Name,
		AnnualIncome: req.AnnualIncome,
		Debt:         req.Debt,
		AssetsValue:  req.AssetsValue,
	}

	if err := s.repos.User.Create(db, &user); err != nil {
		return dto.RegisterResponse{}, errors.Wrap(err, "failed to create user")
	}

	s.emitUserCreated(user)

	accessToken, err := s.token.GenerateAccessToken(user.ID.String())
	if err != nil {
		return dto.RegisterResponse{}, errors.Wrap(err, "failed to generate JWT token")
	}

	refreshToken, err := s.token.GenerateRefreshToken(db, user.ID.String())
	if err != nil {
		return dto.RegisterResponse{}, errors.Wrap(err, "failed to generate refresh token")
	}

	return dto.RegisterResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) Login(db *gorm.DB, req dto.LoginRequest) (dto.LoginResponse, error) {
	user, err := s.repos.User.FindByEmail(db, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.LoginResponse{}, ErrInvalidCredentials
		}
		return dto.LoginResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return dto.LoginResponse{}, ErrInvalidCredentials
	}

	if err := s.token.RevokeAllUserRefreshTokens(db, user.ID.String()); err != nil {
		return dto.LoginResponse{}, errors.Wrap(err, "failed to revoke existing refresh tokens")
	}

	accessToken, err := s.token.GenerateAccessToken(user.ID.String())
	if err != nil {
		return dto.LoginResponse{}, errors.Wrap(err, "failed to generate JWT token")
	}

	refreshToken, err := s.token.GenerateRefreshToken(db, user.ID.String())
	if err != nil {
		return dto.LoginResponse{}, errors.Wrap(err, "failed to generate refresh token")
	}

	return dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) Refresh(db *gorm.DB, refreshTokenString string) (dto.RefreshResponse, error) {
	claims, err := s.jwt.ParseRefreshToken(refreshTokenString)
	if err != nil {
		return dto.RefreshResponse{}, ErrInvalidCredentials
	}

	tokenHash := hashToken(refreshTokenString)

	token, err := s.repos.RefreshToken.FindByTokenHash(db, tokenHash)
	if err != nil {
		return dto.RefreshResponse{}, mapNotFound(err, ErrRefreshTokenNotFound)
	}

	if token.IsExpired() {
		return dto.RefreshResponse{}, ErrRefreshTokenExpired
	}

	userID := claims.Subject

	accessToken, err := s.token.GenerateAccessToken(userID)
	if err != nil {
		return dto.RefreshResponse{}, errors.Wrap(err, "failed to generate JWT token")
	}

	newRefreshToken, err := s.token.GenerateRefreshToken(db, userID)
	if err != nil {
		return dto.RefreshResponse{}, errors.Wrap(err, "failed to generate refresh token")
	}

	if err := s.repos.RefreshToken.Delete(db, token); err != nil {
		return dto.RefreshResponse{}, errors.Wrap(err, "failed to revoke old refresh token")
	}

	return dto.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *authService) Logout(db *gorm.DB, refreshTokenString string) error {
	tokenHash := hashToken(refreshTokenString)

	token, err := s.repos.RefreshToken.FindByTokenHash(db, tokenHash)
	if err != nil {
		return mapNotFound(err, ErrRefreshTokenNotFound)
	}

	return s.repos.RefreshToken.Delete(db, token)
}

func (s *authService) emitUserCreated(user models.User) {
	data := map[string]interface{}{
		"id":            user.ID.String(),
		"username":      user.Username,
		"email":         user.Email,
		"annual_income": user.AnnualIncome,
		"debt":          user.Debt,
		"assets_value":  user.AssetsValue,
		"created_at":    user.CreatedAt,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	s.rmq.Produce("user.created", string(jsonData))
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

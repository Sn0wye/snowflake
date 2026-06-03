package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getsnowflake/snowflake/helium/pkg/messaging"
	"github.com/getsnowflake/snowflake/helium/pkg/repository"
	"github.com/getsnowflake/snowflake/helium/src/dto"
	"github.com/getsnowflake/snowflake/helium/src/models"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type OAuthService interface {
	UpsertOAuthUser(db *gorm.DB, googleSub, email, name string) (string, error)
	GenerateAuthResponse(db *gorm.DB, userID string) (dto.OAuthResponse, error)
}

type oauthService struct {
	repos *repository.Factory
	token TokenService
	rmq   *messaging.MessagingService
}

func newOAuthService(repos *repository.Factory, token TokenService, rmq *messaging.MessagingService) OAuthService {
	return &oauthService{repos: repos, token: token, rmq: rmq}
}

func (s *oauthService) UpsertOAuthUser(db *gorm.DB, googleSub, email, name string) (string, error) {
	return s.upsertOAuthUserInTx(db, googleSub, email, name)
}

func (s *oauthService) upsertOAuthUserInTx(db *gorm.DB, googleSub, email, name string) (userID string, err error) {
	tx := db.Begin()
	if tx.Error != nil {
		return "", errors.Wrap(tx.Error, "failed to begin transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	oauthAccount, err := s.repos.OAuthAccount.FindByProviderAndProviderID(tx, models.ProviderGoogle, googleSub)
	if err == nil {
		if commitErr := tx.Commit().Error; commitErr != nil {
			return "", errors.Wrap(commitErr, "failed to commit transaction")
		}
		return oauthAccount.UserID.String(), nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		tx.Rollback()
		return "", errors.Wrap(err, "failed to query oauth account")
	}

	user, err := s.repos.User.FindByEmail(tx, email)
	if err == nil {
		oauthAccount := &models.OAuthAccount{
			UserID:     user.ID,
			Provider:   models.ProviderGoogle,
			ProviderID: googleSub,
		}
		if createErr := s.repos.OAuthAccount.Create(tx, oauthAccount); createErr != nil {
			tx.Rollback()
			return "", errors.Wrap(createErr, "failed to create oauth account")
		}

		if commitErr := tx.Commit().Error; commitErr != nil {
			return "", errors.Wrap(commitErr, "failed to commit transaction")
		}
		return user.ID.String(), nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		tx.Rollback()
		return "", errors.Wrap(err, "failed to query user")
	}

	username := generateUsername(tx, s.repos, email)

	newUser := models.User{
		Name:     name,
		Email:    email,
		Username: username,
		Password: "",
	}

	if createErr := s.repos.User.Create(tx, &newUser); createErr != nil {
		tx.Rollback()
		return "", errors.Wrap(createErr, "failed to create user")
	}

	newOAuthAccount := models.OAuthAccount{
		UserID:     newUser.ID,
		Provider:   models.ProviderGoogle,
		ProviderID: googleSub,
	}
	if createErr := s.repos.OAuthAccount.Create(tx, &newOAuthAccount); createErr != nil {
		tx.Rollback()
		return "", errors.Wrap(createErr, "failed to create oauth account")
	}

	if commitErr := tx.Commit().Error; commitErr != nil {
		return "", errors.Wrap(commitErr, "failed to commit transaction")
	}

	publishUserCreated(s.rmq, newUser)

	return newUser.ID.String(), nil
}

func (s *oauthService) GenerateAuthResponse(db *gorm.DB, userID string) (dto.OAuthResponse, error) {
	if err := s.token.RevokeAllUserRefreshTokens(db, userID); err != nil {
		return dto.OAuthResponse{}, errors.Wrap(err, "failed to revoke existing refresh tokens")
	}

	accessToken, err := s.token.GenerateAccessToken(userID)
	if err != nil {
		return dto.OAuthResponse{}, errors.Wrap(err, "failed to generate JWT token")
	}

	refreshToken, err := s.token.GenerateRefreshToken(db, userID)
	if err != nil {
		return dto.OAuthResponse{}, errors.Wrap(err, "failed to generate refresh token")
	}

	return dto.OAuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func generateUsername(db *gorm.DB, repos *repository.Factory, email string) string {
	parts := strings.Split(email, "@")
	base := strings.ToLower(parts[0])
	username := base

	for attempt := 0; attempt < 10; attempt++ {
		count, err := repos.User.CountByUsername(db, username)
		if err != nil {
			username = fmt.Sprintf("%s_%s", base, uuid.New().String()[:8])
			continue
		}
		if count == 0 {
			return username
		}
		username = fmt.Sprintf("%s_%s", base, uuid.New().String()[:8])
	}

	return fmt.Sprintf("%s_%s", base, uuid.New().String()[:8])
}

func buildUserCreatedJSON(user models.User) (string, error) {
	data := map[string]interface{}{
		"id":         user.ID.String(),
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func publishUserCreated(rmq *messaging.MessagingService, user models.User) {
	userJSON, marshalErr := buildUserCreatedJSON(user)
	if marshalErr == nil {
		rmq.Produce("user.created", userJSON)
	}
}

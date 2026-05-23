package controllers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/getsnowflake/snowflake/helium/pkg/exceptions"
	"github.com/getsnowflake/snowflake/helium/pkg/jwt"
	"github.com/getsnowflake/snowflake/helium/pkg/messaging"
	"github.com/getsnowflake/snowflake/helium/src/dto"
	"github.com/getsnowflake/snowflake/helium/src/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

type OAuthController interface {
	Redirect(c *fiber.Ctx) error
	Callback(c *fiber.Ctx) error
}

type oauthController struct {
	db     *gorm.DB
	jwt    *jwt.JWT
	rmq    *messaging.MessagingService
	oauth  *oauth2.Config
	secure bool
}

func NewOAuthController(db *gorm.DB, jwt *jwt.JWT, rmq *messaging.MessagingService, conf *viper.Viper) OAuthController {
	return &oauthController{
		db:  db,
		jwt: jwt,
		rmq: rmq,
		oauth: &oauth2.Config{
			ClientID:     conf.GetString("google.client_id"),
			ClientSecret: conf.GetString("google.client_secret"),
			RedirectURL:  conf.GetString("google.redirect_url"),
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		secure: conf.GetBool("http.secure"),
	}
}

// Redirect godoc
//
//	@Summary		/auth/oauth
//	@Description	Initiate Google OAuth login
//	@Tags			Auth
//	@Produce		json
//	@Success		307	"Redirects to Google OAuth consent screen"
//	@Failure		500	{object}	exceptions.InternalServerError	"Failed to generate state"
//	@Router			/auth/oauth [get]
//	@OperationId	oauthRedirect
func (s *oauthController) Redirect(c *fiber.Ctx) error {
	state, err := generateState()
	if err != nil {
		return exceptions.InternalServer(c, "failed to generate state")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HTTPOnly: true,
		SameSite: "lax",
		Secure:   s.secure,
		MaxAge:   600,
	})

	url := s.oauth.AuthCodeURL(state, oauth2.AccessTypeOnline)

	return c.Redirect(url, fiber.StatusTemporaryRedirect)
}

// Callback godoc
//
//	@Summary		/auth/oauth/callback
//	@Description	Google OAuth callback — exchanges code for token, creates/links account, returns access and refresh tokens
//	@Description	Emits: `user.created` event upon new user registration.
//	@Tags			Auth
//	@Produce		json
//	@Param			code	query		string							true	"Authorization code from Google"
//	@Param			state	query		string							true	"CSRF state token"
//	@Success		200		{object}	dto.OAuthResponse				"OAuthResponse"
//	@Failure		400		{object}	exceptions.BadRequestError		"Missing code or state"
//	@Failure		401		{object}	exceptions.UnauthorizedError	"Invalid state"
//	@Failure		500		{object}	exceptions.InternalServerError	"Failed to exchange code OR fetch user info OR generate token"
//	@Router			/auth/oauth/callback [get]
//	@OperationId	oauthCallback
func (s *oauthController) Callback(c *fiber.Ctx) error {
	state := c.Query("state")
	code := c.Query("code")

	if code == "" || state == "" {
		return exceptions.BadRequest(c, "missing code or state")
	}

	cookieState := c.Cookies("oauth_state")
	if cookieState == "" || cookieState != state {
		return exceptions.Unauthorized(c)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    "",
		HTTPOnly: true,
		SameSite: "lax",
		Secure:   s.secure,
		MaxAge:   -1,
	})

	token, err := s.oauth.Exchange(c.Context(), code)
	if err != nil {
		return exceptions.InternalServer(c, "failed to exchange code")
	}

	client := s.oauth.Client(c.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return exceptions.InternalServer(c, "failed to get user info")
	}
	defer resp.Body.Close()

	var googleUser struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return exceptions.InternalServer(c, "failed to decode user info")
	}

	var oauthAccount models.OAuthAccount
	result := s.db.Where("provider = ? AND provider_id = ?", models.ProviderGoogle, googleUser.Sub).First(&oauthAccount)
	if result.Error == nil {
		return s.authResponse(c, oauthAccount.UserID.String())
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return exceptions.InternalServer(c, "failed to query oauth account")
	}

	var user models.User
	result = s.db.Where("email = ?", googleUser.Email).First(&user)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return exceptions.InternalServer(c, "failed to query user")
	}

	if result.Error == nil {
		oauthAccount = models.OAuthAccount{
			UserID:     user.ID,
			Provider:   models.ProviderGoogle,
			ProviderID: googleUser.Sub,
		}
		if err := s.db.Create(&oauthAccount).Error; err != nil {
			return exceptions.InternalServer(c, "failed to create oauth account")
		}

		return s.authResponse(c, user.ID.String())
	}

	username := s.generateUsername(googleUser.Email)
	user = models.User{
		Name:     googleUser.Name,
		Email:    googleUser.Email,
		Username: username,
		Password: "",
	}
	if err := s.db.Create(&user).Error; err != nil {
		return exceptions.InternalServer(c, "failed to create user")
	}

	oauthAccount = models.OAuthAccount{
		UserID:     user.ID,
		Provider:   models.ProviderGoogle,
		ProviderID: googleUser.Sub,
	}
	if err := s.db.Create(&oauthAccount).Error; err != nil {
		return exceptions.InternalServer(c, "failed to create oauth account")
	}

	data := map[string]interface{}{
		"id":         user.ID.String(),
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt,
	}
	jsonData, marshalErr := json.Marshal(data)
	if marshalErr == nil {
		s.rmq.Produce("user.created", string(jsonData))
	}

	return s.authResponse(c, user.ID.String())
}

func (s *oauthController) authResponse(ctx *fiber.Ctx, userID string) error {
	accessToken, err := s.GenerateToken(userID)
	if err != nil {
		return exceptions.InternalServer(ctx, "failed to generate JWT token")
	}

	refreshToken, err := s.storeRefreshToken(userID)
	if err != nil {
		return exceptions.InternalServer(ctx, "failed to generate refresh token")
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.OAuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (s *oauthController) GenerateToken(userId string) (string, error) {
	token, err := s.jwt.GenToken(userId, time.Now().Add(accessTokenDuration))
	if err != nil {
		return "", errors.Wrap(err, "failed to generate JWT token")
	}
	return token, nil
}

func (s *oauthController) storeRefreshToken(userId string) (string, error) {
	parsedUserID, err := uuid.Parse(userId)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse user ID")
	}

	s.db.Where("user_id = ?", parsedUserID).Delete(&models.RefreshToken{})

	tokenString, err := s.jwt.GenRefreshToken(userId, time.Now().Add(refreshTokenDuration))
	if err != nil {
		return "", errors.Wrap(err, "failed to generate refresh token")
	}

	hash := sha256.Sum256([]byte(tokenString))
	tokenHash := hex.EncodeToString(hash[:])

	refreshToken := models.RefreshToken{
		UserID:    parsedUserID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(refreshTokenDuration),
	}

	if err := s.db.Create(&refreshToken).Error; err != nil {
		return "", errors.Wrap(err, "failed to store refresh token")
	}

	return tokenString, nil
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *oauthController) generateUsername(email string) string {
	parts := strings.Split(email, "@")
	base := strings.ToLower(parts[0])
	username := base

	for {
		var count int64
		s.db.Model(&models.User{}).Where("username = ?", username).Count(&count)
		if count == 0 {
			return username
		}
		username = fmt.Sprintf("%s_%s", base, uuid.New().String()[:8])
	}
}

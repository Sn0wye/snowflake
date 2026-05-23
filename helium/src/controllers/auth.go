package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/getsnowflake/snowflake/helium/pkg/exceptions"
	"github.com/getsnowflake/snowflake/helium/pkg/jwt"
	"github.com/getsnowflake/snowflake/helium/pkg/logger"
	"github.com/getsnowflake/snowflake/helium/pkg/messaging"
	"github.com/getsnowflake/snowflake/helium/src/dto"
	"github.com/getsnowflake/snowflake/helium/src/models"
	"github.com/getsnowflake/snowflake/helium/src/services"
	"github.com/getsnowflake/snowflake/helium/src/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthController interface {
	Profile(ctx *fiber.Ctx) error
	Register(ctx *fiber.Ctx) error
	Login(ctx *fiber.Ctx) error
	Refresh(ctx *fiber.Ctx) error
	Logout(ctx *fiber.Ctx) error
}

type authController struct {
	db    *gorm.DB
	jwt   *jwt.JWT
	rmq   *messaging.MessagingService
	token *services.TokenService
	log   *logger.Logger
}

func NewAuthController(db *gorm.DB, jwt *jwt.JWT, rmq *messaging.MessagingService, log *logger.Logger) AuthController {
	return &authController{
		db:    db,
		jwt:   jwt,
		rmq:   rmq,
		token: services.NewTokenService(db, jwt),
		log:   log,
	}
}

// Profile godoc
//
//	@Summary		/auth/profile
//	@Description	Get user profile
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.ProfileResponse				"ProfileResponse"
//	@Failure		401	{object}	exceptions.UnauthorizedError	"Unauthorized"
//	@Failure		500	{object}	exceptions.InternalServerError	"Failed to fetch user"
//	@Security		Bearer
//	@Router			/auth/profile [get]
//	@OperationId	profile
func (s *authController) Profile(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*jwt.Claims)
	user := models.User{}

	result := s.db.Where("id = ?", claims.Subject).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return exceptions.Unauthorized(c)
		}
		return exceptions.InternalServer(c, "failed to fetch user")
	}

	return c.Status(fiber.StatusOK).JSON(dto.ProfileResponse{
		ID:           user.ID.String(),
		Name:         user.Name,
		Username:     user.Username,
		Email:        user.Email,
		AnnualIncome: user.AnnualIncome,
		Debt:         user.Debt,
		AssetsValue:  user.AssetsValue,
	})
}

// Register godoc
//
//	@Summary		/auth/register
//	@Description	Register a new user.
//	@Description	Emits: `user.created` event upon successful registration.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.RegisterRequest					true	"Register Request"
//	@Success		200		{object}	dto.RegisterResponse				"RegisterResponse"
//	@Failure		400		{object}	exceptions.BadRequestError			"Invalid request body"
//	@Failure		422		{object}	exceptions.UnprocessableEntityError	"Email already taken"
//	@Failure		500		{object}	exceptions.InternalServerError		"Failed to hash password OR Failed to marshal data OR Failed to generate JWT token"
//	@Router			/auth/register [post]
//	@OperationId	register
func (s *authController) Register(c *fiber.Ctx) error {
	db := s.db
	body := new(dto.RegisterRequest)
	if err := utils.ParseRequest(c, body); err != nil {
		return err
	}

	var user models.User
	exists := db.Where("email = ?", body.Email).First(&user).RowsAffected
	if exists > 0 {
		return exceptions.UnprocessableEntity(c, "Email already taken")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return exceptions.InternalServer(c, "failed to hash password")
	}

	user = models.User{
		Username:     body.Username,
		Password:     string(hashedPassword),
		Email:        body.Email,
		Name:         body.Name,
		AnnualIncome: body.AnnualIncome,
		Debt:         body.Debt,
		AssetsValue:  body.AssetsValue,
	}

	db.Create(&user)

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
		return exceptions.InternalServer(c, "failed to marshal data")
	}

	s.rmq.Produce("user.created", string(jsonData))

	accessToken, err := s.token.GenerateAccessToken(user.ID.String())
	if err != nil {
		return exceptions.InternalServer(c, "failed to generate JWT token")
	}

	refreshToken, err := s.token.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return exceptions.InternalServer(c, "failed to generate refresh token")
	}

	return c.Status(fiber.StatusOK).JSON(dto.RegisterResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// Login godoc
//
//	@Summary		/auth/login
//	@Description	Login
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.LoginRequest	true	"Login Request"
//	@Success		200		{object}	dto.LoginResponse
//	@Failure		400		{object}	exceptions.BadRequestError		"Invalid request body"
//	@Failure		401		{object}	exceptions.UnauthorizedError	"Wrong email or password"
//	@Failure		500		{object}	exceptions.InternalServerError	"Failed to generate JWT token"
//	@Router			/auth/login [post]
//	@OperationId	login
func (s *authController) Login(c *fiber.Ctx) error {
	db := s.db

	body := new(dto.LoginRequest)

	if err := utils.ParseRequest(c, body); err != nil {
		s.log.Warn("login parse error", zap.Error(err))

		return err
	}

	var user models.User
	db.Where("email = ?", body.Email).First(&user)

	if user.Password == "" {
		return exceptions.Unauthorized(c)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		return exceptions.Unauthorized(c)
	}

	accessToken, err := s.token.GenerateAccessToken(user.ID.String())
	if err != nil {
		return exceptions.InternalServer(c, "failed to generate JWT token")
	}

	s.token.RevokeAllUserRefreshTokens(user.ID.String())

	refreshToken, err := s.token.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return exceptions.InternalServer(c, "failed to generate refresh token")
	}

	return c.Status(fiber.StatusOK).JSON(dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// Refresh godoc
//
//	@Summary		/auth/refresh
//	@Description	Exchange a refresh token for a new access token and refresh token (token rotation).
//	@Description	The provided refresh token is revoked and a new pair is issued.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.RefreshRequest				true	"Refresh Request"
//	@Success		200		{object}	dto.RefreshResponse				"Refreshed tokens"
//	@Failure		400		{object}	exceptions.BadRequestError		"Invalid request body"
//	@Failure		401		{object}	exceptions.UnauthorizedError	"Invalid, revoked, or expired refresh token"
//	@Failure		500		{object}	exceptions.InternalServerError	"Failed to generate tokens"
//	@Router			/auth/refresh [post]
//	@OperationId	refresh
func (s *authController) Refresh(c *fiber.Ctx) error {
	body := new(dto.RefreshRequest)
	if err := utils.ParseRequest(c, body); err != nil {
		return err
	}

	claims, err := s.jwt.ParseRefreshToken(body.RefreshToken)
	if err != nil {
		return exceptions.Unauthorized(c)
	}

	hash := sha256.Sum256([]byte(body.RefreshToken))
	tokenHash := hex.EncodeToString(hash[:])

	var token models.RefreshToken
	result := s.db.Where("token_hash = ?", tokenHash).First(&token)
	if result.Error != nil {
		return exceptions.Unauthorized(c)
	}

	if token.IsExpired() {
		return exceptions.Unauthorized(c)
	}

	s.db.Delete(&token)

	userID := claims.Subject

	accessToken, err := s.token.GenerateAccessToken(userID)
	if err != nil {
		return exceptions.InternalServer(c, "failed to generate JWT token")
	}

	newRefreshToken, err := s.token.GenerateRefreshToken(userID)
	if err != nil {
		return exceptions.InternalServer(c, "failed to generate refresh token")
	}

	return c.Status(fiber.StatusOK).JSON(dto.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	})
}

// Logout godoc
//
//	@Summary		/auth/logout
//	@Description	Revoke a refresh token, invalidating the session.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body	dto.LogoutRequest	true	"Logout Request"
//	@Success		204		"No Content"
//	@Failure		400		{object}	exceptions.BadRequestError		"Invalid request body"
//	@Failure		401		{object}	exceptions.UnauthorizedError	"Invalid or already revoked refresh token"
//	@Router			/auth/logout [post]
//	@OperationId	logout
func (s *authController) Logout(c *fiber.Ctx) error {
	body := new(dto.LogoutRequest)
	if err := utils.ParseRequest(c, body); err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(body.RefreshToken))
	tokenHash := hex.EncodeToString(hash[:])

	var token models.RefreshToken
	result := s.db.Where("token_hash = ?", tokenHash).First(&token)
	if result.Error != nil {
		return exceptions.Unauthorized(c)
	}

	s.db.Delete(&token)

	return c.SendStatus(fiber.StatusNoContent)
}

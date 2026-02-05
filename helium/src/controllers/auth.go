package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	"whatisthis/pkg/exceptions"
	"whatisthis/pkg/jwt"
	"whatisthis/pkg/messaging"
	"whatisthis/src/dto"
	"whatisthis/src/models"
	"whatisthis/src/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthController interface {
	Profile(ctx *fiber.Ctx) error
	Register(ctx *fiber.Ctx) error
	Login(ctx *fiber.Ctx) error
	GenerateToken(userId string) (string, error)
}

type authController struct {
	db  *gorm.DB
	jwt *jwt.JWT
	rmq *messaging.MessagingService
}

func NewAuthController(db *gorm.DB, jwt *jwt.JWT, rmq *messaging.MessagingService) AuthController {
	return &authController{
		db:  db,
		jwt: jwt,
		rmq: rmq,
	}
}

// Profile godoc
//
//	@Summary		/auth/profile
//	@Description	Get user profile
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.ProfileResponse					"ProfileResponse"
//	@Failure		401	{object}	exceptions.UnauthorizedError		"Unauthorized"
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
//	@Description Emits: `user.created` event upon successful registration.
//	@Tags			  Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.RegisterRequest	true	"Register Request"
//	@Success		200		{object}	dto.RegisterResponse "RegisterResponse"
//	@Failure		400		{object}	exceptions.BadRequestError			"Invalid request body"
//	@Failure		422		{object}	exceptions.UnprocessableEntityError	"Email already taken"
//	@Failure		500		{object}	exceptions.InternalServerError	"Failed to hash password OR Failed to marshal data OR Failed to generate JWT token"
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

	token, err := s.GenerateToken(user.ID.String())
	if err != nil {
		return exceptions.InternalServer(c, "failed to generate JWT token")
	}

	return c.Status(fiber.StatusOK).JSON(dto.RegisterResponse{
		Token: token,
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
//	@Failure		400		{object}	exceptions.BadRequestError			"Invalid request body"
//	@Failure		401		{object}	exceptions.UnauthorizedError		"Wrong email or password"
//	@Failure		500		{object}	exceptions.InternalServerError	"Failed to generate JWT token"
//	@Router			/auth/login [post]
//	@OperationId	login
func (s *authController) Login(c *fiber.Ctx) error {
	db := s.db

	body := new(dto.LoginRequest)

	if err := utils.ParseRequest(c, body); err != nil {
		fmt.Println(err)

		return err
	}

	var user models.User
	db.Where("email = ?", body.Email).First(&user)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		return exceptions.Unauthorized(c)
	}

	token, err := s.GenerateToken(user.ID.String())
	if err != nil {
		return exceptions.InternalServer(c, "failed to generate JWT token")
	}

	return c.Status(fiber.StatusOK).JSON(dto.LoginResponse{
		Token: token,
	})
}

func (s *authController) GenerateToken(userId string) (string, error) {
	token, err := s.jwt.GenToken(userId, time.Now().Add(time.Hour*24*90)) // 90 days
	if err != nil {
		return "", errors.Wrap(err, "failed to generate JWT token")
	}

	return token, nil
}

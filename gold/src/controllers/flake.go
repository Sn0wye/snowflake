package controllers

import (
	"net/http"

	"github.com/Sn0wye/snowflake/gold/pkg/exceptions"
	"github.com/Sn0wye/snowflake/gold/pkg/jwt"
	"github.com/Sn0wye/snowflake/gold/src/dto"
	"github.com/Sn0wye/snowflake/gold/src/models"
	"github.com/Sn0wye/snowflake/gold/src/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// maxTypedFlakeKeys is the maximum number of non-unlimited flake keys per account.
// random and handle types are excluded from this limit.
const maxTypedFlakeKeys = 5

// unlimitedKeyTypes are key types not subject to the per-account limit.
var unlimitedKeyTypes = map[models.FlakeType]bool{
	models.FlakeTypeRandom: true,
	models.FlakeTypeHandle: true,
}

type FlakeController interface {
	CreateFlake(ctx *fiber.Ctx) error
	GetFlakes(ctx *fiber.Ctx) error
	DeleteFlake(ctx *fiber.Ctx) error
	PublicLookupFlake(ctx *fiber.Ctx) error
}

type flakeController struct {
	db  *gorm.DB
	jwt *jwt.JWT
}

func NewFlakeController(db *gorm.DB, jwt *jwt.JWT) FlakeController {
	return &flakeController{
		db:  db,
		jwt: jwt,
	}
}

func (s *flakeController) CreateFlake(ctx *fiber.Ctx) error {
	claims := ctx.Locals("claims").(*jwt.Claims)

	body := new(dto.CreateFlakeRequest)
	if err := utils.ParseRequest(ctx, body); err != nil {
		return err
	}

	// Validate key value format for the given key type
	if err := utils.ValidateFlakeKeyValue(body.KeyType, body.KeyValue); err != nil {
		return exceptions.BadRequest(ctx, err.Error())
	}

	var account models.Account
	if err := s.db.Where("user_id = ?", claims.Subject).First(&account).Error; err != nil {
		return exceptions.NotFound(ctx, "Account not found")
	}

	// Block transactions on accounts with reconciliation discrepancies
	if account.ReconciliationStatus == models.AccountReconciliationStatusDiscrepancy {
		return exceptions.BadRequest(ctx, "Account is under reconciliation review and cannot be modified")
	}

	// Enforce key limits for non-unlimited types
	if !unlimitedKeyTypes[body.KeyType] {
		// Check one-per-type rule
		var existing models.Flake
		if s.db.Where("key_type = ? AND account_id = ? AND status = ?",
			body.KeyType, account.ID, models.FlakeStatusActive).First(&existing).Error == nil {
			return exceptions.BadRequest(ctx, "You already have an active flake key of this type")
		}

		// Check total limit (count only non-unlimited active keys)
		var count int64
		s.db.Model(&models.Flake{}).
			Where("account_id = ? AND status = ? AND key_type NOT IN ?",
				account.ID, models.FlakeStatusActive,
				[]models.FlakeType{models.FlakeTypeRandom, models.FlakeTypeHandle}).
			Count(&count)
		if count >= maxTypedFlakeKeys {
			return exceptions.BadRequest(ctx, "Maximum number of flake keys reached (5 typed keys per account)")
		}
	}

	// Check global uniqueness of key_value
	var conflict models.Flake
	if s.db.Where("key_value = ? AND status = ?", body.KeyValue, models.FlakeStatusActive).
		First(&conflict).Error == nil {
		return exceptions.BadRequest(ctx, "This key value is already registered to another account")
	}

	flake := models.Flake{
		AccountID: account.ID,
		KeyType:   body.KeyType,
		KeyValue:  body.KeyValue,
		Status:    models.FlakeStatusActive,
	}

	if err := s.db.Create(&flake).Error; err != nil {
		return exceptions.InternalServer(ctx, "Failed to create flake key")
	}

	return ctx.Status(http.StatusCreated).JSON(dto.FlakeToResponse(flake))
}

func (s *flakeController) GetFlakes(ctx *fiber.Ctx) error {
	claims := ctx.Locals("claims").(*jwt.Claims)

	var account models.Account
	if err := s.db.Where("user_id = ?", claims.Subject).First(&account).Error; err != nil {
		return exceptions.NotFound(ctx, "Account not found")
	}

	var flakes []models.Flake
	if err := s.db.Where("account_id = ?", account.ID).Find(&flakes).Error; err != nil {
		return exceptions.InternalServer(ctx, "Failed to fetch flake keys")
	}

	response := make([]dto.FlakeResponse, len(flakes))
	for i, f := range flakes {
		response[i] = dto.FlakeToResponse(f)
	}

	return ctx.Status(http.StatusOK).JSON(response)
}

func (s *flakeController) DeleteFlake(ctx *fiber.Ctx) error {
	claims := ctx.Locals("claims").(*jwt.Claims)
	id := ctx.Params("id")

	var account models.Account
	if err := s.db.Where("user_id = ?", claims.Subject).First(&account).Error; err != nil {
		return exceptions.NotFound(ctx, "Account not found")
	}

	var flake models.Flake
	if err := s.db.Where("id = ? AND account_id = ?", id, account.ID).First(&flake).Error; err != nil {
		return exceptions.NotFound(ctx, "Flake key not found")
	}

	if flake.Status == models.FlakeStatusInactive {
		return exceptions.BadRequest(ctx, "Flake key is already inactive")
	}

	if err := s.db.Model(&flake).Update("status", models.FlakeStatusInactive).Error; err != nil {
		return exceptions.InternalServer(ctx, "Failed to deactivate flake key")
	}

	return ctx.Status(http.StatusOK).JSON(dto.FlakeToResponse(flake))
}

func (s *flakeController) PublicLookupFlake(ctx *fiber.Ctx) error {
	keyValue := ctx.Query("key_value")
	if keyValue == "" {
		return exceptions.BadRequest(ctx, "key_value query parameter is required")
	}

	var flake models.Flake
	if err := s.db.Where("key_value = ? AND status = ?", keyValue, models.FlakeStatusActive).
		First(&flake).Error; err != nil {
		return exceptions.NotFound(ctx, "Flake key not found")
	}

	var account models.Account
	if err := s.db.Where("id = ?", flake.AccountID).First(&account).Error; err != nil {
		return exceptions.NotFound(ctx, "Account not found")
	}

	return ctx.Status(http.StatusOK).JSON(dto.LookupFlakeResponse{
		AccountID: account.ID,
		KeyType:   flake.KeyType,
		KeyValue:  flake.KeyValue,
	})
}

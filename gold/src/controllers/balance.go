package controllers

import (
	"net/http"
	"strconv"

	"github.com/Sn0wye/snowflake/gold/pkg/exceptions"
	"github.com/Sn0wye/snowflake/gold/pkg/jwt"
	"github.com/Sn0wye/snowflake/gold/src/dto"
	"github.com/Sn0wye/snowflake/gold/src/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const (
	defaultBalanceHistoryLimit = 20
	maxBalanceHistoryLimit     = 100
)

type BalanceController interface {
	GetBalance(ctx *fiber.Ctx) error
	GetBalanceHistory(ctx *fiber.Ctx) error
}

type balanceController struct {
	db  *gorm.DB
	jwt *jwt.JWT
}

func NewBalanceController(db *gorm.DB, jwt *jwt.JWT) BalanceController {
	return &balanceController{
		db:  db,
		jwt: jwt,
	}
}

func (s *balanceController) GetBalance(ctx *fiber.Ctx) error {
	claims := ctx.Locals("claims").(*jwt.Claims)

	var account models.Account
	if err := s.db.Where("user_id = ?", claims.Subject).First(&account).Error; err != nil {
		return exceptions.NotFound(ctx, "Account not found")
	}

	return ctx.Status(http.StatusOK).JSON(dto.AccountToBalanceResponse(account))
}

func (s *balanceController) GetBalanceHistory(ctx *fiber.Ctx) error {
	claims := ctx.Locals("claims").(*jwt.Claims)

	var account models.Account
	if err := s.db.Where("user_id = ?", claims.Subject).First(&account).Error; err != nil {
		return exceptions.NotFound(ctx, "Account not found")
	}

	page := 1
	limit := defaultBalanceHistoryLimit

	if p := ctx.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := ctx.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			if v > maxBalanceHistoryLimit {
				v = maxBalanceHistoryLimit
			}
			limit = v
		}
	}

	offset := (page - 1) * limit

	var total int64
	s.db.Model(&models.TransactionHistory{}).
		Where("account_id = ?", account.ID).
		Count(&total)

	var entries []models.TransactionHistory
	if err := s.db.
		Where("account_id = ?", account.ID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entries).Error; err != nil {
		return exceptions.InternalServer(ctx, "Failed to fetch balance history")
	}

	response := make([]dto.BalanceHistoryEntry, len(entries))
	for i, e := range entries {
		response[i] = dto.TransactionHistoryToEntry(e)
	}

	return ctx.Status(http.StatusOK).JSON(dto.BalanceHistoryResponse{
		Entries: response,
		Total:   total,
		Page:    page,
		Limit:   limit,
	})
}

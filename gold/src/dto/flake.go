package dto

import (
	"time"

	"github.com/Sn0wye/snowflake/gold/src/models"
	"github.com/google/uuid"
)

type CreateFlakeRequest struct {
	KeyType  models.FlakeType `json:"key_type" validate:"required,oneof=email phone cpf cnpj random handle" example:"email"`
	KeyValue string           `json:"key_value" validate:"required" example:"my@example.com"`
} // @name CreateFlakeRequest

type FlakeResponse struct {
	ID        uuid.UUID          `json:"id"`
	AccountID uuid.UUID          `json:"account_id"`
	KeyType   models.FlakeType   `json:"key_type"`
	KeyValue  string             `json:"key_value"`
	Status    models.FlakeStatus `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
} // @name FlakeResponse

type LookupFlakeResponse struct {
	AccountID uuid.UUID        `json:"account_id"`
	KeyType   models.FlakeType `json:"key_type"`
	KeyValue  string           `json:"key_value"`
} // @name LookupFlakeResponse

func FlakeToResponse(f models.Flake) FlakeResponse {
	return FlakeResponse{
		ID:        f.ID,
		AccountID: f.AccountID,
		KeyType:   f.KeyType,
		KeyValue:  f.KeyValue,
		Status:    f.Status,
		CreatedAt: f.CreatedAt,
	}
}

package exceptions

import (
	"github.com/gofiber/fiber/v2"
)

type Error struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

type BadRequestError struct {
	StatusCode int    `json:"status_code" example:"400"`
	Message    string `json:"message" example:"Something went wrong :("`
} // @name BadRequestError

type InternalServerError struct {
	StatusCode int    `json:"status_code" example:"500"`
	Message    string `json:"message" example:"Internal server error :("`
} // @name InternalServerError

type UnauthorizedError struct {
	StatusCode int    `json:"status_code" example:"401"`
	Message    string `json:"message" example:"Unauthorized :("`
} // @name UnauthorizedError

type NotFoundError struct {
	StatusCode int    `json:"status_code" example:"404"`
	Message    string `json:"message" example:"Not found :("`
} // @name NotFoundError

type UnprocessableEntityError struct {
	StatusCode int    `json:"status_code" example:"422"`
	Message    string `json:"message" example:"Unprocessable entity :("`
} // @name UnprocessableEntityError

func writeError(c *fiber.Ctx, message string, statusCode int) error {
	resp := Error{
		StatusCode: statusCode,
		Message:    message,
	}

	return c.Status(statusCode).JSON(resp)
}

var (
	BadRequest = func(c *fiber.Ctx, message string) error {
		return writeError(c, message, fiber.StatusBadRequest)
	}
	InternalServer = func(c *fiber.Ctx, message string) error {
		return writeError(c, message, fiber.StatusInternalServerError)
	}
	Unauthorized = func(c *fiber.Ctx) error {
		return writeError(c, "Unauthorized", fiber.StatusUnauthorized)
	}
	NotFound = func(c *fiber.Ctx, message string) error {
		return writeError(c, message, fiber.StatusNotFound)
	}
	UnprocessableEntity = func(c *fiber.Ctx, message string) error {
		return writeError(c, message, fiber.StatusUnprocessableEntity)
	}
)

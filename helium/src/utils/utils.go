package utils

import (
	"fmt"
	"reflect"
	"snowflake/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

func ParseRequest(c *fiber.Ctx, input interface{}) error {
	if err := c.BodyParser(input); err != nil {
		fmt.Println(err)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	validationErrors := validator.ValidatorInstance.Validate(input)
	if len(validationErrors) > 0 {
		errorMap := make(map[string]string)

		for _, vErr := range validationErrors {
			fieldName := vErr.FailedField
			jsonName := getJSONFieldName(input, fieldName)

			errorMap[jsonName] = vErr.Tag
		}

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": fiber.StatusBadRequest,
			"message":     "Validation error",
			"errors":      errorMap,
		})
	}

	return nil
}

func getJSONFieldName(input interface{}, fieldName string) string {
	val := reflect.ValueOf(input)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Name == fieldName {
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" {
				return jsonTag
			}
			return field.Name
		}
	}
	return fieldName
}

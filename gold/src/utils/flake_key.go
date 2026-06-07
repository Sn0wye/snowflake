package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/getsnowflake/snowflake/gold/src/models"
)

var (
	rePhone  = regexp.MustCompile(`^\+55\d{11}$`)
	reHandle = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{2,29}$`)
	reUUIDv4 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	reEmail  = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
)

// ValidateFlakeKeyValue validates the key_value format based on key_type.
func ValidateFlakeKeyValue(keyType models.FlakeType, keyValue string) error {
	switch keyType {
	case models.FlakeTypeEmail:
		if !reEmail.MatchString(keyValue) {
			return fmt.Errorf("invalid email format")
		}
	case models.FlakeTypePhone:
		if !rePhone.MatchString(keyValue) {
			return fmt.Errorf("invalid phone format: must be +55 followed by 11 digits (e.g. +5511987654321)")
		}
	case models.FlakeTypeCPF:
		if err := validateCPF(keyValue); err != nil {
			return err
		}
	case models.FlakeTypeCNPJ:
		if err := validateCNPJ(keyValue); err != nil {
			return err
		}
	case models.FlakeTypeRandom:
		if !reUUIDv4.MatchString(strings.ToLower(keyValue)) {
			return fmt.Errorf("invalid random key: must be a valid UUID v4")
		}
	case models.FlakeTypeHandle:
		if !reHandle.MatchString(keyValue) {
			return fmt.Errorf("invalid handle: must be 3–30 characters, start with a letter or underscore, and contain only letters, digits, or underscores")
		}
	default:
		return fmt.Errorf("unsupported key type: %s", keyType)
	}
	return nil
}

// validateCPF checks the 11-digit CPF checksum.
func validateCPF(cpf string) error {
	digits := stripNonDigits(cpf)
	if len(digits) != 11 {
		return fmt.Errorf("invalid CPF: must have 11 digits")
	}
	// Reject all-same-digit CPFs (e.g. 111.111.111-11)
	if strings.Count(digits, string(digits[0])) == 11 {
		return fmt.Errorf("invalid CPF")
	}

	d := make([]int, 11)
	for i, ch := range digits {
		d[i], _ = strconv.Atoi(string(ch))
	}

	// First check digit
	sum := 0
	for i := 0; i < 9; i++ {
		sum += d[i] * (10 - i)
	}
	rem := sum % 11
	check1 := 0
	if rem >= 2 {
		check1 = 11 - rem
	}
	if d[9] != check1 {
		return fmt.Errorf("invalid CPF")
	}

	// Second check digit
	sum = 0
	for i := 0; i < 10; i++ {
		sum += d[i] * (11 - i)
	}
	rem = sum % 11
	check2 := 0
	if rem >= 2 {
		check2 = 11 - rem
	}
	if d[10] != check2 {
		return fmt.Errorf("invalid CPF")
	}

	return nil
}

// validateCNPJ checks the 14-digit CNPJ checksum.
func validateCNPJ(cnpj string) error {
	digits := stripNonDigits(cnpj)
	if len(digits) != 14 {
		return fmt.Errorf("invalid CNPJ: must have 14 digits")
	}
	if strings.Count(digits, string(digits[0])) == 14 {
		return fmt.Errorf("invalid CNPJ")
	}

	d := make([]int, 14)
	for i, ch := range digits {
		d[i], _ = strconv.Atoi(string(ch))
	}

	weights1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := 0; i < 12; i++ {
		sum += d[i] * weights1[i]
	}
	rem := sum % 11
	check1 := 0
	if rem >= 2 {
		check1 = 11 - rem
	}
	if d[12] != check1 {
		return fmt.Errorf("invalid CNPJ")
	}

	weights2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum = 0
	for i := 0; i < 13; i++ {
		sum += d[i] * weights2[i]
	}
	rem = sum % 11
	check2 := 0
	if rem >= 2 {
		check2 = 11 - rem
	}
	if d[13] != check2 {
		return fmt.Errorf("invalid CNPJ")
	}

	return nil
}

func stripNonDigits(s string) string {
	var b strings.Builder
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

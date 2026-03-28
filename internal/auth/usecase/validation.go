package usecase

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

const (
	minPasswordLength = 4
)

func initValidator(validate *validator.Validate) error {
	err := validate.RegisterValidation("username", validateUsername)
	if err != nil {
		return err
	}

	err = validate.RegisterValidation("password", validatePassword)
	if err != nil {
		return err
	}

	return nil
}

func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	validUsernameRegex := regexp.MustCompile(`^[a-zA-Zа-яА-Я0-9_.]+$`)
	if !validUsernameRegex.MatchString(username) {
		return false
	}

	if strings.HasPrefix(username, "_") || strings.HasPrefix(username, ".") ||
		strings.HasSuffix(username, "_") || strings.HasSuffix(username, ".") {
		return false
	}

	if strings.Contains(username, "__") || strings.Contains(username, "..") ||
		strings.Contains(username, "_.") || strings.Contains(username, "._") {
		return false
	}

	return true
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < minPasswordLength {
		return false
	}

	hasUppercase := regexp.MustCompile(`[A-ZА-Я]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasUppercase || !hasDigit {
		return false
	}
	return true
}

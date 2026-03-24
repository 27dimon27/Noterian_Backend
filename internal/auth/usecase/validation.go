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
	err := validate.RegisterValidation("login", validateLogin)
	if err != nil {
		return err
	}

	err = validate.RegisterValidation("password", validatePassword)
	if err != nil {
		return err
	}

	return nil
}

func validateLogin(fl validator.FieldLevel) bool {
	login := fl.Field().String()

	validLoginRegex := regexp.MustCompile(`^[a-zA-Zа-яА-Я0-9_.]+$`)
	if !validLoginRegex.MatchString(login) {
		return false
	}

	if strings.HasPrefix(login, "_") || strings.HasPrefix(login, ".") ||
		strings.HasSuffix(login, "_") || strings.HasSuffix(login, ".") {
		return false
	}

	if strings.Contains(login, "__") || strings.Contains(login, "..") ||
		strings.Contains(login, "_.") || strings.Contains(login, "._") {
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

package usecase

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestValidateUsername(t *testing.T) {
	validate := validator.New()
	err := initValidator(validate)
	if err != nil {
		t.Fatalf("Failed to init validator: %v", err)
	}

	testCases := []struct {
		name     string
		username string
		expected bool
	}{
		{"valid alphanumeric", "user123", true},
		{"valid with underscore middle", "user_name", true},
		{"valid with dot middle", "user.name", true},
		{"valid cyrillic", "пользователь", true},
		{"valid mixed", "User_123.name", true},
		{"valid single char", "a", true},

		{"starts with underscore", "_user", false},
		{"starts with dot", ".user", false},
		{"ends with underscore", "user_", false},
		{"ends with dot", "user.", false},
		{"double underscore", "user__name", false},
		{"double dot", "user..name", false},
		{"underscore dot", "user_.name", false},
		{"dot underscore", "user._name", false},
		{"contains space", "user name", false},
		{"contains special char", "user@name", false},
		{"empty string", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validate.Var(tc.username, "username")
			isValid := err == nil
			if isValid != tc.expected {
				t.Errorf("Username '%s': expected %v, got %v", tc.username, tc.expected, isValid)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	validate := validator.New()
	err := initValidator(validate)
	if err != nil {
		t.Fatalf("Failed to init validator: %v", err)
	}

	testCases := []struct {
		name     string
		password string
		expected bool
	}{
		{"valid with uppercase and digit", "Test1234", true},
		{"valid cyrillic uppercase", "Тест1234", true},
		{"valid longer password", "MyPassword123", true},
		{"valid with special chars", "Test123!@#", true},
		{"valid min length 4", "Te1s", true},

		{"too short", "Tes", false},
		{"no uppercase", "test1234", false},
		{"no digit", "Testtest", false},
		{"only lowercase", "test", false},
		{"only uppercase", "TEST", false},
		{"only digits", "123456", false},
		{"empty string", "", false},
		{"exactly min length but missing uppercase", "tes1", false},
		{"exactly min length but missing digit", "Test", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validate.Var(tc.password, "password")
			isValid := err == nil
			if isValid != tc.expected {
				t.Errorf("Password '%s': expected %v, got %v (error: %v)", tc.password, tc.expected, isValid, err)
			}
		})
	}
}

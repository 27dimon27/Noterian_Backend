package body

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

type TestUser struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Age     int    `json:"age"`
	Bio     string `json:"bio"`
	Address string `json:"address"`
}

func TestGetBody_Success(t *testing.T) {
	user := TestUser{
		Name:    "John Doe",
		Email:   "john@example.com",
		Age:     30,
		Bio:     "Normal bio",
		Address: "123 Main St",
	}

	bodyBytes, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("POST", "/test", bytes.NewReader(bodyBytes))

	var result TestUser
	err = GetBody(req, &result)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Name != user.Name {
		t.Errorf("Expected Name=%s, got %s", user.Name, result.Name)
	}
	if result.Email != user.Email {
		t.Errorf("Expected Email=%s, got %s", user.Email, result.Email)
	}
	if result.Age != user.Age {
		t.Errorf("Expected Age=%d, got %d", user.Age, result.Age)
	}
}

func TestGetBody_InvalidJSON(t *testing.T) {
	invalidJSON := `{"name": "John", "email": "john@example.com"`

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(invalidJSON))

	var result TestUser
	err := GetBody(req, &result)

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestGetBody_EmptyBody(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(""))

	var result TestUser
	err := GetBody(req, &result)

	if err == nil {
		t.Error("Expected error for empty body, got nil")
	}
}

func TestGetBody_XSSSanitization(t *testing.T) {
	user := TestUser{
		Name:    "<script>alert('xss')</script>John",
		Email:   "john@example.com",
		Bio:     "<img src=x onerror=alert(1)>",
		Address: "<b>bold</b>",
	}

	bodyBytes, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/test", bytes.NewReader(bodyBytes))

	var result TestUser
	err := GetBody(req, &result)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if strings.Contains(result.Name, "<script>") {
		t.Errorf("XSS not sanitized in Name field: %s", result.Name)
	}
	if strings.Contains(result.Bio, "onerror=") {
		t.Errorf("XSS not sanitized in Bio field: %s", result.Bio)
	}
}

func TestGetBody_NilPointer(t *testing.T) {
	user := TestUser{Name: "Test"}
	bodyBytes, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/test", bytes.NewReader(bodyBytes))

	err := GetBody(req, (*TestUser)(nil))

	if err == nil {
		t.Error("Expected error when passing nil pointer, got nil")
	}

	if !strings.Contains(err.Error(), "json") && !strings.Contains(err.Error(), "nil") {
		t.Logf("Got expected error: %v", err)
	}
}

func TestGetBody_PartialData(t *testing.T) {
	partialJSON := `{"name": "John Doe"}`

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(partialJSON))

	var result TestUser
	err := GetBody(req, &result)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Name != "John Doe" {
		t.Errorf("Expected Name='John Doe', got '%s'", result.Name)
	}

	if result.Email != "" {
		t.Errorf("Expected Email='', got '%s'", result.Email)
	}
	if result.Age != 0 {
		t.Errorf("Expected Age=0, got %d", result.Age)
	}
}

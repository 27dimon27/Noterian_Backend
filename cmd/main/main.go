package main

import (
	"log"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/run"

	_ "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/docs"
)

// @title WHITECROWSOFT API
// @version 1.0
// @description API for note-taking application

// @host localhost:8000
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in cookie
// @name token
// @description JWT token stored in cookie

// @securitydefinitions.apikey CsrfToken
// @in header
// @name X-CSRF-Token

// @accept json
// @produce json

func main() {
	if err := run.Run(); err != nil {
		log.Fatalf("Error to run application: %v", err)
	}
}

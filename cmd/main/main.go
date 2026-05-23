package main

import (
	"log"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/run"

	_ "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/docs"
)

// @title                       WHITECROWSOFT API
// @version                     1.0
// @description                 API for the Noterian note-taking application.
// @description                 Authentication is performed via a JWT-cookie set by /signup and /signin.
// @description                 State-changing endpoints additionally require an X-CSRF-Token header (obtain it from /csrf-token).

// @host                        localhost:8000
// @BasePath                    /
// @schemes                     http https

// @accept                      json
// @produce                     json

// @securityDefinitions.apikey  ApiKeyAuth
// @in                          cookie
// @name                        token
// @description                 JWT-токен сессии, выставляется сервером после /signup или /signin.

// @securityDefinitions.apikey  CsrfToken
// @in                          header
// @name                        X-CSRF-Token
// @description                 CSRF-токен, полученный из /csrf-token. Требуется для POST/PUT/DELETE.

// @tag.name auth
// @tag.description Регистрация, вход и выход пользователя
// @tag.name csrf
// @tag.description Выдача CSRF-токена
// @tag.name notes
// @tag.description Управление заметками
// @tag.name subnotes
// @tag.description Подзаметки
// @tag.name blocks
// @tag.description Блоки внутри заметки
// @tag.name attachments
// @tag.description Аттачи (изображения, gif, аудио, видео)
// @tag.name profile
// @tag.description Профиль и аватар пользователя

func main() {
	if err := run.Run(); err != nil {
		log.Fatalf("Error to run application: %v", err)
	}
}

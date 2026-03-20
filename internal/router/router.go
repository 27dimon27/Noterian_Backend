package router

import (
	"database/sql"
	"net/http"

	authHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/handler"
	authRepo "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/repository"
	authUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/usecase"

	notesHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/handler"
	notesRepo "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/repository"
	notesUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/usecase"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/middleware"
)

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func New(cfg *config.Config, db *sql.DB) (http.Handler, error) {
	userRepo, err := authRepo.NewUserRepository(db)
	if err != nil {
		return nil, err
	}

	authUsecase := authUsecase.NewAuthUsecase(userRepo, cfg.JWT)
	authHandler := authHandler.NewAuthHandler(authUsecase, cfg.JWT)

	noteRepo, err := notesRepo.NewNoteRepository(db)
	if err != nil {
		return nil, err
	}

	noteUsecase := notesUsecase.NewNoteUsecase(noteRepo)
	noteHandler := notesHandler.NewNoteHandler(noteUsecase)

	r := http.NewServeMux()

	r.HandleFunc("GET /ping", pingHandler)

	r.HandleFunc("POST /signup", authHandler.SignupUser)
	r.HandleFunc("POST /signin", authHandler.SigninUser)
	r.HandleFunc("POST /logout", authHandler.LogOutUser)

	r.Handle("GET /notes", middleware.Auth(http.HandlerFunc(noteHandler.GetAllNotes), cfg.JWT))
	r.Handle("GET /notes/{id}", middleware.Auth(http.HandlerFunc(noteHandler.GetNote), cfg.JWT))

	return middleware.Logger(r), nil
}

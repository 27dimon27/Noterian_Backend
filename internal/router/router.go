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

	profilesHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/handler"
	profilesRepo "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/repository"
	profilesUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/usecase"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/middleware"
)

func New(cfg *config.Config, db *sql.DB) (http.Handler, error) {
	userRepo := authRepo.NewUserRepository(db)

	authUsecase, err := authUsecase.NewAuthUsecase(userRepo, cfg.JWT)
	if err != nil {
		return nil, err
	}

	authHandler := authHandler.NewAuthHandler(authUsecase, cfg.JWT)

	noteRepo := notesRepo.NewNoteRepository(db)
	noteUsecase := notesUsecase.NewNoteUsecase(noteRepo)
	noteHandler := notesHandler.NewNoteHandler(noteUsecase)

	profileRepo := profilesRepo.NewProfileRepository(db)
	profileUsecase := profilesUsecase.NewProfileUsecase(profileRepo)
	profileHandler := profilesHandler.NewProfileHandler(profileUsecase, cfg.JWT)

	r := http.NewServeMux()

	r.HandleFunc("POST /signup", authHandler.SignupUser)
	r.HandleFunc("POST /signin", authHandler.SigninUser)
	r.HandleFunc("POST /logout", authHandler.LogOutUser)

	r.Handle("GET /notes", middleware.Auth(http.HandlerFunc(noteHandler.GetNotes), cfg.JWT))
	r.Handle("GET /notes/{noteId}", middleware.Auth(http.HandlerFunc(noteHandler.GetNote), cfg.JWT))
	r.Handle("POST /notes", middleware.Auth(http.HandlerFunc(noteHandler.CreateNote), cfg.JWT))
	r.Handle("PUT /notes/{noteId}", middleware.Auth(http.HandlerFunc(noteHandler.UpdateNote), cfg.JWT))
	r.Handle("DELETE /notes/{noteId}", middleware.Auth(http.HandlerFunc(noteHandler.DeleteNote), cfg.JWT))

	r.Handle("GET /notes/{noteId}/blocks/{blockId}", middleware.Auth(http.HandlerFunc(noteHandler.GetBlock), cfg.JWT))
	r.Handle("POST /notes/{noteId}/blocks", middleware.Auth(http.HandlerFunc(noteHandler.CreateBlock), cfg.JWT))
	r.Handle("PUT /notes/{noteId}/blocks/{blockId}/content", middleware.Auth(http.HandlerFunc(noteHandler.UpdateBlockContent), cfg.JWT))
	r.Handle("PUT /notes/{noteId}/blocks/{blockId}/move", middleware.Auth(http.HandlerFunc(noteHandler.MoveBlock), cfg.JWT))
	r.Handle("DELETE /notes/{noteId}/blocks/{blockId}", middleware.Auth(http.HandlerFunc(noteHandler.DeleteBlock), cfg.JWT))
	r.Handle("PUT /notes/{noteId}/blocks/{blockId}/formatting", middleware.Auth(http.HandlerFunc(noteHandler.UpdateBlockFormatting), cfg.JWT))
	r.Handle("DELETE /notes/{noteId}/blocks/{blockId}/formatting", middleware.Auth(http.HandlerFunc(noteHandler.ResetBlockFormatting), cfg.JWT))
	r.Handle("GET /notes/{noteId}/blocks/{blockId}/formatting", middleware.Auth(http.HandlerFunc(noteHandler.GetBlockFormatting), cfg.JWT))

	r.Handle("GET /profile", middleware.Auth(http.HandlerFunc(profileHandler.GetProfile), cfg.JWT))
	r.Handle("PUT /profile", middleware.Auth(http.HandlerFunc(profileHandler.UpdateProfile), cfg.JWT))
	r.Handle("DELETE /profile", middleware.Auth(http.HandlerFunc(profileHandler.DeleteProfile), cfg.JWT))

	return middleware.Logger(r), nil
}

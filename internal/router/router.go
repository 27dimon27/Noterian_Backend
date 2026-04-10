package router

import (
	"database/sql"
	"net/http"

	authHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/handler"
	authRepo "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/repository"
	authUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/usecase"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/minio"

	attachmentsHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/handler"
	attachmentsRepo "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/repository"
	attachmentsUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/usecase"

	notesHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/handler"
	notesRepo "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/repository"
	notesUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/usecase"

	profilesHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/handler"
	profilesRepo "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/repository"
	profilesUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/usecase"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/csrf"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/middleware"
)

func New(cfg *config.Config, db *sql.DB, minioService *minio.MinIOService) (http.Handler, error) {
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

	csrfHandler := csrf.NewHandler(cfg.CSRF)

	authMiddleware := func(handler http.Handler) http.Handler {
		return middleware.Auth(handler, cfg.JWT)
	}

	csrfMiddleware := func(handler http.Handler) http.Handler {
		return csrf.NewMiddleware(cfg.CSRF).Protect(handler)
	}

	attachmentRepo := attachmentsRepo.NewAttachmentRepository(db, minioService, cfg.MinIO.AttachmentsBucket)
	attachmentUsecase := attachmentsUsecase.NewAttachmentUsecase(attachmentRepo, noteRepo)
	attachmentHandler := attachmentsHandler.NewAttachmentHandler(attachmentUsecase)

	r := http.NewServeMux()

	r.HandleFunc("GET /csrf-token", csrfHandler.GetToken)
	r.HandleFunc("POST /csrf-token/refresh", csrfHandler.RefreshToken)

	r.HandleFunc("POST /signup", authHandler.SignupUser)
	r.HandleFunc("POST /signin", authHandler.SigninUser)
	r.HandleFunc("POST /logout", authHandler.LogOutUser)

	r.Handle("GET /notes", authMiddleware(http.HandlerFunc(noteHandler.GetNotes)))
	r.Handle("GET /notes/{noteId}", authMiddleware(http.HandlerFunc(noteHandler.GetNote)))
	r.Handle("POST /notes", authMiddleware(csrfMiddleware(http.HandlerFunc(noteHandler.CreateNote))))
	r.Handle("PUT /notes/{noteId}", authMiddleware(csrfMiddleware(http.HandlerFunc(noteHandler.UpdateNote))))
	r.Handle("DELETE /notes/{noteId}", authMiddleware(csrfMiddleware(http.HandlerFunc(noteHandler.DeleteNote))))

	r.Handle("GET /notes/{noteId}/blocks/{blockId}", authMiddleware(http.HandlerFunc(noteHandler.GetBlock)))
	r.Handle("POST /notes/{noteId}/blocks", authMiddleware(csrfMiddleware(http.HandlerFunc(noteHandler.CreateBlock))))
	r.Handle("PUT /notes/{noteId}/blocks/{blockId}/content", authMiddleware(csrfMiddleware(http.HandlerFunc(noteHandler.UpdateBlockContent))))
	r.Handle("PUT /notes/{noteId}/blocks/{blockId}/move", authMiddleware(csrfMiddleware(http.HandlerFunc(noteHandler.MoveBlock))))
	r.Handle("DELETE /notes/{noteId}/blocks/{blockId}", authMiddleware(csrfMiddleware(http.HandlerFunc(noteHandler.DeleteBlock))))
	r.Handle("GET /notes/{noteId}/blocks/{blockId}/formatting", authMiddleware(http.HandlerFunc(noteHandler.GetBlockFormatting)))
	r.Handle("PUT /notes/{noteId}/blocks/{blockId}/formatting", authMiddleware(csrfMiddleware(http.HandlerFunc(noteHandler.UpdateBlockFormatting))))
	r.Handle("DELETE /notes/{noteId}/blocks/{blockId}/formatting", authMiddleware(csrfMiddleware(http.HandlerFunc(noteHandler.ResetBlockFormatting))))
	r.Handle("GET /notes/{noteId}/blocks/{blockId}/attachments", middleware.Auth(http.HandlerFunc(attachmentHandler.GetAttachment), cfg.JWT))
	r.Handle("POST /notes/{noteId}/blocks/{blockId}/attachments", middleware.Auth(http.HandlerFunc(attachmentHandler.UploadAttachment), cfg.JWT))
	r.Handle("DELETE /notes/{noteId}/blocks/{blockId}/attachments", middleware.Auth(http.HandlerFunc(attachmentHandler.DeleteAttachment), cfg.JWT))

	r.Handle("GET /profile", authMiddleware(http.HandlerFunc(profileHandler.GetProfile)))
	r.Handle("PUT /profile", authMiddleware(csrfMiddleware(http.HandlerFunc(profileHandler.UpdateProfile))))
	r.Handle("DELETE /profile", authMiddleware(csrfMiddleware(http.HandlerFunc(profileHandler.DeleteProfile))))

	r.Handle("GET /profile", middleware.Auth(http.HandlerFunc(profileHandler.GetProfile), cfg.JWT))
	r.Handle("PUT /profile", middleware.Auth(http.HandlerFunc(profileHandler.UpdateProfile), cfg.JWT))
	r.Handle("DELETE /profile", middleware.Auth(http.HandlerFunc(profileHandler.DeleteProfile), cfg.JWT))

	return middleware.Logger(r), nil
}

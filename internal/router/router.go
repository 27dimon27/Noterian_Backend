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

	supportHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/support/handler"
	supportRepo "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/support/repository"
	supportUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/support/usecase"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/csrf"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/middleware"

	httpSwagger "github.com/swaggo/http-swagger"
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

	profileRepo := profilesRepo.NewProfileRepository(db, minioService, cfg.MinIO.AvatarsBucket)

	profileUsecase, err := profilesUsecase.NewProfileUsecase(profileRepo)
	if err != nil {
		return nil, err
	}

	profileHandler := profilesHandler.NewProfileHandler(profileUsecase, cfg.JWT)

	attachmentRepo := attachmentsRepo.NewAttachmentRepository(db, minioService, cfg.MinIO.AttachmentsBucket)
	attachmentUsecase := attachmentsUsecase.NewAttachmentUsecase(attachmentRepo, noteRepo)
	attachmentHandler := attachmentsHandler.NewAttachmentHandler(attachmentUsecase)

	supportRepository := supportRepo.NewSupportRepository(db)
	supportUsecase := supportUsecase.NewSupportUsecase(supportRepository)
	supportHandlerInstance := supportHandler.NewSupportHandler(supportUsecase)

	csrfHandler := csrf.NewHandler(cfg.CSRF)

	authMiddleware := func(handler http.Handler) http.Handler {
		return middleware.Auth(handler, cfg.JWT)
	}

	csrfMiddleware := func(handler http.Handler) http.Handler {
		return middleware.CSRF(handler, cfg.CSRF)
	}

	xssMiddleware := func(handler http.Handler) http.Handler {
		return middleware.XSS(handler)
	}

	securityMiddleware := func(handler http.Handler) http.Handler {
		return csrfMiddleware(xssMiddleware(handler))
	}

	r := http.NewServeMux()

	r.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("doc.json"),
	))

	r.HandleFunc("GET /csrf-token", csrfHandler.GetToken)

	r.HandleFunc("POST /signup", authHandler.SignupUser)
	r.HandleFunc("POST /signin", authHandler.SigninUser)
	r.HandleFunc("POST /logout", authHandler.LogOutUser)

	r.Handle("GET /notes", authMiddleware(http.HandlerFunc(noteHandler.GetNotes)))
	r.Handle("GET /notes/{noteId}", authMiddleware(http.HandlerFunc(noteHandler.GetNote)))
	r.Handle("POST /notes", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.CreateNote))))
	r.Handle("PUT /notes/{noteId}", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.UpdateNote))))
	r.Handle("DELETE /notes/{noteId}", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.DeleteNote))))

	r.Handle("GET /notes/{noteId}/blocks/{blockId}", authMiddleware(http.HandlerFunc(noteHandler.GetBlock)))
	r.Handle("POST /notes/{noteId}/blocks", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.CreateBlock))))
	r.Handle("PUT /notes/{noteId}/blocks/{blockId}/content", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.UpdateBlockContent))))
	r.Handle("PUT /notes/{noteId}/blocks/{blockId}/move", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.MoveBlock))))
	r.Handle("DELETE /notes/{noteId}/blocks/{blockId}", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.DeleteBlock))))
	r.Handle("GET /notes/{noteId}/blocks/{blockId}/formatting", authMiddleware(http.HandlerFunc(noteHandler.GetBlockFormatting)))
	r.Handle("PUT /notes/{noteId}/blocks/{blockId}/formatting", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.UpdateBlockFormatting))))
	r.Handle("DELETE /notes/{noteId}/blocks/{blockId}/formatting", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.ResetBlockFormatting))))
	r.Handle("GET /notes/{noteId}/blocks/{blockId}/attachments", authMiddleware(http.HandlerFunc(attachmentHandler.GetAttachment)))
	r.Handle("POST /notes/{noteId}/blocks/{blockId}/attachments", authMiddleware(securityMiddleware(http.HandlerFunc(attachmentHandler.UploadAttachment))))
	r.Handle("DELETE /notes/{noteId}/blocks/{blockId}/attachments", authMiddleware(securityMiddleware(http.HandlerFunc(attachmentHandler.DeleteAttachment))))

	r.Handle("GET /profile", authMiddleware(http.HandlerFunc(profileHandler.GetProfile)))
	r.Handle("PUT /profile", authMiddleware(securityMiddleware(http.HandlerFunc(profileHandler.UpdateProfile))))
	r.Handle("DELETE /profile", authMiddleware(securityMiddleware(http.HandlerFunc(profileHandler.DeleteProfile))))
	r.Handle("GET /profile/avatar", authMiddleware(http.HandlerFunc(profileHandler.GetAvatar)))
	r.Handle("POST /profile/avatar", authMiddleware(securityMiddleware(http.HandlerFunc(profileHandler.UploadAvatar))))
	r.Handle("DELETE /profile/avatar", authMiddleware(securityMiddleware(http.HandlerFunc(profileHandler.DeleteAvatar))))
	r.Handle("PUT /profile/password", authMiddleware(securityMiddleware(http.HandlerFunc(profileHandler.ChangePassword))))

	r.Handle("GET /support/categories", http.HandlerFunc(supportHandlerInstance.GetCategories))
	r.Handle("GET /support/statuses", http.HandlerFunc(supportHandlerInstance.GetStatuses))
	r.Handle("POST /support/tickets", authMiddleware(securityMiddleware(http.HandlerFunc(supportHandlerInstance.CreateTicket))))
	r.Handle("GET /support/tickets", authMiddleware(http.HandlerFunc(supportHandlerInstance.GetUserTickets)))
	r.Handle("GET /support/tickets/{id}", authMiddleware(http.HandlerFunc(supportHandlerInstance.GetTicket)))
	r.Handle("PATCH /support/tickets/{id}", authMiddleware(securityMiddleware(http.HandlerFunc(supportHandlerInstance.UpdateTicket))))
	r.Handle("POST /support/tickets/{id}/close", authMiddleware(securityMiddleware(http.HandlerFunc(supportHandlerInstance.CloseTicket))))
	r.Handle("POST /support/tickets/{ticket_id}/messages", authMiddleware(securityMiddleware(http.HandlerFunc(supportHandlerInstance.CreateMessage))))
	r.Handle("GET /support/tickets/{ticket_id}/messages", authMiddleware(http.HandlerFunc(supportHandlerInstance.GetTicketMessages)))
	r.Handle("POST /support/tickets/{ticket_id}/rating", authMiddleware(securityMiddleware(http.HandlerFunc(supportHandlerInstance.CreateRating))))
	r.Handle("GET /support/stats", authMiddleware(http.HandlerFunc(supportHandlerInstance.GetStats)))
	// r.Handle("GET /support/user-stats", authMiddleware(http.HandlerFunc(supportHandlerInstance.GetUserStats)))

	return middleware.Logger(r), nil
}

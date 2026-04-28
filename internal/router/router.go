package router

import (
	"context"
	"database/sql"
	"net/http"

	authHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/handler"
	authRepo "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/repository"
	authUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/usecase"

	attachmentsHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/handler"
	attachmentsRepo "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/repository"
	attachmentsUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/usecase"

	notesHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/handler"
	notesRepo "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/repository"
	notesUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/usecase"

	profilesHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/handler"
	profilesRepo "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/repository"
	profilesUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/usecase"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/csrf"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/metrics"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/minio"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/websocket"
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

	wsHub := websocket.NewHub(noteUsecase, profileUsecase)
	go wsHub.Run(context.Background())

	wsHandler := websocket.NewWebSocketHandler(wsHub, noteUsecase, profileUsecase)

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

	r.Handle("GET /metrics", metrics.MetricsHandler())

	r.HandleFunc("GET /csrf-token", csrfHandler.GetToken)

	r.HandleFunc("POST /signup", authHandler.SignupUser)
	r.HandleFunc("POST /signin", authHandler.SigninUser)
	r.HandleFunc("POST /logout", authHandler.LogOutUser)

	r.Handle("GET /notes", authMiddleware(http.HandlerFunc(noteHandler.GetNotes)))
	r.Handle("GET /notes/{noteId}", authMiddleware(http.HandlerFunc(noteHandler.GetNote)))
	r.Handle("POST /notes", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.CreateNote))))
	r.Handle("PUT /notes/{noteId}", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.UpdateNote))))
	r.Handle("DELETE /notes/{noteId}", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.DeleteNote))))

	r.Handle("GET /notes/{noteId}/subnote", authMiddleware(http.HandlerFunc(noteHandler.GetSubnotes)))
	r.Handle("POST /notes/{noteId}/subnote", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.CreateSubnote))))
	r.Handle("DELETE /notes/{noteId}/subnote/{subnoteId}", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.DeleteSubnote))))

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

	r.Handle("GET /ws/notes/{noteId}", authMiddleware(http.HandlerFunc(wsHandler.ServeWS)))

	return metrics.MetricsMiddleware(middleware.Logger(r)), nil
}

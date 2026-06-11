package router

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"

	attachmentsGrpcClient "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/grpcclient"
	attachmentsHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/handler/http"
	attachmentsRepository "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/repository"
	attachmentsUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/usecase"
	"google.golang.org/grpc"

	authGrpcClient "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/grpcclient"
	authHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/handler/http"
	authUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/usecase"

	notesGrpcClient "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/grpcclient"
	notesHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/handler/http"
	notesRepository "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/repository"
	notesUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/usecase"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/onboarding"

	profilesHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/handler/http"
	profilesRepository "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/repository"
	profilesUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/usecase"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/csrf"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/metrics"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/minio"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/websocket"
)

func New(cfg *config.Config, logger *slog.Logger, db *sql.DB, minioService *minio.MinIOService, attachmentsConn, notesConn, profilesConn *grpc.ClientConn) (http.Handler, error) {
	attachmentRepository := attachmentsRepository.NewAttachmentRepository(db, minioService, cfg.MinIO.AttachmentsBucket, cfg.MinIO.HeadersBucket)
	noteRepository := notesRepository.NewNoteRepository(db)
	profileRepository := profilesRepository.NewProfileRepository(db, minioService, cfg.MinIO.AvatarsBucket)

	attachmentRemoteRepository, err := notesGrpcClient.NewAttachmentsServiceClient(cfg.Services.AttachmentsAddr)
	if err != nil {
		logger.Error("Failed to init attachments grpc client", "error", err)
		return nil, err
	}

	noteRemoteRepository, err := attachmentsGrpcClient.NewNotesServiceClient(cfg.Services.NotesAddr)
	if err != nil {
		logger.Error("Failed to init notes grpc client", "error", err)
		return nil, err
	}

	profileRemoteRepository, err := authGrpcClient.NewProfilesServiceClient(cfg.Services.ProfilesAddr)
	if err != nil {
		logger.Error("Failed to init profiles grpc client", "error", err)
		return nil, err
	}

	attachmentUsecase := attachmentsUsecase.NewAttachmentUsecase(attachmentRepository, noteRemoteRepository)
	onboardingSeeder := onboarding.NewSeeder(noteRepository)
	authUsecase, err := authUsecase.NewAuthUsecase(profileRemoteRepository, cfg.JWT, onboardingSeeder)
	if err != nil {
		logger.Error("Failed to init auth usecase", "error", err)
		return nil, err
	}
	noteUsecase := notesUsecase.NewNoteUsecase(noteRepository, attachmentRemoteRepository)
	profileUsecase, err := profilesUsecase.NewProfileUsecase(profileRepository)
	if err != nil {
		logger.Error("Failed to init profiles usecase", "error", err)
		return nil, err
	}

	attachmentHandler := attachmentsHandler.NewAttachmentHandler(attachmentUsecase, logger)
	authHandler := authHandler.NewAuthHandler(authUsecase, cfg.JWT, logger)
	noteHandler := notesHandler.NewNoteHandler(noteUsecase, logger)
	profileHandler := profilesHandler.NewProfileHandler(profileUsecase, cfg.JWT, logger)

	wsHub := websocket.NewHub(noteUsecase, profileUsecase, attachmentUsecase)
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
	r.HandleFunc("POST /logout", authHandler.LogoutUser)

	r.HandleFunc("GET /public/notes/{noteId}", noteHandler.GetPublicNote)

	r.Handle("GET /notes", authMiddleware(http.HandlerFunc(noteHandler.GetNotes)))
	r.Handle("GET /notes/{noteId}", authMiddleware(http.HandlerFunc(noteHandler.GetNote)))
	r.Handle("GET /notes/{noteId}/pdf", authMiddleware(http.HandlerFunc(noteHandler.GetNotePDF)))
	r.Handle("POST /notes", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.CreateNote))))
	r.Handle("PUT /notes/{noteId}", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.UpdateNote))))
	r.Handle("DELETE /notes/{noteId}", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.DeleteNote))))

	r.Handle("GET /notes/{noteId}/subnote", authMiddleware(http.HandlerFunc(noteHandler.GetSubnotes)))
	r.Handle("POST /notes/{noteId}/subnote", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.CreateSubnote))))
	r.Handle("DELETE /notes/{noteId}/subnote/{subnoteId}", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.DeleteSubnote))))

	r.Handle("POST /notes/{noteId}/blocks", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.CreateBlock))))
	r.Handle("PUT /notes/{noteId}/blocks/{blockId}/content", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.UpdateBlockContent))))
	r.Handle("PUT /notes/{noteId}/blocks/{blockId}/move", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.MoveBlock))))
	r.Handle("DELETE /notes/{noteId}/blocks/{blockId}", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.DeleteBlock))))
	r.Handle("PUT /notes/{noteId}/blocks/{blockId}/formatting", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandler.UpdateBlockFormatting))))

	r.Handle("GET /notes/{noteId}/header", authMiddleware(http.HandlerFunc(attachmentHandler.GetHeader)))
	r.Handle("POST /notes/{noteId}/header", authMiddleware(securityMiddleware(http.HandlerFunc(attachmentHandler.UploadHeader))))
	r.Handle("DELETE /notes/{noteId}/header", authMiddleware(securityMiddleware(http.HandlerFunc(attachmentHandler.DeleteHeader))))

	// для обратной совместимости
	r.Handle("GET /notes/{noteId}/blocks/{blockId}/attachments", authMiddleware(http.HandlerFunc(attachmentHandler.GetAttachment)))
	r.Handle("POST /notes/{noteId}/attachments", authMiddleware(securityMiddleware(http.HandlerFunc(attachmentHandler.UploadAttachment))))
	r.Handle("DELETE /notes/{noteId}/blocks/{blockId}/attachments", authMiddleware(securityMiddleware(http.HandlerFunc(attachmentHandler.DeleteAttachment))))

	r.Handle("GET /profile", authMiddleware(http.HandlerFunc(profileHandler.GetProfile)))
	r.Handle("PUT /profile", authMiddleware(securityMiddleware(http.HandlerFunc(profileHandler.UpdateProfile))))
	r.Handle("DELETE /profile", authMiddleware(securityMiddleware(http.HandlerFunc(profileHandler.DeleteProfile))))

	// для обратной совместимости
	r.Handle("GET /profile/avatar", authMiddleware(http.HandlerFunc(profileHandler.GetAvatar)))

	r.Handle("POST /profile/avatar", authMiddleware(securityMiddleware(http.HandlerFunc(profileHandler.UploadAvatar))))
	r.Handle("DELETE /profile/avatar", authMiddleware(securityMiddleware(http.HandlerFunc(profileHandler.DeleteAvatar))))
	r.Handle("PUT /profile/password", authMiddleware(securityMiddleware(http.HandlerFunc(profileHandler.ChangePassword))))

	r.Handle("GET /ws/notes/{noteId}", authMiddleware(http.HandlerFunc(wsHandler.ServeWS)))

	return metrics.MetricsMiddleware(middleware.Logger(r, logger)), nil
}

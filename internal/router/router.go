package router

import (
	"context"
	"net/http"

	authGrpcClient "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/grpc"
	authHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/handler"

	attachmentsGrpcClient "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/grpc"
	attachmentsHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/handler"

	notesGrpcClient "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/grpc"
	notesHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/handler"

	profilesGrpcClient "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/grpc"
	profilesHandler "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/handler"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/csrf"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/metrics"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/websocket"
	"google.golang.org/grpc"
)

func New(cfg *config.Config) (http.Handler, error) {
	authClient, err := authGrpcClient.NewAuthGrpcClient(
		cfg.AuthService.Address(),
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	noteClient, err := notesGrpcClient.NewNoteGrpcClient(
		cfg.NotesService.Address(),
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	profileClient, err := profilesGrpcClient.NewProfileGrpcClient(
		cfg.ProfilesService.Address(),
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	attachmentClient, err := attachmentsGrpcClient.NewAttachmentGrpcClient(
		cfg.AttachmentsService.Address(),
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	authHandlerObj := authHandler.NewAuthHandler(authClient, cfg.JWT)
	noteHandlerObj := notesHandler.NewNoteHandler(noteClient)
	profileHandlerObj := profilesHandler.NewProfileHandler(profileClient, cfg.JWT)
	attachmentHandlerObj := attachmentsHandler.NewAttachmentHandler(attachmentClient)

	wsHub := websocket.NewHub(noteClient, profileClient)
	go wsHub.Run(context.Background())

	wsHandler := websocket.NewWebSocketHandler(wsHub, noteClient, profileClient)

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

	r.HandleFunc("POST /signup", authHandlerObj.SignupUser)
	r.HandleFunc("POST /signin", authHandlerObj.SigninUser)
	r.HandleFunc("POST /logout", authHandlerObj.LogOutUser)

	r.Handle("GET /notes", authMiddleware(http.HandlerFunc(noteHandlerObj.GetNotes)))
	r.Handle("GET /notes/{noteId}", authMiddleware(http.HandlerFunc(noteHandlerObj.GetNote)))
	r.Handle("POST /notes", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandlerObj.CreateNote))))
	r.Handle("PUT /notes/{noteId}", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandlerObj.UpdateNote))))
	r.Handle("DELETE /notes/{noteId}", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandlerObj.DeleteNote))))

	r.Handle("GET /notes/{noteId}/subnote", authMiddleware(http.HandlerFunc(noteHandlerObj.GetSubnotes)))
	r.Handle("POST /notes/{noteId}/subnote", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandlerObj.CreateSubnote))))
	r.Handle("DELETE /notes/{noteId}/subnote/{subnoteId}", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandlerObj.DeleteSubnote))))

	r.Handle("GET /notes/{noteId}/blocks/{blockId}", authMiddleware(http.HandlerFunc(noteHandlerObj.GetBlock)))
	r.Handle("POST /notes/{noteId}/blocks", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandlerObj.CreateBlock))))
	r.Handle("PUT /notes/{noteId}/blocks/{blockId}/content", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandlerObj.UpdateBlockContent))))
	r.Handle("PUT /notes/{noteId}/blocks/{blockId}/move", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandlerObj.MoveBlock))))
	r.Handle("DELETE /notes/{noteId}/blocks/{blockId}", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandlerObj.DeleteBlock))))
	r.Handle("GET /notes/{noteId}/blocks/{blockId}/formatting", authMiddleware(http.HandlerFunc(noteHandlerObj.GetBlockFormatting)))
	r.Handle("PUT /notes/{noteId}/blocks/{blockId}/formatting", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandlerObj.UpdateBlockFormatting))))
	r.Handle("DELETE /notes/{noteId}/blocks/{blockId}/formatting", authMiddleware(securityMiddleware(http.HandlerFunc(noteHandlerObj.ResetBlockFormatting))))
	r.Handle("GET /notes/{noteId}/blocks/{blockId}/attachments", authMiddleware(http.HandlerFunc(attachmentHandlerObj.GetAttachment)))
	r.Handle("POST /notes/{noteId}/blocks/{blockId}/attachments", authMiddleware(securityMiddleware(http.HandlerFunc(attachmentHandlerObj.UploadAttachment))))
	r.Handle("DELETE /notes/{noteId}/blocks/{blockId}/attachments", authMiddleware(securityMiddleware(http.HandlerFunc(attachmentHandlerObj.DeleteAttachment))))

	r.Handle("GET /profile", authMiddleware(http.HandlerFunc(profileHandlerObj.GetProfile)))
	r.Handle("PUT /profile", authMiddleware(securityMiddleware(http.HandlerFunc(profileHandlerObj.UpdateProfile))))
	r.Handle("DELETE /profile", authMiddleware(securityMiddleware(http.HandlerFunc(profileHandlerObj.DeleteProfile))))
	r.Handle("GET /profile/avatar", authMiddleware(http.HandlerFunc(profileHandlerObj.GetAvatar)))
	r.Handle("POST /profile/avatar", authMiddleware(securityMiddleware(http.HandlerFunc(profileHandlerObj.UploadAvatar))))
	r.Handle("DELETE /profile/avatar", authMiddleware(securityMiddleware(http.HandlerFunc(profileHandlerObj.DeleteAvatar))))
	r.Handle("PUT /profile/password", authMiddleware(securityMiddleware(http.HandlerFunc(profileHandlerObj.ChangePassword))))

	r.Handle("GET /ws/notes/{noteId}", authMiddleware(http.HandlerFunc(wsHandler.ServeWS)))

	return metrics.MetricsMiddleware(middleware.Logger(r)), nil
}

package middleware

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// AuthInterceptor извлекает userID из metadata и добавляет в контекст
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Пропускаем методы, не требующие авторизации
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		userIDs := md.Get("user-id")
		if len(userIDs) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing user-id")
		}

		userID, err := uuid.Parse(userIDs[0])
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid user-id")
		}

		ctx = context.WithValue(ctx, UserIDKey, userID)
		return handler(ctx, req)
	}
}

// AuthStreamInterceptor для потоковых методов
func AuthStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if isPublicMethod(info.FullMethod) {
			return handler(srv, ss)
		}

		ctx := ss.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Error(codes.Unauthenticated, "missing metadata")
		}

		userIDs := md.Get("user-id")
		if len(userIDs) == 0 {
			return status.Error(codes.Unauthenticated, "missing user-id")
		}

		userID, err := uuid.Parse(userIDs[0])
		if err != nil {
			return status.Error(codes.Unauthenticated, "invalid user-id")
		}

		ctx = context.WithValue(ctx, UserIDKey, userID)
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}
		return handler(srv, wrappedStream)
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

func isPublicMethod(method string) bool {
	publicMethods := map[string]bool{
		"/auth.AuthService/SignupUser": true,
		"/auth.AuthService/SigninUser": true,
	}
	return publicMethods[method]
}

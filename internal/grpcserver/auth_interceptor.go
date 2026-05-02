package grpcserver

import (
	"context"
	"strings"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/jwt"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryAuthInterceptor(secret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx, err := authorize(ctx, secret)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func StreamAuthInterceptor(secret string) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx, err := authorize(stream.Context(), secret)
		if err != nil {
			return err
		}

		wrapped := &wrappedServerStream{ServerStream: stream, ctx: ctx}
		return handler(srv, wrapped)
	}
}

func authorize(ctx context.Context, secret string) (context.Context, error) {
	token, err := extractBearerToken(ctx)
	if err != nil {
		return nil, err
	}

	payload, err := jwt.ValidateToken(token, secret)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid auth token")
	}

	uid, err := uuid.Parse(payload.UserID)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid user id in token")
	}

	return context.WithValue(ctx, types.UserIDKey, uid), nil
}

func extractBearerToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	if authorization := md.Get("authorization"); len(authorization) > 0 {
		return parseBearerToken(authorization[0])
	}

	token := md.Get("token")
	if len(token) > 0 {
		return token[0], nil
	}

	return "", status.Error(codes.Unauthenticated, "authorization token is required")
}

func parseBearerToken(header string) (string, error) {
	header = strings.TrimSpace(header)
	if len(header) > 7 && strings.EqualFold(header[:7], "bearer ") {
		return strings.TrimSpace(header[7:]), nil
	}
	return "", status.Error(codes.Unauthenticated, "authorization header must contain Bearer token")
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

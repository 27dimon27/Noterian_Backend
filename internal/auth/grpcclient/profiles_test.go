package grpcclient

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	profilesgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/profiles/grpc/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeProfileServiceClient struct {
	SignupUserFunc func(ctx context.Context, in *profilesgrpc.SignupUserRequest, opts ...grpc.CallOption) (*profilesgrpc.ProfileResponse, error)
	SigninUserFunc func(ctx context.Context, in *profilesgrpc.SigninUserRequest, opts ...grpc.CallOption) (*profilesgrpc.ProfileResponse, error)
}

func (f *fakeProfileServiceClient) SignupUser(ctx context.Context, in *profilesgrpc.SignupUserRequest, opts ...grpc.CallOption) (*profilesgrpc.ProfileResponse, error) {
	return f.SignupUserFunc(ctx, in, opts...)
}

func (f *fakeProfileServiceClient) SigninUser(ctx context.Context, in *profilesgrpc.SigninUserRequest, opts ...grpc.CallOption) (*profilesgrpc.ProfileResponse, error) {
	return f.SigninUserFunc(ctx, in, opts...)
}

func TestSignupUser(t *testing.T) {
	t.Run("forwards args and returns response", func(t *testing.T) {
		want := &profilesgrpc.ProfileResponse{Id: "id-1", Username: "alice"}
		fake := &fakeProfileServiceClient{
			SignupUserFunc: func(ctx context.Context, in *profilesgrpc.SignupUserRequest, opts ...grpc.CallOption) (*profilesgrpc.ProfileResponse, error) {
				if in.Username != "alice" || in.Password != "GoodPass1" {
					t.Errorf("unexpected args: %+v", in)
				}
				return want, nil
			},
		}
		c := &profilesServiceClient{client: fake}

		got, err := c.SignupUser(context.Background(), "alice", "GoodPass1")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if got != want {
			t.Errorf("expected %+v, got %+v", want, got)
		}
	})

	t.Run("AlreadyExists -> ErrUserExist", func(t *testing.T) {
		fake := &fakeProfileServiceClient{
			SignupUserFunc: func(ctx context.Context, in *profilesgrpc.SignupUserRequest, opts ...grpc.CallOption) (*profilesgrpc.ProfileResponse, error) {
				return nil, status.Error(codes.AlreadyExists, "exists")
			},
		}
		c := &profilesServiceClient{client: fake}

		_, err := c.SignupUser(context.Background(), "alice", "GoodPass1")
		if !errors.Is(err, auth.ErrUserExist) {
			t.Fatalf("expected ErrUserExist, got %v", err)
		}
	})

	t.Run("InvalidArgument -> ErrBadCredentials", func(t *testing.T) {
		fake := &fakeProfileServiceClient{
			SignupUserFunc: func(ctx context.Context, in *profilesgrpc.SignupUserRequest, opts ...grpc.CallOption) (*profilesgrpc.ProfileResponse, error) {
				return nil, status.Error(codes.InvalidArgument, "bad")
			},
		}
		c := &profilesServiceClient{client: fake}

		_, err := c.SignupUser(context.Background(), "alice", "GoodPass1")
		if !errors.Is(err, auth.ErrBadCredentials) {
			t.Fatalf("expected ErrBadCredentials, got %v", err)
		}
	})

	t.Run("other code -> raw error", func(t *testing.T) {
		st := status.Error(codes.Internal, "boom")
		fake := &fakeProfileServiceClient{
			SignupUserFunc: func(ctx context.Context, in *profilesgrpc.SignupUserRequest, opts ...grpc.CallOption) (*profilesgrpc.ProfileResponse, error) {
				return nil, st
			},
		}
		c := &profilesServiceClient{client: fake}

		_, err := c.SignupUser(context.Background(), "alice", "GoodPass1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if errors.Is(err, auth.ErrUserExist) || errors.Is(err, auth.ErrBadCredentials) {
			t.Fatalf("did not expect mapped error, got %v", err)
		}
	})

	t.Run("non-status error returned as-is", func(t *testing.T) {
		raw := errors.New("network")
		fake := &fakeProfileServiceClient{
			SignupUserFunc: func(ctx context.Context, in *profilesgrpc.SignupUserRequest, opts ...grpc.CallOption) (*profilesgrpc.ProfileResponse, error) {
				return nil, raw
			},
		}
		c := &profilesServiceClient{client: fake}

		_, err := c.SignupUser(context.Background(), "alice", "GoodPass1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestSigninUser(t *testing.T) {
	t.Run("forwards args and returns response", func(t *testing.T) {
		want := &profilesgrpc.ProfileResponse{Id: "id-1", Username: "alice"}
		fake := &fakeProfileServiceClient{
			SigninUserFunc: func(ctx context.Context, in *profilesgrpc.SigninUserRequest, opts ...grpc.CallOption) (*profilesgrpc.ProfileResponse, error) {
				if in.Username != "alice" {
					t.Errorf("unexpected args: %+v", in)
				}
				return want, nil
			},
		}
		c := &profilesServiceClient{client: fake}

		got, err := c.SigninUser(context.Background(), "alice")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if got != want {
			t.Errorf("expected %+v, got %+v", want, got)
		}
	})

	t.Run("NotFound -> ErrUserNotExist", func(t *testing.T) {
		fake := &fakeProfileServiceClient{
			SigninUserFunc: func(ctx context.Context, in *profilesgrpc.SigninUserRequest, opts ...grpc.CallOption) (*profilesgrpc.ProfileResponse, error) {
				return nil, status.Error(codes.NotFound, "not found")
			},
		}
		c := &profilesServiceClient{client: fake}

		_, err := c.SigninUser(context.Background(), "alice")
		if !errors.Is(err, auth.ErrUserNotExist) {
			t.Fatalf("expected ErrUserNotExist, got %v", err)
		}
	})

	t.Run("InvalidArgument -> ErrBadCredentials", func(t *testing.T) {
		fake := &fakeProfileServiceClient{
			SigninUserFunc: func(ctx context.Context, in *profilesgrpc.SigninUserRequest, opts ...grpc.CallOption) (*profilesgrpc.ProfileResponse, error) {
				return nil, status.Error(codes.InvalidArgument, "bad")
			},
		}
		c := &profilesServiceClient{client: fake}

		_, err := c.SigninUser(context.Background(), "alice")
		if !errors.Is(err, auth.ErrBadCredentials) {
			t.Fatalf("expected ErrBadCredentials, got %v", err)
		}
	})

	t.Run("other code -> raw error", func(t *testing.T) {
		fake := &fakeProfileServiceClient{
			SigninUserFunc: func(ctx context.Context, in *profilesgrpc.SigninUserRequest, opts ...grpc.CallOption) (*profilesgrpc.ProfileResponse, error) {
				return nil, status.Error(codes.Internal, "boom")
			},
		}
		c := &profilesServiceClient{client: fake}

		_, err := c.SigninUser(context.Background(), "alice")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if errors.Is(err, auth.ErrUserNotExist) || errors.Is(err, auth.ErrBadCredentials) {
			t.Fatalf("did not expect mapped error, got %v", err)
		}
	})

	t.Run("non-status error returned as-is", func(t *testing.T) {
		raw := errors.New("network")
		fake := &fakeProfileServiceClient{
			SigninUserFunc: func(ctx context.Context, in *profilesgrpc.SigninUserRequest, opts ...grpc.CallOption) (*profilesgrpc.ProfileResponse, error) {
				return nil, raw
			},
		}
		c := &profilesServiceClient{client: fake}

		_, err := c.SigninUser(context.Background(), "alice")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestNewProfilesServiceClientAndClose(t *testing.T) {
	client, err := NewProfilesServiceClient("passthrough:///localhost:0")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if err := client.Close(); err != nil {
		t.Fatalf("unexpected close err: %v", err)
	}
}

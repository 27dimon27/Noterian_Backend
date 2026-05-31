package usecase

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/grpcclient/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	profilesgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/profiles/grpc/gen"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

var testJWT = config.JWTConfig{
	Secret:     "test-secret",
	CookieName: "session",
	CookieTime: time.Hour,
	Secure:     false,
}

func newUsecase(t *testing.T, c *mocks.MockProfilesServiceClient) *authUsecase {
	t.Helper()
	u, err := NewAuthUsecase(c, testJWT)
	if err != nil {
		t.Fatalf("NewAuthUsecase: %v", err)
	}
	return u
}

func TestNewAuthUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mocks.NewMockProfilesServiceClient(ctrl)

	u, err := NewAuthUsecase(client, testJWT)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u == nil {
		t.Fatal("expected non-nil usecase")
	}
}

func TestSignupUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("invalid username", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		client := mocks.NewMockProfilesServiceClient(ctrl)
		u := newUsecase(t, client)

		_, err := u.SignupUser(ctx, "_bad", "GoodPass1")
		if !errors.Is(err, auth.ErrInvalidUsername) {
			t.Fatalf("expected ErrInvalidUsername, got %v", err)
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		client := mocks.NewMockProfilesServiceClient(ctrl)
		u := newUsecase(t, client)

		_, err := u.SignupUser(ctx, "alice", "short")
		if !errors.Is(err, auth.ErrInvalidPassword) {
			t.Fatalf("expected ErrInvalidPassword, got %v", err)
		}
	})

	t.Run("user already exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		client := mocks.NewMockProfilesServiceClient(ctrl)
		u := newUsecase(t, client)

		client.EXPECT().SignupUser(ctx, "alice", "GoodPass1").Return(nil, auth.ErrUserExist)

		_, err := u.SignupUser(ctx, "alice", "GoodPass1")
		if !errors.Is(err, auth.ErrUserExist) {
			t.Fatalf("expected ErrUserExist, got %v", err)
		}
	})

	t.Run("client error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		client := mocks.NewMockProfilesServiceClient(ctrl)
		u := newUsecase(t, client)

		boom := errors.New("boom")
		client.EXPECT().SignupUser(ctx, "alice", "GoodPass1").Return(nil, boom)

		_, err := u.SignupUser(ctx, "alice", "GoodPass1")
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})

	t.Run("invalid uuid in profile", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		client := mocks.NewMockProfilesServiceClient(ctrl)
		u := newUsecase(t, client)

		client.EXPECT().SignupUser(ctx, "alice", "GoodPass1").Return(
			&profilesgrpc.ProfileResponse{Id: "not-a-uuid", Username: "alice"}, nil,
		)

		if _, err := u.SignupUser(ctx, "alice", "GoodPass1"); err == nil {
			t.Fatal("expected uuid parse error, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		client := mocks.NewMockProfilesServiceClient(ctrl)
		u := newUsecase(t, client)

		client.EXPECT().SignupUser(ctx, "alice", "GoodPass1").Return(
			&profilesgrpc.ProfileResponse{Id: userID.String(), Username: "alice", Avatar: "ava"}, nil,
		)

		got, err := u.SignupUser(ctx, "alice", "GoodPass1")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if got.ID != userID || got.Username != "alice" || got.Avatar != "ava" {
			t.Errorf("unexpected profile: %+v", got)
		}
	})
}

func TestSigninUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	rawPassword := "GoodPass1"
	hashed, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}

	t.Run("user not exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		client := mocks.NewMockProfilesServiceClient(ctrl)
		u := newUsecase(t, client)

		client.EXPECT().SigninUser(ctx, "alice").Return(nil, auth.ErrUserNotExist)

		_, err := u.SigninUser(ctx, "alice", rawPassword)
		if !errors.Is(err, auth.ErrUserNotExist) {
			t.Fatalf("expected ErrUserNotExist, got %v", err)
		}
	})

	t.Run("client error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		client := mocks.NewMockProfilesServiceClient(ctrl)
		u := newUsecase(t, client)

		boom := errors.New("boom")
		client.EXPECT().SigninUser(ctx, "alice").Return(nil, boom)

		_, err := u.SigninUser(ctx, "alice", rawPassword)
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})

	t.Run("bad password", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		client := mocks.NewMockProfilesServiceClient(ctrl)
		u := newUsecase(t, client)

		client.EXPECT().SigninUser(ctx, "alice").Return(
			&profilesgrpc.ProfileResponse{Id: userID.String(), Username: "alice", Password: string(hashed)}, nil,
		)

		_, err := u.SigninUser(ctx, "alice", "WrongPass1")
		if !errors.Is(err, auth.ErrBadCredentials) {
			t.Fatalf("expected ErrBadCredentials, got %v", err)
		}
	})

	t.Run("invalid uuid in profile", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		client := mocks.NewMockProfilesServiceClient(ctrl)
		u := newUsecase(t, client)

		client.EXPECT().SigninUser(ctx, "alice").Return(
			&profilesgrpc.ProfileResponse{Id: "not-a-uuid", Username: "alice", Password: string(hashed)}, nil,
		)

		if _, err := u.SigninUser(ctx, "alice", rawPassword); err == nil {
			t.Fatal("expected uuid parse error, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		client := mocks.NewMockProfilesServiceClient(ctrl)
		u := newUsecase(t, client)

		client.EXPECT().SigninUser(ctx, "alice").Return(
			&profilesgrpc.ProfileResponse{Id: userID.String(), Username: "alice", Avatar: "ava", Password: string(hashed)}, nil,
		)

		got, err := u.SigninUser(ctx, "alice", rawPassword)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if got.ID != userID || got.Username != "alice" || got.Avatar != "ava" {
			t.Errorf("unexpected profile: %+v", got)
		}
	})
}

func TestLogout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mocks.NewMockProfilesServiceClient(ctrl)
	u := newUsecase(t, client)

	w := httptest.NewRecorder()
	u.Logout(context.Background(), w)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	c := cookies[0]
	if c.Name != testJWT.CookieName {
		t.Errorf("expected name=%q, got %q", testJWT.CookieName, c.Name)
	}
	if c.MaxAge != -1 {
		t.Errorf("expected MaxAge=-1, got %d", c.MaxAge)
	}
}

package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/handler/http/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	profilesdto "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/dto"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

var testJWT = config.JWTConfig{
	Secret:     "test-secret",
	CookieName: "session",
	CookieTime: time.Hour,
	Secure:     false,
}

func newHandler(t *testing.T) (*AuthHandler, *mocks.MockAuthUsecase, *gomock.Controller) {
	t.Helper()
	ctrl := gomock.NewController(t)
	uc := mocks.NewMockAuthUsecase(ctrl)
	return NewAuthHandler(uc, testJWT), uc, ctrl
}

func TestNewAuthHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	uc := mocks.NewMockAuthUsecase(ctrl)
	h := NewAuthHandler(uc, testJWT)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestSignupUserHandler(t *testing.T) {
	userID := uuid.New()
	validBody := `{"username":"alice","password":"GoodPass1"}`

	t.Run("nil body", func(t *testing.T) {
		h, _, ctrl := newHandler(t)
		defer ctrl.Finish()

		req := httptest.NewRequest(http.MethodPost, "/signup", nil)
		req.Body = nil
		w := httptest.NewRecorder()
		h.SignupUser(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", w.Code)
		}
	})

	t.Run("bad json", func(t *testing.T) {
		h, _, ctrl := newHandler(t)
		defer ctrl.Finish()

		req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader("not-json"))
		w := httptest.NewRecorder()
		h.SignupUser(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("user already exists -> 409", func(t *testing.T) {
		h, uc, ctrl := newHandler(t)
		defer ctrl.Finish()

		uc.EXPECT().SignupUser(gomock.Any(), "alice", "GoodPass1").Return(nil, auth.ErrUserExist)

		req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(validBody))
		w := httptest.NewRecorder()
		h.SignupUser(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("expected 409, got %d", w.Code)
		}
	})

	t.Run("invalid username -> 400", func(t *testing.T) {
		h, uc, ctrl := newHandler(t)
		defer ctrl.Finish()

		uc.EXPECT().SignupUser(gomock.Any(), "alice", "GoodPass1").Return(nil, auth.ErrInvalidUsername)

		req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(validBody))
		w := httptest.NewRecorder()
		h.SignupUser(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid password -> 400", func(t *testing.T) {
		h, uc, ctrl := newHandler(t)
		defer ctrl.Finish()

		uc.EXPECT().SignupUser(gomock.Any(), "alice", "GoodPass1").Return(nil, auth.ErrInvalidPassword)

		req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(validBody))
		w := httptest.NewRecorder()
		h.SignupUser(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("internal error -> 500", func(t *testing.T) {
		h, uc, ctrl := newHandler(t)
		defer ctrl.Finish()

		uc.EXPECT().SignupUser(gomock.Any(), "alice", "GoodPass1").Return(nil, errors.New("boom"))

		req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(validBody))
		w := httptest.NewRecorder()
		h.SignupUser(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("trims whitespace before forwarding", func(t *testing.T) {
		h, uc, ctrl := newHandler(t)
		defer ctrl.Finish()

		uc.EXPECT().SignupUser(gomock.Any(), "alice", "GoodPass1").Return(
			&profilesdto.Profile{ID: userID, Username: "alice"}, nil,
		)

		body := `{"username":"  alice  ","password":"  GoodPass1  "}`
		req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(body))
		w := httptest.NewRecorder()
		h.SignupUser(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d (%s)", w.Code, w.Body.String())
		}
	})

	t.Run("success sets cookie", func(t *testing.T) {
		h, uc, ctrl := newHandler(t)
		defer ctrl.Finish()

		uc.EXPECT().SignupUser(gomock.Any(), "alice", "GoodPass1").Return(
			&profilesdto.Profile{ID: userID, Username: "alice", Avatar: "ava"}, nil,
		)

		req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(validBody))
		w := httptest.NewRecorder()
		h.SignupUser(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
		}
		cookies := w.Result().Cookies()
		if len(cookies) != 1 || cookies[0].Name != testJWT.CookieName {
			t.Fatalf("expected session cookie, got %+v", cookies)
		}
		if cookies[0].Value == "" {
			t.Error("expected non-empty JWT cookie value")
		}
		if !strings.Contains(w.Body.String(), userID.String()) {
			t.Errorf("expected userID in body, got %s", w.Body.String())
		}
	})
}

func TestSigninUserHandler(t *testing.T) {
	userID := uuid.New()
	validBody := `{"username":"alice","password":"GoodPass1"}`

	t.Run("nil body", func(t *testing.T) {
		h, _, ctrl := newHandler(t)
		defer ctrl.Finish()

		req := httptest.NewRequest(http.MethodPost, "/signin", nil)
		req.Body = nil
		w := httptest.NewRecorder()
		h.SigninUser(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", w.Code)
		}
	})

	t.Run("bad json", func(t *testing.T) {
		h, _, ctrl := newHandler(t)
		defer ctrl.Finish()

		req := httptest.NewRequest(http.MethodPost, "/signin", strings.NewReader("not-json"))
		w := httptest.NewRecorder()
		h.SigninUser(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("bad credentials -> 401", func(t *testing.T) {
		h, uc, ctrl := newHandler(t)
		defer ctrl.Finish()

		uc.EXPECT().SigninUser(gomock.Any(), "alice", "GoodPass1").Return(nil, auth.ErrBadCredentials)

		req := httptest.NewRequest(http.MethodPost, "/signin", strings.NewReader(validBody))
		w := httptest.NewRecorder()
		h.SigninUser(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("user not exists -> 401", func(t *testing.T) {
		h, uc, ctrl := newHandler(t)
		defer ctrl.Finish()

		uc.EXPECT().SigninUser(gomock.Any(), "alice", "GoodPass1").Return(nil, auth.ErrUserNotExist)

		req := httptest.NewRequest(http.MethodPost, "/signin", strings.NewReader(validBody))
		w := httptest.NewRecorder()
		h.SigninUser(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("internal error -> 500", func(t *testing.T) {
		h, uc, ctrl := newHandler(t)
		defer ctrl.Finish()

		uc.EXPECT().SigninUser(gomock.Any(), "alice", "GoodPass1").Return(nil, errors.New("boom"))

		req := httptest.NewRequest(http.MethodPost, "/signin", strings.NewReader(validBody))
		w := httptest.NewRecorder()
		h.SigninUser(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("success sets cookie", func(t *testing.T) {
		h, uc, ctrl := newHandler(t)
		defer ctrl.Finish()

		uc.EXPECT().SigninUser(gomock.Any(), "alice", "GoodPass1").Return(
			&profilesdto.Profile{ID: userID, Username: "alice"}, nil,
		)

		req := httptest.NewRequest(http.MethodPost, "/signin", strings.NewReader(validBody))
		w := httptest.NewRecorder()
		h.SigninUser(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
		}
		cookies := w.Result().Cookies()
		if len(cookies) != 1 || cookies[0].Name != testJWT.CookieName {
			t.Fatalf("expected session cookie, got %+v", cookies)
		}
	})
}

func TestLogoutUserHandler(t *testing.T) {
	h, uc, ctrl := newHandler(t)
	defer ctrl.Finish()

	uc.EXPECT().Logout(gomock.Any(), gomock.Any()).Do(func(_ context.Context, w http.ResponseWriter) {
		http.SetCookie(w, &http.Cookie{Name: testJWT.CookieName, Value: "", MaxAge: -1, Path: "/"})
	})

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()
	h.LogoutUser(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge != -1 {
		t.Errorf("expected cleared cookie, got %+v", cookies)
	}
}

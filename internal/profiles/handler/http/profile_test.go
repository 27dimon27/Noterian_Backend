package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/logger"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/handler/http/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

var log = logger.Init()

var testJWT = config.JWTConfig{
	Secret:     "secret",
	CookieName: "auth",
	Secure:     false,
}

func withUserID(r *http.Request, userID uuid.UUID) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, userID))
}

// pngBytes returns 512+ bytes that begin with a PNG magic header so
// http.DetectContentType returns "image/png".
func pngBytes(t *testing.T) []byte {
	t.Helper()
	// PNG signature
	header := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	payload := bytes.Repeat([]byte{0}, 600)
	return append(header, payload...)
}

func buildMultipartRequest(t *testing.T, fieldName, filename string, content []byte) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	fw, err := mw.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/profile/avatar", body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func TestGetProfileHandler(t *testing.T) {
	userID := uuid.New()

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		w := httptest.NewRecorder()
		h.GetProfile(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().GetProfile(gomock.Any(), userID).Return(nil, profiles.ErrUserNotExist)

		req := withUserID(httptest.NewRequest(http.MethodGet, "/profile", nil), userID)
		w := httptest.NewRecorder()
		h.GetProfile(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().GetProfile(gomock.Any(), userID).Return(nil, errors.New("boom"))

		req := withUserID(httptest.NewRequest(http.MethodGet, "/profile", nil), userID)
		w := httptest.NewRecorder()
		h.GetProfile(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().GetProfile(gomock.Any(), userID).Return(&models.Profile{ID: userID, Username: "alice"}, nil)

		req := withUserID(httptest.NewRequest(http.MethodGet, "/profile", nil), userID)
		w := httptest.NewRecorder()
		h.GetProfile(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d (%s)", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "alice") {
			t.Errorf("expected username in body, got %s", w.Body.String())
		}
	})
}

func TestUpdateProfileHandler(t *testing.T) {
	userID := uuid.New()
	validBody := `{"username":"alice"}`

	t.Run("nil body", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		req := httptest.NewRequest(http.MethodPut, "/profile", nil)
		req.Body = nil
		req = withUserID(req, userID)
		w := httptest.NewRecorder()
		h.UpdateProfile(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		req := httptest.NewRequest(http.MethodPut, "/profile", strings.NewReader(validBody))
		w := httptest.NewRecorder()
		h.UpdateProfile(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("bad json", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		req := withUserID(httptest.NewRequest(http.MethodPut, "/profile", strings.NewReader("not-json")), userID)
		w := httptest.NewRecorder()
		h.UpdateProfile(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("usecase invalid data -> 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().UpdateProfile(gomock.Any(), userID, gomock.Any()).Return(nil, profiles.ErrInvalidProfileData)

		req := withUserID(httptest.NewRequest(http.MethodPut, "/profile", strings.NewReader(validBody)), userID)
		w := httptest.NewRecorder()
		h.UpdateProfile(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("usecase internal error -> 500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().UpdateProfile(gomock.Any(), userID, gomock.Any()).Return(nil, errors.New("boom"))

		req := withUserID(httptest.NewRequest(http.MethodPut, "/profile", strings.NewReader(validBody)), userID)
		w := httptest.NewRecorder()
		h.UpdateProfile(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().UpdateProfile(gomock.Any(), userID, gomock.Any()).Return(&models.Profile{ID: userID, Username: "alice"}, nil)

		req := withUserID(httptest.NewRequest(http.MethodPut, "/profile", strings.NewReader(validBody)), userID)
		w := httptest.NewRecorder()
		h.UpdateProfile(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}

func TestDeleteProfileHandler(t *testing.T) {
	userID := uuid.New()

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		req := httptest.NewRequest(http.MethodDelete, "/profile", nil)
		w := httptest.NewRecorder()
		h.DeleteProfile(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("not found -> 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().DeleteProfile(gomock.Any(), userID).Return(profiles.ErrUserNotExist)

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/profile", nil), userID)
		w := httptest.NewRecorder()
		h.DeleteProfile(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().DeleteProfile(gomock.Any(), userID).Return(errors.New("boom"))

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/profile", nil), userID)
		w := httptest.NewRecorder()
		h.DeleteProfile(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("success clears cookie", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().DeleteProfile(gomock.Any(), userID).Return(nil)

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/profile", nil), userID)
		w := httptest.NewRecorder()
		h.DeleteProfile(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d", w.Code)
		}
		if !strings.Contains(w.Header().Get("Set-Cookie"), testJWT.CookieName+"=") {
			t.Errorf("expected cleared cookie, got %q", w.Header().Get("Set-Cookie"))
		}
	})
}

func TestGetAvatarHandler(t *testing.T) {
	userID := uuid.New()

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		req := httptest.NewRequest(http.MethodGet, "/profile/avatar", nil)
		w := httptest.NewRecorder()
		h.GetAvatar(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().GetAvatar(gomock.Any(), userID).Return(nil, profiles.ErrAvatarNotFound)

		req := withUserID(httptest.NewRequest(http.MethodGet, "/profile/avatar", nil), userID)
		w := httptest.NewRecorder()
		h.GetAvatar(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().GetAvatar(gomock.Any(), userID).Return(nil, errors.New("boom"))

		req := withUserID(httptest.NewRequest(http.MethodGet, "/profile/avatar", nil), userID)
		w := httptest.NewRecorder()
		h.GetAvatar(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().GetAvatar(gomock.Any(), userID).Return(&models.Avatar{ID: uuid.New(), ProfileID: userID, AvatarURL: "https://x"}, nil)

		req := withUserID(httptest.NewRequest(http.MethodGet, "/profile/avatar", nil), userID)
		w := httptest.NewRecorder()
		h.GetAvatar(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}

func TestUploadAvatarHandler(t *testing.T) {
	userID := uuid.New()

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		req := buildMultipartRequest(t, "file", "x.png", pngBytes(t))
		w := httptest.NewRecorder()
		h.UploadAvatar(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("file too large", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		// Build content larger than MAX_FILE_SIZE
		big := bytes.Repeat([]byte("A"), profiles.MAX_FILE_SIZE+1024)
		req := withUserID(buildMultipartRequest(t, "file", "big.png", big), userID)
		w := httptest.NewRecorder()
		h.UploadAvatar(w, req)

		if w.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("expected 413, got %d", w.Code)
		}
	})

	t.Run("invalid mime", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		req := withUserID(buildMultipartRequest(t, "file", "f.txt", bytes.Repeat([]byte("plain text content "), 50)), userID)
		w := httptest.NewRecorder()
		h.UploadAvatar(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("missing file form field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		req := withUserID(buildMultipartRequest(t, "other", "f.png", pngBytes(t)), userID)
		w := httptest.NewRecorder()
		h.UploadAvatar(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().
			UploadAvatar(gomock.Any(), userID, gomock.Any(), gomock.Any(), "image/png", gomock.Any()).
			Return(&models.Avatar{ID: uuid.New(), ProfileID: userID, AvatarURL: "https://x"}, nil)

		req := withUserID(buildMultipartRequest(t, "file", "f.png", pngBytes(t)), userID)
		w := httptest.NewRecorder()
		h.UploadAvatar(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d (%s)", w.Code, w.Body.String())
		}
	})
}

func TestDeleteAvatarHandler(t *testing.T) {
	userID := uuid.New()

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		req := httptest.NewRequest(http.MethodDelete, "/profile/avatar", nil)
		w := httptest.NewRecorder()
		h.DeleteAvatar(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().DeleteAvatar(gomock.Any(), userID).Return(profiles.ErrAvatarNotFound)

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/profile/avatar", nil), userID)
		w := httptest.NewRecorder()
		h.DeleteAvatar(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("internal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().DeleteAvatar(gomock.Any(), userID).Return(errors.New("boom"))

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/profile/avatar", nil), userID)
		w := httptest.NewRecorder()
		h.DeleteAvatar(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().DeleteAvatar(gomock.Any(), userID).Return(nil)

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/profile/avatar", nil), userID)
		w := httptest.NewRecorder()
		h.DeleteAvatar(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d", w.Code)
		}
	})
}

func TestChangePasswordHandler(t *testing.T) {
	userID := uuid.New()
	validBody := `{"old_password":"Old1","new_password":"New1"}`

	t.Run("nil body", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		req := httptest.NewRequest(http.MethodPut, "/profile/password", nil)
		req.Body = nil
		req = withUserID(req, userID)
		w := httptest.NewRecorder()
		h.ChangePassword(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		req := httptest.NewRequest(http.MethodPut, "/profile/password", strings.NewReader(validBody))
		w := httptest.NewRecorder()
		h.ChangePassword(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("bad json", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		req := withUserID(httptest.NewRequest(http.MethodPut, "/profile/password", strings.NewReader("garbage")), userID)
		w := httptest.NewRecorder()
		h.ChangePassword(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("user not found -> 404", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().ChangePassword(gomock.Any(), userID, "Old1", "New1").Return(nil, profiles.ErrUserNotExist)

		req := withUserID(httptest.NewRequest(http.MethodPut, "/profile/password", strings.NewReader(validBody)), userID)
		w := httptest.NewRecorder()
		h.ChangePassword(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("wrong password -> 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().ChangePassword(gomock.Any(), userID, "Old1", "New1").Return(nil, profiles.ErrWrongPassword)

		req := withUserID(httptest.NewRequest(http.MethodPut, "/profile/password", strings.NewReader(validBody)), userID)
		w := httptest.NewRecorder()
		h.ChangePassword(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("internal error -> 500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().ChangePassword(gomock.Any(), userID, "Old1", "New1").Return(nil, errors.New("boom"))

		req := withUserID(httptest.NewRequest(http.MethodPut, "/profile/password", strings.NewReader(validBody)), userID)
		w := httptest.NewRecorder()
		h.ChangePassword(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockProfileUsecase(ctrl)
		h := NewProfileHandler(uc, testJWT, log)

		uc.EXPECT().ChangePassword(gomock.Any(), userID, "Old1", "New1").
			Return(&models.Profile{ID: userID, Username: "alice"}, nil)

		req := withUserID(httptest.NewRequest(http.MethodPut, "/profile/password", strings.NewReader(validBody)), userID)
		w := httptest.NewRecorder()
		h.ChangePassword(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}

// guard against unused import in some builds
var _ = io.EOF

package usecase

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/logger"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/usecase/mocks"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

var log = logger.Init()

func newUsecase(t *testing.T, repo *mocks.MockProfileRepository) *profileUsecase {
	t.Helper()
	u, err := NewProfileUsecase(repo, log)
	if err != nil {
		t.Fatalf("NewProfileUsecase: %v", err)
	}
	return u
}

func TestNewProfileUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockProfileRepository(ctrl)
	u, err := NewProfileUsecase(repo, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u == nil {
		t.Fatal("expected non-nil usecase")
	}
}

func TestGetProfile(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success with avatar", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		profile := &models.Profile{ID: userID, Username: "alice"}
		avatar := &models.Avatar{ID: uuid.New(), ProfileID: userID, AvatarURL: "https://avatars/foo"}

		repo.EXPECT().GetProfile(ctx, userID).Return(profile, nil)
		repo.EXPECT().GetAvatar(ctx, userID).Return(avatar, nil)

		got, err := u.GetProfile(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Avatar != avatar.AvatarURL {
			t.Errorf("expected Avatar=%q, got %q", avatar.AvatarURL, got.Avatar)
		}
	})

	t.Run("success without avatar", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		profile := &models.Profile{ID: userID, Username: "alice"}

		repo.EXPECT().GetProfile(ctx, userID).Return(profile, nil)
		repo.EXPECT().GetAvatar(ctx, userID).Return(nil, profiles.ErrAvatarNotFound)

		got, err := u.GetProfile(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Avatar != "" {
			t.Errorf("expected empty Avatar, got %q", got.Avatar)
		}
	})

	t.Run("profile error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		repo.EXPECT().GetProfile(ctx, userID).Return(nil, profiles.ErrUserNotExist)

		_, err := u.GetProfile(ctx, userID)
		if !errors.Is(err, profiles.ErrUserNotExist) {
			t.Fatalf("expected ErrUserNotExist, got %v", err)
		}
	})

	t.Run("avatar error other than not-found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		profile := &models.Profile{ID: userID, Username: "alice"}
		boom := errors.New("boom")

		repo.EXPECT().GetProfile(ctx, userID).Return(profile, nil)
		repo.EXPECT().GetAvatar(ctx, userID).Return(nil, boom)

		_, err := u.GetProfile(ctx, userID)
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestGetProfileByUsername(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		profile := &models.Profile{ID: userID, Username: "bob"}
		avatar := &models.Avatar{ID: uuid.New(), ProfileID: userID, AvatarURL: "https://x"}

		repo.EXPECT().GetProfileByUsername(ctx, "bob").Return(profile, nil)
		repo.EXPECT().GetAvatar(ctx, userID).Return(avatar, nil)

		got, err := u.GetProfileByUsername(ctx, "bob")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Avatar != avatar.AvatarURL {
			t.Errorf("expected Avatar set, got %q", got.Avatar)
		}
	})

	t.Run("profile not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		repo.EXPECT().GetProfileByUsername(ctx, "missing").Return(nil, profiles.ErrUserNotExist)

		_, err := u.GetProfileByUsername(ctx, "missing")
		if !errors.Is(err, profiles.ErrUserNotExist) {
			t.Fatalf("expected ErrUserNotExist, got %v", err)
		}
	})

	t.Run("avatar not found ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		profile := &models.Profile{ID: userID, Username: "bob"}
		repo.EXPECT().GetProfileByUsername(ctx, "bob").Return(profile, nil)
		repo.EXPECT().GetAvatar(ctx, userID).Return(nil, profiles.ErrAvatarNotFound)

		got, err := u.GetProfileByUsername(ctx, "bob")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Avatar != "" {
			t.Errorf("expected empty avatar")
		}
	})

	t.Run("avatar unexpected error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		profile := &models.Profile{ID: userID, Username: "bob"}
		boom := errors.New("boom")
		repo.EXPECT().GetProfileByUsername(ctx, "bob").Return(profile, nil)
		repo.EXPECT().GetAvatar(ctx, userID).Return(nil, boom)

		_, err := u.GetProfileByUsername(ctx, "bob")
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestUpdateProfile(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("invalid username", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		_, err := u.UpdateProfile(ctx, userID, models.Profile{Username: "_bad"})
		if !errors.Is(err, profiles.ErrInvalidProfileData) {
			t.Fatalf("expected ErrInvalidProfileData, got %v", err)
		}
	})

	t.Run("username taken by other user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		other := &models.Profile{ID: uuid.New(), Username: "alice"}
		repo.EXPECT().GetProfileByUsername(ctx, "alice").Return(other, nil)

		_, err := u.UpdateProfile(ctx, userID, models.Profile{Username: "alice"})
		if !errors.Is(err, profiles.ErrUsernameExists) {
			t.Fatalf("expected ErrUsernameExists, got %v", err)
		}
	})

	t.Run("get by username fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		boom := errors.New("boom")
		repo.EXPECT().GetProfileByUsername(ctx, "alice").Return(nil, boom)

		_, err := u.UpdateProfile(ctx, userID, models.Profile{Username: "alice"})
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})

	t.Run("success same username", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		// returning current user with same ID -> not "taken"
		self := &models.Profile{ID: userID, Username: "alice"}
		updated := &models.Profile{ID: userID, Username: "alice"}
		avatar := &models.Avatar{ID: uuid.New(), AvatarURL: "https://av"}

		repo.EXPECT().GetProfileByUsername(ctx, "alice").Return(self, nil)
		repo.EXPECT().UpdateProfile(ctx, userID, models.Profile{Username: "alice"}).Return(updated, nil)
		repo.EXPECT().GetAvatar(ctx, userID).Return(avatar, nil)

		got, err := u.UpdateProfile(ctx, userID, models.Profile{Username: "alice"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Avatar != avatar.AvatarURL {
			t.Errorf("expected avatar URL set")
		}
	})

	t.Run("update fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		boom := errors.New("boom")
		repo.EXPECT().GetProfileByUsername(ctx, "alice").Return(nil, profiles.ErrUserNotExist)
		repo.EXPECT().UpdateProfile(ctx, userID, models.Profile{Username: "alice"}).Return(nil, boom)

		_, err := u.UpdateProfile(ctx, userID, models.Profile{Username: "alice"})
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})

	t.Run("avatar not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		updated := &models.Profile{ID: userID, Username: "alice"}
		repo.EXPECT().GetProfileByUsername(ctx, "alice").Return(nil, profiles.ErrUserNotExist)
		repo.EXPECT().UpdateProfile(ctx, userID, models.Profile{Username: "alice"}).Return(updated, nil)
		repo.EXPECT().GetAvatar(ctx, userID).Return(nil, profiles.ErrAvatarNotFound)

		got, err := u.UpdateProfile(ctx, userID, models.Profile{Username: "alice"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Avatar != "" {
			t.Errorf("expected empty avatar")
		}
	})

	t.Run("avatar fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		updated := &models.Profile{ID: userID, Username: "alice"}
		boom := errors.New("boom")
		repo.EXPECT().GetProfileByUsername(ctx, "alice").Return(nil, profiles.ErrUserNotExist)
		repo.EXPECT().UpdateProfile(ctx, userID, models.Profile{Username: "alice"}).Return(updated, nil)
		repo.EXPECT().GetAvatar(ctx, userID).Return(nil, boom)

		_, err := u.UpdateProfile(ctx, userID, models.Profile{Username: "alice"})
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestDeleteProfile(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		repo.EXPECT().DeleteAvatar(ctx, userID).Return(nil)
		repo.EXPECT().DeleteProfile(ctx, userID).Return(nil)

		if err := u.DeleteProfile(ctx, userID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("avatar not found, profile deleted", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		repo.EXPECT().DeleteAvatar(ctx, userID).Return(profiles.ErrAvatarNotFound)
		repo.EXPECT().DeleteProfile(ctx, userID).Return(nil)

		if err := u.DeleteProfile(ctx, userID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("delete avatar fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		boom := errors.New("boom")
		repo.EXPECT().DeleteAvatar(ctx, userID).Return(boom)

		if err := u.DeleteProfile(ctx, userID); !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})

	t.Run("delete profile fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		boom := errors.New("boom")
		repo.EXPECT().DeleteAvatar(ctx, userID).Return(nil)
		repo.EXPECT().DeleteProfile(ctx, userID).Return(boom)

		if err := u.DeleteProfile(ctx, userID); !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestDeleteProfileWithCookie(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success clears cookie", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		repo.EXPECT().DeleteAvatar(ctx, userID).Return(nil)
		repo.EXPECT().DeleteProfile(ctx, userID).Return(nil)

		w := httptest.NewRecorder()
		cfg := config.JWTConfig{CookieName: "auth", Secure: true}
		if err := u.DeleteProfileWithCookie(ctx, userID, w, cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		setCookie := w.Header().Get("Set-Cookie")
		if !strings.Contains(setCookie, "auth=") {
			t.Errorf("expected cookie cleared, got %q", setCookie)
		}
	})

	t.Run("delete fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		boom := errors.New("boom")
		repo.EXPECT().DeleteAvatar(ctx, userID).Return(boom)

		w := httptest.NewRecorder()
		if err := u.DeleteProfileWithCookie(ctx, userID, w, config.JWTConfig{CookieName: "auth"}); !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestGetAvatar(t *testing.T) {
	ctx := context.Background()
	profileID := uuid.New()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		avatar := &models.Avatar{ID: uuid.New(), ProfileID: profileID}
		repo.EXPECT().GetAvatar(ctx, profileID).Return(avatar, nil)

		got, err := u.GetAvatar(ctx, profileID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != avatar {
			t.Errorf("expected the same avatar")
		}
	})

	t.Run("not found error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		repo.EXPECT().GetAvatar(ctx, profileID).Return(nil, profiles.ErrAvatarNotFound)

		_, err := u.GetAvatar(ctx, profileID)
		if !errors.Is(err, profiles.ErrAvatarNotFound) {
			t.Fatalf("expected ErrAvatarNotFound, got %v", err)
		}
	})

	t.Run("nil avatar maps to not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		repo.EXPECT().GetAvatar(ctx, profileID).Return(nil, nil)

		_, err := u.GetAvatar(ctx, profileID)
		if !errors.Is(err, profiles.ErrAvatarNotFound) {
			t.Fatalf("expected ErrAvatarNotFound, got %v", err)
		}
	})
}

func TestUploadAvatar(t *testing.T) {
	ctx := context.Background()
	profileID := uuid.New()
	reader := strings.NewReader("file-content")

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		avatar := &models.Avatar{ID: uuid.New(), ProfileID: profileID, AvatarURL: "https://x"}
		repo.EXPECT().
			UploadAvatar(ctx, profileID, "file.png", int64(12), "image/png", gomock.Any()).
			Return(avatar, nil)

		got, err := u.UploadAvatar(ctx, profileID, "file.png", 12, "image/png", reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != avatar {
			t.Errorf("expected returned avatar")
		}
	})

	t.Run("error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		boom := errors.New("boom")
		repo.EXPECT().
			UploadAvatar(ctx, profileID, "f", int64(1), "image/png", gomock.Any()).
			Return(nil, boom)

		_, err := u.UploadAvatar(ctx, profileID, "f", 1, "image/png", reader)
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestDeleteAvatar(t *testing.T) {
	ctx := context.Background()
	profileID := uuid.New()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		repo.EXPECT().DeleteAvatar(ctx, profileID).Return(nil)
		if err := u.DeleteAvatar(ctx, profileID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		boom := errors.New("boom")
		repo.EXPECT().DeleteAvatar(ctx, profileID).Return(boom)
		if err := u.DeleteAvatar(ctx, profileID); !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestChangePassword(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	oldPassword := "OldPass1"
	newPassword := "NewPass1"

	makeHash := func(pw string) []byte {
		h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
		if err != nil {
			t.Fatalf("hash: %v", err)
		}
		return h
	}

	t.Run("invalid new password", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		_, err := u.ChangePassword(ctx, userID, oldPassword, "weak")
		if !errors.Is(err, profiles.ErrInvalidPasswordData) {
			t.Fatalf("expected ErrInvalidPasswordData, got %v", err)
		}
	})

	t.Run("get password fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		boom := errors.New("boom")
		repo.EXPECT().GetPassword(ctx, userID).Return(nil, boom)

		_, err := u.ChangePassword(ctx, userID, oldPassword, newPassword)
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})

	t.Run("wrong old password", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		repo.EXPECT().GetPassword(ctx, userID).Return(makeHash("OtherPass1"), nil)

		_, err := u.ChangePassword(ctx, userID, oldPassword, newPassword)
		if !errors.Is(err, profiles.ErrWrongPassword) {
			t.Fatalf("expected ErrWrongPassword, got %v", err)
		}
	})

	t.Run("change fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		boom := errors.New("boom")
		repo.EXPECT().GetPassword(ctx, userID).Return(makeHash(oldPassword), nil)
		repo.EXPECT().ChangePassword(ctx, userID, newPassword).Return(nil, boom)

		_, err := u.ChangePassword(ctx, userID, oldPassword, newPassword)
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})

	t.Run("success with avatar", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		updated := &models.Profile{ID: userID, Username: "alice"}
		avatar := &models.Avatar{ID: uuid.New(), AvatarURL: "https://x"}

		repo.EXPECT().GetPassword(ctx, userID).Return(makeHash(oldPassword), nil)
		repo.EXPECT().ChangePassword(ctx, userID, newPassword).Return(updated, nil)
		repo.EXPECT().GetAvatar(ctx, userID).Return(avatar, nil)

		got, err := u.ChangePassword(ctx, userID, oldPassword, newPassword)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Avatar != avatar.AvatarURL {
			t.Errorf("expected avatar url set")
		}
	})

	t.Run("success without avatar", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		updated := &models.Profile{ID: userID, Username: "alice"}

		repo.EXPECT().GetPassword(ctx, userID).Return(makeHash(oldPassword), nil)
		repo.EXPECT().ChangePassword(ctx, userID, newPassword).Return(updated, nil)
		repo.EXPECT().GetAvatar(ctx, userID).Return(nil, profiles.ErrAvatarNotFound)

		got, err := u.ChangePassword(ctx, userID, oldPassword, newPassword)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Avatar != "" {
			t.Errorf("expected empty avatar")
		}
	})

	t.Run("avatar lookup fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		updated := &models.Profile{ID: userID, Username: "alice"}
		boom := errors.New("boom")

		repo.EXPECT().GetPassword(ctx, userID).Return(makeHash(oldPassword), nil)
		repo.EXPECT().ChangePassword(ctx, userID, newPassword).Return(updated, nil)
		repo.EXPECT().GetAvatar(ctx, userID).Return(nil, boom)

		_, err := u.ChangePassword(ctx, userID, oldPassword, newPassword)
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestGetPassword(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		want := []byte("hashedpw")
		repo.EXPECT().GetPassword(ctx, userID).Return(want, nil)

		got, err := u.GetPassword(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got) != string(want) {
			t.Errorf("expected %q, got %q", want, got)
		}
	})

	t.Run("error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		boom := errors.New("boom")
		repo.EXPECT().GetPassword(ctx, userID).Return(nil, boom)

		_, err := u.GetPassword(ctx, userID)
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestSignupUser(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid username", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		_, err := u.SignupUser(ctx, "_bad", "GoodPass1")
		if !errors.Is(err, profiles.ErrInvalidProfileData) {
			t.Fatalf("expected ErrInvalidProfileData, got %v", err)
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		_, err := u.SignupUser(ctx, "alice", "weak")
		if !errors.Is(err, profiles.ErrInvalidProfileData) {
			t.Fatalf("expected ErrInvalidProfileData, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		want := &models.Profile{ID: uuid.New(), Username: "alice"}
		repo.EXPECT().SignupUser(ctx, "alice", "GoodPass1").Return(want, nil)

		got, err := u.SignupUser(ctx, "alice", "GoodPass1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("expected returned profile")
		}
	})
}

func TestSigninUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("user not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		repo.EXPECT().SigninUser(ctx, "alice").Return(nil, profiles.ErrUserNotExist)

		_, err := u.SigninUser(ctx, "alice")
		if !errors.Is(err, profiles.ErrUserNotExist) {
			t.Fatalf("expected ErrUserNotExist, got %v", err)
		}
	})

	t.Run("success with avatar", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		profile := &models.Profile{ID: userID, Username: "alice"}
		avatar := &models.Avatar{ID: uuid.New(), AvatarURL: "https://x"}

		repo.EXPECT().SigninUser(ctx, "alice").Return(profile, nil)
		repo.EXPECT().GetAvatar(ctx, userID).Return(avatar, nil)

		got, err := u.SigninUser(ctx, "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Avatar != avatar.AvatarURL {
			t.Errorf("expected avatar URL set")
		}
	})

	t.Run("success no avatar", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		profile := &models.Profile{ID: userID, Username: "alice"}
		repo.EXPECT().SigninUser(ctx, "alice").Return(profile, nil)
		repo.EXPECT().GetAvatar(ctx, userID).Return(nil, profiles.ErrAvatarNotFound)

		got, err := u.SigninUser(ctx, "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Avatar != "" {
			t.Errorf("expected empty avatar")
		}
	})

	t.Run("avatar unexpected error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockProfileRepository(ctrl)
		u := newUsecase(t, repo)

		profile := &models.Profile{ID: userID, Username: "alice"}
		boom := errors.New("boom")
		repo.EXPECT().SigninUser(ctx, "alice").Return(profile, nil)
		repo.EXPECT().GetAvatar(ctx, userID).Return(nil, boom)

		_, err := u.SigninUser(ctx, "alice")
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name     string
		username string
		valid    bool
	}{
		{"valid simple", "alice", true},
		{"valid digits", "alice123", true},
		{"valid mixed", "al_ice.test", true},
		{"leading underscore", "_alice", false},
		{"leading dot", ".alice", false},
		{"trailing underscore", "alice_", false},
		{"trailing dot", "alice.", false},
		{"double underscore", "al__ice", false},
		{"double dot", "al..ice", false},
		{"underscore dot", "al_.ice", false},
		{"dot underscore", "al._ice", false},
		{"invalid chars", "alice!", false},
		{"empty", "", false},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockProfileRepository(ctrl)
	u := newUsecase(t, repo)

	for _, tc := range tests {
		t.Run("username/"+tc.name, func(t *testing.T) {
			err := u.validate.Var(tc.username, "required,username")
			if (err == nil) != tc.valid {
				t.Errorf("for %q expected valid=%v, got err=%v", tc.username, tc.valid, err)
			}
		})
	}

	passwordTests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"valid", "GoodPass1", true},
		{"short", "G1a", false},
		{"no upper", "goodpass1", false},
		{"no digit", "GoodPass", false},
	}

	for _, tc := range passwordTests {
		t.Run("password/"+tc.name, func(t *testing.T) {
			err := u.validate.Var(tc.password, "required,password")
			if (err == nil) != tc.valid {
				t.Errorf("for %q expected valid=%v, got err=%v", tc.password, tc.valid, err)
			}
		})
	}
}

// guard against unused import in some builds
var (
	_ = io.EOF
	_ = time.Now
)

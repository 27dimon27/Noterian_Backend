package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/repository/mocks"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

const testBucket = "avatars"

type repoEnv struct {
	repo    *profileRepository
	db      *sql.DB
	sqlMock sqlmock.Sqlmock
	minio   *mocks.MockMinIOService
	ctrl    *gomock.Controller
}

func newRepoEnv(t *testing.T) *repoEnv {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	ctrl := gomock.NewController(t)
	minio := mocks.NewMockMinIOService(ctrl)
	repo := NewProfileRepository(db, minio, testBucket)

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close db: %v", err)
		}
		ctrl.Finish()
	})

	return &repoEnv{
		repo:    repo,
		db:      db,
		sqlMock: mock,
		minio:   minio,
		ctrl:    ctrl,
	}
}

func TestGetProfile(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		env := newRepoEnv(t)

		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "username", "created_at", "updated_at"}).
			AddRow(userID, "alice", now, now)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(GET_PROFILE_BY_USER_ID)).
			WithArgs(userID).
			WillReturnRows(rows)

		got, err := env.repo.GetProfile(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != userID || got.Username != "alice" {
			t.Errorf("unexpected profile: %+v", got)
		}
		if err := env.sqlMock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		env := newRepoEnv(t)

		env.sqlMock.ExpectQuery(regexp.QuoteMeta(GET_PROFILE_BY_USER_ID)).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		_, err := env.repo.GetProfile(ctx, userID)
		if !errors.Is(err, profiles.ErrUserNotExist) {
			t.Fatalf("expected ErrUserNotExist, got %v", err)
		}
	})

	t.Run("db error", func(t *testing.T) {
		env := newRepoEnv(t)

		boom := errors.New("boom")
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(GET_PROFILE_BY_USER_ID)).
			WithArgs(userID).
			WillReturnError(boom)

		_, err := env.repo.GetProfile(ctx, userID)
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestGetProfileByUsername(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		env := newRepoEnv(t)

		id := uuid.New()
		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "username", "created_at", "updated_at"}).
			AddRow(id, "alice", now, now)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(GET_PROFILE_BY_USERNAME)).
			WithArgs("alice").
			WillReturnRows(rows)

		got, err := env.repo.GetProfileByUsername(ctx, "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != id {
			t.Errorf("unexpected id")
		}
	})

	t.Run("not found", func(t *testing.T) {
		env := newRepoEnv(t)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(GET_PROFILE_BY_USERNAME)).
			WithArgs("missing").
			WillReturnError(sql.ErrNoRows)

		_, err := env.repo.GetProfileByUsername(ctx, "missing")
		if !errors.Is(err, profiles.ErrUserNotExist) {
			t.Fatalf("expected ErrUserNotExist, got %v", err)
		}
	})

	t.Run("db error", func(t *testing.T) {
		env := newRepoEnv(t)
		boom := errors.New("boom")
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(GET_PROFILE_BY_USERNAME)).
			WithArgs("alice").
			WillReturnError(boom)

		_, err := env.repo.GetProfileByUsername(ctx, "alice")
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestUpdateProfile(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		env := newRepoEnv(t)

		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "username", "created_at", "updated_at"}).
			AddRow(userID, "newname", now, now)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(UPDATE_PROFILE_BY_USER_ID)).
			WithArgs(userID, "newname").
			WillReturnRows(rows)

		got, err := env.repo.UpdateProfile(ctx, userID, models.Profile{Username: "newname"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Username != "newname" {
			t.Errorf("unexpected username: %q", got.Username)
		}
	})

	t.Run("not found", func(t *testing.T) {
		env := newRepoEnv(t)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(UPDATE_PROFILE_BY_USER_ID)).
			WithArgs(userID, "x").
			WillReturnError(sql.ErrNoRows)

		_, err := env.repo.UpdateProfile(ctx, userID, models.Profile{Username: "x"})
		if !errors.Is(err, profiles.ErrUserNotExist) {
			t.Fatalf("expected ErrUserNotExist, got %v", err)
		}
	})

	t.Run("db error", func(t *testing.T) {
		env := newRepoEnv(t)
		boom := errors.New("boom")
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(UPDATE_PROFILE_BY_USER_ID)).
			WithArgs(userID, "x").
			WillReturnError(boom)

		_, err := env.repo.UpdateProfile(ctx, userID, models.Profile{Username: "x"})
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestDeleteProfile(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		env := newRepoEnv(t)
		rows := sqlmock.NewRows([]string{"id"}).AddRow(userID)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(DELETE_PROFILE_BY_USER_ID)).
			WithArgs(userID).
			WillReturnRows(rows)

		if err := env.repo.DeleteProfile(ctx, userID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		env := newRepoEnv(t)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(DELETE_PROFILE_BY_USER_ID)).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		err := env.repo.DeleteProfile(ctx, userID)
		if !errors.Is(err, profiles.ErrUserNotExist) {
			t.Fatalf("expected ErrUserNotExist, got %v", err)
		}
	})

	t.Run("db error", func(t *testing.T) {
		env := newRepoEnv(t)
		boom := errors.New("boom")
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(DELETE_PROFILE_BY_USER_ID)).
			WithArgs(userID).
			WillReturnError(boom)

		err := env.repo.DeleteProfile(ctx, userID)
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestGetAvatar(t *testing.T) {
	ctx := context.Background()
	profileID := uuid.New()
	avatarID := uuid.New()

	t.Run("success not expired", func(t *testing.T) {
		env := newRepoEnv(t)

		now := time.Now()
		expires := now.Add(time.Hour)
		rows := sqlmock.NewRows([]string{
			"id", "profile_id", "minio_key", "avatar_url", "url_expires_at", "created_at", "updated_at",
		}).AddRow(avatarID, profileID, "key1", "https://x", expires, now, now)

		env.sqlMock.ExpectQuery(regexp.QuoteMeta(GET_AVATAR_BY_PROFILE_ID)).
			WithArgs(profileID).
			WillReturnRows(rows)

		got, err := env.repo.GetAvatar(ctx, profileID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.AvatarURL != "https://x" {
			t.Errorf("unexpected URL: %q", got.AvatarURL)
		}
	})

	t.Run("success regenerates expired url", func(t *testing.T) {
		env := newRepoEnv(t)

		now := time.Now()
		expires := now.Add(-time.Hour)
		rows := sqlmock.NewRows([]string{
			"id", "profile_id", "minio_key", "avatar_url", "url_expires_at", "created_at", "updated_at",
		}).AddRow(avatarID, profileID, "key1", "old-url", expires, now, now)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(GET_AVATAR_BY_PROFILE_ID)).
			WithArgs(profileID).
			WillReturnRows(rows)

		env.minio.EXPECT().
			GeneratePresignedURL(ctx, testBucket, "key1", profiles.PRESIGNED_URL_EXPIRY).
			Return("new-url", nil)

		updRows := sqlmock.NewRows([]string{"avatar_url", "url_expires_at", "updated_at"}).
			AddRow("new-url", now.Add(profiles.PRESIGNED_URL_EXPIRY), now)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(UPDATE_AVATAR_URL)).
			WithArgs(avatarID, "new-url", sqlmock.AnyArg()).
			WillReturnRows(updRows)

		got, err := env.repo.GetAvatar(ctx, profileID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.AvatarURL != "new-url" {
			t.Errorf("expected new-url, got %q", got.AvatarURL)
		}
	})

	t.Run("regenerate presign fails", func(t *testing.T) {
		env := newRepoEnv(t)

		now := time.Now()
		expires := now.Add(-time.Hour)
		rows := sqlmock.NewRows([]string{
			"id", "profile_id", "minio_key", "avatar_url", "url_expires_at", "created_at", "updated_at",
		}).AddRow(avatarID, profileID, "key1", "old", expires, now, now)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(GET_AVATAR_BY_PROFILE_ID)).
			WithArgs(profileID).
			WillReturnRows(rows)

		boom := errors.New("boom")
		env.minio.EXPECT().
			GeneratePresignedURL(ctx, testBucket, "key1", profiles.PRESIGNED_URL_EXPIRY).
			Return("", boom)

		_, err := env.repo.GetAvatar(ctx, profileID)
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})

	t.Run("update url fails", func(t *testing.T) {
		env := newRepoEnv(t)

		now := time.Now()
		expires := now.Add(-time.Hour)
		rows := sqlmock.NewRows([]string{
			"id", "profile_id", "minio_key", "avatar_url", "url_expires_at", "created_at", "updated_at",
		}).AddRow(avatarID, profileID, "key1", "old", expires, now, now)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(GET_AVATAR_BY_PROFILE_ID)).
			WithArgs(profileID).
			WillReturnRows(rows)

		env.minio.EXPECT().
			GeneratePresignedURL(ctx, testBucket, "key1", profiles.PRESIGNED_URL_EXPIRY).
			Return("new-url", nil)

		boom := errors.New("update failed")
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(UPDATE_AVATAR_URL)).
			WithArgs(avatarID, "new-url", sqlmock.AnyArg()).
			WillReturnError(boom)

		_, err := env.repo.GetAvatar(ctx, profileID)
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		env := newRepoEnv(t)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(GET_AVATAR_BY_PROFILE_ID)).
			WithArgs(profileID).
			WillReturnError(sql.ErrNoRows)

		_, err := env.repo.GetAvatar(ctx, profileID)
		if !errors.Is(err, profiles.ErrAvatarNotFound) {
			t.Fatalf("expected ErrAvatarNotFound, got %v", err)
		}
	})

	t.Run("db error", func(t *testing.T) {
		env := newRepoEnv(t)
		boom := errors.New("boom")
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(GET_AVATAR_BY_PROFILE_ID)).
			WithArgs(profileID).
			WillReturnError(boom)

		_, err := env.repo.GetAvatar(ctx, profileID)
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestUploadAvatar(t *testing.T) {
	ctx := context.Background()
	profileID := uuid.New()
	fileContent := "abcdef"

	t.Run("success no previous avatar", func(t *testing.T) {
		env := newRepoEnv(t)

		// DeleteAvatar inside UploadAvatar - no prior row
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(DELETE_AVATAR_BY_ID)).
			WithArgs(profileID).
			WillReturnError(sql.ErrNoRows)

		env.minio.EXPECT().
			UploadFile(ctx, testBucket, gomock.Any(), gomock.Any(), int64(6), "image/png").
			Return(nil)
		env.minio.EXPECT().
			GeneratePresignedURL(ctx, testBucket, gomock.Any(), profiles.PRESIGNED_URL_EXPIRY).
			Return("https://signed", nil)

		now := time.Now()
		insertedID := uuid.New()
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(CREATE_AVATAR)).
			WithArgs(sqlmock.AnyArg(), profileID, sqlmock.AnyArg(), "https://signed", sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "profile_id", "minio_key", "avatar_url", "url_expires_at", "created_at", "updated_at",
			}).AddRow(insertedID, profileID, "k", "https://signed", now.Add(profiles.PRESIGNED_URL_EXPIRY), now, now))

		got, err := env.repo.UploadAvatar(ctx, profileID, "file.png", int64(len(fileContent)), "image/png", strings.NewReader(fileContent))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.AvatarURL != "https://signed" {
			t.Errorf("expected signed URL")
		}
	})

	t.Run("delete avatar unexpected db error", func(t *testing.T) {
		env := newRepoEnv(t)

		boom := errors.New("boom")
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(DELETE_AVATAR_BY_ID)).
			WithArgs(profileID).
			WillReturnError(boom)

		_, err := env.repo.UploadAvatar(ctx, profileID, "f", int64(len(fileContent)), "image/png", strings.NewReader(fileContent))
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})

	t.Run("minio upload fails", func(t *testing.T) {
		env := newRepoEnv(t)

		env.sqlMock.ExpectQuery(regexp.QuoteMeta(DELETE_AVATAR_BY_ID)).
			WithArgs(profileID).
			WillReturnError(sql.ErrNoRows)
		env.minio.EXPECT().
			UploadFile(ctx, testBucket, gomock.Any(), gomock.Any(), int64(6), "image/png").
			Return(errors.New("s3 down"))

		_, err := env.repo.UploadAvatar(ctx, profileID, "f", 6, "image/png", strings.NewReader(fileContent))
		if !errors.Is(err, profiles.ErrFailedToUpload) {
			t.Fatalf("expected ErrFailedToUpload, got %v", err)
		}
	})

	t.Run("presign fails", func(t *testing.T) {
		env := newRepoEnv(t)

		env.sqlMock.ExpectQuery(regexp.QuoteMeta(DELETE_AVATAR_BY_ID)).
			WithArgs(profileID).
			WillReturnError(sql.ErrNoRows)
		env.minio.EXPECT().
			UploadFile(ctx, testBucket, gomock.Any(), gomock.Any(), int64(6), "image/png").
			Return(nil)
		env.minio.EXPECT().
			GeneratePresignedURL(ctx, testBucket, gomock.Any(), profiles.PRESIGNED_URL_EXPIRY).
			Return("", errors.New("nope"))
		env.minio.EXPECT().
			DeleteFile(ctx, testBucket, gomock.Any()).
			Return(nil)

		_, err := env.repo.UploadAvatar(ctx, profileID, "f", 6, "image/png", strings.NewReader(fileContent))
		if !errors.Is(err, profiles.ErrFailedToGenerateURL) {
			t.Fatalf("expected ErrFailedToGenerateURL, got %v", err)
		}
	})

	t.Run("insert fails rolls back minio", func(t *testing.T) {
		env := newRepoEnv(t)

		env.sqlMock.ExpectQuery(regexp.QuoteMeta(DELETE_AVATAR_BY_ID)).
			WithArgs(profileID).
			WillReturnError(sql.ErrNoRows)
		env.minio.EXPECT().
			UploadFile(ctx, testBucket, gomock.Any(), gomock.Any(), int64(6), "image/png").
			Return(nil)
		env.minio.EXPECT().
			GeneratePresignedURL(ctx, testBucket, gomock.Any(), profiles.PRESIGNED_URL_EXPIRY).
			Return("https://signed", nil)

		boom := errors.New("insert fail")
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(CREATE_AVATAR)).
			WillReturnError(boom)
		env.minio.EXPECT().
			DeleteFile(ctx, testBucket, gomock.Any()).
			Return(nil)

		_, err := env.repo.UploadAvatar(ctx, profileID, "f", 6, "image/png", strings.NewReader(fileContent))
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestDeleteAvatar(t *testing.T) {
	ctx := context.Background()
	profileID := uuid.New()

	t.Run("success", func(t *testing.T) {
		env := newRepoEnv(t)
		rows := sqlmock.NewRows([]string{"minio_key"}).AddRow("the-key")
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(DELETE_AVATAR_BY_ID)).
			WithArgs(profileID).
			WillReturnRows(rows)
		env.minio.EXPECT().DeleteFile(ctx, testBucket, "the-key").Return(nil)

		if err := env.repo.DeleteAvatar(ctx, profileID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		env := newRepoEnv(t)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(DELETE_AVATAR_BY_ID)).
			WithArgs(profileID).
			WillReturnError(sql.ErrNoRows)

		if err := env.repo.DeleteAvatar(ctx, profileID); !errors.Is(err, profiles.ErrAvatarNotFound) {
			t.Fatalf("expected ErrAvatarNotFound, got %v", err)
		}
	})

	t.Run("db error", func(t *testing.T) {
		env := newRepoEnv(t)
		boom := errors.New("boom")
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(DELETE_AVATAR_BY_ID)).
			WithArgs(profileID).
			WillReturnError(boom)

		if err := env.repo.DeleteAvatar(ctx, profileID); !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})

	t.Run("minio delete fails", func(t *testing.T) {
		env := newRepoEnv(t)
		rows := sqlmock.NewRows([]string{"minio_key"}).AddRow("the-key")
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(DELETE_AVATAR_BY_ID)).
			WithArgs(profileID).
			WillReturnRows(rows)
		boom := errors.New("s3 down")
		env.minio.EXPECT().DeleteFile(ctx, testBucket, "the-key").Return(boom)

		if err := env.repo.DeleteAvatar(ctx, profileID); !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestChangePassword(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		env := newRepoEnv(t)

		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "username", "created_at", "updated_at"}).
			AddRow(userID, "alice", now, now)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(CHANGE_PASSWORD_BY_USER_ID)).
			WithArgs(userID, sqlmock.AnyArg()).
			WillReturnRows(rows)

		got, err := env.repo.ChangePassword(ctx, userID, "NewPass1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Username != "alice" {
			t.Errorf("unexpected: %+v", got)
		}
	})

	t.Run("not found", func(t *testing.T) {
		env := newRepoEnv(t)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(CHANGE_PASSWORD_BY_USER_ID)).
			WithArgs(userID, sqlmock.AnyArg()).
			WillReturnError(sql.ErrNoRows)

		_, err := env.repo.ChangePassword(ctx, userID, "NewPass1")
		if !errors.Is(err, profiles.ErrUserNotExist) {
			t.Fatalf("expected ErrUserNotExist, got %v", err)
		}
	})

	t.Run("db error", func(t *testing.T) {
		env := newRepoEnv(t)
		boom := errors.New("boom")
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(CHANGE_PASSWORD_BY_USER_ID)).
			WithArgs(userID, sqlmock.AnyArg()).
			WillReturnError(boom)

		_, err := env.repo.ChangePassword(ctx, userID, "NewPass1")
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestGetPassword(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		env := newRepoEnv(t)
		want := []byte("hashed")
		rows := sqlmock.NewRows([]string{"password"}).AddRow(want)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(GET_PASSWORD_BY_USER_ID)).
			WithArgs(userID).
			WillReturnRows(rows)

		got, err := env.repo.GetPassword(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got) != string(want) {
			t.Errorf("expected %q, got %q", want, got)
		}
	})

	t.Run("not found", func(t *testing.T) {
		env := newRepoEnv(t)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(GET_PASSWORD_BY_USER_ID)).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		_, err := env.repo.GetPassword(ctx, userID)
		if !errors.Is(err, profiles.ErrUserNotExist) {
			t.Fatalf("expected ErrUserNotExist, got %v", err)
		}
	})

	t.Run("db error", func(t *testing.T) {
		env := newRepoEnv(t)
		boom := errors.New("boom")
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(GET_PASSWORD_BY_USER_ID)).
			WithArgs(userID).
			WillReturnError(boom)

		_, err := env.repo.GetPassword(ctx, userID)
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestSignupUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		env := newRepoEnv(t)

		env.sqlMock.ExpectQuery(regexp.QuoteMeta(CHECK_USER_EXISTS)).
			WithArgs("alice").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
		env.sqlMock.ExpectExec(regexp.QuoteMeta(CREATE_USER)).
			WithArgs(sqlmock.AnyArg(), "alice", sqlmock.AnyArg(), 1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		got, err := env.repo.SignupUser(ctx, "alice", "NewPass1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Username != "alice" {
			t.Errorf("expected alice, got %q", got.Username)
		}
		if err := bcrypt.CompareHashAndPassword(got.Password, []byte("NewPass1")); err != nil {
			t.Errorf("password hash should match: %v", err)
		}
	})

	t.Run("check exists fails", func(t *testing.T) {
		env := newRepoEnv(t)
		boom := errors.New("boom")
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(CHECK_USER_EXISTS)).
			WithArgs("alice").
			WillReturnError(boom)

		_, err := env.repo.SignupUser(ctx, "alice", "x")
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})

	t.Run("username exists", func(t *testing.T) {
		env := newRepoEnv(t)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(CHECK_USER_EXISTS)).
			WithArgs("alice").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		_, err := env.repo.SignupUser(ctx, "alice", "x")
		if !errors.Is(err, profiles.ErrUsernameExists) {
			t.Fatalf("expected ErrUsernameExists, got %v", err)
		}
	})

	t.Run("insert fails", func(t *testing.T) {
		env := newRepoEnv(t)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(CHECK_USER_EXISTS)).
			WithArgs("alice").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		boom := errors.New("insert fail")
		env.sqlMock.ExpectExec(regexp.QuoteMeta(CREATE_USER)).
			WillReturnError(boom)

		_, err := env.repo.SignupUser(ctx, "alice", "NewPass1")
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

func TestSigninUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		env := newRepoEnv(t)
		id := uuid.New()
		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"id", "username", "password", "token_version", "created_at", "updated_at",
		}).AddRow(id, "alice", []byte("pw"), 1, now, now)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(GET_USER_BY_USERNAME)).
			WithArgs("alice").
			WillReturnRows(rows)

		got, err := env.repo.SigninUser(ctx, "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != id {
			t.Errorf("unexpected id")
		}
	})

	t.Run("not found", func(t *testing.T) {
		env := newRepoEnv(t)
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(GET_USER_BY_USERNAME)).
			WithArgs("alice").
			WillReturnError(sql.ErrNoRows)

		_, err := env.repo.SigninUser(ctx, "alice")
		if !errors.Is(err, profiles.ErrUserNotExist) {
			t.Fatalf("expected ErrUserNotExist, got %v", err)
		}
	})

	t.Run("db error", func(t *testing.T) {
		env := newRepoEnv(t)
		boom := errors.New("boom")
		env.sqlMock.ExpectQuery(regexp.QuoteMeta(GET_USER_BY_USERNAME)).
			WithArgs("alice").
			WillReturnError(boom)

		_, err := env.repo.SigninUser(ctx, "alice")
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom, got %v", err)
		}
	})
}

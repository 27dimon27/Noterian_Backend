package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/google/uuid"
)

func TestGetProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewProfileRepository(db)
	userID := uuid.New()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "created_at", "updated_at"}).
			AddRow(userID, "testuser", now, now)

		mock.ExpectQuery("SELECT id, username, created_at, updated_at FROM profiles WHERE id").
			WithArgs(userID).
			WillReturnRows(rows)

		profile, err := repo.GetProfile(context.Background(), userID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if profile.ID != userID {
			t.Errorf("expected ID %v, got %v", userID, profile.ID)
		}
		if profile.Username != "testuser" {
			t.Errorf("expected Username 'testuser', got '%s'", profile.Username)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, username, created_at, updated_at FROM profiles WHERE id").
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetProfile(context.Background(), userID)

		if !errors.Is(err, profiles.ErrUserNotExist) {
			t.Errorf("expected ErrUserNotExist, got %v", err)
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, username, created_at, updated_at FROM profiles WHERE id").
			WithArgs(userID).
			WillReturnError(errors.New("db error"))

		_, err := repo.GetProfile(context.Background(), userID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestUpdateProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewProfileRepository(db)
	userID := uuid.New()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		profile := models.Profile{
			Username: "newusername",
		}
		rows := sqlmock.NewRows([]string{"id", "username", "created_at", "updated_at"}).
			AddRow(userID, "newusername", now, now)

		mock.ExpectQuery("UPDATE profiles SET username").
			WithArgs(userID, "newusername").
			WillReturnRows(rows)

		updated, err := repo.UpdateProfile(context.Background(), userID, profile)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if updated.Username != "newusername" {
			t.Errorf("expected Username 'newusername', got '%s'", updated.Username)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		profile := models.Profile{
			Username: "newusername",
		}

		mock.ExpectQuery("UPDATE profiles SET username").
			WithArgs(userID, "newusername").
			WillReturnError(sql.ErrNoRows)

		_, err := repo.UpdateProfile(context.Background(), userID, profile)

		if !errors.Is(err, profiles.ErrUserNotExist) {
			t.Errorf("expected ErrUserNotExist, got %v", err)
		}
	})

	t.Run("query error", func(t *testing.T) {
		profile := models.Profile{
			Username: "newusername",
		}

		mock.ExpectQuery("UPDATE profiles SET username").
			WithArgs(userID, "newusername").
			WillReturnError(errors.New("db error"))

		_, err := repo.UpdateProfile(context.Background(), userID, profile)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestDeleteProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewProfileRepository(db)
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id"}).AddRow(userID)

		mock.ExpectQuery("DELETE FROM profiles WHERE id").
			WithArgs(userID).
			WillReturnRows(rows)

		err := repo.DeleteProfile(context.Background(), userID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mock.ExpectQuery("DELETE FROM profiles WHERE id").
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		err := repo.DeleteProfile(context.Background(), userID)

		if !errors.Is(err, profiles.ErrUserNotExist) {
			t.Errorf("expected ErrUserNotExist, got %v", err)
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("DELETE FROM profiles WHERE id").
			WithArgs(userID).
			WillReturnError(errors.New("db error"))

		err := repo.DeleteProfile(context.Background(), userID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

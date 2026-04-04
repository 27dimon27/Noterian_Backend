package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/google/uuid"
)

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)
	if repo == nil {
		t.Errorf("expected non-nil repository")
	}
	if repo.db != db {
		t.Errorf("expected db to be set")
	}

	login := "testuser"
	password := "validPassword123"

	t.Run("success creation", func(t *testing.T) {
		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(login).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		mock.ExpectExec("INSERT INTO profiles").
			WithArgs(sqlmock.AnyArg(), login, sqlmock.AnyArg(), 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		user, err := repo.CreateUser(context.Background(), login, password)
		if err != nil {
			t.Errorf("unexpected err: %s", err)
			return
		}

		if user.Username != login {
			t.Errorf("expected username %s, got %s", login, user.Username)
		}
		if user.TokenVersion != 1 {
			t.Errorf("expected token_version 1, got %d", user.TokenVersion)
		}
		if user.ID == uuid.Nil {
			t.Errorf("expected non-nil id")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("user already exists", func(t *testing.T) {
		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(login).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		_, err := repo.CreateUser(context.Background(), login, password)
		if !errors.Is(err, auth.ErrUserExist) {
			t.Errorf("expected ErrUserExist, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("check query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(login).
			WillReturnError(errors.New("db error"))

		_, err := repo.CreateUser(context.Background(), login, password)
		if err == nil {
			t.Errorf("expected error, got nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("insert query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(login).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		mock.ExpectExec("INSERT INTO profiles").
			WithArgs(sqlmock.AnyArg(), login, sqlmock.AnyArg(), 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("insert error"))

		_, err := repo.CreateUser(context.Background(), login, password)
		if err == nil {
			t.Errorf("expected error, got nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestGetUserByLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)
	if repo == nil {
		t.Errorf("expected non-nil repository")
	}
	if repo.db != db {
		t.Errorf("expected db to be set")
	}

	login := "testuser"
	userID := uuid.New()
	hashedPassword := []byte("hashedpassword")
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "password", "token_version", "created_at", "updated_at"}).
			AddRow(userID, login, hashedPassword, 1, now, now)

		mock.ExpectQuery("SELECT id, username, password, token_version, created_at, updated_at FROM profiles").
			WithArgs(login).
			WillReturnRows(rows)

		user, err := repo.GetUserByUsername(context.Background(), login)
		if err != nil {
			t.Errorf("unexpected err: %s", err)
			return
		}

		if user.ID != userID {
			t.Errorf("expected id %v, got %v", userID, user.ID)
		}
		if user.Username != login {
			t.Errorf("expected username %s, got %s", login, user.Username)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, username, password, token_version, created_at, updated_at FROM profiles").
			WithArgs(login).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetUserByUsername(context.Background(), login)
		if !errors.Is(err, auth.ErrUserNotExist) {
			t.Errorf("expected ErrUserNotExist, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, username, password, token_version, created_at, updated_at FROM profiles").
			WithArgs(login).
			WillReturnError(errors.New("db error"))

		_, err := repo.GetUserByUsername(context.Background(), login)
		if err == nil {
			t.Errorf("expected error, got nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("scan error - wrong columns", func(t *testing.T) {
		// возвращаем неправильное количество колонок
		rows := sqlmock.NewRows([]string{"id", "username"}).
			AddRow(userID, login)

		mock.ExpectQuery("SELECT id, username, password, token_version, created_at, updated_at FROM profiles").
			WithArgs(login).
			WillReturnRows(rows)

		_, err := repo.GetUserByUsername(context.Background(), login)
		if err == nil {
			t.Errorf("expected error, got nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

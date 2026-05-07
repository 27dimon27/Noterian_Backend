package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepository(t *testing.T) (*userRepository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := NewUserRepository(db)
	return repo, mock
}

func TestUserRepository_SignupUser_Success(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	username := "testuser"
	password := "Test1234"

	// Expect check for existing user
	mock.ExpectQuery(regexp.QuoteMeta(CHECK_USER_EXISTS)).
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Expect insert
	mock.ExpectExec(regexp.QuoteMeta(CREATE_USER)).
		WithArgs(sqlmock.AnyArg(), username, sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	user, err := repo.SignupUser(context.Background(), username, password)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, 1, user.TokenVersion)
	assert.NotEmpty(t, user.ID)
	assert.NotEmpty(t, user.Password)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_SignupUser_UserAlreadyExists(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	username := "existinguser"
	password := "Test1234"

	// Expect check for existing user - returns true
	mock.ExpectQuery(regexp.QuoteMeta(CHECK_USER_EXISTS)).
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	user, err := repo.SignupUser(context.Background(), username, password)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, auth.ErrUserExist)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_SignupUser_CheckExistsQueryError(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	username := "testuser"
	password := "Test1234"

	mock.ExpectQuery(regexp.QuoteMeta(CHECK_USER_EXISTS)).
		WithArgs(username).
		WillReturnError(sql.ErrConnDone)

	user, err := repo.SignupUser(context.Background(), username, password)

	assert.Nil(t, user)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_SignupUser_InsertError(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	username := "testuser"
	password := "Test1234"

	mock.ExpectQuery(regexp.QuoteMeta(CHECK_USER_EXISTS)).
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectExec(regexp.QuoteMeta(CREATE_USER)).
		WithArgs(sqlmock.AnyArg(), username, sqlmock.AnyArg(), 1).
		WillReturnError(sql.ErrTxDone)

	user, err := repo.SignupUser(context.Background(), username, password)

	assert.Nil(t, user)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_SigninUser_Success(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	userID := uuid.New()
	username := "testuser"
	hashedPassword := []byte("hashedpassword")
	tokenVersion := 1
	now := time.Now()

	// Используем правильные типы данных для time.Time
	rows := sqlmock.NewRows([]string{"id", "username", "password", "token_version", "created_at", "updated_at"}).
		AddRow(userID.String(), username, hashedPassword, tokenVersion, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(GET_USER_BY_USERNAME)).
		WithArgs(username).
		WillReturnRows(rows)

	user, err := repo.SigninUser(context.Background(), username)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, hashedPassword, user.Password)
	assert.Equal(t, tokenVersion, user.TokenVersion)
	assert.WithinDuration(t, now, user.CreatedAt, time.Second)
	assert.WithinDuration(t, now, user.UpdatedAt, time.Second)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_SigninUser_NotFound(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	username := "nonexistent"

	mock.ExpectQuery(regexp.QuoteMeta(GET_USER_BY_USERNAME)).
		WithArgs(username).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.SigninUser(context.Background(), username)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, auth.ErrUserNotExist)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_SigninUser_QueryError(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	username := "testuser"

	mock.ExpectQuery(regexp.QuoteMeta(GET_USER_BY_USERNAME)).
		WithArgs(username).
		WillReturnError(sql.ErrConnDone)

	user, err := repo.SigninUser(context.Background(), username)

	assert.Nil(t, user)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_SigninUser_ScanError(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	username := "testuser"

	// Return wrong number of columns to cause scan error
	rows := sqlmock.NewRows([]string{"id", "username"}).
		AddRow(uuid.New().String(), username)

	mock.ExpectQuery(regexp.QuoteMeta(GET_USER_BY_USERNAME)).
		WithArgs(username).
		WillReturnRows(rows)

	user, err := repo.SigninUser(context.Background(), username)

	assert.Nil(t, user)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_SigninUser_WithNullTimestamps(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	userID := uuid.New()
	username := "testuser"
	hashedPassword := []byte("hashedpassword")
	tokenVersion := 1

	// Тестируем случай, когда в БД могут быть NULL значения (хотя по схеме они NOT NULL)
	rows := sqlmock.NewRows([]string{"id", "username", "password", "token_version", "created_at", "updated_at"}).
		AddRow(userID.String(), username, hashedPassword, tokenVersion, nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta(GET_USER_BY_USERNAME)).
		WithArgs(username).
		WillReturnRows(rows)

	user, err := repo.SigninUser(context.Background(), username)

	// Это должно вызвать ошибку сканирования, так как NULL нельзя присвоить time.Time
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_SignupUser_WithSpecialCharacters(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"with underscore", "test_user", "Test1234"},
		{"with dot", "test.user", "Test1234"},
		{"with numbers", "test123", "Test1234"},
		{"russian", "тест", "Test1234"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(CHECK_USER_EXISTS)).
				WithArgs(tc.username).
				WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

			mock.ExpectExec(regexp.QuoteMeta(CREATE_USER)).
				WithArgs(sqlmock.AnyArg(), tc.username, sqlmock.AnyArg(), 1).
				WillReturnResult(sqlmock.NewResult(1, 1))

			user, err := repo.SignupUser(context.Background(), tc.username, tc.password)

			assert.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, tc.username, user.Username)
		})
	}
}

func TestUserRepository_ConcurrentSignupUser(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	username := "testuser"
	password := "Test1234"

	// Simulate concurrent access by checking expectations after each call
	for i := 0; i < 3; i++ {
		mock.ExpectQuery(regexp.QuoteMeta(CHECK_USER_EXISTS)).
			WithArgs(username).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		mock.ExpectExec(regexp.QuoteMeta(CREATE_USER)).
			WithArgs(sqlmock.AnyArg(), username, sqlmock.AnyArg(), 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		user, err := repo.SignupUser(context.Background(), username, password)
		assert.NoError(t, err)
		assert.NotNil(t, user)
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_SignupUser_WithEmptyFields(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"empty username", "", "Test1234"},
		{"empty password", "testuser", ""},
		{"both empty", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(CHECK_USER_EXISTS)).
				WithArgs(tc.username).
				WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

			mock.ExpectExec(regexp.QuoteMeta(CREATE_USER)).
				WithArgs(sqlmock.AnyArg(), tc.username, sqlmock.AnyArg(), 1).
				WillReturnResult(sqlmock.NewResult(1, 1))

			user, err := repo.SignupUser(context.Background(), tc.username, tc.password)

			// Репозиторий не валидирует поля, он просто передает их в БД
			// Поэтому ошибки не будет, даже если поля пустые
			assert.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, tc.username, user.Username)
		})
	}
}

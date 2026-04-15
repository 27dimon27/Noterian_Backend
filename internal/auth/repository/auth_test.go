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
	"golang.org/x/crypto/bcrypt"
)

func setupTestRepository(t *testing.T) (*userRepository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := NewUserRepository(db)
	return repo, mock
}

func TestCreateUser_Success(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	ctx := context.Background()
	username := "testuser"
	password := "Test1234"

	mock.ExpectQuery(regexp.QuoteMeta(CHECK_USER_EXISTS)).
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectExec(regexp.QuoteMeta(CREATE_USER)).
		WithArgs(sqlmock.AnyArg(), username, sqlmock.AnyArg(), 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	user, err := repo.CreateUser(ctx, username, password)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, 1, user.TokenVersion)
	assert.NotEmpty(t, user.ID)
	assert.NotEmpty(t, user.Password)

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_UserAlreadyExists(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	ctx := context.Background()
	username := "existinguser"
	password := "Test1234"

	mock.ExpectQuery(regexp.QuoteMeta(CHECK_USER_EXISTS)).
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	user, err := repo.CreateUser(ctx, username, password)

	assert.Error(t, err)
	assert.Equal(t, auth.ErrUserExist, err)
	assert.Nil(t, user)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_CheckExistsError(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	ctx := context.Background()
	username := "testuser"
	password := "Test1234"

	mock.ExpectQuery(regexp.QuoteMeta(CHECK_USER_EXISTS)).
		WithArgs(username).
		WillReturnError(sql.ErrConnDone)

	user, err := repo.CreateUser(ctx, username, password)

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.Nil(t, user)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_InsertError(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	ctx := context.Background()
	username := "testuser"
	password := "Test1234"

	mock.ExpectQuery(regexp.QuoteMeta(CHECK_USER_EXISTS)).
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectExec(regexp.QuoteMeta(CREATE_USER)).
		WithArgs(sqlmock.AnyArg(), username, sqlmock.AnyArg(), 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrTxDone)

	user, err := repo.CreateUser(ctx, username, password)

	assert.Error(t, err)
	assert.Equal(t, sql.ErrTxDone, err)
	assert.Nil(t, user)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByUsername_Success(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	ctx := context.Background()
	username := "testuser"
	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Test1234"), bcrypt.DefaultCost)

	now := time.Now()
	createdAt := now
	updatedAt := now

	rows := sqlmock.NewRows([]string{"id", "username", "password", "token_version", "created_at", "updated_at"}).
		AddRow(userID, username, hashedPassword, 1, createdAt, updatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(GET_USER_BY_USERNAME)).
		WithArgs(username).
		WillReturnRows(rows)

	user, err := repo.GetUserByUsername(ctx, username)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, hashedPassword, user.Password)
	assert.Equal(t, 1, user.TokenVersion)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByUsername_UserNotFound(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	ctx := context.Background()
	username := "nonexistent"

	mock.ExpectQuery(regexp.QuoteMeta(GET_USER_BY_USERNAME)).
		WithArgs(username).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByUsername(ctx, username)

	assert.Error(t, err)
	assert.Equal(t, auth.ErrUserNotExist, err)
	assert.Nil(t, user)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByUsername_DatabaseError(t *testing.T) {
	repo, mock := setupTestRepository(t)
	defer repo.db.Close()

	ctx := context.Background()
	username := "testuser"

	mock.ExpectQuery(regexp.QuoteMeta(GET_USER_BY_USERNAME)).
		WithArgs(username).
		WillReturnError(sql.ErrConnDone)

	user, err := repo.GetUserByUsername(ctx, username)

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.Nil(t, user)

	assert.NoError(t, mock.ExpectationsWereMet())
}

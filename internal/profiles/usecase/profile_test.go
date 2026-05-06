package usecase

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockProfileRepo struct {
	getProfileFunc           func(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
	getProfileByUsernameFunc func(ctx context.Context, username string) (*models.Profile, error)
	updateProfileFunc        func(ctx context.Context, userID uuid.UUID, profile models.Profile) (*models.Profile, error)
	deleteProfileFunc        func(ctx context.Context, userID uuid.UUID) error
	getAvatarFunc            func(ctx context.Context, profileID uuid.UUID) (*models.Avatar, error)
	uploadAvatarFunc         func(ctx context.Context, profileID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Avatar, error)
	deleteAvatarFunc         func(ctx context.Context, profileID uuid.UUID) error
	changePasswordFunc       func(ctx context.Context, userID uuid.UUID, newPassword string) (*models.Profile, error)
	getPasswordFunc          func(ctx context.Context, userID uuid.UUID) ([]byte, error)
}

func (m *mockProfileRepo) GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	if m.getProfileFunc != nil {
		return m.getProfileFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockProfileRepo) GetProfileByUsername(ctx context.Context, username string) (*models.Profile, error) {
	if m.getProfileByUsernameFunc != nil {
		return m.getProfileByUsernameFunc(ctx, username)
	}
	return nil, nil
}

func (m *mockProfileRepo) UpdateProfile(ctx context.Context, userID uuid.UUID, profile models.Profile) (*models.Profile, error) {
	if m.updateProfileFunc != nil {
		return m.updateProfileFunc(ctx, userID, profile)
	}
	return nil, nil
}

func (m *mockProfileRepo) DeleteProfile(ctx context.Context, userID uuid.UUID) error {
	if m.deleteProfileFunc != nil {
		return m.deleteProfileFunc(ctx, userID)
	}
	return nil
}

func (m *mockProfileRepo) GetAvatar(ctx context.Context, profileID uuid.UUID) (*models.Avatar, error) {
	if m.getAvatarFunc != nil {
		return m.getAvatarFunc(ctx, profileID)
	}
	return nil, nil
}

func (m *mockProfileRepo) UploadAvatar(ctx context.Context, profileID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Avatar, error) {
	if m.uploadAvatarFunc != nil {
		return m.uploadAvatarFunc(ctx, profileID, fileName, fileSize, mimeType, fileReader)
	}
	return nil, nil
}

func (m *mockProfileRepo) DeleteAvatar(ctx context.Context, profileID uuid.UUID) error {
	if m.deleteAvatarFunc != nil {
		return m.deleteAvatarFunc(ctx, profileID)
	}
	return nil
}

func (m *mockProfileRepo) ChangePassword(ctx context.Context, userID uuid.UUID, newPassword string) (*models.Profile, error) {
	if m.changePasswordFunc != nil {
		return m.changePasswordFunc(ctx, userID, newPassword)
	}
	return nil, nil
}

func (m *mockProfileRepo) GetPassword(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	if m.getPasswordFunc != nil {
		return m.getPasswordFunc(ctx, userID)
	}
	return nil, nil
}

func TestProfileUsecase_GetProfile(t *testing.T) {
	userID := uuid.New()
	expectedProfile := &models.Profile{
		ID:       userID,
		Username: "testuser",
	}

	tests := []struct {
		name      string
		setupMock func(*mockProfileRepo)
		wantErr   bool
		errType   error
	}{
		{
			name: "Success",
			setupMock: func(m *mockProfileRepo) {
				m.getProfileFunc = func(ctx context.Context, id uuid.UUID) (*models.Profile, error) {
					return expectedProfile, nil
				}
			},
			wantErr: false,
		},
		{
			name: "User Not Found",
			setupMock: func(m *mockProfileRepo) {
				m.getProfileFunc = func(ctx context.Context, id uuid.UUID) (*models.Profile, error) {
					return nil, profiles.ErrUserNotExist
				}
			},
			wantErr: true,
			errType: profiles.ErrUserNotExist,
		},
		{
			name: "Database Error",
			setupMock: func(m *mockProfileRepo) {
				m.getProfileFunc = func(ctx context.Context, id uuid.UUID) (*models.Profile, error) {
					return nil, errors.New("database connection failed")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProfileRepo{}
			tt.setupMock(mockRepo)

			usecase, err := NewProfileUsecase(mockRepo)
			assert.NoError(t, err)

			profile, err := usecase.GetProfile(context.Background(), userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedProfile, profile)
			}
		})
	}
}

func TestProfileUsecase_UpdateProfile(t *testing.T) {
	userID := uuid.New()
	existingUsername := "existinguser"
	newUsername := "newusername"

	tests := []struct {
		name      string
		profile   models.Profile
		setupMock func(*mockProfileRepo)
		wantErr   bool
		errType   error
	}{
		{
			name: "Success",
			profile: models.Profile{
				Username: newUsername,
			},
			setupMock: func(m *mockProfileRepo) {
				m.getProfileByUsernameFunc = func(ctx context.Context, username string) (*models.Profile, error) {
					return nil, profiles.ErrUserNotExist
				}
				m.updateProfileFunc = func(ctx context.Context, id uuid.UUID, profile models.Profile) (*models.Profile, error) {
					return &models.Profile{
						ID:       id,
						Username: newUsername,
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "Invalid Username - Empty",
			profile: models.Profile{
				Username: "",
			},
			setupMock: func(m *mockProfileRepo) {},
			wantErr:   true,
			errType:   profiles.ErrInvalidProfileData,
		},
		{
			name: "Username Already Exists",
			profile: models.Profile{
				Username: existingUsername,
			},
			setupMock: func(m *mockProfileRepo) {
				m.getProfileByUsernameFunc = func(ctx context.Context, username string) (*models.Profile, error) {
					return &models.Profile{
						ID:       uuid.New(),
						Username: existingUsername,
					}, nil
				}
			},
			wantErr: true,
			errType: profiles.ErrUsernameExists,
		},
		{
			name: "Database Error On GetByUsername",
			profile: models.Profile{
				Username: newUsername,
			},
			setupMock: func(m *mockProfileRepo) {
				m.getProfileByUsernameFunc = func(ctx context.Context, username string) (*models.Profile, error) {
					return nil, errors.New("database error")
				}
			},
			wantErr: true,
		},
		{
			name: "User Not Found On Update",
			profile: models.Profile{
				Username: newUsername,
			},
			setupMock: func(m *mockProfileRepo) {
				m.getProfileByUsernameFunc = func(ctx context.Context, username string) (*models.Profile, error) {
					return nil, profiles.ErrUserNotExist
				}
				m.updateProfileFunc = func(ctx context.Context, id uuid.UUID, profile models.Profile) (*models.Profile, error) {
					return nil, profiles.ErrUserNotExist
				}
			},
			wantErr: true,
			errType: profiles.ErrUserNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProfileRepo{}
			tt.setupMock(mockRepo)

			usecase, err := NewProfileUsecase(mockRepo)
			assert.NoError(t, err)

			profile, err := usecase.UpdateProfile(context.Background(), userID, tt.profile)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, profile)
			}
		})
	}
}

func TestProfileUsecase_DeleteProfile(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name      string
		setupMock func(*mockProfileRepo)
		wantErr   bool
		errType   error
	}{
		{
			name: "Success",
			setupMock: func(m *mockProfileRepo) {
				m.deleteProfileFunc = func(ctx context.Context, id uuid.UUID) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "User Not Found",
			setupMock: func(m *mockProfileRepo) {
				m.deleteProfileFunc = func(ctx context.Context, id uuid.UUID) error {
					return profiles.ErrUserNotExist
				}
			},
			wantErr: true,
			errType: profiles.ErrUserNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProfileRepo{}
			tt.setupMock(mockRepo)

			usecase, err := NewProfileUsecase(mockRepo)
			assert.NoError(t, err)

			err = usecase.DeleteProfile(context.Background(), userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProfileUsecase_GetAvatar(t *testing.T) {
	profileID := uuid.New()
	expectedAvatar := &models.Avatar{
		ID:        uuid.New(),
		ProfileID: profileID,
		AvatarURL: "http://example.com/avatar.jpg",
	}

	tests := []struct {
		name      string
		setupMock func(*mockProfileRepo)
		wantErr   bool
		errType   error
	}{
		{
			name: "Success",
			setupMock: func(m *mockProfileRepo) {
				m.getAvatarFunc = func(ctx context.Context, id uuid.UUID) (*models.Avatar, error) {
					return expectedAvatar, nil
				}
			},
			wantErr: false,
		},
		{
			name: "Avatar Not Found",
			setupMock: func(m *mockProfileRepo) {
				m.getAvatarFunc = func(ctx context.Context, id uuid.UUID) (*models.Avatar, error) {
					return nil, nil
				}
			},
			wantErr: true,
			errType: profiles.ErrAvatarNotFound,
		},
		{
			name: "Database Error",
			setupMock: func(m *mockProfileRepo) {
				m.getAvatarFunc = func(ctx context.Context, id uuid.UUID) (*models.Avatar, error) {
					return nil, errors.New("database error")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProfileRepo{}
			tt.setupMock(mockRepo)

			usecase, err := NewProfileUsecase(mockRepo)
			assert.NoError(t, err)

			avatar, err := usecase.GetAvatar(context.Background(), profileID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedAvatar, avatar)
			}
		})
	}
}

func TestProfileUsecase_UploadAvatar(t *testing.T) {
	profileID := uuid.New()
	fileName := "test.jpg"
	fileSize := int64(1024)
	mimeType := "image/jpeg"
	fileReader := strings.NewReader("test image data")

	tests := []struct {
		name      string
		setupMock func(*mockProfileRepo)
		wantErr   bool
	}{
		{
			name: "Success",
			setupMock: func(m *mockProfileRepo) {
				m.uploadAvatarFunc = func(ctx context.Context, id uuid.UUID, fn string, fs int64, mt string, fr io.Reader) (*models.Avatar, error) {
					return &models.Avatar{
						ID:        uuid.New(),
						ProfileID: id,
						AvatarURL: "http://example.com/avatar.jpg",
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "Upload Failed",
			setupMock: func(m *mockProfileRepo) {
				m.uploadAvatarFunc = func(ctx context.Context, id uuid.UUID, fn string, fs int64, mt string, fr io.Reader) (*models.Avatar, error) {
					return nil, profiles.ErrFailedToUpload
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProfileRepo{}
			tt.setupMock(mockRepo)

			usecase, err := NewProfileUsecase(mockRepo)
			assert.NoError(t, err)

			avatar, err := usecase.UploadAvatar(context.Background(), profileID, fileName, fileSize, mimeType, fileReader)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, avatar)
			}
		})
	}
}

func TestProfileUsecase_DeleteAvatar(t *testing.T) {
	profileID := uuid.New()

	tests := []struct {
		name      string
		setupMock func(*mockProfileRepo)
		wantErr   bool
		errType   error
	}{
		{
			name: "Success",
			setupMock: func(m *mockProfileRepo) {
				m.deleteAvatarFunc = func(ctx context.Context, id uuid.UUID) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "Avatar Not Found",
			setupMock: func(m *mockProfileRepo) {
				m.deleteAvatarFunc = func(ctx context.Context, id uuid.UUID) error {
					return profiles.ErrAvatarNotFound
				}
			},
			wantErr: true,
			errType: profiles.ErrAvatarNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProfileRepo{}
			tt.setupMock(mockRepo)

			usecase, err := NewProfileUsecase(mockRepo)
			assert.NoError(t, err)

			err = usecase.DeleteAvatar(context.Background(), profileID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

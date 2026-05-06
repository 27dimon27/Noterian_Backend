package usecase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockAttachmentRepository struct {
	getAttachmentFunc    func(ctx context.Context, blockID uuid.UUID) (*models.Attachment, error)
	uploadAttachmentFunc func(ctx context.Context, blockID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Attachment, error)
	deleteAttachmentFunc func(ctx context.Context, blockID uuid.UUID) error
}

func (m *mockAttachmentRepository) GetAttachment(ctx context.Context, blockID uuid.UUID) (*models.Attachment, error) {
	if m.getAttachmentFunc != nil {
		return m.getAttachmentFunc(ctx, blockID)
	}
	return nil, nil
}

func (m *mockAttachmentRepository) UploadAttachment(ctx context.Context, blockID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Attachment, error) {
	if m.uploadAttachmentFunc != nil {
		return m.uploadAttachmentFunc(ctx, blockID, fileName, fileSize, mimeType, fileReader)
	}
	return nil, nil
}

func (m *mockAttachmentRepository) DeleteAttachment(ctx context.Context, blockID uuid.UUID) error {
	if m.deleteAttachmentFunc != nil {
		return m.deleteAttachmentFunc(ctx, blockID)
	}
	return nil
}

type mockNoteRepository struct {
	getNoteFunc  func(ctx context.Context, noteID uuid.UUID) (*models.Note, error)
	getBlockFunc func(ctx context.Context, blockID uuid.UUID) (*models.Block, error)
}

func (m *mockNoteRepository) GetNote(ctx context.Context, noteID uuid.UUID) (*models.Note, error) {
	if m.getNoteFunc != nil {
		return m.getNoteFunc(ctx, noteID)
	}
	return nil, nil
}

func (m *mockNoteRepository) GetBlock(ctx context.Context, blockID uuid.UUID) (*models.Block, error) {
	if m.getBlockFunc != nil {
		return m.getBlockFunc(ctx, blockID)
	}
	return nil, nil
}

func TestAttachmentUsecase_GetAttachment(t *testing.T) {
	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	tests := []struct {
		name          string
		setupMock     func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository)
		expectedError error
	}{
		{
			name: "success",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: userID}, nil
				}
				noteRepo.getBlockFunc = func(ctx context.Context, id uuid.UUID) (*models.Block, error) {
					return &models.Block{ID: blockID, NoteID: noteID}, nil
				}
				attachmentRepo.getAttachmentFunc = func(ctx context.Context, id uuid.UUID) (*models.Attachment, error) {
					return &models.Attachment{ID: uuid.New(), BlockID: blockID}, nil
				}
			},
			expectedError: nil,
		},
		{
			name: "error - note not found",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return nil, nil
				}
			},
			expectedError: attachments.ErrNoteNotFound,
		},
		{
			name: "error - note repository error",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return nil, errors.New("database error")
				}
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "error - forbidden note access",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: uuid.New()}, nil
				}
			},
			expectedError: attachments.ErrForbidden,
		},
		{
			name: "error - block not found",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: userID}, nil
				}
				noteRepo.getBlockFunc = func(ctx context.Context, id uuid.UUID) (*models.Block, error) {
					return nil, nil
				}
			},
			expectedError: attachments.ErrBlockNotFound,
		},
		{
			name: "error - block repository error",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: userID}, nil
				}
				noteRepo.getBlockFunc = func(ctx context.Context, id uuid.UUID) (*models.Block, error) {
					return nil, errors.New("database error")
				}
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "error - block noteID mismatch",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: userID}, nil
				}
				noteRepo.getBlockFunc = func(ctx context.Context, id uuid.UUID) (*models.Block, error) {
					return &models.Block{ID: blockID, NoteID: uuid.New()}, nil
				}
			},
			expectedError: attachments.ErrForbidden,
		},
		{
			name: "error - attachment not found",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: userID}, nil
				}
				noteRepo.getBlockFunc = func(ctx context.Context, id uuid.UUID) (*models.Block, error) {
					return &models.Block{ID: blockID, NoteID: noteID}, nil
				}
				attachmentRepo.getAttachmentFunc = func(ctx context.Context, id uuid.UUID) (*models.Attachment, error) {
					return nil, nil
				}
			},
			expectedError: attachments.ErrAttachmentNotFound,
		},
		{
			name: "error - attachment repository error",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: userID}, nil
				}
				noteRepo.getBlockFunc = func(ctx context.Context, id uuid.UUID) (*models.Block, error) {
					return &models.Block{ID: blockID, NoteID: noteID}, nil
				}
				attachmentRepo.getAttachmentFunc = func(ctx context.Context, id uuid.UUID) (*models.Attachment, error) {
					return nil, errors.New("database error")
				}
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAttachmentRepo := &mockAttachmentRepository{}
			mockNoteRepo := &mockNoteRepository{}
			tt.setupMock(mockAttachmentRepo, mockNoteRepo)

			usecase := NewAttachmentUsecase(mockAttachmentRepo, mockNoteRepo)
			attachment, err := usecase.GetAttachment(context.Background(), noteID, blockID, userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, attachment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, attachment)
			}
		})
	}
}

func TestAttachmentUsecase_UploadAttachment(t *testing.T) {
	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	fileName := "test.jpg"
	fileSize := int64(1024)
	mimeType := "image/jpeg"
	fileReader := bytes.NewReader([]byte("test data"))

	tests := []struct {
		name          string
		setupMock     func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository)
		expectedError error
	}{
		{
			name: "success",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: userID}, nil
				}
				noteRepo.getBlockFunc = func(ctx context.Context, id uuid.UUID) (*models.Block, error) {
					return &models.Block{ID: blockID, NoteID: noteID}, nil
				}
				attachmentRepo.uploadAttachmentFunc = func(ctx context.Context, id uuid.UUID, fName string, fSize int64, mType string, reader io.Reader) (*models.Attachment, error) {
					return &models.Attachment{ID: uuid.New(), BlockID: blockID}, nil
				}
			},
			expectedError: nil,
		},
		{
			name: "error - note not found",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return nil, nil
				}
			},
			expectedError: attachments.ErrNoteNotFound,
		},
		{
			name: "error - note repository error",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return nil, errors.New("database error")
				}
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "error - forbidden note access",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: uuid.New()}, nil
				}
			},
			expectedError: attachments.ErrForbidden,
		},
		{
			name: "error - block not found",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: userID}, nil
				}
				noteRepo.getBlockFunc = func(ctx context.Context, id uuid.UUID) (*models.Block, error) {
					return nil, nil
				}
			},
			expectedError: attachments.ErrBlockNotFound,
		},
		{
			name: "error - block repository error",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: userID}, nil
				}
				noteRepo.getBlockFunc = func(ctx context.Context, id uuid.UUID) (*models.Block, error) {
					return nil, errors.New("database error")
				}
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "error - block noteID mismatch",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: userID}, nil
				}
				noteRepo.getBlockFunc = func(ctx context.Context, id uuid.UUID) (*models.Block, error) {
					return &models.Block{ID: blockID, NoteID: uuid.New()}, nil
				}
			},
			expectedError: attachments.ErrForbidden,
		},
		{
			name: "error - upload attachment repository error",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: userID}, nil
				}
				noteRepo.getBlockFunc = func(ctx context.Context, id uuid.UUID) (*models.Block, error) {
					return &models.Block{ID: blockID, NoteID: noteID}, nil
				}
				attachmentRepo.uploadAttachmentFunc = func(ctx context.Context, id uuid.UUID, fName string, fSize int64, mType string, reader io.Reader) (*models.Attachment, error) {
					return nil, errors.New("upload error")
				}
			},
			expectedError: errors.New("upload error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAttachmentRepo := &mockAttachmentRepository{}
			mockNoteRepo := &mockNoteRepository{}
			tt.setupMock(mockAttachmentRepo, mockNoteRepo)

			usecase := NewAttachmentUsecase(mockAttachmentRepo, mockNoteRepo)
			attachment, err := usecase.UploadAttachment(context.Background(), noteID, blockID, userID, fileName, fileSize, mimeType, fileReader)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, attachment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, attachment)
			}
		})
	}
}

func TestAttachmentUsecase_DeleteAttachment(t *testing.T) {
	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	tests := []struct {
		name          string
		setupMock     func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository)
		expectedError error
	}{
		{
			name: "success",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: userID}, nil
				}
				noteRepo.getBlockFunc = func(ctx context.Context, id uuid.UUID) (*models.Block, error) {
					return &models.Block{ID: blockID, NoteID: noteID}, nil
				}
				attachmentRepo.deleteAttachmentFunc = func(ctx context.Context, id uuid.UUID) error {
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name: "error - note not found",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return nil, nil
				}
			},
			expectedError: attachments.ErrNoteNotFound,
		},
		{
			name: "error - note repository error",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return nil, errors.New("database error")
				}
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "error - forbidden note access",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: uuid.New()}, nil
				}
			},
			expectedError: attachments.ErrForbidden,
		},
		{
			name: "error - block not found",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: userID}, nil
				}
				noteRepo.getBlockFunc = func(ctx context.Context, id uuid.UUID) (*models.Block, error) {
					return nil, nil
				}
			},
			expectedError: attachments.ErrBlockNotFound,
		},
		{
			name: "error - block repository error",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: userID}, nil
				}
				noteRepo.getBlockFunc = func(ctx context.Context, id uuid.UUID) (*models.Block, error) {
					return nil, errors.New("database error")
				}
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "error - block noteID mismatch",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: userID}, nil
				}
				noteRepo.getBlockFunc = func(ctx context.Context, id uuid.UUID) (*models.Block, error) {
					return &models.Block{ID: blockID, NoteID: uuid.New()}, nil
				}
			},
			expectedError: attachments.ErrForbidden,
		},
		{
			name: "error - delete attachment repository error",
			setupMock: func(attachmentRepo *mockAttachmentRepository, noteRepo *mockNoteRepository) {
				noteRepo.getNoteFunc = func(ctx context.Context, id uuid.UUID) (*models.Note, error) {
					return &models.Note{ID: noteID, UserID: userID}, nil
				}
				noteRepo.getBlockFunc = func(ctx context.Context, id uuid.UUID) (*models.Block, error) {
					return &models.Block{ID: blockID, NoteID: noteID}, nil
				}
				attachmentRepo.deleteAttachmentFunc = func(ctx context.Context, id uuid.UUID) error {
					return errors.New("delete error")
				}
			},
			expectedError: errors.New("delete error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAttachmentRepo := &mockAttachmentRepository{}
			mockNoteRepo := &mockNoteRepository{}
			tt.setupMock(mockAttachmentRepo, mockNoteRepo)

			usecase := NewAttachmentUsecase(mockAttachmentRepo, mockNoteRepo)
			err := usecase.DeleteAttachment(context.Background(), noteID, blockID, userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

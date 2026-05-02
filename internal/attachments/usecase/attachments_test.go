package usecase

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAttachmentUsecase_GetAttachment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAttachmentRepo := mocks.NewMockAttachmentRepository(ctrl)
	mockNoteRepo := mocks.NewMockNoteRepository(ctrl)
	usecase := NewAttachmentUsecase(mockAttachmentRepo, mockNoteRepo)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func() {
				mockNoteRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(&models.Note{ID: noteID, UserID: userID}, nil)
				mockNoteRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(&models.Block{ID: blockID, NoteID: noteID}, nil)
				mockAttachmentRepo.EXPECT().GetAttachment(gomock.Any(), blockID).Return(&models.Attachment{ID: uuid.New(), BlockID: blockID}, nil)
			},
			wantErr: nil,
		},
		{
			name: "note not found",
			setupMock: func() {
				mockNoteRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(nil, nil)
			},
			wantErr: attachments.ErrNoteNotFound,
		},
		{
			name: "forbidden - note belongs to different user",
			setupMock: func() {
				mockNoteRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(&models.Note{ID: noteID, UserID: uuid.New()}, nil)
			},
			wantErr: attachments.ErrForbidden,
		},
		{
			name: "block not found",
			setupMock: func() {
				mockNoteRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(&models.Note{ID: noteID, UserID: userID}, nil)
				mockNoteRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(nil, nil)
			},
			wantErr: attachments.ErrBlockNotFound,
		},
		{
			name: "block does not belong to note",
			setupMock: func() {
				mockNoteRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(&models.Note{ID: noteID, UserID: userID}, nil)
				mockNoteRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(&models.Block{ID: blockID, NoteID: uuid.New()}, nil)
			},
			wantErr: attachments.ErrForbidden,
		},
		{
			name: "attachment not found",
			setupMock: func() {
				mockNoteRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(&models.Note{ID: noteID, UserID: userID}, nil)
				mockNoteRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(&models.Block{ID: blockID, NoteID: noteID}, nil)
				mockAttachmentRepo.EXPECT().GetAttachment(gomock.Any(), blockID).Return(nil, nil)
			},
			wantErr: attachments.ErrAttachmentNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			attachment, err := usecase.GetAttachment(context.Background(), noteID, blockID, userID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, attachment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, attachment)
			}
		})
	}
}

func TestAttachmentUsecase_UploadAttachment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAttachmentRepo := mocks.NewMockAttachmentRepository(ctrl)
	mockNoteRepo := mocks.NewMockNoteRepository(ctrl)
	usecase := NewAttachmentUsecase(mockAttachmentRepo, mockNoteRepo)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	fileName := "test.png"
	fileSize := int64(1024)
	mimeType := "image/png"

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func() {
				mockNoteRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(&models.Note{ID: noteID, UserID: userID}, nil)
				mockNoteRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(&models.Block{ID: blockID, NoteID: noteID}, nil)
				mockAttachmentRepo.EXPECT().UploadAttachment(gomock.Any(), blockID, fileName, fileSize, mimeType, gomock.Any()).Return(&models.Attachment{ID: uuid.New(), BlockID: blockID}, nil)
			},
			wantErr: nil,
		},
		{
			name: "note not found",
			setupMock: func() {
				mockNoteRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(nil, nil)
			},
			wantErr: attachments.ErrNoteNotFound,
		},
		{
			name: "forbidden - note belongs to different user",
			setupMock: func() {
				mockNoteRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(&models.Note{ID: noteID, UserID: uuid.New()}, nil)
			},
			wantErr: attachments.ErrForbidden,
		},
		{
			name: "block not found",
			setupMock: func() {
				mockNoteRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(&models.Note{ID: noteID, UserID: userID}, nil)
				mockNoteRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(nil, nil)
			},
			wantErr: attachments.ErrBlockNotFound,
		},
		{
			name: "block already has attachment",
			setupMock: func() {
				mockNoteRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(&models.Note{ID: noteID, UserID: userID}, nil)
				mockNoteRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(&models.Block{ID: blockID, NoteID: noteID}, nil)
				mockAttachmentRepo.EXPECT().UploadAttachment(gomock.Any(), blockID, fileName, fileSize, mimeType, gomock.Any()).Return(nil, attachments.ErrBlockAlreadyHasAttach)
			},
			wantErr: attachments.ErrBlockAlreadyHasAttach,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			reader := bytes.NewReader([]byte("test content"))

			attachment, err := usecase.UploadAttachment(
				context.Background(),
				noteID,
				blockID,
				userID,
				reader,
				fileName,
				fileSize,
				mimeType,
			)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, attachment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, attachment)
			}
		})
	}
}

func TestAttachmentUsecase_DeleteAttachment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAttachmentRepo := mocks.NewMockAttachmentRepository(ctrl)
	mockNoteRepo := mocks.NewMockNoteRepository(ctrl)
	usecase := NewAttachmentUsecase(mockAttachmentRepo, mockNoteRepo)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func() {
				mockNoteRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(&models.Note{ID: noteID, UserID: userID}, nil)
				mockNoteRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(&models.Block{ID: blockID, NoteID: noteID}, nil)
				mockAttachmentRepo.EXPECT().DeleteAttachment(gomock.Any(), blockID).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "note not found",
			setupMock: func() {
				mockNoteRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(nil, nil)
			},
			wantErr: attachments.ErrNoteNotFound,
		},
		{
			name: "forbidden - note belongs to different user",
			setupMock: func() {
				mockNoteRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(&models.Note{ID: noteID, UserID: uuid.New()}, nil)
			},
			wantErr: attachments.ErrForbidden,
		},
		{
			name: "attachment not found on delete",
			setupMock: func() {
				mockNoteRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(&models.Note{ID: noteID, UserID: userID}, nil)
				mockNoteRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(&models.Block{ID: blockID, NoteID: noteID}, nil)
				mockAttachmentRepo.EXPECT().DeleteAttachment(gomock.Any(), blockID).Return(attachments.ErrAttachmentNotFound)
			},
			wantErr: attachments.ErrAttachmentNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := usecase.DeleteAttachment(context.Background(), noteID, blockID, userID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

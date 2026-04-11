package usecase

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
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
	attachmentID := uuid.New()

	t.Run("success", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}
		block := &models.Block{
			ID:     blockID,
			NoteID: noteID,
		}
		attachment := &models.Attachment{
			ID:        attachmentID,
			BlockID:   blockID,
			MinioKey:  "test-key",
			AttachURL: "https://example.com/file",
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)
		mockAttachmentRepo.EXPECT().
			GetAttachment(gomock.Any(), blockID).
			Return(attachment, nil)

		result, err := usecase.GetAttachment(context.Background(), noteID, blockID, userID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if result.ID != attachmentID {
			t.Errorf("expected ID %v, got %v", attachmentID, result.ID)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(nil, nil)

		_, err := usecase.GetAttachment(context.Background(), noteID, blockID, userID)

		if !errors.Is(err, attachments.ErrNoteNotFound) {
			t.Errorf("expected ErrNoteNotFound, got %v", err)
		}
	})

	t.Run("note repo error", func(t *testing.T) {
		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(nil, errors.New("db error"))

		_, err := usecase.GetAttachment(context.Background(), noteID, blockID, userID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("forbidden - wrong user", func(t *testing.T) {
		otherUserID := uuid.New()
		note := &models.Note{
			ID:     noteID,
			UserID: otherUserID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)

		_, err := usecase.GetAttachment(context.Background(), noteID, blockID, userID)

		if !errors.Is(err, attachments.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("block not found", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(nil, nil)

		_, err := usecase.GetAttachment(context.Background(), noteID, blockID, userID)

		if !errors.Is(err, attachments.ErrBlockNotFound) {
			t.Errorf("expected ErrBlockNotFound, got %v", err)
		}
	})

	t.Run("block repo error", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(nil, errors.New("db error"))

		_, err := usecase.GetAttachment(context.Background(), noteID, blockID, userID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("block belongs to different note", func(t *testing.T) {
		otherNoteID := uuid.New()
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}
		block := &models.Block{
			ID:     blockID,
			NoteID: otherNoteID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)

		_, err := usecase.GetAttachment(context.Background(), noteID, blockID, userID)

		if !errors.Is(err, attachments.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("attachment not found", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}
		block := &models.Block{
			ID:     blockID,
			NoteID: noteID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)
		mockAttachmentRepo.EXPECT().
			GetAttachment(gomock.Any(), blockID).
			Return(nil, nil)

		_, err := usecase.GetAttachment(context.Background(), noteID, blockID, userID)

		if !errors.Is(err, attachments.ErrAttachmentNotFound) {
			t.Errorf("expected ErrAttachmentNotFound, got %v", err)
		}
	})

	t.Run("attachment repo error", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}
		block := &models.Block{
			ID:     blockID,
			NoteID: noteID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)
		mockAttachmentRepo.EXPECT().
			GetAttachment(gomock.Any(), blockID).
			Return(nil, errors.New("db error"))

		_, err := usecase.GetAttachment(context.Background(), noteID, blockID, userID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
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
	attachmentID := uuid.New()
	fileName := "test.jpg"
	fileSize := int64(1024)
	mimeType := "image/jpeg"
	fileReader := bytes.NewReader([]byte{0xFF, 0xD8, 0xFF, 0xE0})

	t.Run("success", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}
		block := &models.Block{
			ID:     blockID,
			NoteID: noteID,
		}
		attachment := &models.Attachment{
			ID:        attachmentID,
			BlockID:   blockID,
			MinioKey:  "test-key",
			AttachURL: "https://example.com/file",
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)
		mockAttachmentRepo.EXPECT().
			UploadAttachment(gomock.Any(), blockID, fileName, fileSize, mimeType, gomock.Any()).
			Return(attachment, nil)

		result, err := usecase.UploadAttachment(context.Background(), noteID, blockID, userID, fileName, fileSize, mimeType, fileReader)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if result.ID != attachmentID {
			t.Errorf("expected ID %v, got %v", attachmentID, result.ID)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(nil, nil)

		_, err := usecase.UploadAttachment(context.Background(), noteID, blockID, userID, fileName, fileSize, mimeType, fileReader)

		if !errors.Is(err, attachments.ErrNoteNotFound) {
			t.Errorf("expected ErrNoteNotFound, got %v", err)
		}
	})

	t.Run("note repo error", func(t *testing.T) {
		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(nil, errors.New("db error"))

		_, err := usecase.UploadAttachment(context.Background(), noteID, blockID, userID, fileName, fileSize, mimeType, fileReader)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("forbidden - wrong user", func(t *testing.T) {
		otherUserID := uuid.New()
		note := &models.Note{
			ID:     noteID,
			UserID: otherUserID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)

		_, err := usecase.UploadAttachment(context.Background(), noteID, blockID, userID, fileName, fileSize, mimeType, fileReader)

		if !errors.Is(err, attachments.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("block not found", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(nil, nil)

		_, err := usecase.UploadAttachment(context.Background(), noteID, blockID, userID, fileName, fileSize, mimeType, fileReader)

		if !errors.Is(err, attachments.ErrBlockNotFound) {
			t.Errorf("expected ErrBlockNotFound, got %v", err)
		}
	})

	t.Run("block repo error", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(nil, errors.New("db error"))

		_, err := usecase.UploadAttachment(context.Background(), noteID, blockID, userID, fileName, fileSize, mimeType, fileReader)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("block belongs to different note", func(t *testing.T) {
		otherNoteID := uuid.New()
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}
		block := &models.Block{
			ID:     blockID,
			NoteID: otherNoteID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)

		_, err := usecase.UploadAttachment(context.Background(), noteID, blockID, userID, fileName, fileSize, mimeType, fileReader)

		if !errors.Is(err, attachments.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("attachment repo error", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}
		block := &models.Block{
			ID:     blockID,
			NoteID: noteID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)
		mockAttachmentRepo.EXPECT().
			UploadAttachment(gomock.Any(), blockID, fileName, fileSize, mimeType, gomock.Any()).
			Return(nil, errors.New("upload error"))

		_, err := usecase.UploadAttachment(context.Background(), noteID, blockID, userID, fileName, fileSize, mimeType, fileReader)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
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

	t.Run("success", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}
		block := &models.Block{
			ID:     blockID,
			NoteID: noteID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)
		mockAttachmentRepo.EXPECT().
			DeleteAttachment(gomock.Any(), blockID).
			Return(nil)

		err := usecase.DeleteAttachment(context.Background(), noteID, blockID, userID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(nil, nil)

		err := usecase.DeleteAttachment(context.Background(), noteID, blockID, userID)

		if !errors.Is(err, attachments.ErrNoteNotFound) {
			t.Errorf("expected ErrNoteNotFound, got %v", err)
		}
	})

	t.Run("note repo error", func(t *testing.T) {
		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(nil, errors.New("db error"))

		err := usecase.DeleteAttachment(context.Background(), noteID, blockID, userID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("forbidden - wrong user", func(t *testing.T) {
		otherUserID := uuid.New()
		note := &models.Note{
			ID:     noteID,
			UserID: otherUserID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)

		err := usecase.DeleteAttachment(context.Background(), noteID, blockID, userID)

		if !errors.Is(err, attachments.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("block not found", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(nil, nil)

		err := usecase.DeleteAttachment(context.Background(), noteID, blockID, userID)

		if !errors.Is(err, attachments.ErrBlockNotFound) {
			t.Errorf("expected ErrBlockNotFound, got %v", err)
		}
	})

	t.Run("block repo error", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(nil, errors.New("db error"))

		err := usecase.DeleteAttachment(context.Background(), noteID, blockID, userID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("block belongs to different note", func(t *testing.T) {
		otherNoteID := uuid.New()
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}
		block := &models.Block{
			ID:     blockID,
			NoteID: otherNoteID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)

		err := usecase.DeleteAttachment(context.Background(), noteID, blockID, userID)

		if !errors.Is(err, attachments.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("attachment not found", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}
		block := &models.Block{
			ID:     blockID,
			NoteID: noteID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)
		mockAttachmentRepo.EXPECT().
			DeleteAttachment(gomock.Any(), blockID).
			Return(attachments.ErrAttachmentNotFound)

		err := usecase.DeleteAttachment(context.Background(), noteID, blockID, userID)

		if !errors.Is(err, attachments.ErrAttachmentNotFound) {
			t.Errorf("expected ErrAttachmentNotFound, got %v", err)
		}
	})

	t.Run("attachment repo error", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}
		block := &models.Block{
			ID:     blockID,
			NoteID: noteID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)
		mockAttachmentRepo.EXPECT().
			DeleteAttachment(gomock.Any(), blockID).
			Return(errors.New("db error"))

		err := usecase.DeleteAttachment(context.Background(), noteID, blockID, userID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestCheckNoteAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAttachmentRepo := mocks.NewMockAttachmentRepository(ctrl)
	mockNoteRepo := mocks.NewMockNoteRepository(ctrl)
	usecase := NewAttachmentUsecase(mockAttachmentRepo, mockNoteRepo)

	userID := uuid.New()
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)

		result, err := usecase.checkNoteAccess(context.Background(), noteID, userID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if result.ID != noteID {
			t.Errorf("expected ID %v, got %v", noteID, result.ID)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(nil, nil)

		_, err := usecase.checkNoteAccess(context.Background(), noteID, userID)

		if !errors.Is(err, attachments.ErrNoteNotFound) {
			t.Errorf("expected ErrNoteNotFound, got %v", err)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		otherUserID := uuid.New()
		note := &models.Note{
			ID:     noteID,
			UserID: otherUserID,
		}

		mockNoteRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)

		_, err := usecase.checkNoteAccess(context.Background(), noteID, userID)

		if !errors.Is(err, attachments.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})
}

func TestCheckBlockAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAttachmentRepo := mocks.NewMockAttachmentRepository(ctrl)
	mockNoteRepo := mocks.NewMockNoteRepository(ctrl)
	usecase := NewAttachmentUsecase(mockAttachmentRepo, mockNoteRepo)

	noteID := uuid.New()
	blockID := uuid.New()

	t.Run("success", func(t *testing.T) {
		block := &models.Block{
			ID:     blockID,
			NoteID: noteID,
		}

		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)

		result, err := usecase.checkBlockAccess(context.Background(), noteID, blockID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if result.ID != blockID {
			t.Errorf("expected ID %v, got %v", blockID, result.ID)
		}
	})

	t.Run("block not found", func(t *testing.T) {
		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(nil, nil)

		_, err := usecase.checkBlockAccess(context.Background(), noteID, blockID)

		if !errors.Is(err, attachments.ErrBlockNotFound) {
			t.Errorf("expected ErrBlockNotFound, got %v", err)
		}
	})

	t.Run("forbidden - wrong note", func(t *testing.T) {
		otherNoteID := uuid.New()
		block := &models.Block{
			ID:     blockID,
			NoteID: otherNoteID,
		}

		mockNoteRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)

		_, err := usecase.checkBlockAccess(context.Background(), noteID, blockID)

		if !errors.Is(err, attachments.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})
}

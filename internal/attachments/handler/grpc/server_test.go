package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/handler/grpc/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	attachmentsgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/attachments/grpc/gen"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServer_GetAttachment(t *testing.T) {
	noteID := uuid.New()
	blockID := uuid.New()
	userID := uuid.New()

	t.Run("invalid block id", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		_, err := s.GetAttachment(context.Background(), &attachmentsgrpc.GetAttachmentRequest{
			BlockId: "not-uuid",
			NoteId:  noteID.String(),
			UserId:  userID.String(),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("invalid note id", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		_, err := s.GetAttachment(context.Background(), &attachmentsgrpc.GetAttachmentRequest{
			BlockId: blockID.String(),
			NoteId:  "not-uuid",
			UserId:  userID.String(),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("invalid user id", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		_, err := s.GetAttachment(context.Background(), &attachmentsgrpc.GetAttachmentRequest{
			BlockId: blockID.String(),
			NoteId:  noteID.String(),
			UserId:  "not-uuid",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("usecase error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		uc.EXPECT().GetAttachment(gomock.Any(), noteID, blockID, userID).Return(nil, errors.New("boom"))

		resp, err := s.GetAttachment(context.Background(), &attachmentsgrpc.GetAttachmentRequest{
			BlockId: blockID.String(),
			NoteId:  noteID.String(),
			UserId:  userID.String(),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if resp != nil {
			t.Fatalf("expected nil resp, got %v", resp)
		}
	})

	t.Run("not found when nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		uc.EXPECT().GetAttachment(gomock.Any(), noteID, blockID, userID).Return(nil, nil)

		_, err := s.GetAttachment(context.Background(), &attachmentsgrpc.GetAttachmentRequest{
			BlockId: blockID.String(),
			NoteId:  noteID.String(),
			UserId:  userID.String(),
		})
		if status.Code(err) != codes.NotFound {
			t.Fatalf("expected codes.NotFound, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		attachmentID := uuid.New()
		now := time.Now()
		expected := &models.Attachment{
			ID:           attachmentID,
			BlockID:      blockID,
			MinioKey:     "key",
			AttachURL:    "https://example.com/file",
			URLExpiresAt: now.Add(time.Hour),
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		uc.EXPECT().GetAttachment(gomock.Any(), noteID, blockID, userID).Return(expected, nil)

		resp, err := s.GetAttachment(context.Background(), &attachmentsgrpc.GetAttachmentRequest{
			BlockId: blockID.String(),
			NoteId:  noteID.String(),
			UserId:  userID.String(),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Id != attachmentID.String() {
			t.Errorf("expected id %s, got %s", attachmentID, resp.Id)
		}
		if resp.BlockId != blockID.String() {
			t.Errorf("expected block id %s, got %s", blockID, resp.BlockId)
		}
		if resp.AttachUrl != "https://example.com/file" {
			t.Errorf("expected URL, got %s", resp.AttachUrl)
		}
	})
}

func TestServer_DeleteAttachment(t *testing.T) {
	noteID := uuid.New()
	blockID := uuid.New()
	userID := uuid.New()

	t.Run("invalid block id", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		_, err := s.DeleteAttachment(context.Background(), &attachmentsgrpc.DeleteAttachmentRequest{
			BlockId: "not-uuid",
			NoteId:  noteID.String(),
			UserId:  userID.String(),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("invalid note id", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		_, err := s.DeleteAttachment(context.Background(), &attachmentsgrpc.DeleteAttachmentRequest{
			BlockId: blockID.String(),
			NoteId:  "not-uuid",
			UserId:  userID.String(),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("invalid user id", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		_, err := s.DeleteAttachment(context.Background(), &attachmentsgrpc.DeleteAttachmentRequest{
			BlockId: blockID.String(),
			NoteId:  noteID.String(),
			UserId:  "not-uuid",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("usecase error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		uc.EXPECT().DeleteAttachment(gomock.Any(), noteID, blockID, userID).Return(errors.New("boom"))

		_, err := s.DeleteAttachment(context.Background(), &attachmentsgrpc.DeleteAttachmentRequest{
			BlockId: blockID.String(),
			NoteId:  noteID.String(),
			UserId:  userID.String(),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		uc.EXPECT().DeleteAttachment(gomock.Any(), noteID, blockID, userID).Return(nil)

		resp, err := s.DeleteAttachment(context.Background(), &attachmentsgrpc.DeleteAttachmentRequest{
			BlockId: blockID.String(),
			NoteId:  noteID.String(),
			UserId:  userID.String(),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
	})
}

func TestServer_GetHeader(t *testing.T) {
	noteID := uuid.New()
	userID := uuid.New()

	t.Run("invalid note id", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		_, err := s.GetHeader(context.Background(), &attachmentsgrpc.GetHeaderRequest{
			NoteId: "not-uuid",
			UserId: userID.String(),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("invalid user id", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		_, err := s.GetHeader(context.Background(), &attachmentsgrpc.GetHeaderRequest{
			NoteId: noteID.String(),
			UserId: "not-uuid",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("header not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		uc.EXPECT().GetHeader(gomock.Any(), noteID, userID).Return(nil, attachments.ErrHeaderNotFound)

		_, err := s.GetHeader(context.Background(), &attachmentsgrpc.GetHeaderRequest{
			NoteId: noteID.String(),
			UserId: userID.String(),
		})
		if status.Code(err) != codes.NotFound {
			t.Fatalf("expected codes.NotFound, got %v", err)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		uc.EXPECT().GetHeader(gomock.Any(), noteID, userID).Return(nil, errors.New("boom"))

		_, err := s.GetHeader(context.Background(), &attachmentsgrpc.GetHeaderRequest{
			NoteId: noteID.String(),
			UserId: userID.String(),
		})
		if status.Code(err) != codes.Internal {
			t.Fatalf("expected codes.Internal, got %v", err)
		}
	})

	t.Run("nil header returns not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		uc.EXPECT().GetHeader(gomock.Any(), noteID, userID).Return(nil, nil)

		_, err := s.GetHeader(context.Background(), &attachmentsgrpc.GetHeaderRequest{
			NoteId: noteID.String(),
			UserId: userID.String(),
		})
		if status.Code(err) != codes.NotFound {
			t.Fatalf("expected codes.NotFound, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		headerID := uuid.New()
		now := time.Now()
		expected := &models.Header{
			ID:           headerID,
			NoteID:       noteID,
			MinioKey:     "key",
			HeaderURL:    "https://example.com/header",
			URLExpiresAt: now.Add(time.Hour),
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		uc.EXPECT().GetHeader(gomock.Any(), noteID, userID).Return(expected, nil)

		resp, err := s.GetHeader(context.Background(), &attachmentsgrpc.GetHeaderRequest{
			NoteId: noteID.String(),
			UserId: userID.String(),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Id != headerID.String() {
			t.Errorf("expected id %s, got %s", headerID, resp.Id)
		}
		if resp.HeaderUrl != "https://example.com/header" {
			t.Errorf("expected URL, got %s", resp.HeaderUrl)
		}
	})
}

func TestServer_DeleteHeader(t *testing.T) {
	noteID := uuid.New()
	userID := uuid.New()

	t.Run("invalid note id", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		_, err := s.DeleteHeader(context.Background(), &attachmentsgrpc.DeleteHeaderRequest{
			NoteId: "not-uuid",
			UserId: userID.String(),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("invalid user id", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		_, err := s.DeleteHeader(context.Background(), &attachmentsgrpc.DeleteHeaderRequest{
			NoteId: noteID.String(),
			UserId: "not-uuid",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("header not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		uc.EXPECT().DeleteHeader(gomock.Any(), noteID, userID).Return(attachments.ErrHeaderNotFound)

		_, err := s.DeleteHeader(context.Background(), &attachmentsgrpc.DeleteHeaderRequest{
			NoteId: noteID.String(),
			UserId: userID.String(),
		})
		if status.Code(err) != codes.NotFound {
			t.Fatalf("expected codes.NotFound, got %v", err)
		}
	})

	t.Run("other error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		uc.EXPECT().DeleteHeader(gomock.Any(), noteID, userID).Return(errors.New("boom"))

		_, err := s.DeleteHeader(context.Background(), &attachmentsgrpc.DeleteHeaderRequest{
			NoteId: noteID.String(),
			UserId: userID.String(),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		s := NewServer(uc)

		uc.EXPECT().DeleteHeader(gomock.Any(), noteID, userID).Return(nil)

		resp, err := s.DeleteHeader(context.Background(), &attachmentsgrpc.DeleteHeaderRequest{
			NoteId: noteID.String(),
			UserId: userID.String(),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
	})
}

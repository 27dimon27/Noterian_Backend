package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	notesgen "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/notes/grpc/gen"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

type fakeNotesClient struct {
	GetBlocksFunc           func(ctx context.Context, noteID, userID uuid.UUID) ([]*notesgen.BlockResponse, error)
	ShiftBlockPositionsFunc func(ctx context.Context, noteID uuid.UUID, fromPosition, direction int) error
	CreateBlockFunc         func(ctx context.Context, userID uuid.UUID, block *notesgen.BlockResponse) (*notesgen.BlockResponse, error)
	GetBlockFunc            func(ctx context.Context, blockID, noteID, userID uuid.UUID) (*notesgen.BlockResponse, error)
	DeleteBlockFunc         func(ctx context.Context, blockID, noteID, userID uuid.UUID) (uuid.UUID, error)
	CloseFunc               func() error
}

func (f *fakeNotesClient) GetNote(ctx context.Context, noteID, userID uuid.UUID) (*notesgen.NoteResponse, error) {
	return nil, nil
}

func (f *fakeNotesClient) GetBlock(ctx context.Context, blockID, noteID, userID uuid.UUID) (*notesgen.BlockResponse, error) {
	if f.GetBlockFunc != nil {
		return f.GetBlockFunc(ctx, blockID, noteID, userID)
	}
	return nil, nil
}

func (f *fakeNotesClient) GetBlocks(ctx context.Context, noteID, userID uuid.UUID) ([]*notesgen.BlockResponse, error) {
	if f.GetBlocksFunc != nil {
		return f.GetBlocksFunc(ctx, noteID, userID)
	}
	return nil, nil
}

func (f *fakeNotesClient) CreateBlock(ctx context.Context, userID uuid.UUID, block *notesgen.BlockResponse) (*notesgen.BlockResponse, error) {
	if f.CreateBlockFunc != nil {
		return f.CreateBlockFunc(ctx, userID, block)
	}
	return nil, nil
}

func (f *fakeNotesClient) ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition, direction int) error {
	if f.ShiftBlockPositionsFunc != nil {
		return f.ShiftBlockPositionsFunc(ctx, noteID, fromPosition, direction)
	}
	return nil
}

func (f *fakeNotesClient) DeleteBlock(ctx context.Context, blockID, noteID, userID uuid.UUID) (uuid.UUID, error) {
	if f.DeleteBlockFunc != nil {
		return f.DeleteBlockFunc(ctx, blockID, noteID, userID)
	}
	return uuid.Nil, nil
}

func (f *fakeNotesClient) Close() error {
	if f.CloseFunc != nil {
		return f.CloseFunc()
	}
	return nil
}

func TestGetAttachment_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	usecase := NewAttachmentUsecase(repoMock, &fakeNotesClient{})

	noteID := uuid.New()
	blockID := uuid.New()
	expected := &models.Attachment{ID: uuid.New(), BlockID: blockID}

	repoMock.EXPECT().GetAttachment(gomock.Any(), blockID).Return(expected, nil)

	attachment, err := usecase.GetAttachment(context.Background(), noteID, blockID, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attachment != expected {
		t.Fatalf("expected %v, got %v", expected, attachment)
	}
}

func TestGetAttachment_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	usecase := NewAttachmentUsecase(repoMock, &fakeNotesClient{})

	blockID := uuid.New()

	repoMock.EXPECT().GetAttachment(gomock.Any(), blockID).Return(nil, nil)

	attachment, err := usecase.GetAttachment(context.Background(), uuid.New(), blockID, uuid.New())
	if !errors.Is(err, attachments.ErrAttachmentNotFound) {
		t.Fatalf("expected ErrAttachmentNotFound, got %v", err)
	}
	if attachment != nil {
		t.Fatalf("expected nil attachment, got %v", attachment)
	}
}

func TestUploadAttachment_InvalidMimeType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	usecase := NewAttachmentUsecase(repoMock, &fakeNotesClient{})

	_, err := usecase.UploadAttachment(context.Background(), uuid.New(), uuid.New(), "file.bin", 10, "application/pdf", nil, false, 0)
	if !errors.Is(err, attachments.ErrInvalidMimeType) {
		t.Fatalf("expected ErrInvalidMimeType, got %v", err)
	}
}

func TestUploadAttachment_GetBlocksError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	notesClient := &fakeNotesClient{
		GetBlocksFunc: func(ctx context.Context, noteID, userID uuid.UUID) ([]*notesgen.BlockResponse, error) {
			return nil, errors.New("grpc error")
		},
	}
	usecase := NewAttachmentUsecase(repoMock, notesClient)

	_, err := usecase.UploadAttachment(context.Background(), uuid.New(), uuid.New(), "file.png", 100, "image/png", nil, false, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUploadAttachment_InvalidPosition(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	notesClient := &fakeNotesClient{
		GetBlocksFunc: func(ctx context.Context, noteID, userID uuid.UUID) ([]*notesgen.BlockResponse, error) {
			return []*notesgen.BlockResponse{{Id: uuid.NewString()}}, nil
		},
	}
	usecase := NewAttachmentUsecase(repoMock, notesClient)

	_, err := usecase.UploadAttachment(context.Background(), uuid.New(), uuid.New(), "file.png", 100, "image/png", nil, true, 2)
	if !errors.Is(err, attachments.ErrInvalidPosition) {
		t.Fatalf("expected ErrInvalidPosition, got %v", err)
	}
}

func TestUploadAttachment_CreateBlockRollback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	shiftCalls := 0
	createdBlockID := uuid.New()
	noteID := uuid.New()
	userID := uuid.New()
	notesClient := &fakeNotesClient{
		GetBlocksFunc: func(ctx context.Context, noteIDParam, userIDParam uuid.UUID) ([]*notesgen.BlockResponse, error) {
			return []*notesgen.BlockResponse{{Id: uuid.NewString(), Position: 0}}, nil
		},
		ShiftBlockPositionsFunc: func(ctx context.Context, noteIDParam uuid.UUID, fromPosition, direction int) error {
			if shiftCalls == 0 && (fromPosition != 1 || direction != 1) {
				t.Fatalf("expected first shift +1 at position 1, got %d %d", fromPosition, direction)
			}
			if shiftCalls == 1 && (fromPosition != 1 || direction != -1) {
				t.Fatalf("expected rollback shift -1 at position 1, got %d %d", fromPosition, direction)
			}
			shiftCalls++
			return nil
		},
		CreateBlockFunc: func(ctx context.Context, userIDParam uuid.UUID, block *notesgen.BlockResponse) (*notesgen.BlockResponse, error) {
			if block.NoteId == "" {
				t.Fatal("expected block NoteId to be set")
			}
			return &notesgen.BlockResponse{Id: createdBlockID.String()}, nil
		},
		DeleteBlockFunc: func(ctx context.Context, blockID, noteIDParam, userIDParam uuid.UUID) (uuid.UUID, error) {
			if blockID != createdBlockID {
				t.Fatalf("expected DeleteBlock with %v, got %v", createdBlockID, blockID)
			}
			return noteIDParam, nil
		},
	}
	usecase := NewAttachmentUsecase(repoMock, notesClient)

	repoMock.EXPECT().UploadAttachment(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("upload failed"))

	_, err := usecase.UploadAttachment(context.Background(), noteID, userID, "file.png", 100, "image/png", nil, false, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if shiftCalls != 2 {
		t.Fatalf("expected 2 shift calls, got %d", shiftCalls)
	}
}

func TestDeleteAttachment_BlockNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	notesClient := &fakeNotesClient{
		GetBlockFunc: func(ctx context.Context, blockID, noteID, userID uuid.UUID) (*notesgen.BlockResponse, error) {
			return nil, nil
		},
	}
	usecase := NewAttachmentUsecase(repoMock, notesClient)

	err := usecase.DeleteAttachment(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if !errors.Is(err, attachments.ErrBlockNotFound) {
		t.Fatalf("expected ErrBlockNotFound, got %v", err)
	}
}

func TestDeleteHeader_ForwardError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	repoMock.EXPECT().DeleteHeader(gomock.Any(), gomock.Any()).Return(errors.New("db failure"))
	usecase := NewAttachmentUsecase(repoMock, &fakeNotesClient{})

	err := usecase.DeleteHeader(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetHeader_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	expected := &models.Header{ID: uuid.New()}
	repoMock.EXPECT().GetHeader(gomock.Any(), gomock.Any()).Return(expected, nil)
	usecase := NewAttachmentUsecase(repoMock, &fakeNotesClient{})

	header, err := usecase.GetHeader(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if header != expected {
		t.Fatalf("expected %v, got %v", expected, header)
	}
}

func TestUploadHeader_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	expected := &models.Header{ID: uuid.New()}
	repoMock.EXPECT().UploadHeader(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(expected, nil)
	usecase := NewAttachmentUsecase(repoMock, &fakeNotesClient{})

	header, err := usecase.UploadHeader(context.Background(), uuid.New(), uuid.New(), "header.png", 100, "image/png", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if header != expected {
		t.Fatalf("expected %v, got %v", expected, header)
	}
}

func TestDeleteAttachment_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	notesClient := &fakeNotesClient{
		GetBlockFunc: func(ctx context.Context, blockID, noteID, userID uuid.UUID) (*notesgen.BlockResponse, error) {
			return &notesgen.BlockResponse{Id: blockID.String()}, nil
		},
	}
	repoMock.EXPECT().DeleteAttachment(gomock.Any(), gomock.Any()).Return(nil)
	usecase := NewAttachmentUsecase(repoMock, notesClient)

	err := usecase.DeleteAttachment(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUploadAttachment_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	createdBlockID := uuid.New()
	noteID := uuid.New()
	userID := uuid.New()
	shiftCalls := 0
	expectedAttachment := &models.Attachment{ID: uuid.New(), BlockID: createdBlockID}
	notesClient := &fakeNotesClient{
		GetBlocksFunc: func(ctx context.Context, noteIDParam, userIDParam uuid.UUID) ([]*notesgen.BlockResponse, error) {
			return []*notesgen.BlockResponse{}, nil
		},
		ShiftBlockPositionsFunc: func(ctx context.Context, noteIDParam uuid.UUID, fromPosition, direction int) error {
			if shiftCalls == 0 && (fromPosition != 0 || direction != 1) {
				t.Fatalf("expected shift +1 at position 0, got %d %d", fromPosition, direction)
			}
			shiftCalls++
			return nil
		},
		CreateBlockFunc: func(ctx context.Context, userIDParam uuid.UUID, block *notesgen.BlockResponse) (*notesgen.BlockResponse, error) {
			return &notesgen.BlockResponse{Id: createdBlockID.String()}, nil
		},
	}
	repoMock.EXPECT().UploadAttachment(gomock.Any(), createdBlockID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedAttachment, nil)
	usecase := NewAttachmentUsecase(repoMock, notesClient)

	attachment, err := usecase.UploadAttachment(context.Background(), noteID, userID, "file.png", 100, "image/png", nil, false, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attachment != expectedAttachment {
		t.Fatalf("expected %v, got %v", expectedAttachment, attachment)
	}
}

func TestUploadAttachment_SuccessWithPosition(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	createdBlockID := uuid.New()
	noteID := uuid.New()
	userID := uuid.New()
	shiftCalls := 0
	expectedAttachment := &models.Attachment{ID: uuid.New(), BlockID: createdBlockID}
	notesClient := &fakeNotesClient{
		GetBlocksFunc: func(ctx context.Context, noteIDParam, userIDParam uuid.UUID) ([]*notesgen.BlockResponse, error) {
			return []*notesgen.BlockResponse{{Id: uuid.NewString(), Position: 0}}, nil
		},
		ShiftBlockPositionsFunc: func(ctx context.Context, noteIDParam uuid.UUID, fromPosition, direction int) error {
			if shiftCalls == 0 && (fromPosition != 1 || direction != 1) {
				t.Fatalf("expected shift +1 at position 1, got %d %d", fromPosition, direction)
			}
			shiftCalls++
			return nil
		},
		CreateBlockFunc: func(ctx context.Context, userIDParam uuid.UUID, block *notesgen.BlockResponse) (*notesgen.BlockResponse, error) {
			return &notesgen.BlockResponse{Id: createdBlockID.String()}, nil
		},
	}
	repoMock.EXPECT().UploadAttachment(gomock.Any(), createdBlockID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedAttachment, nil)
	usecase := NewAttachmentUsecase(repoMock, notesClient)

	attachment, err := usecase.UploadAttachment(context.Background(), noteID, userID, "file.png", 100, "image/png", nil, true, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attachment != expectedAttachment {
		t.Fatalf("expected %v, got %v", expectedAttachment, attachment)
	}
}

func TestDeleteAttachment_GetBlockError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	notesClient := &fakeNotesClient{
		GetBlockFunc: func(ctx context.Context, blockID, noteID, userID uuid.UUID) (*notesgen.BlockResponse, error) {
			return nil, errors.New("grpc fail")
		},
	}
	usecase := NewAttachmentUsecase(repoMock, notesClient)

	err := usecase.DeleteAttachment(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetHeader_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	repoMock.EXPECT().GetHeader(gomock.Any(), gomock.Any()).Return(nil, nil)
	usecase := NewAttachmentUsecase(repoMock, &fakeNotesClient{})

	header, err := usecase.GetHeader(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, attachments.ErrHeaderNotFound) {
		t.Fatalf("expected ErrHeaderNotFound, got %v", err)
	}
	if header != nil {
		t.Fatalf("expected nil header, got %v", header)
	}
}

func TestUploadHeader_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	repoMock.EXPECT().UploadHeader(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("upload error"))
	usecase := NewAttachmentUsecase(repoMock, &fakeNotesClient{})

	header, err := usecase.UploadHeader(context.Background(), uuid.New(), uuid.New(), "header.png", 100, "image/png", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if header != nil {
		t.Fatalf("expected nil header, got %v", header)
	}
}

func TestDeleteHeader_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	repoMock.EXPECT().DeleteHeader(gomock.Any(), gomock.Any()).Return(nil)
	usecase := NewAttachmentUsecase(repoMock, &fakeNotesClient{})

	err := usecase.DeleteHeader(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetBlockTypeByMimeType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	usecase := NewAttachmentUsecase(repoMock, &fakeNotesClient{})

	if got, err := usecase.getBlockTypeByMimeType("video/mp4"); got != 7 || err != nil {
		t.Fatalf("expected video block type 7, got %d, %v", got, err)
	}
	if got, err := usecase.getBlockTypeByMimeType("audio/mpeg"); got != 6 || err != nil {
		t.Fatalf("expected audio block type 6, got %d, %v", got, err)
	}
	if got, err := usecase.getBlockTypeByMimeType("image/png"); got != 2 || err != nil {
		t.Fatalf("expected image block type 2, got %d, %v", got, err)
	}
	if got, err := usecase.getBlockTypeByMimeType("image/gif"); got != 2 || err != nil {
		t.Fatalf("expected gif block type 2, got %d, %v", got, err)
	}
	if _, err := usecase.getBlockTypeByMimeType("application/pdf"); !errors.Is(err, attachments.ErrInvalidMimeType) {
		t.Fatalf("expected ErrInvalidMimeType, got %v", err)
	}
}

func TestGetAttachment_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	usecase := NewAttachmentUsecase(repoMock, &fakeNotesClient{})

	blockID := uuid.New()
	repoMock.EXPECT().GetAttachment(gomock.Any(), blockID).Return(nil, errors.New("db error"))

	attachment, err := usecase.GetAttachment(context.Background(), uuid.New(), blockID, uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if attachment != nil {
		t.Fatalf("expected nil attachment, got %v", attachment)
	}
}

func TestGetHeader_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	repoMock.EXPECT().GetHeader(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
	usecase := NewAttachmentUsecase(repoMock, &fakeNotesClient{})

	header, err := usecase.GetHeader(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if header != nil {
		t.Fatalf("expected nil header, got %v", header)
	}
}

func TestUploadAttachment_NegativePosition(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	notesClient := &fakeNotesClient{
		GetBlocksFunc: func(ctx context.Context, noteID, userID uuid.UUID) ([]*notesgen.BlockResponse, error) {
			return []*notesgen.BlockResponse{}, nil
		},
	}
	usecase := NewAttachmentUsecase(repoMock, notesClient)

	_, err := usecase.UploadAttachment(context.Background(), uuid.New(), uuid.New(), "f.png", 1, "image/png", nil, true, -1)
	if !errors.Is(err, attachments.ErrInvalidPosition) {
		t.Fatalf("expected ErrInvalidPosition, got %v", err)
	}
}

func TestUploadAttachment_ShiftError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	notesClient := &fakeNotesClient{
		GetBlocksFunc: func(ctx context.Context, noteID, userID uuid.UUID) ([]*notesgen.BlockResponse, error) {
			return []*notesgen.BlockResponse{}, nil
		},
		ShiftBlockPositionsFunc: func(ctx context.Context, noteID uuid.UUID, fromPosition, direction int) error {
			return errors.New("shift fail")
		},
	}
	usecase := NewAttachmentUsecase(repoMock, notesClient)

	_, err := usecase.UploadAttachment(context.Background(), uuid.New(), uuid.New(), "f.png", 1, "image/png", nil, false, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUploadAttachment_CreateBlockError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	shiftCalls := 0
	notesClient := &fakeNotesClient{
		GetBlocksFunc: func(ctx context.Context, noteID, userID uuid.UUID) ([]*notesgen.BlockResponse, error) {
			return []*notesgen.BlockResponse{}, nil
		},
		ShiftBlockPositionsFunc: func(ctx context.Context, noteID uuid.UUID, fromPosition, direction int) error {
			shiftCalls++
			return nil
		},
		CreateBlockFunc: func(ctx context.Context, userID uuid.UUID, block *notesgen.BlockResponse) (*notesgen.BlockResponse, error) {
			return nil, errors.New("create fail")
		},
	}
	usecase := NewAttachmentUsecase(repoMock, notesClient)

	_, err := usecase.UploadAttachment(context.Background(), uuid.New(), uuid.New(), "f.png", 1, "image/png", nil, false, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if shiftCalls != 2 {
		t.Fatalf("expected rollback shift, total shift calls = 2, got %d", shiftCalls)
	}
}

func TestUploadAttachment_CreateBlockReturnsInvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	shiftCalls := 0
	notesClient := &fakeNotesClient{
		GetBlocksFunc: func(ctx context.Context, noteID, userID uuid.UUID) ([]*notesgen.BlockResponse, error) {
			return []*notesgen.BlockResponse{}, nil
		},
		ShiftBlockPositionsFunc: func(ctx context.Context, noteID uuid.UUID, fromPosition, direction int) error {
			shiftCalls++
			return nil
		},
		CreateBlockFunc: func(ctx context.Context, userID uuid.UUID, block *notesgen.BlockResponse) (*notesgen.BlockResponse, error) {
			return &notesgen.BlockResponse{Id: "not-a-uuid"}, nil
		},
	}
	usecase := NewAttachmentUsecase(repoMock, notesClient)

	_, err := usecase.UploadAttachment(context.Background(), uuid.New(), uuid.New(), "f.png", 1, "image/png", nil, false, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if shiftCalls != 2 {
		t.Fatalf("expected rollback shift, total shift calls = 2, got %d", shiftCalls)
	}
}

func TestDeleteAttachment_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	notesClient := &fakeNotesClient{
		GetBlockFunc: func(ctx context.Context, blockID, noteID, userID uuid.UUID) (*notesgen.BlockResponse, error) {
			return &notesgen.BlockResponse{Id: blockID.String()}, nil
		},
	}
	repoMock.EXPECT().DeleteAttachment(gomock.Any(), gomock.Any()).Return(errors.New("repo fail"))
	usecase := NewAttachmentUsecase(repoMock, notesClient)

	err := usecase.DeleteAttachment(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	notesMocks "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/grpcclient/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	notesgen "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/notes/grpc/gen"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTestUsecase(t *testing.T) (*attachmentUsecase, *mocks.MockAttachmentRepository, *notesMocks.MockNotesServiceClient) {
	t.Helper()

	ctrl := gomock.NewController(t)
	repoMock := mocks.NewMockAttachmentRepository(ctrl)
	notesMock := notesMocks.NewMockNotesServiceClient(ctrl)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	u := NewAttachmentUsecase(repoMock, notesMock, logger)
	return u, repoMock, notesMock
}

func appErrWith(err error, publicMsg string, code int) types.AppErrorInterface {
	return &types.AppError{
		Err:        err,
		PublicMsg:  publicMsg,
		StatusCode: code,
		Layer:      "repo",
		Op:         "test",
	}
}

func pickAllowedImageMime(t *testing.T) string {
	t.Helper()
	for mime, ok := range attachments.AllowedMimeTypesForImage {
		if ok {
			return mime
		}
	}
	t.Fatal("no allowed image mime types configured in attachments.AllowedMimeTypesForImage")
	return ""
}

func TestUsecase_GetAttachment_Success(t *testing.T) {
	u, repoMock, _ := newTestUsecase(t)

	noteID, blockID, userID := uuid.New(), uuid.New(), uuid.New()
	want := &models.Attachment{ID: uuid.New(), BlockID: blockID}

	repoMock.EXPECT().GetAttachment(gomock.Any(), blockID).Return(want, nil)

	got, customErr := u.GetAttachment(context.Background(), noteID, blockID, userID)
	require.Nil(t, customErr)
	assert.Equal(t, want, got)
}

func TestUsecase_GetAttachment_RepoError(t *testing.T) {
	u, repoMock, _ := newTestUsecase(t)

	noteID, blockID, userID := uuid.New(), uuid.New(), uuid.New()
	repoErr := appErrWith(attachments.ErrAttachmentNotFound, attachments.PublicMsgErrAttachmentNotFound, 404)

	repoMock.EXPECT().GetAttachment(gomock.Any(), blockID).Return(nil, repoErr)

	got, customErr := u.GetAttachment(context.Background(), noteID, blockID, userID)
	require.Nil(t, got)
	require.NotNil(t, customErr)
	assert.True(t, customErr.Is(attachments.ErrAttachmentNotFound))
}

func TestUsecase_GetAttachment_NilAttachmentWithoutError(t *testing.T) {
	u, repoMock, _ := newTestUsecase(t)

	noteID, blockID, userID := uuid.New(), uuid.New(), uuid.New()

	repoMock.EXPECT().GetAttachment(gomock.Any(), blockID).Return(nil, nil)

	got, customErr := u.GetAttachment(context.Background(), noteID, blockID, userID)
	require.Nil(t, got)
	require.NotNil(t, customErr)
	assert.True(t, customErr.Is(attachments.ErrAttachmentNotFound))
}

func TestUsecase_GetHeader_Success(t *testing.T) {
	u, repoMock, _ := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	want := &models.Header{ID: uuid.New(), NoteID: noteID}

	repoMock.EXPECT().GetHeader(gomock.Any(), noteID).Return(want, nil)

	got, customErr := u.GetHeader(context.Background(), noteID, userID)
	require.Nil(t, customErr)
	assert.Equal(t, want, got)
}

func TestUsecase_GetHeader_RepoError(t *testing.T) {
	u, repoMock, _ := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	repoErr := appErrWith(attachments.ErrHeaderNotFound, attachments.PublicMsgErrHeaderNotFound, 404)

	repoMock.EXPECT().GetHeader(gomock.Any(), noteID).Return(nil, repoErr)

	got, customErr := u.GetHeader(context.Background(), noteID, userID)
	require.Nil(t, got)
	require.NotNil(t, customErr)
	assert.True(t, customErr.Is(attachments.ErrHeaderNotFound))
}

func TestUsecase_GetHeader_NilHeaderWithoutError(t *testing.T) {
	u, repoMock, _ := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()

	repoMock.EXPECT().GetHeader(gomock.Any(), noteID).Return(nil, nil)

	got, customErr := u.GetHeader(context.Background(), noteID, userID)
	require.Nil(t, got)
	require.NotNil(t, customErr)
	assert.True(t, customErr.Is(attachments.ErrHeaderNotFound))
}

func TestUsecase_UploadHeader_Success(t *testing.T) {
	u, repoMock, _ := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	want := &models.Header{ID: uuid.New(), NoteID: noteID}
	reader := strings.NewReader("data")

	repoMock.EXPECT().
		UploadHeader(gomock.Any(), noteID, "header.png", int64(4), "image/png", gomock.Any()).
		Return(want, nil)

	got, customErr := u.UploadHeader(context.Background(), noteID, userID, "header.png", 4, "image/png", reader)
	require.Nil(t, customErr)
	assert.Equal(t, want, got)
}

func TestUsecase_UploadHeader_RepoError(t *testing.T) {
	u, repoMock, _ := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	repoErr := appErrWith(errors.New("boom"), attachments.PublicMsgErrInternalServer, 500)

	repoMock.EXPECT().
		UploadHeader(gomock.Any(), noteID, "header.png", int64(4), "image/png", gomock.Any()).
		Return(nil, repoErr)

	got, customErr := u.UploadHeader(context.Background(), noteID, userID, "header.png", 4, "image/png", strings.NewReader("data"))
	require.Nil(t, got)
	require.NotNil(t, customErr)
}

func TestUsecase_DeleteHeader_Success(t *testing.T) {
	u, repoMock, _ := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	repoMock.EXPECT().DeleteHeader(gomock.Any(), noteID).Return(nil)

	customErr := u.DeleteHeader(context.Background(), noteID, userID)
	require.Nil(t, customErr)
}

func TestUsecase_DeleteHeader_RepoError(t *testing.T) {
	u, repoMock, _ := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	repoErr := appErrWith(attachments.ErrHeaderNotFound, attachments.PublicMsgErrHeaderNotFound, 404)
	repoMock.EXPECT().DeleteHeader(gomock.Any(), noteID).Return(repoErr)

	customErr := u.DeleteHeader(context.Background(), noteID, userID)
	require.NotNil(t, customErr)
	assert.True(t, customErr.Is(attachments.ErrHeaderNotFound))
}

func TestUsecase_DeleteAttachment_Success(t *testing.T) {
	u, repoMock, notesMock := newTestUsecase(t)

	noteID, blockID, userID := uuid.New(), uuid.New(), uuid.New()

	notesMock.EXPECT().
		GetBlock(gomock.Any(), blockID, noteID, userID).
		Return(&notesgen.BlockResponse{Id: blockID.String()}, nil)

	repoMock.EXPECT().DeleteAttachment(gomock.Any(), blockID).Return(nil)

	customErr := u.DeleteAttachment(context.Background(), noteID, blockID, userID)
	require.Nil(t, customErr)
}

func TestUsecase_DeleteAttachment_GetBlockGRPCError(t *testing.T) {
	u, _, notesMock := newTestUsecase(t)

	noteID, blockID, userID := uuid.New(), uuid.New(), uuid.New()

	notesMock.EXPECT().
		GetBlock(gomock.Any(), blockID, noteID, userID).
		Return(nil, errors.New("grpc unavailable"))

	customErr := u.DeleteAttachment(context.Background(), noteID, blockID, userID)
	require.NotNil(t, customErr)
}

func TestUsecase_DeleteAttachment_BlockNotFound(t *testing.T) {
	u, _, notesMock := newTestUsecase(t)

	noteID, blockID, userID := uuid.New(), uuid.New(), uuid.New()

	notesMock.EXPECT().
		GetBlock(gomock.Any(), blockID, noteID, userID).
		Return(nil, nil)

	customErr := u.DeleteAttachment(context.Background(), noteID, blockID, userID)
	require.NotNil(t, customErr)
	assert.True(t, customErr.Is(attachments.ErrBlockNotFound))
}

func TestUsecase_DeleteAttachment_RepoError(t *testing.T) {
	u, repoMock, notesMock := newTestUsecase(t)

	noteID, blockID, userID := uuid.New(), uuid.New(), uuid.New()

	notesMock.EXPECT().
		GetBlock(gomock.Any(), blockID, noteID, userID).
		Return(&notesgen.BlockResponse{Id: blockID.String()}, nil)

	repoErr := appErrWith(attachments.ErrAttachmentNotFound, attachments.PublicMsgErrAttachmentNotFound, 404)
	repoMock.EXPECT().DeleteAttachment(gomock.Any(), blockID).Return(repoErr)

	customErr := u.DeleteAttachment(context.Background(), noteID, blockID, userID)
	require.NotNil(t, customErr)
	assert.True(t, customErr.Is(attachments.ErrAttachmentNotFound))
}

func TestUsecase_UploadAttachment_InvalidMimeType(t *testing.T) {
	u, _, _ := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()

	got, customErr := u.UploadAttachment(
		context.Background(), noteID, userID,
		"virus.exe", 10, "application/x-msdownload",
		strings.NewReader("data"), false, 0,
	)

	require.Nil(t, got)
	require.NotNil(t, customErr)
	assert.True(t, customErr.Is(attachments.ErrInvalidMimeType))
}

func TestUsecase_UploadAttachment_Success_NoExplicitPosition(t *testing.T) {
	u, repoMock, notesMock := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	mime := pickAllowedImageMime(t)
	newBlockID := uuid.New()

	existingBlocks := []*notesgen.BlockResponse{{Id: uuid.New().String()}, {Id: uuid.New().String()}}

	notesMock.EXPECT().
		GetBlocks(gomock.Any(), noteID, userID).
		Return(existingBlocks, nil)

	notesMock.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 2, 1).
		Return(nil)

	notesMock.EXPECT().
		CreateBlock(gomock.Any(), userID, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ uuid.UUID, block *notesgen.BlockResponse) (*notesgen.BlockResponse, error) {
			assert.Equal(t, noteID.String(), block.NoteId)
			assert.Equal(t, int32(2), block.Position)
			return &notesgen.BlockResponse{Id: newBlockID.String()}, nil
		})

	want := &models.Attachment{ID: uuid.New(), BlockID: newBlockID}
	repoMock.EXPECT().
		UploadAttachment(gomock.Any(), newBlockID, "img.png", int64(123), mime, gomock.Any()).
		Return(want, nil)

	got, customErr := u.UploadAttachment(
		context.Background(), noteID, userID,
		"img.png", 123, mime,
		strings.NewReader("data"), false, 0,
	)

	require.Nil(t, customErr)
	assert.Equal(t, want, got)
}

func TestUsecase_UploadAttachment_Success_ExplicitPosition(t *testing.T) {
	u, repoMock, notesMock := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	mime := pickAllowedImageMime(t)
	newBlockID := uuid.New()

	existingBlocks := []*notesgen.BlockResponse{{Id: uuid.New().String()}, {Id: uuid.New().String()}, {Id: uuid.New().String()}}

	notesMock.EXPECT().
		GetBlocks(gomock.Any(), noteID, userID).
		Return(existingBlocks, nil)

	notesMock.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 1, 1).
		Return(nil)

	notesMock.EXPECT().
		CreateBlock(gomock.Any(), userID, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ uuid.UUID, block *notesgen.BlockResponse) (*notesgen.BlockResponse, error) {
			assert.Equal(t, int32(1), block.Position)
			return &notesgen.BlockResponse{Id: newBlockID.String()}, nil
		})

	want := &models.Attachment{ID: uuid.New(), BlockID: newBlockID}
	repoMock.EXPECT().
		UploadAttachment(gomock.Any(), newBlockID, "img.png", int64(123), mime, gomock.Any()).
		Return(want, nil)

	got, customErr := u.UploadAttachment(
		context.Background(), noteID, userID,
		"img.png", 123, mime,
		strings.NewReader("data"), true, 1,
	)

	require.Nil(t, customErr)
	assert.Equal(t, want, got)
}

func TestUsecase_UploadAttachment_InvalidPosition_Negative(t *testing.T) {
	u, _, notesMock := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	mime := pickAllowedImageMime(t)

	notesMock.EXPECT().
		GetBlocks(gomock.Any(), noteID, userID).
		Return([]*notesgen.BlockResponse{{Id: uuid.New().String()}}, nil)

	got, customErr := u.UploadAttachment(
		context.Background(), noteID, userID,
		"img.png", 10, mime,
		strings.NewReader("data"), true, -1,
	)

	require.Nil(t, got)
	require.NotNil(t, customErr)
	assert.True(t, customErr.Is(attachments.ErrInvalidPosition))
}

func TestUsecase_UploadAttachment_InvalidPosition_TooLarge(t *testing.T) {
	u, _, notesMock := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	mime := pickAllowedImageMime(t)

	notesMock.EXPECT().
		GetBlocks(gomock.Any(), noteID, userID).
		Return([]*notesgen.BlockResponse{{Id: uuid.New().String()}}, nil)

	got, customErr := u.UploadAttachment(
		context.Background(), noteID, userID,
		"img.png", 10, mime,
		strings.NewReader("data"), true, 2,
	)

	require.Nil(t, got)
	require.NotNil(t, customErr)
	assert.True(t, customErr.Is(attachments.ErrInvalidPosition))
}

func TestUsecase_UploadAttachment_PositionEqualsLen_IsValid(t *testing.T) {
	u, repoMock, notesMock := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	mime := pickAllowedImageMime(t)
	newBlockID := uuid.New()

	notesMock.EXPECT().
		GetBlocks(gomock.Any(), noteID, userID).
		Return([]*notesgen.BlockResponse{{Id: uuid.New().String()}}, nil)

	notesMock.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 1, 1).
		Return(nil)

	notesMock.EXPECT().
		CreateBlock(gomock.Any(), userID, gomock.Any()).
		Return(&notesgen.BlockResponse{Id: newBlockID.String()}, nil)

	want := &models.Attachment{ID: uuid.New(), BlockID: newBlockID}
	repoMock.EXPECT().
		UploadAttachment(gomock.Any(), newBlockID, "img.png", int64(10), mime, gomock.Any()).
		Return(want, nil)

	got, customErr := u.UploadAttachment(
		context.Background(), noteID, userID,
		"img.png", 10, mime,
		strings.NewReader("data"), true, 1,
	)

	require.Nil(t, customErr)
	assert.Equal(t, want, got)
}

func TestUsecase_UploadAttachment_GetBlocksGRPCError(t *testing.T) {
	u, _, notesMock := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	mime := pickAllowedImageMime(t)

	notesMock.EXPECT().
		GetBlocks(gomock.Any(), noteID, userID).
		Return(nil, errors.New("grpc down"))

	got, customErr := u.UploadAttachment(
		context.Background(), noteID, userID,
		"img.png", 10, mime,
		strings.NewReader("data"), false, 0,
	)

	require.Nil(t, got)
	require.NotNil(t, customErr)
}

func TestUsecase_UploadAttachment_ShiftPositionsFails(t *testing.T) {
	u, _, notesMock := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	mime := pickAllowedImageMime(t)

	notesMock.EXPECT().
		GetBlocks(gomock.Any(), noteID, userID).
		Return(nil, nil)

	notesMock.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 0, 1).
		Return(errors.New("shift failed"))

	got, customErr := u.UploadAttachment(
		context.Background(), noteID, userID,
		"img.png", 10, mime,
		strings.NewReader("data"), false, 0,
	)

	require.Nil(t, got)
	require.NotNil(t, customErr)
}

func TestUsecase_UploadAttachment_CreateBlockFails_RollbackShiftSucceeds(t *testing.T) {
	u, _, notesMock := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	mime := pickAllowedImageMime(t)

	notesMock.EXPECT().
		GetBlocks(gomock.Any(), noteID, userID).
		Return(nil, nil)

	gomock.InOrder(
		notesMock.EXPECT().ShiftBlockPositions(gomock.Any(), noteID, 0, 1).Return(nil),
		notesMock.EXPECT().ShiftBlockPositions(gomock.Any(), noteID, 0, -1).Return(nil),
	)

	notesMock.EXPECT().
		CreateBlock(gomock.Any(), userID, gomock.Any()).
		Return(nil, errors.New("create failed"))

	got, customErr := u.UploadAttachment(
		context.Background(), noteID, userID,
		"img.png", 10, mime,
		strings.NewReader("data"), false, 0,
	)

	require.Nil(t, got)
	require.NotNil(t, customErr)
}

func TestUsecase_UploadAttachment_CreateBlockFails_RollbackShiftAlsoFails(t *testing.T) {
	u, _, notesMock := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	mime := pickAllowedImageMime(t)

	notesMock.EXPECT().
		GetBlocks(gomock.Any(), noteID, userID).
		Return(nil, nil)

	gomock.InOrder(
		notesMock.EXPECT().ShiftBlockPositions(gomock.Any(), noteID, 0, 1).Return(nil),
		notesMock.EXPECT().ShiftBlockPositions(gomock.Any(), noteID, 0, -1).Return(errors.New("rollback shift failed")),
	)

	notesMock.EXPECT().
		CreateBlock(gomock.Any(), userID, gomock.Any()).
		Return(nil, errors.New("create failed"))

	got, customErr := u.UploadAttachment(
		context.Background(), noteID, userID,
		"img.png", 10, mime,
		strings.NewReader("data"), false, 0,
	)

	require.Nil(t, got)
	require.NotNil(t, customErr)
}

func TestUsecase_UploadAttachment_BlockIDParseFails_RollbackSucceeds(t *testing.T) {
	u, _, notesMock := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	mime := pickAllowedImageMime(t)

	notesMock.EXPECT().
		GetBlocks(gomock.Any(), noteID, userID).
		Return(nil, nil)

	notesMock.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 0, 1).
		Return(nil)

	notesMock.EXPECT().
		CreateBlock(gomock.Any(), userID, gomock.Any()).
		Return(&notesgen.BlockResponse{Id: "not-a-valid-uuid"}, nil)

	notesMock.EXPECT().
		DeleteBlock(gomock.Any(), uuid.Nil, noteID, userID).
		Return(uuid.Nil, nil)

	notesMock.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 0, -1).
		Return(nil)

	got, customErr := u.UploadAttachment(
		context.Background(), noteID, userID,
		"img.png", 10, mime,
		strings.NewReader("data"), false, 0,
	)

	require.Nil(t, got)
	require.NotNil(t, customErr)
}

func TestUsecase_UploadAttachment_BlockIDParseFails_RollbackDeleteAndShiftFail(t *testing.T) {
	u, _, notesMock := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	mime := pickAllowedImageMime(t)

	notesMock.EXPECT().
		GetBlocks(gomock.Any(), noteID, userID).
		Return(nil, nil)

	notesMock.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 0, 1).
		Return(nil)

	notesMock.EXPECT().
		CreateBlock(gomock.Any(), userID, gomock.Any()).
		Return(&notesgen.BlockResponse{Id: "still-not-a-uuid"}, nil)

	notesMock.EXPECT().
		DeleteBlock(gomock.Any(), uuid.Nil, noteID, userID).
		Return(uuid.Nil, errors.New("delete failed"))

	notesMock.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 0, -1).
		Return(errors.New("rollback shift failed"))

	got, customErr := u.UploadAttachment(
		context.Background(), noteID, userID,
		"img.png", 10, mime,
		strings.NewReader("data"), false, 0,
	)

	require.Nil(t, got)
	require.NotNil(t, customErr)
}

func TestUsecase_UploadAttachment_RepoFails_RollbackSucceeds(t *testing.T) {
	u, repoMock, notesMock := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	mime := pickAllowedImageMime(t)
	newBlockID := uuid.New()

	notesMock.EXPECT().
		GetBlocks(gomock.Any(), noteID, userID).
		Return(nil, nil)

	notesMock.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 0, 1).
		Return(nil)

	notesMock.EXPECT().
		CreateBlock(gomock.Any(), userID, gomock.Any()).
		Return(&notesgen.BlockResponse{Id: newBlockID.String()}, nil)

	notesMock.EXPECT().
		DeleteBlock(gomock.Any(), newBlockID, noteID, userID).
		Return(noteID, nil)

	notesMock.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 0, -1).
		Return(nil)

	repoErr := appErrWith(errors.New("disk full"), attachments.PublicMsgErrInternalServer, 500)
	repoMock.EXPECT().
		UploadAttachment(gomock.Any(), newBlockID, "img.png", int64(10), mime, gomock.Any()).
		Return(nil, repoErr)

	got, customErr := u.UploadAttachment(
		context.Background(), noteID, userID,
		"img.png", 10, mime,
		strings.NewReader("data"), false, 0,
	)

	require.Nil(t, got)
	require.NotNil(t, customErr)

	assert.Equal(t, attachments.PublicMsgErrInternalServer, customErr.PublicMessage())
	assert.Equal(t, 500, customErr.Code())
}

func TestUsecase_UploadAttachment_RepoFails_RollbackDeleteFails(t *testing.T) {
	u, repoMock, notesMock := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	mime := pickAllowedImageMime(t)
	newBlockID := uuid.New()

	notesMock.EXPECT().
		GetBlocks(gomock.Any(), noteID, userID).
		Return(nil, nil)

	notesMock.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 0, 1).
		Return(nil)

	notesMock.EXPECT().
		CreateBlock(gomock.Any(), userID, gomock.Any()).
		Return(&notesgen.BlockResponse{Id: newBlockID.String()}, nil)

	notesMock.EXPECT().
		DeleteBlock(gomock.Any(), newBlockID, noteID, userID).
		Return(uuid.Nil, errors.New("delete also failed"))

	notesMock.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 0, -1).
		Return(nil)

	repoErr := appErrWith(errors.New("disk full"), attachments.PublicMsgErrInternalServer, 500)
	repoMock.EXPECT().
		UploadAttachment(gomock.Any(), newBlockID, "img.png", int64(10), mime, gomock.Any()).
		Return(nil, repoErr)

	got, customErr := u.UploadAttachment(
		context.Background(), noteID, userID,
		"img.png", 10, mime,
		strings.NewReader("data"), false, 0,
	)

	require.Nil(t, got)
	require.NotNil(t, customErr)

	assert.Equal(t, attachments.PublicMsgErrInternalServer, customErr.PublicMessage())
	assert.Equal(t, 500, customErr.Code())
}

func TestUsecase_UploadAttachment_RepoFails_RollbackShiftFails(t *testing.T) {
	u, repoMock, notesMock := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	mime := pickAllowedImageMime(t)
	newBlockID := uuid.New()

	notesMock.EXPECT().
		GetBlocks(gomock.Any(), noteID, userID).
		Return(nil, nil)

	notesMock.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 0, 1).
		Return(nil)

	notesMock.EXPECT().
		CreateBlock(gomock.Any(), userID, gomock.Any()).
		Return(&notesgen.BlockResponse{Id: newBlockID.String()}, nil)

	notesMock.EXPECT().
		DeleteBlock(gomock.Any(), newBlockID, noteID, userID).
		Return(noteID, nil)

	notesMock.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 0, -1).
		Return(errors.New("rollback shift failed"))

	repoErr := appErrWith(errors.New("disk full"), attachments.PublicMsgErrInternalServer, 500)
	repoMock.EXPECT().
		UploadAttachment(gomock.Any(), newBlockID, "img.png", int64(10), mime, gomock.Any()).
		Return(nil, repoErr)

	got, customErr := u.UploadAttachment(
		context.Background(), noteID, userID,
		"img.png", 10, mime,
		strings.NewReader("data"), false, 0,
	)

	require.Nil(t, got)
	require.NotNil(t, customErr)
	assert.Equal(t, attachments.PublicMsgErrInternalServer, customErr.PublicMessage())
	assert.Equal(t, 500, customErr.Code())
}

func TestUsecase_UploadAttachment_ForwardsReader(t *testing.T) {
	u, repoMock, notesMock := newTestUsecase(t)

	noteID, userID := uuid.New(), uuid.New()
	mime := pickAllowedImageMime(t)
	newBlockID := uuid.New()
	content := "raw-bytes"

	notesMock.EXPECT().
		GetBlocks(gomock.Any(), noteID, userID).
		Return(nil, nil)

	notesMock.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 0, 1).
		Return(nil)

	notesMock.EXPECT().
		CreateBlock(gomock.Any(), userID, gomock.Any()).
		Return(&notesgen.BlockResponse{Id: newBlockID.String()}, nil)

	var capturedReader io.Reader
	repoMock.EXPECT().
		UploadAttachment(gomock.Any(), newBlockID, "img.png", int64(len(content)), mime, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ uuid.UUID, _ string, _ int64, _ string, r io.Reader) (*models.Attachment, types.AppErrorInterface) {
			capturedReader = r
			return &models.Attachment{ID: uuid.New(), BlockID: newBlockID}, nil
		})

	_, customErr := u.UploadAttachment(
		context.Background(), noteID, userID,
		"img.png", int64(len(content)), mime,
		strings.NewReader(content), false, 0,
	)

	require.Nil(t, customErr)
	require.NotNil(t, capturedReader)
	b, err := io.ReadAll(capturedReader)
	require.NoError(t, err)
	assert.Equal(t, content, string(b))
}

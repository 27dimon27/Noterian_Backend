package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/handler/http/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ============================================================
// helpers
// ============================================================

func newTestHandler(t *testing.T) (*AttachmentHandler, *mocks.MockAttachmentUsecase) {
	t.Helper()

	ctrl := gomock.NewController(t)
	usecaseMock := mocks.NewMockAttachmentUsecase(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	h := NewAttachmentHandler(usecaseMock, logger)
	return h, usecaseMock
}

// withUser injects userID into the request context, mimicking auth middleware.
func withUser(r *http.Request, userID uuid.UUID) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, userID))
}

type errorBody struct {
	Error string `json:"error"`
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) errorBody {
	t.Helper()
	var body errorBody
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	return body
}

func appErrWith(err error, publicMsg string, code int) types.AppErrorInterface {
	return &types.AppError{
		Err:        err,
		PublicMsg:  publicMsg,
		StatusCode: code,
		Layer:      "usecase",
		Op:         "test",
	}
}

// pngBytes returns a buffer whose first bytes are a valid PNG signature
// (so http.DetectContentType reports "image/png"), padded to the given
// total size with zero bytes.
func pngBytes(totalSize int) []byte {
	sig := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if totalSize < len(sig) {
		totalSize = len(sig)
	}
	buf := make([]byte, totalSize)
	copy(buf, sig)
	return buf
}

// newMultipartRequest builds a multipart/form-data POST request with a single
// file field named "file". Extra query params (e.g. "position") can be added
// via rawQuery.
func newMultipartRequest(t *testing.T, target, rawQuery, fileFieldName, fileName string, fileContent []byte) *http.Request {
	t.Helper()

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)

	if fileFieldName != "" {
		part, err := mw.CreateFormFile(fileFieldName, fileName)
		require.NoError(t, err)
		_, err = part.Write(fileContent)
		require.NoError(t, err)
	}
	require.NoError(t, mw.Close())

	url := target
	if rawQuery != "" {
		url += "?" + rawQuery
	}

	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

// ============================================================
// GetAttachment
// ============================================================

func TestHandler_GetAttachment_Unauthorized(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/notes/x/blocks/y/attachments", nil)
	rec := httptest.NewRecorder()

	h.GetAttachment(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_GetAttachment_MissingNoteID(t *testing.T) {
	h, _ := newTestHandler(t)

	req := withUser(httptest.NewRequest(http.MethodGet, "/notes//blocks/y/attachments", nil), uuid.New())
	rec := httptest.NewRecorder()

	h.GetAttachment(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_GetAttachment_InvalidNoteID(t *testing.T) {
	h, _ := newTestHandler(t)

	req := withUser(httptest.NewRequest(http.MethodGet, "/notes/not-a-uuid/blocks/y/attachments", nil), uuid.New())
	req.SetPathValue("noteId", "not-a-uuid")
	rec := httptest.NewRecorder()

	h.GetAttachment(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_GetAttachment_MissingBlockID(t *testing.T) {
	h, _ := newTestHandler(t)

	noteID := uuid.New()
	req := withUser(httptest.NewRequest(http.MethodGet, "/notes/x/blocks/y/attachments", nil), uuid.New())
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.GetAttachment(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_GetAttachment_InvalidBlockID(t *testing.T) {
	h, _ := newTestHandler(t)

	noteID := uuid.New()
	req := withUser(httptest.NewRequest(http.MethodGet, "/notes/x/blocks/y/attachments", nil), uuid.New())
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", "not-a-uuid")
	rec := httptest.NewRecorder()

	h.GetAttachment(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_GetAttachment_Success(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, blockID, userID := uuid.New(), uuid.New(), uuid.New()
	want := &models.Attachment{ID: uuid.New(), BlockID: blockID, MinioKey: "k", AttachURL: "http://x"}

	usecaseMock.EXPECT().GetAttachment(gomock.Any(), noteID, blockID, userID).Return(want, nil)

	req := withUser(httptest.NewRequest(http.MethodGet, "/notes/x/blocks/y/attachments", nil), userID)
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	rec := httptest.NewRecorder()

	h.GetAttachment(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var got struct {
		ID uuid.UUID `json:"id"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, want.ID, got.ID)
}

func TestHandler_GetAttachment_UsecaseForbidden(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, blockID, userID := uuid.New(), uuid.New(), uuid.New()
	usecaseMock.EXPECT().GetAttachment(gomock.Any(), noteID, blockID, userID).
		Return(nil, appErrWith(attachments.ErrForbidden, attachments.PublicMsgErrForbidden, 403))

	req := withUser(httptest.NewRequest(http.MethodGet, "/notes/x/blocks/y/attachments", nil), userID)
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	rec := httptest.NewRecorder()

	h.GetAttachment(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, attachments.PublicMsgErrForbidden, decodeError(t, rec).Error)
}

func TestHandler_GetAttachment_UsecaseNotFoundVariants(t *testing.T) {
	notFoundErrs := []error{attachments.ErrNoteNotFound, attachments.ErrBlockNotFound, attachments.ErrAttachmentNotFound}

	for _, target := range notFoundErrs {
		t.Run(target.Error(), func(t *testing.T) {
			h, usecaseMock := newTestHandler(t)

			noteID, blockID, userID := uuid.New(), uuid.New(), uuid.New()
			usecaseMock.EXPECT().GetAttachment(gomock.Any(), noteID, blockID, userID).
				Return(nil, appErrWith(target, "not found", 404))

			req := withUser(httptest.NewRequest(http.MethodGet, "/notes/x/blocks/y/attachments", nil), userID)
			req.SetPathValue("noteId", noteID.String())
			req.SetPathValue("blockId", blockID.String())
			rec := httptest.NewRecorder()

			h.GetAttachment(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code)
		})
	}
}

func TestHandler_GetAttachment_UsecaseInternalError(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, blockID, userID := uuid.New(), uuid.New(), uuid.New()
	usecaseMock.EXPECT().GetAttachment(gomock.Any(), noteID, blockID, userID).
		Return(nil, appErrWith(errors.New("boom"), attachments.PublicMsgErrInternalServer, 500))

	req := withUser(httptest.NewRequest(http.MethodGet, "/notes/x/blocks/y/attachments", nil), userID)
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	rec := httptest.NewRecorder()

	h.GetAttachment(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ============================================================
// DeleteAttachment
// ============================================================

func TestHandler_DeleteAttachment_Unauthorized(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/notes/x/blocks/y/attachments", nil)
	rec := httptest.NewRecorder()

	h.DeleteAttachment(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_DeleteAttachment_MissingNoteID(t *testing.T) {
	h, _ := newTestHandler(t)

	req := withUser(httptest.NewRequest(http.MethodDelete, "/notes/x/blocks/y/attachments", nil), uuid.New())
	rec := httptest.NewRecorder()

	h.DeleteAttachment(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_DeleteAttachment_InvalidNoteID(t *testing.T) {
	h, _ := newTestHandler(t)

	req := withUser(httptest.NewRequest(http.MethodDelete, "/notes/x/blocks/y/attachments", nil), uuid.New())
	req.SetPathValue("noteId", "not-a-uuid")
	rec := httptest.NewRecorder()

	h.DeleteAttachment(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_DeleteAttachment_MissingBlockID(t *testing.T) {
	h, _ := newTestHandler(t)

	req := withUser(httptest.NewRequest(http.MethodDelete, "/notes/x/blocks/y/attachments", nil), uuid.New())
	req.SetPathValue("noteId", uuid.New().String())
	rec := httptest.NewRecorder()

	h.DeleteAttachment(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_DeleteAttachment_InvalidBlockID(t *testing.T) {
	h, _ := newTestHandler(t)

	req := withUser(httptest.NewRequest(http.MethodDelete, "/notes/x/blocks/y/attachments", nil), uuid.New())
	req.SetPathValue("noteId", uuid.New().String())
	req.SetPathValue("blockId", "not-a-uuid")
	rec := httptest.NewRecorder()

	h.DeleteAttachment(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_DeleteAttachment_Success(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, blockID, userID := uuid.New(), uuid.New(), uuid.New()
	usecaseMock.EXPECT().DeleteAttachment(gomock.Any(), noteID, blockID, userID).Return(nil)

	req := withUser(httptest.NewRequest(http.MethodDelete, "/notes/x/blocks/y/attachments", nil), userID)
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	rec := httptest.NewRecorder()

	h.DeleteAttachment(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_DeleteAttachment_UsecaseForbidden(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, blockID, userID := uuid.New(), uuid.New(), uuid.New()
	usecaseMock.EXPECT().DeleteAttachment(gomock.Any(), noteID, blockID, userID).
		Return(appErrWith(attachments.ErrForbidden, attachments.PublicMsgErrForbidden, 403))

	req := withUser(httptest.NewRequest(http.MethodDelete, "/notes/x/blocks/y/attachments", nil), userID)
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	rec := httptest.NewRecorder()

	h.DeleteAttachment(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestHandler_DeleteAttachment_UsecaseNotFound(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, blockID, userID := uuid.New(), uuid.New(), uuid.New()
	usecaseMock.EXPECT().DeleteAttachment(gomock.Any(), noteID, blockID, userID).
		Return(appErrWith(attachments.ErrAttachmentNotFound, attachments.PublicMsgErrAttachmentNotFound, 404))

	req := withUser(httptest.NewRequest(http.MethodDelete, "/notes/x/blocks/y/attachments", nil), userID)
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	rec := httptest.NewRecorder()

	h.DeleteAttachment(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_DeleteAttachment_UsecaseInternalError(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, blockID, userID := uuid.New(), uuid.New(), uuid.New()
	usecaseMock.EXPECT().DeleteAttachment(gomock.Any(), noteID, blockID, userID).
		Return(appErrWith(errors.New("boom"), attachments.PublicMsgErrInternalServer, 500))

	req := withUser(httptest.NewRequest(http.MethodDelete, "/notes/x/blocks/y/attachments", nil), userID)
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	rec := httptest.NewRecorder()

	h.DeleteAttachment(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ============================================================
// GetHeader
// ============================================================

func TestHandler_GetHeader_Unauthorized(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/notes/x/header", nil)
	rec := httptest.NewRecorder()

	h.GetHeader(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_GetHeader_MissingNoteID(t *testing.T) {
	h, _ := newTestHandler(t)

	req := withUser(httptest.NewRequest(http.MethodGet, "/notes/x/header", nil), uuid.New())
	rec := httptest.NewRecorder()

	h.GetHeader(rec, req)

	// Note: handler.go uses StatusUnauthorized here (not BadRequest) for this branch.
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_GetHeader_InvalidNoteID(t *testing.T) {
	h, _ := newTestHandler(t)

	req := withUser(httptest.NewRequest(http.MethodGet, "/notes/x/header", nil), uuid.New())
	req.SetPathValue("noteId", "not-a-uuid")
	rec := httptest.NewRecorder()

	h.GetHeader(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_GetHeader_Success(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()
	want := &models.Header{ID: uuid.New(), NoteID: noteID, MinioKey: "k", HeaderURL: "http://x"}

	usecaseMock.EXPECT().GetHeader(gomock.Any(), noteID, userID).Return(want, nil)

	req := withUser(httptest.NewRequest(http.MethodGet, "/notes/x/header", nil), userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.GetHeader(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var got struct {
		ID uuid.UUID `json:"id"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, want.ID, got.ID)
}

func TestHandler_GetHeader_NotFound(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()
	usecaseMock.EXPECT().GetHeader(gomock.Any(), noteID, userID).
		Return(nil, appErrWith(attachments.ErrHeaderNotFound, attachments.PublicMsgErrHeaderNotFound, 404))

	req := withUser(httptest.NewRequest(http.MethodGet, "/notes/x/header", nil), userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.GetHeader(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_GetHeader_InternalError(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()
	usecaseMock.EXPECT().GetHeader(gomock.Any(), noteID, userID).
		Return(nil, appErrWith(errors.New("boom"), attachments.PublicMsgErrInternalServer, 500))

	req := withUser(httptest.NewRequest(http.MethodGet, "/notes/x/header", nil), userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.GetHeader(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ============================================================
// DeleteHeader
// ============================================================

func TestHandler_DeleteHeader_Unauthorized(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/notes/x/header", nil)
	rec := httptest.NewRecorder()

	h.DeleteHeader(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_DeleteHeader_MissingNoteID(t *testing.T) {
	h, _ := newTestHandler(t)

	req := withUser(httptest.NewRequest(http.MethodDelete, "/notes/x/header", nil), uuid.New())
	rec := httptest.NewRecorder()

	h.DeleteHeader(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_DeleteHeader_InvalidNoteID(t *testing.T) {
	h, _ := newTestHandler(t)

	req := withUser(httptest.NewRequest(http.MethodDelete, "/notes/x/header", nil), uuid.New())
	req.SetPathValue("noteId", "not-a-uuid")
	rec := httptest.NewRecorder()

	h.DeleteHeader(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_DeleteHeader_Success(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()
	usecaseMock.EXPECT().DeleteHeader(gomock.Any(), noteID, userID).Return(nil)

	req := withUser(httptest.NewRequest(http.MethodDelete, "/notes/x/header", nil), userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.DeleteHeader(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_DeleteHeader_NotFound(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()
	usecaseMock.EXPECT().DeleteHeader(gomock.Any(), noteID, userID).
		Return(appErrWith(attachments.ErrHeaderNotFound, attachments.PublicMsgErrHeaderNotFound, 404))

	req := withUser(httptest.NewRequest(http.MethodDelete, "/notes/x/header", nil), userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.DeleteHeader(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_DeleteHeader_InternalError(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()
	usecaseMock.EXPECT().DeleteHeader(gomock.Any(), noteID, userID).
		Return(appErrWith(errors.New("boom"), attachments.PublicMsgErrInternalServer, 500))

	req := withUser(httptest.NewRequest(http.MethodDelete, "/notes/x/header", nil), userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.DeleteHeader(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ============================================================
// UploadAttachment
// ============================================================

func TestHandler_UploadAttachment_Unauthorized(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/notes/x/attachments", nil)
	rec := httptest.NewRecorder()

	h.UploadAttachment(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_UploadAttachment_MissingNoteID(t *testing.T) {
	h, _ := newTestHandler(t)

	req := withUser(httptest.NewRequest(http.MethodPost, "/notes/x/attachments", nil), uuid.New())
	rec := httptest.NewRecorder()

	h.UploadAttachment(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_UploadAttachment_InvalidNoteID(t *testing.T) {
	h, _ := newTestHandler(t)

	req := withUser(httptest.NewRequest(http.MethodPost, "/notes/x/attachments", nil), uuid.New())
	req.SetPathValue("noteId", "not-a-uuid")
	rec := httptest.NewRecorder()

	h.UploadAttachment(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_UploadAttachment_InvalidPositionQueryParam(t *testing.T) {
	h, _ := newTestHandler(t)

	noteID := uuid.New()
	req := withUser(httptest.NewRequest(http.MethodPost, "/notes/x/attachments?position=abc", nil), uuid.New())
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadAttachment(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_UploadAttachment_NoFileField(t *testing.T) {
	h, _ := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()

	// build a multipart body with no "file" part at all
	req := newMultipartRequest(t, "/notes/x/attachments", "", "", "", nil)
	req = withUser(req, userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadAttachment(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandler_UploadAttachment_BodyTooLarge(t *testing.T) {
	h, _ := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()

	// MAX_VIDEO_SIZE is the global cap applied by http.MaxBytesReader in
	// UploadAttachment; multipart framing overhead pushes a payload of
	// exactly that size over the limit.
	content := pngBytes(int(attachments.MAX_VIDEO_SIZE) + 1024)

	req := newMultipartRequest(t, "/notes/x/attachments", "", "file", "big.png", content)
	req = withUser(req, userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadAttachment(rec, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	assert.Equal(t, attachments.PublicMsgErrFileTooLarge, decodeError(t, rec).Error)
}

func TestHandler_UploadAttachment_InvalidMimeType(t *testing.T) {
	h, _ := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()

	// Plain text content sniffs to "text/plain; charset=utf-8", which is not
	// in any of the allowed mime maps.
	content := []byte("just some plain text content, nothing special here")

	req := newMultipartRequest(t, "/notes/x/attachments", "", "file", "note.txt", content)
	req = withUser(req, userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadAttachment(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, attachments.PublicMsgErrInvalidMimeType, decodeError(t, rec).Error)
}

func TestHandler_UploadAttachment_FileTooLargeForSpecificMimeType(t *testing.T) {
	h, _ := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()

	// image/png is capped at MAX_IMAGE_SIZE; exceed it while staying far
	// below MAX_VIDEO_SIZE so the global MaxBytesReader doesn't trip first.
	content := pngBytes(int(attachments.MAX_IMAGE_SIZE) + 1024)

	req := newMultipartRequest(t, "/notes/x/attachments", "", "file", "big.png", content)
	req = withUser(req, userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadAttachment(rec, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	assert.Equal(t, attachments.PublicMsgErrSpecificFileTooLarge["IMAGE"], decodeError(t, rec).Error)
}

func TestHandler_UploadAttachment_Success_NoPosition(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, userID, blockID := uuid.New(), uuid.New(), uuid.New()
	content := pngBytes(64)

	want := &models.Attachment{ID: uuid.New(), BlockID: blockID}
	usecaseMock.EXPECT().
		UploadAttachment(gomock.Any(), noteID, userID, "photo.png", int64(len(content)), "image/png", gomock.Any(), false, 0).
		Return(want, nil)

	req := newMultipartRequest(t, "/notes/x/attachments", "", "file", "photo.png", content)
	req = withUser(req, userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadAttachment(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var got struct {
		ID uuid.UUID `json:"id"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, want.ID, got.ID)
}

func TestHandler_UploadAttachment_Success_WithPosition(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, userID, blockID := uuid.New(), uuid.New(), uuid.New()
	content := pngBytes(64)

	want := &models.Attachment{ID: uuid.New(), BlockID: blockID}
	usecaseMock.EXPECT().
		UploadAttachment(gomock.Any(), noteID, userID, "photo.png", int64(len(content)), "image/png", gomock.Any(), true, 3).
		Return(want, nil)

	req := newMultipartRequest(t, "/notes/x/attachments", "position=3", "file", "photo.png", content)
	req = withUser(req, userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadAttachment(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestHandler_UploadAttachment_UsecaseForbidden(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()
	content := pngBytes(64)

	usecaseMock.EXPECT().
		UploadAttachment(gomock.Any(), noteID, userID, gomock.Any(), gomock.Any(), "image/png", gomock.Any(), false, 0).
		Return(nil, appErrWith(attachments.ErrForbidden, attachments.PublicMsgErrForbidden, 403))

	req := newMultipartRequest(t, "/notes/x/attachments", "", "file", "photo.png", content)
	req = withUser(req, userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadAttachment(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestHandler_UploadAttachment_UsecaseNotFound(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()
	content := pngBytes(64)

	usecaseMock.EXPECT().
		UploadAttachment(gomock.Any(), noteID, userID, gomock.Any(), gomock.Any(), "image/png", gomock.Any(), false, 0).
		Return(nil, appErrWith(attachments.ErrBlockNotFound, attachments.PublicMsgErrBlockNotFound, 404))

	req := newMultipartRequest(t, "/notes/x/attachments", "", "file", "photo.png", content)
	req = withUser(req, userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadAttachment(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_UploadAttachment_UsecaseConflict(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()
	content := pngBytes(64)

	usecaseMock.EXPECT().
		UploadAttachment(gomock.Any(), noteID, userID, gomock.Any(), gomock.Any(), "image/png", gomock.Any(), false, 0).
		Return(nil, appErrWith(attachments.ErrBlockAlreadyHasAttach, attachments.PublicMsgErrBlockAlreadyHasAttach, 409))

	req := newMultipartRequest(t, "/notes/x/attachments", "", "file", "photo.png", content)
	req = withUser(req, userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadAttachment(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestHandler_UploadAttachment_UsecaseInternalError(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()
	content := pngBytes(64)

	usecaseMock.EXPECT().
		UploadAttachment(gomock.Any(), noteID, userID, gomock.Any(), gomock.Any(), "image/png", gomock.Any(), false, 0).
		Return(nil, appErrWith(errors.New("boom"), attachments.PublicMsgErrInternalServer, 500))

	req := newMultipartRequest(t, "/notes/x/attachments", "", "file", "photo.png", content)
	req = withUser(req, userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadAttachment(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandler_UploadAttachment_NilAttachmentWithoutError(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()
	content := pngBytes(64)

	usecaseMock.EXPECT().
		UploadAttachment(gomock.Any(), noteID, userID, gomock.Any(), gomock.Any(), "image/png", gomock.Any(), false, 0).
		Return(nil, nil)

	req := newMultipartRequest(t, "/notes/x/attachments", "", "file", "photo.png", content)
	req = withUser(req, userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadAttachment(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// sanity: confirms the full file content (sniff buffer + remainder) reaches
// the usecase undamaged, not just the first 512 sniffed bytes.
func TestHandler_UploadAttachment_FullContentForwarded(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()
	content := pngBytes(2000) // larger than the 512-byte sniff buffer

	var captured []byte
	usecaseMock.EXPECT().
		UploadAttachment(gomock.Any(), noteID, userID, "photo.png", int64(len(content)), "image/png", gomock.Any(), false, 0).
		DoAndReturn(func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ string, _ int64, _ string, r io.Reader, _ bool, _ int) (*models.Attachment, types.AppErrorInterface) {
			var err error
			captured, err = io.ReadAll(r)
			require.NoError(t, err)
			return &models.Attachment{ID: uuid.New()}, nil
		})

	req := newMultipartRequest(t, "/notes/x/attachments", "", "file", "photo.png", content)
	req = withUser(req, userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadAttachment(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, content, captured)
}

// ============================================================
// UploadHeader
// ============================================================

func TestHandler_UploadHeader_Unauthorized(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/notes/x/header", nil)
	rec := httptest.NewRecorder()

	h.UploadHeader(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_UploadHeader_MissingNoteID(t *testing.T) {
	h, _ := newTestHandler(t)

	req := withUser(httptest.NewRequest(http.MethodPost, "/notes/x/header", nil), uuid.New())
	rec := httptest.NewRecorder()

	h.UploadHeader(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_UploadHeader_InvalidNoteID(t *testing.T) {
	h, _ := newTestHandler(t)

	req := withUser(httptest.NewRequest(http.MethodPost, "/notes/x/header", nil), uuid.New())
	req.SetPathValue("noteId", "not-a-uuid")
	rec := httptest.NewRecorder()

	h.UploadHeader(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_UploadHeader_NoFileField(t *testing.T) {
	h, _ := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()

	req := newMultipartRequest(t, "/notes/x/header", "", "", "", nil)
	req = withUser(req, userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadHeader(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandler_UploadHeader_BodyTooLarge(t *testing.T) {
	h, _ := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()

	// MAX_IMAGE_SIZE is the global cap for headers; comfortably exceed it.
	content := pngBytes(int(attachments.MAX_IMAGE_SIZE) + 1024)

	req := newMultipartRequest(t, "/notes/x/header", "", "file", "big.png", content)
	req = withUser(req, userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadHeader(rec, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	assert.Equal(t, attachments.PublicMsgErrFileTooLarge, decodeError(t, rec).Error)
}

func TestHandler_UploadHeader_InvalidMimeType(t *testing.T) {
	h, _ := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()
	content := []byte("just some plain text content, nothing special here")

	req := newMultipartRequest(t, "/notes/x/header", "", "file", "note.txt", content)
	req = withUser(req, userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadHeader(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, attachments.PublicMsgErrInvalidMimeType, decodeError(t, rec).Error)
}

func TestHandler_UploadHeader_Success(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()
	content := pngBytes(64)

	want := &models.Header{ID: uuid.New(), NoteID: noteID}
	usecaseMock.EXPECT().
		UploadHeader(gomock.Any(), noteID, userID, "logo.png", int64(len(content)), "image/png", gomock.Any()).
		Return(want, nil)

	req := newMultipartRequest(t, "/notes/x/header", "", "file", "logo.png", content)
	req = withUser(req, userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadHeader(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var got struct {
		ID uuid.UUID `json:"id"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, want.ID, got.ID)
}

func TestHandler_UploadHeader_UsecaseError(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()
	content := pngBytes(64)

	// UploadHeader's error branch is a flat 500 regardless of underlying error.
	usecaseMock.EXPECT().
		UploadHeader(gomock.Any(), noteID, userID, gomock.Any(), gomock.Any(), "image/png", gomock.Any()).
		Return(nil, appErrWith(errors.New("boom"), "something broke", 500))

	req := newMultipartRequest(t, "/notes/x/header", "", "file", "logo.png", content)
	req = withUser(req, userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadHeader(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "something broke", decodeError(t, rec).Error)
}

func TestHandler_UploadHeader_NilHeaderWithoutError(t *testing.T) {
	h, usecaseMock := newTestHandler(t)

	noteID, userID := uuid.New(), uuid.New()
	content := pngBytes(64)

	usecaseMock.EXPECT().
		UploadHeader(gomock.Any(), noteID, userID, gomock.Any(), gomock.Any(), "image/png", gomock.Any()).
		Return(nil, nil)

	req := newMultipartRequest(t, "/notes/x/header", "", "file", "logo.png", content)
	req = withUser(req, userID)
	req.SetPathValue("noteId", noteID.String())
	rec := httptest.NewRecorder()

	h.UploadHeader(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

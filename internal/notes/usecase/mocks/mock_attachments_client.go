package mocks

import (
	"context"
	"errors"
	reflect "reflect"

	attachmentsgen "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/attachments/grpc/gen"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

type MockAttachmentsServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockAttachmentsServiceClientMockRecorder
}

type MockAttachmentsServiceClientMockRecorder struct {
	mock *MockAttachmentsServiceClient
}

func NewMockAttachmentsServiceClient(ctrl *gomock.Controller) *MockAttachmentsServiceClient {
	mock := &MockAttachmentsServiceClient{ctrl: ctrl}
	mock.recorder = &MockAttachmentsServiceClientMockRecorder{mock}
	return mock
}

func (m *MockAttachmentsServiceClient) EXPECT() *MockAttachmentsServiceClientMockRecorder {
	return m.recorder
}

func (m *MockAttachmentsServiceClient) GetAttachment(ctx context.Context, blockID, noteID, userID uuid.UUID) (*attachmentsgen.AttachmentResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAttachment", ctx, blockID, noteID, userID)
	ret0, ok := ret[0].(*attachmentsgen.AttachmentResponse)
	if !ok {
		return nil, errors.New("failed to retrieve attachment: ret0")
	}
	ret1, ok := ret[1].(error)
	if !ok {
		return nil, errors.New("failed to retrieve attachment: ret1")
	}
	return ret0, ret1
}

func (mr *MockAttachmentsServiceClientMockRecorder) GetAttachment(ctx, blockID, noteID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAttachment", reflect.TypeOf((*MockAttachmentsServiceClient)(nil).GetAttachment), ctx, blockID, noteID, userID)
}

func (m *MockAttachmentsServiceClient) DeleteAttachment(ctx context.Context, blockID, noteID, userID uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAttachment", ctx, blockID, noteID, userID)
	ret0, ok := ret[0].(error)
	if !ok {
		return errors.New("failed to delete attachment: ret0")
	}
	return ret0
}

func (mr *MockAttachmentsServiceClientMockRecorder) DeleteAttachment(ctx, blockID, noteID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAttachment", reflect.TypeOf((*MockAttachmentsServiceClient)(nil).DeleteAttachment), ctx, blockID, noteID, userID)
}

func (m *MockAttachmentsServiceClient) GetHeader(ctx context.Context, noteID, userID uuid.UUID) (*attachmentsgen.HeaderResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHeader", ctx, noteID, userID)
	ret0, ok := ret[0].(*attachmentsgen.HeaderResponse)
	if !ok {
		return nil, errors.New("failed to retrieve header: ret0")
	}
	ret1, ok := ret[1].(error)
	if !ok {
		return nil, errors.New("failed to retrieve header: ret1")
	}
	return ret0, ret1
}

func (mr *MockAttachmentsServiceClientMockRecorder) GetHeader(ctx, noteID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHeader", reflect.TypeOf((*MockAttachmentsServiceClient)(nil).GetHeader), ctx, noteID, userID)
}

func (m *MockAttachmentsServiceClient) DeleteHeader(ctx context.Context, noteID, userID uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteHeader", ctx, noteID, userID)
	ret0, ok := ret[0].(error)
	if !ok {
		return errors.New("failed to delete header: ret0")
	}
	return ret0
}

func (mr *MockAttachmentsServiceClientMockRecorder) DeleteHeader(ctx, noteID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteHeader", reflect.TypeOf((*MockAttachmentsServiceClient)(nil).DeleteHeader), ctx, noteID, userID)
}

func (m *MockAttachmentsServiceClient) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, ok := ret[0].(error)
	if !ok {
		return errors.New("failed to close attachment client: ret0")
	}
	return ret0
}

func (mr *MockAttachmentsServiceClientMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockAttachmentsServiceClient)(nil).Close))
}

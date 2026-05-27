// usecase/mocks/mock_attachments_client.go
package mocks

import (
	"context"
	reflect "reflect"

	attachmentsgen "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/attachments/grpc/gen"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

// MockAttachmentsServiceClient is a mock of AttachmentsServiceClient interface
type MockAttachmentsServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockAttachmentsServiceClientMockRecorder
}

// MockAttachmentsServiceClientMockRecorder is the mock recorder for MockAttachmentsServiceClient
type MockAttachmentsServiceClientMockRecorder struct {
	mock *MockAttachmentsServiceClient
}

// NewMockAttachmentsServiceClient creates a new mock instance
func NewMockAttachmentsServiceClient(ctrl *gomock.Controller) *MockAttachmentsServiceClient {
	mock := &MockAttachmentsServiceClient{ctrl: ctrl}
	mock.recorder = &MockAttachmentsServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAttachmentsServiceClient) EXPECT() *MockAttachmentsServiceClientMockRecorder {
	return m.recorder
}

// GetAttachment mocks base method
func (m *MockAttachmentsServiceClient) GetAttachment(ctx context.Context, blockID, noteID, userID uuid.UUID) (*attachmentsgen.AttachmentResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAttachment", ctx, blockID, noteID, userID)
	ret0, _ := ret[0].(*attachmentsgen.AttachmentResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAttachment indicates an expected call of GetAttachment
func (mr *MockAttachmentsServiceClientMockRecorder) GetAttachment(ctx, blockID, noteID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAttachment", reflect.TypeOf((*MockAttachmentsServiceClient)(nil).GetAttachment), ctx, blockID, noteID, userID)
}

// DeleteAttachment mocks base method
func (m *MockAttachmentsServiceClient) DeleteAttachment(ctx context.Context, blockID, noteID, userID uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAttachment", ctx, blockID, noteID, userID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAttachment indicates an expected call of DeleteAttachment
func (mr *MockAttachmentsServiceClientMockRecorder) DeleteAttachment(ctx, blockID, noteID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAttachment", reflect.TypeOf((*MockAttachmentsServiceClient)(nil).DeleteAttachment), ctx, blockID, noteID, userID)
}

// GetHeader mocks base method
func (m *MockAttachmentsServiceClient) GetHeader(ctx context.Context, noteID, userID uuid.UUID) (*attachmentsgen.HeaderResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHeader", ctx, noteID, userID)
	ret0, _ := ret[0].(*attachmentsgen.HeaderResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHeader indicates an expected call of GetHeader
func (mr *MockAttachmentsServiceClientMockRecorder) GetHeader(ctx, noteID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHeader", reflect.TypeOf((*MockAttachmentsServiceClient)(nil).GetHeader), ctx, noteID, userID)
}

// DeleteHeader mocks base method
func (m *MockAttachmentsServiceClient) DeleteHeader(ctx context.Context, noteID, userID uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteHeader", ctx, noteID, userID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteHeader indicates an expected call of DeleteHeader
func (mr *MockAttachmentsServiceClientMockRecorder) DeleteHeader(ctx, noteID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteHeader", reflect.TypeOf((*MockAttachmentsServiceClient)(nil).DeleteHeader), ctx, noteID, userID)
}

// Close mocks base method
func (m *MockAttachmentsServiceClient) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (mr *MockAttachmentsServiceClientMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockAttachmentsServiceClient)(nil).Close))
}

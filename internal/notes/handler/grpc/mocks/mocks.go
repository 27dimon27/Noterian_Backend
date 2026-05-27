// handler/grpc/mocks/mock_usecase.go
package mocks

import (
	"context"
	"reflect"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

type MockNoteUsecase struct {
	ctrl     *gomock.Controller
	recorder *MockNoteUsecaseMockRecorder
}

type MockNoteUsecaseMockRecorder struct {
	mock *MockNoteUsecase
}

func NewMockNoteUsecase(ctrl *gomock.Controller) *MockNoteUsecase {
	mock := &MockNoteUsecase{ctrl: ctrl}
	mock.recorder = &MockNoteUsecaseMockRecorder{mock}
	return mock
}

func (m *MockNoteUsecase) EXPECT() *MockNoteUsecaseMockRecorder {
	return m.recorder
}

func (m *MockNoteUsecase) GetNote(ctx context.Context, noteID, userID uuid.UUID) (*models.Note, []models.Block, map[string]models.BlockFormatting, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNote", ctx, noteID, userID)
	ret0, _ := ret[0].(*models.Note)
	ret1, _ := ret[1].([]models.Block)
	ret2, _ := ret[2].(map[string]models.BlockFormatting)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

func (mr *MockNoteUsecaseMockRecorder) GetNote(ctx, noteID, userID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNote", reflect.TypeOf((*MockNoteUsecase)(nil).GetNote), ctx, noteID, userID)
}

func (m *MockNoteUsecase) GetBlock(ctx context.Context, blockID, noteID, userID uuid.UUID) (*models.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlock", ctx, blockID, noteID, userID)
	ret0, _ := ret[0].(*models.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockNoteUsecaseMockRecorder) GetBlock(ctx, blockID, noteID, userID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlock", reflect.TypeOf((*MockNoteUsecase)(nil).GetBlock), ctx, blockID, noteID, userID)
}

func (m *MockNoteUsecase) CreateBlock(ctx context.Context, noteID, userID uuid.UUID, block models.Block) (*models.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateBlock", ctx, noteID, userID, block)
	ret0, _ := ret[0].(*models.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockNoteUsecaseMockRecorder) CreateBlock(ctx, noteID, userID, block interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateBlock", reflect.TypeOf((*MockNoteUsecase)(nil).CreateBlock), ctx, noteID, userID, block)
}

func (m *MockNoteUsecase) ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition, direction int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ShiftBlockPositions", ctx, noteID, fromPosition, direction)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockNoteUsecaseMockRecorder) ShiftBlockPositions(ctx, noteID, fromPosition, direction interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShiftBlockPositions", reflect.TypeOf((*MockNoteUsecase)(nil).ShiftBlockPositions), ctx, noteID, fromPosition, direction)
}

func (m *MockNoteUsecase) DeleteBlock(ctx context.Context, blockID, noteID, userID uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteBlock", ctx, blockID, noteID, userID)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockNoteUsecaseMockRecorder) DeleteBlock(ctx, blockID, noteID, userID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteBlock", reflect.TypeOf((*MockNoteUsecase)(nil).DeleteBlock), ctx, blockID, noteID, userID)
}

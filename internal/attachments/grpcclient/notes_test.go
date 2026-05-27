package grpcclient

import (
	"context"
	"errors"
	"testing"

	notesgen "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/notes/grpc/gen"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type fakeNoteServiceClient struct {
	GetNoteFunc             func(ctx context.Context, in *notesgen.GetNoteRequest, opts ...grpc.CallOption) (*notesgen.NoteResponse, error)
	GetBlockFunc            func(ctx context.Context, in *notesgen.GetBlockRequest, opts ...grpc.CallOption) (*notesgen.BlockResponse, error)
	GetBlocksFunc           func(ctx context.Context, in *notesgen.GetBlocksRequest, opts ...grpc.CallOption) (*notesgen.GetBlocksResponse, error)
	CreateBlockFunc         func(ctx context.Context, in *notesgen.CreateBlockRequest, opts ...grpc.CallOption) (*notesgen.BlockResponse, error)
	ShiftBlockPositionsFunc func(ctx context.Context, in *notesgen.ShiftBlockPositionsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	DeleteBlockFunc         func(ctx context.Context, in *notesgen.DeleteBlockRequest, opts ...grpc.CallOption) (*notesgen.DeleteBlockResponse, error)
}

func (f *fakeNoteServiceClient) GetNote(ctx context.Context, in *notesgen.GetNoteRequest, opts ...grpc.CallOption) (*notesgen.NoteResponse, error) {
	return f.GetNoteFunc(ctx, in, opts...)
}

func (f *fakeNoteServiceClient) GetBlock(ctx context.Context, in *notesgen.GetBlockRequest, opts ...grpc.CallOption) (*notesgen.BlockResponse, error) {
	return f.GetBlockFunc(ctx, in, opts...)
}

func (f *fakeNoteServiceClient) GetBlocks(ctx context.Context, in *notesgen.GetBlocksRequest, opts ...grpc.CallOption) (*notesgen.GetBlocksResponse, error) {
	return f.GetBlocksFunc(ctx, in, opts...)
}

func (f *fakeNoteServiceClient) CreateBlock(ctx context.Context, in *notesgen.CreateBlockRequest, opts ...grpc.CallOption) (*notesgen.BlockResponse, error) {
	return f.CreateBlockFunc(ctx, in, opts...)
}

func (f *fakeNoteServiceClient) ShiftBlockPositions(ctx context.Context, in *notesgen.ShiftBlockPositionsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return f.ShiftBlockPositionsFunc(ctx, in, opts...)
}

func (f *fakeNoteServiceClient) DeleteBlock(ctx context.Context, in *notesgen.DeleteBlockRequest, opts ...grpc.CallOption) (*notesgen.DeleteBlockResponse, error) {
	return f.DeleteBlockFunc(ctx, in, opts...)
}

func TestGetNoteForwardsArgs(t *testing.T) {
	noteID := uuid.New()
	userID := uuid.New()
	fake := &fakeNoteServiceClient{
		GetNoteFunc: func(ctx context.Context, in *notesgen.GetNoteRequest, opts ...grpc.CallOption) (*notesgen.NoteResponse, error) {
			if in.NoteId != noteID.String() {
				t.Errorf("expected NoteId %s, got %s", noteID, in.NoteId)
			}
			if in.UserId != userID.String() {
				t.Errorf("expected UserId %s, got %s", userID, in.UserId)
			}
			return &notesgen.NoteResponse{Id: noteID.String()}, nil
		},
	}
	c := &notesServiceClient{client: fake}

	resp, err := c.GetNote(context.Background(), noteID, userID)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp.Id != noteID.String() {
		t.Fatalf("expected id %s, got %s", noteID, resp.Id)
	}
}

func TestGetBlockForwardsArgs(t *testing.T) {
	blockID := uuid.New()
	noteID := uuid.New()
	userID := uuid.New()
	fake := &fakeNoteServiceClient{
		GetBlockFunc: func(ctx context.Context, in *notesgen.GetBlockRequest, opts ...grpc.CallOption) (*notesgen.BlockResponse, error) {
			if in.BlockId != blockID.String() || in.NoteId != noteID.String() || in.UserId != userID.String() {
				t.Errorf("unexpected args: %+v", in)
			}
			return &notesgen.BlockResponse{Id: blockID.String()}, nil
		},
	}
	c := &notesServiceClient{client: fake}

	resp, err := c.GetBlock(context.Background(), blockID, noteID, userID)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp.Id != blockID.String() {
		t.Fatalf("expected id %s, got %s", blockID, resp.Id)
	}
}

func TestGetBlocksReturnsBlocks(t *testing.T) {
	noteID := uuid.New()
	userID := uuid.New()
	want := []*notesgen.BlockResponse{{Id: uuid.NewString()}, {Id: uuid.NewString()}}
	fake := &fakeNoteServiceClient{
		GetBlocksFunc: func(ctx context.Context, in *notesgen.GetBlocksRequest, opts ...grpc.CallOption) (*notesgen.GetBlocksResponse, error) {
			return &notesgen.GetBlocksResponse{Blocks: want}, nil
		},
	}
	c := &notesServiceClient{client: fake}

	got, err := c.GetBlocks(context.Background(), noteID, userID)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d blocks, got %d", len(want), len(got))
	}
}

func TestGetBlocksError(t *testing.T) {
	fake := &fakeNoteServiceClient{
		GetBlocksFunc: func(ctx context.Context, in *notesgen.GetBlocksRequest, opts ...grpc.CallOption) (*notesgen.GetBlocksResponse, error) {
			return nil, errors.New("boom")
		},
	}
	c := &notesServiceClient{client: fake}

	got, err := c.GetBlocks(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestCreateBlockForwardsArgs(t *testing.T) {
	userID := uuid.New()
	input := &notesgen.BlockResponse{NoteId: uuid.NewString(), Position: 1}
	fake := &fakeNoteServiceClient{
		CreateBlockFunc: func(ctx context.Context, in *notesgen.CreateBlockRequest, opts ...grpc.CallOption) (*notesgen.BlockResponse, error) {
			if in.UserId != userID.String() {
				t.Errorf("expected userId %s, got %s", userID, in.UserId)
			}
			if in.Block != input {
				t.Errorf("expected input block forwarded")
			}
			return &notesgen.BlockResponse{Id: uuid.NewString()}, nil
		},
	}
	c := &notesServiceClient{client: fake}

	resp, err := c.CreateBlock(context.Background(), userID, input)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestShiftBlockPositionsForwardsArgs(t *testing.T) {
	noteID := uuid.New()
	fake := &fakeNoteServiceClient{
		ShiftBlockPositionsFunc: func(ctx context.Context, in *notesgen.ShiftBlockPositionsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
			if in.NoteId != noteID.String() || in.FromPosition != 3 || in.Direction != -1 {
				t.Errorf("unexpected args: %+v", in)
			}
			return &emptypb.Empty{}, nil
		},
	}
	c := &notesServiceClient{client: fake}

	if err := c.ShiftBlockPositions(context.Background(), noteID, 3, -1); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestShiftBlockPositionsError(t *testing.T) {
	fake := &fakeNoteServiceClient{
		ShiftBlockPositionsFunc: func(ctx context.Context, in *notesgen.ShiftBlockPositionsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
			return nil, errors.New("boom")
		},
	}
	c := &notesServiceClient{client: fake}
	if err := c.ShiftBlockPositions(context.Background(), uuid.New(), 0, 1); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteBlockReturnsNoteUUID(t *testing.T) {
	noteID := uuid.New()
	blockID := uuid.New()
	userID := uuid.New()
	fake := &fakeNoteServiceClient{
		DeleteBlockFunc: func(ctx context.Context, in *notesgen.DeleteBlockRequest, opts ...grpc.CallOption) (*notesgen.DeleteBlockResponse, error) {
			if in.BlockId != blockID.String() || in.NoteId != noteID.String() || in.UserId != userID.String() {
				t.Errorf("unexpected args: %+v", in)
			}
			return &notesgen.DeleteBlockResponse{NoteId: noteID.String()}, nil
		},
	}
	c := &notesServiceClient{client: fake}

	got, err := c.DeleteBlock(context.Background(), blockID, noteID, userID)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != noteID {
		t.Fatalf("expected %s, got %s", noteID, got)
	}
}

func TestDeleteBlockError(t *testing.T) {
	fake := &fakeNoteServiceClient{
		DeleteBlockFunc: func(ctx context.Context, in *notesgen.DeleteBlockRequest, opts ...grpc.CallOption) (*notesgen.DeleteBlockResponse, error) {
			return nil, errors.New("boom")
		},
	}
	c := &notesServiceClient{client: fake}

	got, err := c.DeleteBlock(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %v", got)
	}
}

func TestDeleteBlockInvalidUUIDInResponse(t *testing.T) {
	fake := &fakeNoteServiceClient{
		DeleteBlockFunc: func(ctx context.Context, in *notesgen.DeleteBlockRequest, opts ...grpc.CallOption) (*notesgen.DeleteBlockResponse, error) {
			return &notesgen.DeleteBlockResponse{NoteId: "not-a-uuid"}, nil
		},
	}
	c := &notesServiceClient{client: fake}

	_, err := c.DeleteBlock(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected uuid parse error, got nil")
	}
}

func TestNewNotesServiceClientAndClose(t *testing.T) {
	client, err := NewNotesServiceClient("passthrough:///localhost:0")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if err := client.Close(); err != nil {
		t.Fatalf("unexpected close err: %v", err)
	}
}

package websocket

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	attachMocks "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/handler/http/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	noteMocks "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/handler/http/mocks"
	profileMocks "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/handler/http/mocks"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

type hubTestEnv struct {
	hub       *Hub
	noteUC    *noteMocks.MockNoteUsecase
	profileUC *profileMocks.MockProfileUsecase
	attachUC  *attachMocks.MockAttachmentUsecase
}

func newHubTestEnv(ctrl *gomock.Controller) *hubTestEnv {
	noteUC := noteMocks.NewMockNoteUsecase(ctrl)
	profileUC := profileMocks.NewMockProfileUsecase(ctrl)
	attachUC := attachMocks.NewMockAttachmentUsecase(ctrl)

	adapter := NewAttachmentUsecaseAdapter(attachUC.UploadAttachment)
	hub := NewHub(noteUC, profileUC, adapter)

	return &hubTestEnv{
		hub:       hub,
		noteUC:    noteUC,
		profileUC: profileUC,
		attachUC:  attachUC,
	}
}

func newTestClient(userID, userName, noteID string, bufSize int) *ClientInfo {
	return &ClientInfo{
		UserID:   userID,
		UserName: userName,
		NoteID:   noteID,
		Send:     make(chan WebSocketMessage, bufSize),
	}
}

func drainBroadcast(hub *Hub) []*BroadcastMessage {
	msgs := []*BroadcastMessage{}
	for {
		select {
		case m := <-hub.broadcast:
			msgs = append(msgs, m)
		default:
			return msgs
		}
	}
}

func collectMessages(ch chan WebSocketMessage, timeout time.Duration) []WebSocketMessage {
	msgs := []WebSocketMessage{}
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	for {
		select {
		case m, ok := <-ch:
			if !ok {
				return msgs
			}
			msgs = append(msgs, m)
		case <-deadline.C:
			return msgs
		}
	}
}

func TestNewHub(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	if env.hub.rooms == nil {
		t.Fatal("rooms map should be initialized")
	}
	if env.hub.register == nil || env.hub.unregister == nil || env.hub.broadcast == nil {
		t.Fatal("channels should be initialized")
	}
	if env.hub.storage == nil {
		t.Fatal("storage should be initialized")
	}
	if cap(env.hub.broadcast) != 512 {
		t.Errorf("expected broadcast cap 512, got %d", cap(env.hub.broadcast))
	}
}

func TestHub_HandleRegister_DeletedRoom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	room.IsDeleted = true
	env.hub.rooms[noteID] = room

	client := newTestClient(uuid.New().String(), "Alice", noteID, 4)
	env.hub.handleRegister(client)

	select {
	case msg := <-client.Send:
		if msg.Type != MsgError {
			t.Errorf("expected MsgError, got %s", msg.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("expected error message")
	}
}

func TestHub_HandleRegister_NewRoom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	blockID := uuid.New()
	blocks := []models.Block{{ID: blockID, BlockTypeID: 1, Content: "hi"}}
	note := &models.Note{ID: noteID, UserID: userID, Title: "t"}

	done := make(chan struct{})
	env.noteUC.EXPECT().
		GetNote(gomock.Any(), noteID, userID).
		DoAndReturn(func(_ context.Context, _, _ uuid.UUID) (*models.Note, []models.Block, map[string]models.BlockFormatting, error) {
			close(done)
			return note, blocks, map[string]models.BlockFormatting{}, nil
		})

	client := newTestClient(userID.String(), "Alice", noteID.String(), 8)
	env.hub.handleRegister(client)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("GetNote was not called from sendSyncState")
	}

	env.hub.mu.RLock()
	room, ok := env.hub.rooms[noteID.String()]
	env.hub.mu.RUnlock()
	if !ok {
		t.Fatal("room should be created")
	}
	if _, ok := room.GetClient(userID.String()); !ok {
		t.Fatal("client should be added to room")
	}

	msgs := drainBroadcast(env.hub)
	foundJoined := false
	for _, m := range msgs {
		if m.Message.Type == MsgUserJoined {
			foundJoined = true
		}
	}
	if !foundJoined {
		t.Fatal("expected MsgUserJoined broadcast")
	}
}

func TestHub_HandleUnregister_NoRoom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	client := newTestClient(uuid.New().String(), "x", uuid.New().String(), 1)
	env.hub.handleUnregister(client)
}

func TestHub_HandleUnregister_RemovesClientAndDeletesEmptyRoom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room
	client := newTestClient(uuid.New().String(), "Alice", noteID, 4)
	room.AddClient(client)

	env.hub.handleUnregister(client)

	if _, ok := room.GetClient(client.UserID); ok {
		t.Fatal("client should be removed")
	}

	env.hub.mu.RLock()
	_, exists := env.hub.rooms[noteID]
	env.hub.mu.RUnlock()
	if exists {
		t.Fatal("empty room should be deleted")
	}

	if _, ok := <-client.Send; ok {
		t.Fatal("send channel should be closed")
	}

	msgs := drainBroadcast(env.hub)
	foundLeft := false
	for _, m := range msgs {
		if m.Message.Type == MsgUserLeft {
			foundLeft = true
		}
	}
	if !foundLeft {
		t.Fatal("expected MsgUserLeft broadcast")
	}
}

func TestHub_HandleUnregister_KeepsRoomWithOtherClients(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room
	c1 := newTestClient(uuid.New().String(), "A", noteID, 4)
	c2 := newTestClient(uuid.New().String(), "B", noteID, 4)
	room.AddClient(c1)
	room.AddClient(c2)

	env.hub.handleUnregister(c1)

	env.hub.mu.RLock()
	_, exists := env.hub.rooms[noteID]
	env.hub.mu.RUnlock()
	if !exists {
		t.Fatal("room should remain because of other clients")
	}
}

func TestHub_HandleBroadcast_NonExistentRoom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	env.hub.handleBroadcast(&BroadcastMessage{NoteID: "nope", Message: WebSocketMessage{Type: MsgHeartbeat}})
}

func TestHub_HandleBroadcast_ExcludesUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room

	c1 := newTestClient("u1", "A", noteID, 4)
	c2 := newTestClient("u2", "B", noteID, 4)
	room.AddClient(c1)
	room.AddClient(c2)

	env.hub.handleBroadcast(&BroadcastMessage{
		NoteID:  noteID,
		Message: WebSocketMessage{Type: MsgHeartbeat},
		Exclude: "u1",
	})

	if len(c1.Send) != 0 {
		t.Fatal("excluded client should not receive")
	}
	if len(c2.Send) != 1 {
		t.Fatal("non-excluded client should receive")
	}
}

func TestHub_BroadcastToRoom_ChannelFullDoesNotPanic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	for i := 0; i < cap(env.hub.broadcast); i++ {
		env.hub.broadcast <- &BroadcastMessage{NoteID: "x"}
	}
	env.hub.broadcastToRoom("x", WebSocketMessage{Type: MsgHeartbeat}, "")
}

func TestHub_SendSyncState_InvalidNoteID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := "not-a-uuid"
	room := NewNoteRoom(noteID)
	client := newTestClient(uuid.New().String(), "A", noteID, 4)
	env.hub.sendSyncState(client, room)

	msg := <-client.Send
	if msg.Type != MsgError {
		t.Fatalf("expected MsgError, got %s", msg.Type)
	}
}

func TestHub_SendSyncState_InvalidUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	client := newTestClient("not-a-uuid", "A", noteID, 4)
	env.hub.sendSyncState(client, room)

	msg := <-client.Send
	if msg.Type != MsgError {
		t.Fatalf("expected MsgError, got %s", msg.Type)
	}
}

func TestHub_SendSyncState_GetNoteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	env.noteUC.EXPECT().
		GetNote(gomock.Any(), noteID, userID).
		Return(nil, nil, nil, errors.New("boom"))

	room := NewNoteRoom(noteID.String())
	client := newTestClient(userID.String(), "A", noteID.String(), 4)
	env.hub.sendSyncState(client, room)

	msg := <-client.Send
	if msg.Type != MsgError {
		t.Fatalf("expected MsgError, got %s", msg.Type)
	}
}

func TestHub_SendSyncState_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	blockID := uuid.New()
	blocks := []models.Block{{ID: blockID, BlockTypeID: 1, Content: "hello"}}

	env.noteUC.EXPECT().
		GetNote(gomock.Any(), noteID, userID).
		Return(&models.Note{ID: noteID, UserID: userID}, blocks, map[string]models.BlockFormatting{}, nil)

	room := NewNoteRoom(noteID.String())
	client := newTestClient(userID.String(), "A", noteID.String(), 4)
	room.AddClient(client)

	env.hub.sendSyncState(client, room)

	msg := <-client.Send
	if msg.Type != MsgSyncState {
		t.Fatalf("expected MsgSyncState, got %s", msg.Type)
	}

	if _, ok := room.GetCRDTDocument(blockID.String()); !ok {
		t.Fatal("expected CRDT document to be created for text block")
	}

	cursor := client.GetCursor()
	if cursor.BlockID != blockID.String() {
		t.Errorf("expected cursor on block %s, got %s", blockID, cursor.BlockID)
	}
}

func TestHub_HandleOperation_NoRoom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	env.hub.HandleOperation("nope", "u1", WebSocketMessage{Type: MsgCursorMove})
}

func TestHub_HandleOperation_DeletedRoom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	room.IsDeleted = true
	env.hub.rooms[noteID] = room

	env.hub.HandleOperation(noteID, "u1", WebSocketMessage{Type: MsgCursorMove})
}

func TestHub_HandleCursorMove(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room
	c := newTestClient("u1", "A", noteID, 4)
	room.AddClient(c)

	cursor := CursorPosition{BlockID: "b1", Position: 7}
	env.hub.HandleOperation(noteID, "u1", WebSocketMessage{
		Type: MsgCursorMove,
		Msg:  cursor,
	})

	if c.GetCursor().Position != 7 {
		t.Errorf("expected position 7, got %d", c.GetCursor().Position)
	}

	msgs := drainBroadcast(env.hub)
	if len(msgs) == 0 {
		t.Fatal("expected broadcast")
	}
}

func TestHub_HandleCursorMove_ClientNotInRoom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room

	env.hub.HandleOperation(noteID, "ghost", WebSocketMessage{
		Type: MsgCursorMove,
		Msg:  CursorPosition{Position: 1},
	})
}

func TestHub_HandleInsertChar_Local(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	blockID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room
	c := newTestClient("u1", "A", noteID, 4)
	room.AddClient(c)

	op := InsertCharOperation{
		BlockID: blockID,
		Char:    "x",
	}

	env.hub.HandleOperation(noteID, "u1", WebSocketMessage{
		Type:    MsgInsertChar,
		IsLocal: true,
		Msg:     op,
	})

	doc, ok := room.GetCRDTDocument(blockID)
	if !ok {
		t.Fatal("expected CRDT doc to be created")
	}
	if doc.GetText() != "x" {
		t.Errorf("expected text 'x', got %q", doc.GetText())
	}

	msgs := drainBroadcast(env.hub)
	if len(msgs) < 1 {
		t.Fatal("expected at least one broadcast")
	}
}

func TestHub_HandleInsertChar_Remote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	blockID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room
	doc := NewCRDTDocument("u1")
	room.SetCRDTDocument(blockID, doc)

	op := InsertCharOperation{
		BlockID:  blockID,
		Char:     "y",
		UniqueID: "u1:1:1",
		PrevID:   "root:0:0",
		Lamport:  10,
		UserID:   "u1",
	}

	env.hub.HandleOperation(noteID, "u1", WebSocketMessage{
		Type:    MsgInsertChar,
		IsLocal: false,
		Msg:     op,
	})

	if doc.GetText() != "y" {
		t.Errorf("expected 'y', got %q", doc.GetText())
	}
}

func TestHub_HandleInsertChar_RemoteNoDocNoop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room

	env.hub.HandleOperation(noteID, "u1", WebSocketMessage{
		Type:    MsgInsertChar,
		IsLocal: false,
		Msg:     InsertCharOperation{BlockID: "no-doc", Char: "y", UniqueID: "z", PrevID: "root:0:0"},
	})
}

func TestHub_HandleInsertChar_LocalClientMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room

	env.hub.HandleOperation(noteID, "ghost", WebSocketMessage{
		Type:    MsgInsertChar,
		IsLocal: true,
		Msg:     InsertCharOperation{BlockID: "b", Char: "z"},
	})
}

func TestHub_HandleDeleteChar_Local(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	blockID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room
	c := newTestClient("u1", "A", noteID, 4)
	room.AddClient(c)

	doc := NewCRDTDocument("u1")
	doc.LoadText("ab", "u1")
	room.SetCRDTDocument(blockID, doc)
	c.UpdateCursor(CursorPosition{BlockID: blockID, Position: 2})

	env.hub.HandleOperation(noteID, "u1", WebSocketMessage{
		Type:    MsgDeleteChar,
		IsLocal: true,
		Msg:     DeleteCharOperation{BlockID: blockID},
	})

	if doc.GetText() != "a" {
		t.Errorf("expected 'a', got %q", doc.GetText())
	}
}

func TestHub_HandleDeleteChar_Local_NoCursor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	blockID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room
	c := newTestClient("u1", "A", noteID, 4)
	room.AddClient(c)
	doc := NewCRDTDocument("u1")
	room.SetCRDTDocument(blockID, doc)

	env.hub.HandleOperation(noteID, "u1", WebSocketMessage{
		Type:    MsgDeleteChar,
		IsLocal: true,
		Msg:     DeleteCharOperation{BlockID: blockID},
	})
}

func TestHub_HandleDeleteChar_LocalNoDoc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room
	c := newTestClient("u1", "A", noteID, 4)
	room.AddClient(c)

	env.hub.HandleOperation(noteID, "u1", WebSocketMessage{
		Type:    MsgDeleteChar,
		IsLocal: true,
		Msg:     DeleteCharOperation{BlockID: "no-doc"},
	})
}

func TestHub_HandleDeleteChar_Remote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	blockID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room
	doc := NewCRDTDocument("u1")
	doc.LoadText("abc", "u1")
	room.SetCRDTDocument(blockID, doc)

	state := doc.GetCRDTState()
	delID := state[0].ID

	env.hub.HandleOperation(noteID, "u1", WebSocketMessage{
		Type:    MsgDeleteChar,
		IsLocal: false,
		Msg:     DeleteCharOperation{BlockID: blockID, UniqueID: delID, Lamport: 100},
	})

	if doc.GetText() != "bc" {
		t.Errorf("expected 'bc', got %q", doc.GetText())
	}
}

func TestHub_HandleApplyFormatting_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	blockID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	c := newTestClient(userID.String(), "A", noteID.String(), 4)
	room.AddClient(c)

	done := make(chan struct{})
	env.noteUC.EXPECT().
		UpdateBlockFormatting(gomock.Any(), blockID, noteID, userID, gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _, _ uuid.UUID, _ models.FormattingRange) (*models.BlockFormatting, error) {
			close(done)
			return &models.BlockFormatting{}, nil
		})

	boldTrue := true
	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgApplyFormatting,
		Msg: FormattingOperation{
			BlockID:  blockID.String(),
			StartPos: 0,
			EndPos:   3,
			Bold:     &boldTrue,
		},
	})

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("UpdateBlockFormatting was not called")
	}

	if room.SequenceID != 1 {
		t.Errorf("expected SequenceID 1, got %d", room.SequenceID)
	}
}

func TestHub_HandleApplyFormatting_ErrorSentToClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	blockID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	c := newTestClient(userID.String(), "A", noteID.String(), 8)
	room.AddClient(c)

	env.noteUC.EXPECT().
		UpdateBlockFormatting(gomock.Any(), blockID, noteID, userID, gomock.Any()).
		Return(nil, errors.New("nope"))

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgApplyFormatting,
		Msg: FormattingOperation{
			BlockID: blockID.String(),
		},
	})

	select {
	case msg := <-c.Send:
		if msg.Type != MsgError {
			t.Errorf("expected MsgError, got %s", msg.Type)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected error message")
	}
}

func TestHub_HandleApplyFormatting_ClientMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room

	env.hub.HandleOperation(noteID.String(), uuid.New().String(), WebSocketMessage{
		Type: MsgApplyFormatting,
		Msg:  FormattingOperation{BlockID: uuid.New().String()},
	})
}

func TestHub_HandleCreateBlock_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	createdBlockID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	c := newTestClient(userID.String(), "A", noteID.String(), 4)
	room.AddClient(c)

	env.noteUC.EXPECT().
		CreateBlock(gomock.Any(), noteID, userID, gomock.Any()).
		Return(&models.Block{ID: createdBlockID, BlockTypeID: 1, Position: 0}, nil)

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgCreateBlock,
		Msg:  CreateBlockOperation{BlockTypeID: 1, Position: 0},
	})

	if _, ok := room.GetCRDTDocument(createdBlockID.String()); !ok {
		t.Fatal("expected CRDT doc for text block")
	}
}

func TestHub_HandleCreateBlock_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	c := newTestClient(userID.String(), "A", noteID.String(), 4)
	room.AddClient(c)

	env.noteUC.EXPECT().
		CreateBlock(gomock.Any(), noteID, userID, gomock.Any()).
		Return(nil, errors.New("boom"))

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgCreateBlock,
		Msg:  CreateBlockOperation{BlockTypeID: 1},
	})

	msg := <-c.Send
	if msg.Type != MsgError {
		t.Fatalf("expected MsgError, got %s", msg.Type)
	}
}

func TestHub_HandleDeleteBlock_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	blockID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	c := newTestClient(userID.String(), "A", noteID.String(), 4)
	room.AddClient(c)
	room.SetCRDTDocument(blockID.String(), NewCRDTDocument(userID.String()))

	env.noteUC.EXPECT().
		DeleteBlock(gomock.Any(), blockID, noteID, userID).
		Return(nil)

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgDeleteBlock,
		Msg:  DeleteBlockOperation{BlockID: blockID.String()},
	})

	if _, ok := room.GetCRDTDocument(blockID.String()); ok {
		t.Fatal("CRDT doc should be deleted")
	}
}

func TestHub_HandleDeleteBlock_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	blockID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	c := newTestClient(userID.String(), "A", noteID.String(), 4)
	room.AddClient(c)

	env.noteUC.EXPECT().
		DeleteBlock(gomock.Any(), blockID, noteID, userID).
		Return(errors.New("nope"))

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgDeleteBlock,
		Msg:  DeleteBlockOperation{BlockID: blockID.String()},
	})

	msg := <-c.Send
	if msg.Type != MsgError {
		t.Fatalf("expected MsgError, got %s", msg.Type)
	}
}

func TestHub_HandleMoveBlock_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	blockID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	c := newTestClient(userID.String(), "A", noteID.String(), 4)
	room.AddClient(c)

	env.noteUC.EXPECT().
		MoveBlock(gomock.Any(), blockID, noteID, userID, 3).
		Return(&models.Block{}, nil)

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgMoveBlock,
		Msg:  MoveBlockOperation{BlockID: blockID.String(), NewPosition: 3},
	})
}

func TestHub_HandleMoveBlock_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	blockID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	c := newTestClient(userID.String(), "A", noteID.String(), 4)
	room.AddClient(c)

	env.noteUC.EXPECT().
		MoveBlock(gomock.Any(), blockID, noteID, userID, 3).
		Return(nil, errors.New("nope"))

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgMoveBlock,
		Msg:  MoveBlockOperation{BlockID: blockID.String(), NewPosition: 3},
	})

	msg := <-c.Send
	if msg.Type != MsgError {
		t.Fatalf("expected MsgError, got %s", msg.Type)
	}
}

func TestHub_HandleUpdateNoteTitle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	c := newTestClient(userID.String(), "A", noteID.String(), 4)
	room.AddClient(c)

	note := &models.Note{ID: noteID, UserID: userID, Title: "old"}
	env.noteUC.EXPECT().GetNote(gomock.Any(), noteID, userID).Return(note, nil, nil, nil)
	env.noteUC.EXPECT().
		UpdateNote(gomock.Any(), noteID, userID, gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ uuid.UUID, n models.Note) (*models.Note, error) {
			if n.Title != "new title" {
				t.Errorf("expected updated title, got %q", n.Title)
			}
			return &n, nil
		})

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgUpdateNoteTitle,
		Msg:  "new title",
	})
}

func TestHub_HandleUpdateNoteTitle_GetNoteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	c := newTestClient(userID.String(), "A", noteID.String(), 4)
	room.AddClient(c)

	env.noteUC.EXPECT().GetNote(gomock.Any(), noteID, userID).Return(nil, nil, nil, errors.New("boom"))

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgUpdateNoteTitle,
		Msg:  "new title",
	})

	msg := <-c.Send
	if msg.Type != MsgError {
		t.Fatalf("expected MsgError, got %s", msg.Type)
	}
}

func TestHub_HandleUpdateNoteTitle_UpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	c := newTestClient(userID.String(), "A", noteID.String(), 4)
	room.AddClient(c)

	env.noteUC.EXPECT().GetNote(gomock.Any(), noteID, userID).Return(&models.Note{ID: noteID, UserID: userID}, nil, nil, nil)
	env.noteUC.EXPECT().UpdateNote(gomock.Any(), noteID, userID, gomock.Any()).Return(nil, errors.New("nope"))

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgUpdateNoteTitle,
		Msg:  "x",
	})

	msg := <-c.Send
	if msg.Type != MsgError {
		t.Fatalf("expected MsgError, got %s", msg.Type)
	}
}

func TestHub_HandleUpdateNotePublic_NotOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	otherUser := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	c := newTestClient(userID.String(), "A", noteID.String(), 4)
	room.AddClient(c)

	env.noteUC.EXPECT().
		GetNote(gomock.Any(), noteID, userID).
		Return(&models.Note{ID: noteID, UserID: otherUser}, nil, nil, nil)

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgUpdateNotePublic,
		Msg:  false,
	})
}

func TestHub_HandleUpdateNotePublic_OwnerMakePrivate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	owner := newTestClient(userID.String(), "Owner", noteID.String(), 4)
	other := newTestClient(uuid.New().String(), "Guest", noteID.String(), 4)
	room.AddClient(owner)
	room.AddClient(other)

	note := &models.Note{ID: noteID, UserID: userID, IsPublic: true}
	env.noteUC.EXPECT().GetNote(gomock.Any(), noteID, userID).Return(note, nil, nil, nil).Times(2)
	env.noteUC.EXPECT().
		UpdateNote(gomock.Any(), noteID, userID, gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ uuid.UUID, n models.Note) (*models.Note, error) {
			if n.IsPublic {
				t.Error("note should be marked private")
			}
			return &n, nil
		})

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgUpdateNotePublic,
		Msg:  false,
	})

	for _, c := range []*ClientInfo{owner, other} {
		msgs := collectMessages(c.Send, time.Second)
		foundPrivate := false
		for _, m := range msgs {
			if m.Type == MsgNotePrivate {
				foundPrivate = true
			}
		}
		if !foundPrivate {
			t.Errorf("client %s did not receive MsgNotePrivate", c.UserID)
		}
	}
}

func TestHub_HandleUpdateNotePublic_OwnerMakePublic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	owner := newTestClient(userID.String(), "Owner", noteID.String(), 4)
	room.AddClient(owner)

	note := &models.Note{ID: noteID, UserID: userID, IsPublic: false}
	env.noteUC.EXPECT().GetNote(gomock.Any(), noteID, userID).Return(note, nil, nil, nil).Times(2)
	env.noteUC.EXPECT().UpdateNote(gomock.Any(), noteID, userID, gomock.Any()).Return(note, nil)

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgUpdateNotePublic,
		Msg:  true,
	})

	if _, ok := room.GetClient(userID.String()); !ok {
		t.Fatal("client should remain when making public")
	}
}

func TestHub_HandleUpdateNotePublic_OwnerGetNoteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	owner := newTestClient(userID.String(), "Owner", noteID.String(), 4)
	room.AddClient(owner)

	first := env.noteUC.EXPECT().
		GetNote(gomock.Any(), noteID, userID).
		Return(&models.Note{ID: noteID, UserID: userID}, nil, nil, nil)
	env.noteUC.EXPECT().
		GetNote(gomock.Any(), noteID, userID).
		Return(nil, nil, nil, errors.New("boom")).
		After(first)

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgUpdateNotePublic,
		Msg:  false,
	})
}

func TestHub_HandleUpdateNotePublic_OwnerUpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	owner := newTestClient(userID.String(), "Owner", noteID.String(), 4)
	room.AddClient(owner)

	note := &models.Note{ID: noteID, UserID: userID}
	env.noteUC.EXPECT().GetNote(gomock.Any(), noteID, userID).Return(note, nil, nil, nil).Times(2)
	env.noteUC.EXPECT().UpdateNote(gomock.Any(), noteID, userID, gomock.Any()).Return(nil, errors.New("nope"))

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgUpdateNotePublic,
		Msg:  false,
	})
}

func TestHub_HandleUploadAttachment_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	c := newTestClient(userID.String(), "A", noteID.String(), 4)
	room.AddClient(c)

	pngHeader := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	data := append(pngHeader, make([]byte, 200)...)

	env.attachUC.EXPECT().
		UploadAttachment(gomock.Any(), noteID, userID, "f.png", int64(len(data)), "image/png", gomock.Any(), true, 1).
		DoAndReturn(func(_ context.Context, _, _ uuid.UUID, _ string, _ int64, _ string, r io.Reader, _ bool, _ int) (*models.Attachment, error) {
			_, _ = io.ReadAll(r)
			return &models.Attachment{ID: uuid.New(), BlockID: uuid.New()}, nil
		})

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgUploadAttachment,
		Msg: UploadAttachmentOperation{
			FileName:    "f.png",
			FileData:    data,
			HasPosition: true,
			Position:    1,
		},
	})
}

func TestHub_HandleUploadAttachment_InvalidMime(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	c := newTestClient(userID.String(), "A", noteID.String(), 4)
	room.AddClient(c)

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgUploadAttachment,
		Msg: UploadAttachmentOperation{
			FileName: "a.bin",
			FileData: []byte("not a valid file with html<html><body>test</body></html> contents"),
		},
	})

	msg := <-c.Send
	if msg.Type != MsgError {
		t.Fatalf("expected error, got %s", msg.Type)
	}
}

func TestHub_HandleUploadAttachment_FileTooLarge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	c := newTestClient(userID.String(), "A", noteID.String(), 4)
	room.AddClient(c)

	data := make([]byte, MAX_IMAGE_SIZE+1)
	copy(data, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'})

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgUploadAttachment,
		Msg: UploadAttachmentOperation{
			FileName: "big.png",
			FileData: data,
		},
	})

	msg := <-c.Send
	if msg.Type != MsgError {
		t.Fatalf("expected error, got %s", msg.Type)
	}
}

func TestHub_HandleUploadAttachment_UploadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	c := newTestClient(userID.String(), "A", noteID.String(), 4)
	room.AddClient(c)

	pngHeader := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	data := append(pngHeader, make([]byte, 100)...)

	env.attachUC.EXPECT().
		UploadAttachment(gomock.Any(), noteID, userID, "f.png", int64(len(data)), "image/png", gomock.Any(), false, 0).
		Return(nil, errors.New("upload fail"))

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgUploadAttachment,
		Msg: UploadAttachmentOperation{
			FileName: "f.png",
			FileData: data,
		},
	})

	msg := <-c.Send
	if msg.Type != MsgError {
		t.Fatalf("expected error, got %s", msg.Type)
	}
}

func TestHub_HandleUploadAttachment_ClientMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room

	env.hub.HandleOperation(noteID.String(), uuid.New().String(), WebSocketMessage{
		Type: MsgUploadAttachment,
		Msg:  UploadAttachmentOperation{FileName: "f.png", FileData: []byte("data")},
	})
}

func TestHub_HandleDeleteNote_NotOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	otherUser := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room

	env.noteUC.EXPECT().
		GetNote(gomock.Any(), noteID, userID).
		Return(&models.Note{ID: noteID, UserID: otherUser}, nil, nil, nil)

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgDeleteNote,
	})

	env.hub.mu.RLock()
	_, exists := env.hub.rooms[noteID.String()]
	env.hub.mu.RUnlock()
	if !exists {
		t.Fatal("room should remain when caller is not owner")
	}
}

func TestHub_HandleDeleteNote_OwnerSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room
	c1 := newTestClient(userID.String(), "A", noteID.String(), 4)
	c2 := newTestClient(uuid.New().String(), "B", noteID.String(), 4)
	room.AddClient(c1)
	room.AddClient(c2)

	env.noteUC.EXPECT().
		GetNote(gomock.Any(), noteID, userID).
		Return(&models.Note{ID: noteID, UserID: userID}, nil, nil, nil)
	env.noteUC.EXPECT().DeleteNote(gomock.Any(), noteID, userID).Return(nil)

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgDeleteNote,
	})

	env.hub.mu.RLock()
	_, exists := env.hub.rooms[noteID.String()]
	env.hub.mu.RUnlock()
	if exists {
		t.Fatal("room should be removed after delete")
	}

	for _, c := range []*ClientInfo{c1, c2} {
		msgs := collectMessages(c.Send, time.Second)
		foundDeleted := false
		for _, m := range msgs {
			if m.Type == MsgNoteDeleted {
				foundDeleted = true
			}
		}
		if !foundDeleted {
			t.Errorf("client %s did not get MsgNoteDeleted", c.UserID)
		}
	}
}

func TestHub_HandleDeleteNote_OwnerDeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	room := NewNoteRoom(noteID.String())
	env.hub.rooms[noteID.String()] = room

	env.noteUC.EXPECT().
		GetNote(gomock.Any(), noteID, userID).
		Return(&models.Note{ID: noteID, UserID: userID}, nil, nil, nil)
	env.noteUC.EXPECT().DeleteNote(gomock.Any(), noteID, userID).Return(errors.New("nope"))

	env.hub.HandleOperation(noteID.String(), userID.String(), WebSocketMessage{
		Type: MsgDeleteNote,
	})

	env.hub.mu.RLock()
	_, exists := env.hub.rooms[noteID.String()]
	env.hub.mu.RUnlock()
	if !exists {
		t.Fatal("room should remain when delete fails")
	}
}

func TestHub_HandleHeartbeat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room
	c := newTestClient("u1", "A", noteID, 4)
	room.AddClient(c)

	before := c.LastPing
	env.hub.HandleOperation(noteID, "u1", WebSocketMessage{Type: MsgHeartbeat})
	if c.LastPing <= before {
		t.Fatal("expected LastPing to advance")
	}
}

func TestHub_HandleHeartbeat_NoClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room

	env.hub.HandleOperation(noteID, "ghost", WebSocketMessage{Type: MsgHeartbeat})
}

func TestHub_IsNoteOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	other := uuid.New()

	env.noteUC.EXPECT().
		GetNote(gomock.Any(), noteID, userID).
		Return(&models.Note{ID: noteID, UserID: userID}, nil, nil, nil)
	if !env.hub.isNoteOwner(noteID.String(), userID.String()) {
		t.Fatal("expected true when matching")
	}

	env.noteUC.EXPECT().
		GetNote(gomock.Any(), noteID, userID).
		Return(&models.Note{ID: noteID, UserID: other}, nil, nil, nil)
	if env.hub.isNoteOwner(noteID.String(), userID.String()) {
		t.Fatal("expected false when not matching")
	}

	env.noteUC.EXPECT().
		GetNote(gomock.Any(), noteID, userID).
		Return(nil, nil, nil, errors.New("boom"))
	if env.hub.isNoteOwner(noteID.String(), userID.String()) {
		t.Fatal("expected false on error")
	}
}

func TestHub_CleanupInactiveRooms(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	oldEmpty := NewNoteRoom("old-empty")
	oldEmpty.CreatedAt = time.Now().Unix() - 1000
	env.hub.rooms["old-empty"] = oldEmpty

	oldWithClient := NewNoteRoom("old-with-client")
	oldWithClient.CreatedAt = time.Now().Unix() - 1000
	oldWithClient.AddClient(newTestClient("u", "U", "old-with-client", 1))
	env.hub.rooms["old-with-client"] = oldWithClient

	fresh := NewNoteRoom("fresh")
	env.hub.rooms["fresh"] = fresh

	env.hub.cleanupInactiveRooms()

	env.hub.mu.RLock()
	defer env.hub.mu.RUnlock()
	if _, ok := env.hub.rooms["old-empty"]; ok {
		t.Error("old empty room should be cleaned up")
	}
	if _, ok := env.hub.rooms["old-with-client"]; !ok {
		t.Error("old room with clients should not be cleaned up")
	}
	if _, ok := env.hub.rooms["fresh"]; !ok {
		t.Error("fresh room should not be cleaned up")
	}
}

func TestHub_SendHeartbeats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room
	c := newTestClient("u1", "A", noteID, 4)
	room.AddClient(c)

	env.hub.sendHeartbeats()

	select {
	case msg := <-c.Send:
		if msg.Type != MsgHeartbeat {
			t.Errorf("expected heartbeat, got %s", msg.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("expected heartbeat message")
	}
}

func TestHub_SendHeartbeats_FullBufferDropped(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room
	c := newTestClient("u1", "A", noteID, 1)
	room.AddClient(c)
	c.Send <- WebSocketMessage{Type: MsgError}

	env.hub.sendHeartbeats()

	if len(c.Send) != 1 {
		t.Errorf("expected the existing message to remain (full buffer drops heartbeat), got %d", len(c.Send))
	}
}

func TestHub_Shutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room
	c := newTestClient("u1", "A", noteID, 4)
	room.AddClient(c)

	env.hub.shutdown()

	if _, ok := <-c.Send; ok {
		t.Fatal("send channel should be closed")
	}
}

func TestHub_Run_ShutsDownOnCtxDone(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room
	c := newTestClient("u1", "A", noteID, 4)
	room.AddClient(c)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		env.hub.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after ctx cancel")
	}

	if _, ok := <-c.Send; ok {
		t.Fatal("client send should be closed after shutdown")
	}
}

func TestHub_Run_RegisterUnregisterRoundTrip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New()
	userID := uuid.New()
	blockID := uuid.New()

	env.noteUC.EXPECT().
		GetNote(gomock.Any(), noteID, userID).
		Return(
			&models.Note{ID: noteID, UserID: userID},
			[]models.Block{{ID: blockID, BlockTypeID: 1}},
			map[string]models.BlockFormatting{},
			nil,
		).
		AnyTimes()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go env.hub.Run(ctx)

	client := newTestClient(userID.String(), "Alice", noteID.String(), 8)
	env.hub.register <- client

	// Wait for the room to exist AND for sendSyncState to finish delivering
	// (MsgSyncState in the buffer) so we don't race the goroutine when we
	// unregister the client and the hub closes Send.
	deadline := time.Now().Add(2 * time.Second)
	sawSync := false
	for time.Now().Before(deadline) && !sawSync {
		select {
		case msg := <-client.Send:
			if msg.Type == MsgSyncState {
				sawSync = true
			}
		case <-time.After(10 * time.Millisecond):
		}
	}
	if !sawSync {
		t.Fatal("did not observe sync state delivery")
	}
	env.hub.mu.RLock()
	_, ok := env.hub.rooms[noteID.String()]
	env.hub.mu.RUnlock()
	if !ok {
		t.Fatal("room was not created via register channel")
	}

	env.hub.unregister <- client
	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		env.hub.mu.RLock()
		_, stillThere := env.hub.rooms[noteID.String()]
		env.hub.mu.RUnlock()
		if !stillThere {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	env.hub.mu.RLock()
	_, stillThere := env.hub.rooms[noteID.String()]
	env.hub.mu.RUnlock()
	if stillThere {
		t.Fatal("room should be removed after last client unregister")
	}
}

func TestHub_ErrorMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	c := &ClientInfo{UserID: "u1", NoteID: "n1"}
	msg := env.hub.errorMessage("oops", c)

	if msg.Type != MsgError {
		t.Errorf("expected MsgError, got %s", msg.Type)
	}
	if msg.UserID != "u1" || msg.NoteID != "n1" {
		t.Error("expected user/note IDs to be set")
	}
	m, ok := msg.Msg.(map[string]string)
	if !ok || m["error"] != "oops" {
		t.Error("expected error string")
	}
}

func TestMapToStruct_Success(t *testing.T) {
	type tgt struct {
		Foo string `json:"foo"`
	}
	src := map[string]string{"foo": "bar"}
	var dest tgt
	if err := mapToStruct(src, &dest); err != nil {
		t.Fatal(err)
	}
	if dest.Foo != "bar" {
		t.Errorf("expected 'bar', got %q", dest.Foo)
	}
}

func TestMapToStruct_MarshalError(t *testing.T) {
	src := make(chan int)
	var dest map[string]string
	if err := mapToStruct(src, &dest); err == nil {
		t.Fatal("expected error marshaling channel")
	}
}

func TestGetMaxSizeByMimeType(t *testing.T) {
	cases := map[string]int64{
		"image/png":  MAX_IMAGE_SIZE,
		"image/gif":  MAX_GIF_SIZE,
		"audio/mpeg": MAX_AUDIO_SIZE,
		"video/mp4":  MAX_VIDEO_SIZE,
	}
	for mime, want := range cases {
		got, err := getMaxSizeByMimeType(mime)
		if err != nil {
			t.Errorf("mime %s: unexpected error %v", mime, err)
		}
		if got != want {
			t.Errorf("mime %s: expected %d, got %d", mime, want, got)
		}
	}

	if _, err := getMaxSizeByMimeType("application/x-bogus"); err == nil {
		t.Fatal("expected ErrInvalidMimeType for unsupported type")
	}
}

func TestHub_UpdateCursorsAfterOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	blockID := "b1"
	room := NewNoteRoom(noteID)
	c := newTestClient("u1", "A", noteID, 4)
	c.UpdateCursor(CursorPosition{BlockID: blockID, Position: 5})
	room.AddClient(c)

	op := &InsertCharOperation{BlockID: blockID, Position: 2}
	env.hub.updateCursorsAfterOperation(room, MsgInsertChar, op, blockID)

	if c.GetCursor().Position != 6 {
		t.Errorf("expected cursor at 6 after insert at 2, got %d", c.GetCursor().Position)
	}
}

// Sanity check that the Hub satisfies the websocket interfaces by exercising
// HandleOperation across all message types with prepared rooms; mostly serves as
// a compile-time / smoke test that no panic occurs on unknown messages.
func TestHub_HandleOperation_UnknownTypeNoop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room

	env.hub.HandleOperation(noteID, "u1", WebSocketMessage{Type: MessageType("not_a_real_type")})
}

// Guarantees that AttachmentUsecaseAdapter and ProfileUsecaseInterface are wired in
// the test env even though we don't call them in most tests.
func TestHub_AttachmentAdapterAndProfileWired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	if env.hub.attachmentUsecase == nil {
		t.Fatal("attachment usecase should be set")
	}
	if env.hub.profileUsecase == nil {
		t.Fatal("profile usecase should be set")
	}

	// Trigger profile mock once so generated EXPECT helpers are usable from tests in the future.
	env.profileUC.EXPECT().GetProfile(gomock.Any(), gomock.Any()).Return(&models.Profile{}, nil)
	if _, err := env.hub.profileUsecase.GetProfile(context.Background(), uuid.New()); err != nil {
		t.Fatal(err)
	}
}

// Race smoke test: concurrent broadcasts to the same room shouldn't panic.
func TestHub_HandleBroadcast_Concurrent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	noteID := uuid.New().String()
	room := NewNoteRoom(noteID)
	env.hub.rooms[noteID] = room
	c := newTestClient("u1", "A", noteID, 32)
	room.AddClient(c)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			env.hub.handleBroadcast(&BroadcastMessage{
				NoteID:  noteID,
				Message: WebSocketMessage{Type: MsgHeartbeat},
			})
		}()
	}
	wg.Wait()

	if len(c.Send) == 0 {
		t.Fatal("expected at least one broadcast delivered")
	}
}

func TestHub_ErrorMessageContainsErrorText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newHubTestEnv(ctrl)

	c := &ClientInfo{UserID: "u1", NoteID: "n1"}
	msg := env.hub.errorMessage("specific reason", c)
	m, ok := msg.Msg.(map[string]string)
	if !ok || !strings.Contains(m["error"], "specific reason") {
		t.Errorf("expected error to contain 'specific reason', got %v", msg.Msg)
	}
}

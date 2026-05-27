package websocket

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

// Mock implementations

type MockNoteUsecase struct {
	GetNoteFunc               func(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, []models.Block, map[string]models.BlockFormatting, error)
	UpdateNoteFunc            func(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, note models.Note) (*models.Note, error)
	DeleteNoteFunc            func(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) error
	CreateBlockFunc           func(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, block models.Block) (*models.Block, error)
	DeleteBlockFunc           func(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) error
	MoveBlockFunc             func(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, newPosition int) (*models.Block, error)
	UpdateBlockContentFunc    func(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, content string) (*models.Block, error)
	UpdateBlockFormattingFunc func(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, formattingRange models.FormattingRange) (*models.BlockFormatting, error)
}

func (m *MockNoteUsecase) GetNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, []models.Block, map[string]models.BlockFormatting, error) {
	if m.GetNoteFunc != nil {
		return m.GetNoteFunc(ctx, noteID, userID)
	}
	return nil, nil, nil, nil
}

func (m *MockNoteUsecase) UpdateNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, note models.Note) (*models.Note, error) {
	if m.UpdateNoteFunc != nil {
		return m.UpdateNoteFunc(ctx, noteID, userID, note)
	}
	return nil, nil
}

func (m *MockNoteUsecase) DeleteNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) error {
	if m.DeleteNoteFunc != nil {
		return m.DeleteNoteFunc(ctx, noteID, userID)
	}
	return nil
}

func (m *MockNoteUsecase) CreateBlock(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, block models.Block) (*models.Block, error) {
	if m.CreateBlockFunc != nil {
		return m.CreateBlockFunc(ctx, noteID, userID, block)
	}
	return nil, nil
}

func (m *MockNoteUsecase) DeleteBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) error {
	if m.DeleteBlockFunc != nil {
		return m.DeleteBlockFunc(ctx, blockID, noteID, userID)
	}
	return nil
}

func (m *MockNoteUsecase) MoveBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, newPosition int) (*models.Block, error) {
	if m.MoveBlockFunc != nil {
		return m.MoveBlockFunc(ctx, blockID, noteID, userID, newPosition)
	}
	return nil, nil
}

func (m *MockNoteUsecase) UpdateBlockContent(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, content string) (*models.Block, error) {
	if m.UpdateBlockContentFunc != nil {
		return m.UpdateBlockContentFunc(ctx, blockID, noteID, userID, content)
	}
	return nil, nil
}

func (m *MockNoteUsecase) UpdateBlockFormatting(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, formattingRange models.FormattingRange) (*models.BlockFormatting, error) {
	if m.UpdateBlockFormattingFunc != nil {
		return m.UpdateBlockFormattingFunc(ctx, blockID, noteID, userID, formattingRange)
	}
	return nil, nil
}

type MockProfileUsecase struct {
	GetProfileFunc func(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
}

func (m *MockProfileUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	if m.GetProfileFunc != nil {
		return m.GetProfileFunc(ctx, userID)
	}
	return &models.Profile{Username: "testuser"}, nil
}

type MockAttachmentUsecase struct {
	UploadAttachmentFunc func(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, fileName string, fileSize int64, mimeType string, fileData []byte, hasPosition bool, position int) (*models.Attachment, error)
}

func (m *MockAttachmentUsecase) UploadAttachment(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, fileName string, fileSize int64, mimeType string, fileData []byte, hasPosition bool, position int) (*models.Attachment, error) {
	if m.UploadAttachmentFunc != nil {
		return m.UploadAttachmentFunc(ctx, noteID, userID, fileName, fileSize, mimeType, fileData, hasPosition, position)
	}
	return &models.Attachment{
		ID:       uuid.New(),
		BlockID:  uuid.New(),
		MinioKey: "test-key",
	}, nil
}

// Helper functions

func createTestHub() *Hub {
	noteUsecase := &MockNoteUsecase{}
	profileUsecase := &MockProfileUsecase{}
	attachmentUsecase := &MockAttachmentUsecase{}
	return NewHub(noteUsecase, profileUsecase, attachmentUsecase)
}

func createTestClient(userID, userName, noteID string) *ClientInfo {
	return &ClientInfo{
		UserID:   userID,
		UserName: userName,
		NoteID:   noteID,
		Send:     make(chan WebSocketMessage, 256),
		LastPing: time.Now().Unix(),
	}
}

// Tests

func TestNewHub(t *testing.T) {
	hub := createTestHub()

	if hub == nil {
		t.Fatal("Expected hub to be created")
	}

	if hub.rooms == nil {
		t.Error("Rooms map should be initialized")
	}

	if hub.register == nil {
		t.Error("Register channel should be initialized")
	}

	if hub.unregister == nil {
		t.Error("Unregister channel should be initialized")
	}

	if hub.broadcast == nil {
		t.Error("Broadcast channel should be initialized")
	}

	if hub.storage == nil {
		t.Error("Storage should be initialized")
	}
}

func TestHub_HandleRegister_NewRoom(t *testing.T) {
	hub := createTestHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	client := createTestClient("user1", "Test User", "note123")

	hub.register <- client

	// Give some time for processing
	time.Sleep(50 * time.Millisecond)

	hub.mu.RLock()
	room, exists := hub.rooms["note123"]
	hub.mu.RUnlock()

	if !exists {
		t.Error("Room should be created")
	}

	if room.NoteID != "note123" {
		t.Errorf("Expected NoteID 'note123', got '%s'", room.NoteID)
	}

	// Client should be added to the room
	_, ok := room.GetClient("user1")
	if !ok {
		t.Error("Client should be added to room")
	}
}

func TestHub_HandleRegister_ExistingRoom(t *testing.T) {
	hub := createTestHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	// First client creates room
	client1 := createTestClient("user1", "User One", "note123")
	hub.register <- client1

	time.Sleep(50 * time.Millisecond)

	// Second client joins existing room
	client2 := createTestClient("user2", "User Two", "note123")
	hub.register <- client2

	time.Sleep(50 * time.Millisecond)

	hub.mu.RLock()
	room, exists := hub.rooms["note123"]
	hub.mu.RUnlock()

	if !exists {
		t.Fatal("Room should exist")
	}

	_, ok1 := room.GetClient("user1")
	_, ok2 := room.GetClient("user2")

	if !ok1 || !ok2 {
		t.Error("Both clients should be in the room")
	}
}

func TestHub_HandleRegister_DeletedRoom(t *testing.T) {
	hub := createTestHub()

	// Create room and mark as deleted
	room := NewNoteRoom("note123")
	room.IsDeleted = true
	hub.mu.Lock()
	hub.rooms["note123"] = room
	hub.mu.Unlock()

	client := createTestClient("user1", "Test User", "note123")

	// Start reading from client's Send channel
	var receivedMsg WebSocketMessage
	go func() {
		select {
		case msg := <-client.Send:
			receivedMsg = msg
		case <-time.After(100 * time.Millisecond):
		}
	}()

	hub.handleRegister(client)

	time.Sleep(50 * time.Millisecond)

	if receivedMsg.Type != MsgError {
		t.Errorf("Expected error message, got %s", receivedMsg.Type)
	}
}

func TestHub_HandleUnregister(t *testing.T) {
	hub := createTestHub()

	// Create room with client
	room := NewNoteRoom("note123")
	client := createTestClient("user1", "Test User", "note123")
	room.AddClient(client)
	hub.mu.Lock()
	hub.rooms["note123"] = room
	hub.mu.Unlock()

	hub.handleUnregister(client)

	// Client should be removed
	_, ok := room.GetClient("user1")
	if ok {
		t.Error("Client should be removed from room")
	}

	// Send channel should be closed
	_, ok = <-client.Send
	if ok {
		t.Error("Send channel should be closed")
	}
}

func TestHub_HandleUnregister_EmptyRoomRemoval(t *testing.T) {
	hub := createTestHub()

	// Create room with one client
	room := NewNoteRoom("note123")
	client := createTestClient("user1", "Test User", "note123")
	room.AddClient(client)
	hub.mu.Lock()
	hub.rooms["note123"] = room
	hub.mu.Unlock()

	hub.handleUnregister(client)

	hub.mu.RLock()
	_, exists := hub.rooms["note123"]
	hub.mu.RUnlock()

	if exists {
		t.Error("Empty room should be removed from hub")
	}
}

func TestHub_HandleBroadcast(t *testing.T) {
	hub := createTestHub()

	// Create room with two clients
	room := NewNoteRoom("note123")
	client1 := createTestClient("user1", "User One", "note123")
	client2 := createTestClient("user2", "User Two", "note123")
	room.AddClient(client1)
	room.AddClient(client2)
	hub.mu.Lock()
	hub.rooms["note123"] = room
	hub.mu.Unlock()

	msg := WebSocketMessage{
		Type: MsgCursorMove,
		Msg:  map[string]string{"test": "data"},
	}

	broadcastMsg := &BroadcastMessage{
		NoteID:  "note123",
		Message: msg,
		Exclude: "user1",
	}

	// Read from client2's channel (should receive message)
	var receivedMsg WebSocketMessage
	go func() {
		select {
		case receivedMsg = <-client2.Send:
		case <-time.After(100 * time.Millisecond):
		}
	}()

	hub.handleBroadcast(broadcastMsg)

	time.Sleep(50 * time.Millisecond)

	if receivedMsg.Type != MsgCursorMove {
		t.Errorf("Expected broadcast message, got %+v", receivedMsg)
	}
}

func TestHub_BroadcastToRoom(t *testing.T) {
	hub := createTestHub()

	// Create room with two clients
	room := NewNoteRoom("note123")
	client1 := createTestClient("user1", "User One", "note123")
	client2 := createTestClient("user2", "User Two", "note123")
	room.AddClient(client1)
	room.AddClient(client2)
	hub.mu.Lock()
	hub.rooms["note123"] = room
	hub.mu.Unlock()

	msg := WebSocketMessage{
		Type: MsgUserJoined,
		Msg:  map[string]string{"message": "joined"},
	}

	// Start broadcast goroutine
	go hub.Run(context.Background())

	hub.broadcastToRoom("note123", msg, "user1")

	time.Sleep(50 * time.Millisecond)

	// Check that client2 received message
	select {
	case received := <-client2.Send:
		if received.Type != MsgUserJoined {
			t.Errorf("Expected MsgUserJoined, got %s", received.Type)
		}
	default:
		t.Error("Client2 should have received the message")
	}

	// Client1 should NOT have received message
	select {
	case <-client1.Send:
		t.Error("Client1 should not have received message (excluded)")
	default:
		// Expected
	}
}

func TestHub_HandleCursorMove(t *testing.T) {
	hub := createTestHub()

	room := NewNoteRoom("note123")
	client := createTestClient("user1", "Test User", "note123")
	room.AddClient(client)
	hub.mu.Lock()
	hub.rooms["note123"] = room
	hub.mu.Unlock()

	cursor := CursorPosition{
		BlockID:  "block1",
		Position: 42,
	}

	hub.handleCursorMove(room, "user1", cursor)

	storedCursor := client.GetCursor()
	if storedCursor.Position != 42 {
		t.Errorf("Expected cursor position 42, got %d", storedCursor.Position)
	}
	if storedCursor.BlockID != "block1" {
		t.Errorf("Expected BlockID 'block1', got '%s'", storedCursor.BlockID)
	}
}

func TestHub_HandleInsertChar_Local(t *testing.T) {
	hub := createTestHub()

	room := NewNoteRoom("note123")
	client := createTestClient("user1", "Test User", "note123")
	room.AddClient(client)
	hub.mu.Lock()
	hub.rooms["note123"] = room
	hub.mu.Unlock()

	msg := WebSocketMessage{
		Type:    MsgInsertChar,
		IsLocal: true,
	}

	op := &InsertCharOperation{
		BlockID: "block1",
		Char:    "a",
	}

	hub.handleInsertChar(room, "user1", msg, op)

	doc, exists := room.GetCRDTDocument("block1")
	if !exists {
		t.Error("CRDT document should be created")
	}

	text := doc.GetText()
	if text != "a" {
		t.Errorf("Expected text 'a', got '%s'", text)
	}
}

func TestHub_HandleInsertChar_Remote(t *testing.T) {
	hub := createTestHub()

	room := NewNoteRoom("note123")
	doc := NewCRDTDocument("user2")
	doc.InsertChar(0, 'b', "user2")
	room.SetCRDTDocument("block1", doc)
	hub.mu.Lock()
	hub.rooms["note123"] = room
	hub.mu.Unlock()

	msg := WebSocketMessage{
		Type:    MsgInsertChar,
		IsLocal: false,
	}

	op := &InsertCharOperation{
		BlockID:   "block1",
		UniqueID:  "user2:2:2",
		Char:      "c",
		PrevID:    "user2:1:1",
		Lamport:   2,
		UserID:    "user2",
		Timestamp: time.Now().UnixNano(),
	}

	hub.handleInsertChar(room, "user2", msg, op)

	text := doc.GetText()
	// Should have 'b' and 'c' (order depends on CRDT)
	if len(text) != 2 {
		t.Errorf("Expected 2 characters, got %d", len(text))
	}
}

func TestHub_HandleDeleteChar_Local(t *testing.T) {
	hub := createTestHub()

	room := NewNoteRoom("note123")
	client := createTestClient("user1", "Test User", "note123")
	room.AddClient(client)

	// Create CRDT doc with content "ab"
	doc := NewCRDTDocument("user1")
	doc.InsertChar(0, 'a', "user1")
	doc.InsertChar(1, 'b', "user1")
	room.SetCRDTDocument("block1", doc)

	hub.mu.Lock()
	hub.rooms["note123"] = room
	hub.mu.Unlock()

	// Set cursor position to 2 (after 'b')
	client.UpdateCursor(CursorPosition{
		BlockID:  "block1",
		Position: 2,
	})

	msg := WebSocketMessage{
		Type:    MsgDeleteChar,
		IsLocal: true,
	}

	op := &DeleteCharOperation{
		BlockID: "block1",
	}

	hub.handleDeleteChar(room, "user1", msg, op)

	text := doc.GetText()
	if text != "a" {
		t.Errorf("Expected text 'a', got '%s'", text)
	}
}

func TestHub_HandleDeleteChar_Remote(t *testing.T) {
	hub := createTestHub()

	room := NewNoteRoom("note123")
	doc := NewCRDTDocument("user1")
	id := doc.InsertChar(0, 'a', "user1")
	room.SetCRDTDocument("block1", doc)
	hub.mu.Lock()
	hub.rooms["note123"] = room
	hub.mu.Unlock()

	msg := WebSocketMessage{
		Type:    MsgDeleteChar,
		IsLocal: false,
	}

	op := &DeleteCharOperation{
		BlockID:  "block1",
		UniqueID: id,
		Lamport:  2,
	}

	hub.handleDeleteChar(room, "user2", msg, op)

	text := doc.GetText()
	if text != "" {
		t.Errorf("Expected empty text, got '%s'", text)
	}
}

func TestHub_HandleCreateBlock(t *testing.T) {
	hub := createTestHub()

	// Setup mock
	createdBlock := &models.Block{
		ID:          uuid.New(),
		BlockTypeID: 1,
		Position:    0,
	}

	mockNoteUsecase := &MockNoteUsecase{
		CreateBlockFunc: func(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, block models.Block) (*models.Block, error) {
			return createdBlock, nil
		},
	}
	hub.noteUsecase = mockNoteUsecase

	room := NewNoteRoom("note123")
	client := createTestClient("user1", "Test User", "note123")
	room.AddClient(client)
	hub.mu.Lock()
	hub.rooms["note123"] = room
	hub.mu.Unlock()

	op := &CreateBlockOperation{
		BlockTypeID: 1,
		Position:    0,
	}

	hub.handleCreateBlock(room, "user1", op)

	// Should create CRDT document for text block
	_, exists := room.GetCRDTDocument(createdBlock.ID.String())
	if !exists {
		t.Error("CRDT document should be created for text block")
	}
}

func TestHub_HandleCreateBlock_Error(t *testing.T) {
	hub := createTestHub()

	mockNoteUsecase := &MockNoteUsecase{
		CreateBlockFunc: func(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, block models.Block) (*models.Block, error) {
			return nil, errors.New("create block failed")
		},
	}
	hub.noteUsecase = mockNoteUsecase

	room := NewNoteRoom("note123")
	client := createTestClient("user1", "Test User", "note123")
	room.AddClient(client)
	hub.mu.Lock()
	hub.rooms["note123"] = room
	hub.mu.Unlock()

	// Read error message
	var receivedMsg WebSocketMessage
	go func() {
		select {
		case receivedMsg = <-client.Send:
		case <-time.After(100 * time.Millisecond):
		}
	}()

	op := &CreateBlockOperation{
		BlockTypeID: 1,
		Position:    0,
	}

	hub.handleCreateBlock(room, "user1", op)

	time.Sleep(50 * time.Millisecond)

	if receivedMsg.Type != MsgError {
		t.Errorf("Expected error message, got %s", receivedMsg.Type)
	}
}

func TestHub_HandleDeleteBlock(t *testing.T) {
	hub := createTestHub()

	deleteCalled := false
	mockNoteUsecase := &MockNoteUsecase{
		DeleteBlockFunc: func(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) error {
			deleteCalled = true
			return nil
		},
	}
	hub.noteUsecase = mockNoteUsecase

	room := NewNoteRoom("note123")
	client := createTestClient("user1", "Test User", "note123")
	room.AddClient(client)
	room.SetCRDTDocument("block1", NewCRDTDocument("user1"))
	hub.mu.Lock()
	hub.rooms["note123"] = room
	hub.mu.Unlock()

	op := &DeleteBlockOperation{
		BlockID: "block1",
	}

	hub.handleDeleteBlock(room, "user1", op)

	if !deleteCalled {
		t.Error("DeleteBlock should be called")
	}

	_, exists := room.GetCRDTDocument("block1")
	if exists {
		t.Error("CRDT document should be deleted")
	}
}

func TestHub_HandleMoveBlock(t *testing.T) {
	hub := createTestHub()

	moveCalled := false
	mockNoteUsecase := &MockNoteUsecase{
		MoveBlockFunc: func(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, newPosition int) (*models.Block, error) {
			moveCalled = true
			return &models.Block{Position: newPosition}, nil
		},
	}
	hub.noteUsecase = mockNoteUsecase

	room := NewNoteRoom("note123")
	client := createTestClient("user1", "Test User", "note123")
	room.AddClient(client)
	hub.mu.Lock()
	hub.rooms["note123"] = room
	hub.mu.Unlock()

	op := &MoveBlockOperation{
		BlockID:     "block1",
		NewPosition: 5,
	}

	hub.handleMoveBlock(room, "user1", op)

	if !moveCalled {
		t.Error("MoveBlock should be called")
	}
}

func TestHub_HandleUpdateNoteTitle(t *testing.T) {
	hub := createTestHub()

	updateCalled := false
	mockNoteUsecase := &MockNoteUsecase{
		GetNoteFunc: func(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, []models.Block, map[string]models.BlockFormatting, error) {
			return &models.Note{Title: "Old Title"}, nil, nil, nil
		},
		UpdateNoteFunc: func(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, note models.Note) (*models.Note, error) {
			updateCalled = true
			if note.Title != "New Title" {
				t.Errorf("Expected title 'New Title', got '%s'", note.Title)
			}
			return &note, nil
		},
	}
	hub.noteUsecase = mockNoteUsecase

	room := NewNoteRoom("note123")
	client := createTestClient("user1", "Test User", "note123")
	room.AddClient(client)
	hub.mu.Lock()
	hub.rooms["note123"] = room
	hub.mu.Unlock()

	hub.handleUpdateNoteTitle(room, "user1", "New Title")

	if !updateCalled {
		t.Error("UpdateNote should be called")
	}
}

func TestHub_HandleUpdateNotePublic_Owner(t *testing.T) {
	hub := createTestHub()

	userUUID := uuid.New()
	noteUUID := uuid.New()
	noteIDStr := noteUUID.String()
	userIDStr := userUUID.String()

	updateCalled := false
	mockNoteUsecase := &MockNoteUsecase{
		GetNoteFunc: func(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, []models.Block, map[string]models.BlockFormatting, error) {
			return &models.Note{
				ID:       noteUUID,
				UserID:   userUUID,
				IsPublic: false,
			}, nil, nil, nil
		},
		UpdateNoteFunc: func(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, note models.Note) (*models.Note, error) {
			updateCalled = true
			if !note.IsPublic {
				t.Error("IsPublic should be true")
			}
			return &note, nil
		},
	}
	hub.noteUsecase = mockNoteUsecase

	room := NewNoteRoom(noteIDStr)
	client := createTestClient(userIDStr, "Test User", noteIDStr)
	room.AddClient(client)
	hub.mu.Lock()
	hub.rooms[noteIDStr] = room
	hub.mu.Unlock()

	hub.handleUpdateNotePublic(room, userIDStr, true)

	if !updateCalled {
		t.Error("UpdateNote should be called for owner")
	}
}

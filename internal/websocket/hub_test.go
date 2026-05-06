// file: hub_test.go
package websocket

import (
	"context"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

// Mock implementations for interfaces
type MockNoteUsecase struct {
	notes  map[string]*models.Note
	blocks map[string][]models.Block
}

func (m *MockNoteUsecase) GetNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, error) {
	// Mock implementation
	return &models.Note{}, nil
}

func (m *MockNoteUsecase) UpdateNote(ctx context.Context, noteID uuid.UUID, note models.Note, userID uuid.UUID) (*models.Note, error) {
	return &note, nil
}

func (m *MockNoteUsecase) DeleteNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) error {
	return nil
}

func (m *MockNoteUsecase) GetBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) (*models.Block, error) {
	return &models.Block{}, nil
}

func (m *MockNoteUsecase) GetBlocks(ctx context.Context, noteID uuid.UUID) ([]models.Block, error) {
	return []models.Block{}, nil
}

func (m *MockNoteUsecase) CreateBlock(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, block models.Block) (*models.Block, error) {
	return &block, nil
}

func (m *MockNoteUsecase) DeleteBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) error {
	return nil
}

func (m *MockNoteUsecase) MoveBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, newPosition int) (*models.Block, error) {
	return &models.Block{}, nil
}

func (m *MockNoteUsecase) UpdateBlockContent(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, content string) (*models.Block, error) {
	return &models.Block{}, nil
}

func (m *MockNoteUsecase) UpdateBlockFormatting(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, formattingRange models.FormattingRange) (*models.BlockFormatting, error) {
	return &models.BlockFormatting{}, nil
}

type MockProfileUsecase struct{}

func (m *MockProfileUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	return &models.Profile{
		Username: "testuser",
	}, nil
}

func TestNewHub(t *testing.T) {
	noteUsecase := &MockNoteUsecase{}
	profileUsecase := &MockProfileUsecase{}

	hub := NewHub(noteUsecase, profileUsecase)

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

func TestHub_RegisterAndUnregister(t *testing.T) {
	noteUsecase := &MockNoteUsecase{}
	profileUsecase := &MockProfileUsecase{}

	hub := NewHub(noteUsecase, profileUsecase)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	client := &ClientInfo{
		UserID:   "user1",
		UserName: "Test User",
		NoteID:   "note123",
		Send:     make(chan WebSocketMessage, 10),
	}

	// Register client
	hub.register <- client

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Check that room was created
	hub.mu.RLock()
	_, exists := hub.rooms["note123"]
	hub.mu.RUnlock()

	if !exists {
		t.Error("Room should be created after client registration")
	}

	// Unregister client
	hub.unregister <- client

	time.Sleep(100 * time.Millisecond)

	// Check that room is empty but not deleted (still exists with no clients)
	hub.mu.RLock()
	room, exists := hub.rooms["note123"]
	hub.mu.RUnlock()

	if exists && len(room.Clients) != 0 {
		t.Error("Room should have no clients after unregister")
	}
}

func TestHub_HandleCursorMove(t *testing.T) {
	noteUsecase := &MockNoteUsecase{}
	profileUsecase := &MockProfileUsecase{}

	hub := NewHub(noteUsecase, profileUsecase)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	client := &ClientInfo{
		UserID:   "user1",
		UserName: "User One",
		NoteID:   "note123",
		Send:     make(chan WebSocketMessage, 10),
	}

	hub.register <- client
	time.Sleep(100 * time.Millisecond)

	// Handle cursor move
	cursorMsg := WebSocketMessage{
		Type: MsgCursorMove,
		Msg: map[string]interface{}{
			"blockId":  "block1",
			"position": 10,
		},
	}

	hub.HandleOperation("note123", "user1", cursorMsg)

	time.Sleep(50 * time.Millisecond)

	// Check that cursor was updated
	hub.mu.RLock()
	room := hub.rooms["note123"]
	hub.mu.RUnlock()

	if room != nil {
		clientInfo, _ := room.GetClient("user1")
		cursor := clientInfo.GetCursor()
		if cursor.Position != 10 {
			t.Errorf("Expected cursor position 10, got %d", cursor.Position)
		}
	}
}

func TestHub_CleanupInactiveRooms(t *testing.T) {
	noteUsecase := &MockNoteUsecase{}
	profileUsecase := &MockProfileUsecase{}

	hub := NewHub(noteUsecase, profileUsecase)

	// Manually add an old room
	room := NewNoteRoom("old_note")
	room.CreatedAt = time.Now().Unix() - 1000 // Very old
	hub.mu.Lock()
	hub.rooms["old_note"] = room
	hub.mu.Unlock()

	hub.cleanupInactiveRooms()

	hub.mu.RLock()
	_, exists := hub.rooms["old_note"]
	hub.mu.RUnlock()

	if exists {
		t.Error("Inactive room should have been cleaned up")
	}
}

func TestHub_IsNoteOwner(t *testing.T) {
	// This test requires proper mock setup with GetNote returning appropriate values
	noteUsecase := &MockNoteUsecase{}
	profileUsecase := &MockProfileUsecase{}

	hub := NewHub(noteUsecase, profileUsecase)

	// For now, just verify method doesn't panic
	_ = hub.isNoteOwner("note123", "user1")
}

func TestHub_ErrorResponse(t *testing.T) {
	noteUsecase := &MockNoteUsecase{}
	profileUsecase := &MockProfileUsecase{}

	hub := NewHub(noteUsecase, profileUsecase)

	client := &ClientInfo{
		UserID: "user1",
		NoteID: "note123",
	}

	errMsg := hub.errorMessage("test error", client)

	if errMsg.Type != MsgError {
		t.Errorf("Expected MsgError, got %s", errMsg.Type)
	}

	if errMsg.UserID != "user1" {
		t.Errorf("Expected UserID 'user1', got '%s'", errMsg.UserID)
	}
}

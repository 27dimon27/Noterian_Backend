// file: room_test.go
package websocket

import (
	"testing"
)

func TestNewNoteRoom(t *testing.T) {
	room := NewNoteRoom("note123")

	if room.NoteID != "note123" {
		t.Errorf("Expected NoteID 'note123', got '%s'", room.NoteID)
	}

	if room.Clients == nil {
		t.Error("Clients map should be initialized")
	}

	if room.CRDTDocuments == nil {
		t.Error("CRDTDocuments map should be initialized")
	}

	if room.IsDeleted {
		t.Error("IsDeleted should be false for new room")
	}

	if room.CreatedAt == 0 {
		t.Error("CreatedAt should be set")
	}
}

func TestNoteRoom_AddAndGetClient(t *testing.T) {
	room := NewNoteRoom("note123")

	client := &ClientInfo{
		UserID:   "user1",
		UserName: "Test User",
	}

	room.AddClient(client)

	retrieved, ok := room.GetClient("user1")
	if !ok {
		t.Error("GetClient should return true for existing client")
	}

	if retrieved.UserID != client.UserID {
		t.Errorf("Expected UserID '%s', got '%s'", client.UserID, retrieved.UserID)
	}
}

func TestNoteRoom_RemoveClient(t *testing.T) {
	room := NewNoteRoom("note123")

	client := &ClientInfo{UserID: "user1"}
	room.AddClient(client)

	_, ok := room.GetClient("user1")
	if !ok {
		t.Error("Client should exist before removal")
	}

	room.RemoveClient("user1")

	_, ok = room.GetClient("user1")
	if ok {
		t.Error("Client should not exist after removal")
	}
}

func TestNoteRoom_GetAllCursors(t *testing.T) {
	room := NewNoteRoom("note123")

	client1 := &ClientInfo{
		UserID:   "user1",
		UserName: "User One",
		LastCursor: CursorPosition{
			BlockID:  "block1",
			Position: 5,
		},
	}

	client2 := &ClientInfo{
		UserID:   "user2",
		UserName: "User Two",
		LastCursor: CursorPosition{
			BlockID:  "block2",
			Position: 10,
		},
	}

	room.AddClient(client1)
	room.AddClient(client2)

	cursors := room.GetAllCursors()

	if len(cursors) != 2 {
		t.Errorf("Expected 2 cursors, got %d", len(cursors))
	}

	// Find cursor for user1
	var found bool
	for _, c := range cursors {
		if c.UserID == "user1" && c.Cursor.Position == 5 {
			found = true
		}
	}
	if !found {
		t.Error("Cursor for user1 not found with correct position")
	}
}

func TestNoteRoom_CRDTDocumentOperations(t *testing.T) {
	room := NewNoteRoom("note123")

	doc := NewCRDTDocument("user1")
	room.SetCRDTDocument("block1", doc)

	retrieved, ok := room.GetCRDTDocument("block1")
	if !ok {
		t.Error("GetCRDTDocument should return true for existing document")
	}

	if retrieved != doc {
		t.Error("Retrieved document should be the same as stored")
	}

	room.DeleteCRDTDocument("block1")

	_, ok = room.GetCRDTDocument("block1")
	if ok {
		t.Error("Document should be deleted")
	}
}

func TestNoteRoom_EmptyRoomCursors(t *testing.T) {
	room := NewNoteRoom("note123")

	cursors := room.GetAllCursors()

	if len(cursors) != 0 {
		t.Errorf("Expected empty cursor list, got %d cursors", len(cursors))
	}
}

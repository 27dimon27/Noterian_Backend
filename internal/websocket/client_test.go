// file: client_test.go
package websocket

import (
	"testing"
)

func TestClientInfo_UpdateCursor(t *testing.T) {
	client := &ClientInfo{
		UserID:   "user1",
		UserName: "Test User",
		LastCursor: CursorPosition{
			BlockID:  "block1",
			Position: 0,
		},
	}

	cursor := CursorPosition{
		BlockID:  "block2",
		Position: 10,
	}

	client.UpdateCursor(cursor)

	result := client.GetCursor()

	if result.BlockID != "block2" {
		t.Errorf("Expected BlockID 'block2', got '%s'", result.BlockID)
	}

	if result.Position != 10 {
		t.Errorf("Expected Position 10, got %d", result.Position)
	}

	if result.Timestamp == 0 {
		t.Error("Timestamp should be set")
	}
}

func TestClientInfo_GetCursor(t *testing.T) {
	client := &ClientInfo{
		UserID: "user1",
		LastCursor: CursorPosition{
			BlockID:  "block1",
			Position: 5,
		},
	}

	cursor := client.GetCursor()

	if cursor.BlockID != "block1" {
		t.Errorf("Expected BlockID 'block1', got '%s'", cursor.BlockID)
	}

	if cursor.Position != 5 {
		t.Errorf("Expected Position 5, got %d", cursor.Position)
	}
}

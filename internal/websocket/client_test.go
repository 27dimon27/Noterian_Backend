package websocket

import (
	"testing"
)

func TestClientInfo_UpdateCursor(t *testing.T) {
	client := &ClientInfo{
		UserID:   "user1",
		UserName: "Test User",
		LastCursor: CursorPosition{
			BlockID:       "block1",
			StartPosition: 0,
			EndPosition:   0,
		},
	}

	cursor := CursorPosition{
		BlockID:       "block2",
		StartPosition: 10,
		EndPosition:   10,
	}

	client.UpdateCursor(cursor)

	result := client.GetCursor()

	if result.BlockID != "block2" {
		t.Errorf("Expected BlockID 'block2', got '%s'", result.BlockID)
	}

	if result.StartPosition != 10 {
		t.Errorf("Expected StartPosition 10, got %d", result.StartPosition)
	}

	if result.Timestamp == 0 {
		t.Error("Timestamp should be set")
	}
}

func TestClientInfo_GetCursor(t *testing.T) {
	client := &ClientInfo{
		UserID: "user1",
		LastCursor: CursorPosition{
			BlockID:       "block1",
			StartPosition: 5,
			EndPosition:   5,
		},
	}

	cursor := client.GetCursor()

	if cursor.BlockID != "block1" {
		t.Errorf("Expected BlockID 'block1', got '%s'", cursor.BlockID)
	}

	if cursor.StartPosition != 5 {
		t.Errorf("Expected StartPosition 5, got %d", cursor.StartPosition)
	}
}

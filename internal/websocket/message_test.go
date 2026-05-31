package websocket

import (
	"encoding/json"
	"testing"
)

func TestWebSocketMessage_JSONSerialization(t *testing.T) {
	msg := WebSocketMessage{
		Type:      MsgInsertChars,
		UserID:    "user1",
		UserName:  "Test User",
		NoteID:    "note123",
		Timestamp: 1234567890,
		Msg: map[string]interface{}{
			"blockId":       "block1",
			"startPosition": 5,
			"endPosition":   5,
			"char":          "a",
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	var unmarshaled WebSocketMessage
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if unmarshaled.Type != msg.Type {
		t.Errorf("Expected Type '%s', got '%s'", msg.Type, unmarshaled.Type)
	}

	if unmarshaled.UserID != msg.UserID {
		t.Errorf("Expected UserID '%s', got '%s'", msg.UserID, unmarshaled.UserID)
	}
}

func TestCursorPosition_JSONSerialization(t *testing.T) {
	cursor := CursorPosition{
		BlockID:       "block1",
		StartPosition: 42,
		EndPosition:   42,
		UserID:        "user1",
		UserName:      "Test User",
		Timestamp:     1234567890,
	}

	data, err := json.Marshal(cursor)
	if err != nil {
		t.Fatalf("Failed to marshal cursor: %v", err)
	}

	var unmarshaled CursorPosition
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal cursor: %v", err)
	}

	if unmarshaled.BlockID != cursor.BlockID {
		t.Errorf("Expected BlockID '%s', got '%s'", cursor.BlockID, unmarshaled.BlockID)
	}

	if unmarshaled.StartPosition != cursor.StartPosition {
		t.Errorf("Expected StartPosition %d, got %d", cursor.StartPosition, unmarshaled.StartPosition)
	}

	if unmarshaled.EndPosition != cursor.EndPosition {
		t.Errorf("Expected EndPosition %d, got %d", cursor.EndPosition, unmarshaled.EndPosition)
	}
}

func TestInsertCharOperation_JSONSerialization(t *testing.T) {
	op := InsertCharsOperation{
		ID:        "op1",
		BlockID:   "block1",
		Position:  5,
		Char:      "a",
		Lamport:   100,
		UniqueIDs: []string{"user1:1:100"},
		PrevID:    "root:0:0",
		UserID:    "user1",
		Timestamp: 1234567890,
	}

	data, err := json.Marshal(op)
	if err != nil {
		t.Fatalf("Failed to marshal operation: %v", err)
	}

	var unmarshaled InsertCharsOperation
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal operation: %v", err)
	}

	if unmarshaled.Char != op.Char {
		t.Errorf("Expected Char '%s', got '%s'", op.Char, unmarshaled.Char)
	}

	if unmarshaled.Position != op.Position {
		t.Errorf("Expected Position %d, got %d", op.Position, unmarshaled.Position)
	}
}

func TestDeleteCharOperation_JSONSerialization(t *testing.T) {
	op := DeleteCharsOperation{
		ID:            "op1",
		BlockID:       "block1",
		StartPosition: 5,
		EndPosition:   5,
		UniqueIDs:     []string{"user1:1:100"},
		Lamport:       101,
		UserID:        "user1",
		Timestamp:     1234567890,
	}

	data, err := json.Marshal(op)
	if err != nil {
		t.Fatalf("Failed to marshal operation: %v", err)
	}

	var unmarshaled DeleteCharsOperation
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal operation: %v", err)
	}

	if unmarshaled.UniqueIDs[0] != op.UniqueIDs[0] {
		t.Errorf("Expected UniqueID '%s', got '%s'", op.UniqueIDs[0], unmarshaled.UniqueIDs[0])
	}
}

func TestFormattingOperation_JSONSerialization(t *testing.T) {
	boldTrue := true
	italicFalse := false

	op := FormattingOperation{
		ID:         "op1",
		BlockID:    "block1",
		StartPos:   0,
		EndPos:     10,
		Bold:       &boldTrue,
		Italic:     &italicFalse,
		SequenceID: 1,
		Lamport:    100,
		UserID:     "user1",
		Timestamp:  1234567890,
	}

	data, err := json.Marshal(op)
	if err != nil {
		t.Fatalf("Failed to marshal operation: %v", err)
	}

	var unmarshaled FormattingOperation
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal operation: %v", err)
	}

	if unmarshaled.Bold == nil || *unmarshaled.Bold != true {
		t.Error("Bold should be true")
	}

	if unmarshaled.Italic == nil || *unmarshaled.Italic != false {
		t.Error("Italic should be false")
	}
}

func TestAllMessageTypes(t *testing.T) {
	types := []MessageType{
		MsgUserJoined,
		MsgUserLeft,
		MsgError,
		MsgSyncState,
		MsgHeartbeat,
		MsgCursorMove,
		MsgInsertChars,
		MsgDeleteChars,
		MsgApplyFormatting,
		MsgCreateBlock,
		MsgDeleteBlock,
		MsgMoveBlock,
		MsgUpdateNoteTitle,
		MsgUpdateNotePublic,
		MsgDeleteNote,
		MsgNotePrivate,
		MsgNoteDeleted,
	}

	for _, msgType := range types {
		if string(msgType) == "" {
			t.Errorf("Message type %s has empty string value", msgType)
		}
	}
}

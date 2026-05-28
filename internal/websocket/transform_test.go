package websocket

import (
	"testing"
)

func TestTransformCursorPosition_InsertBefore(t *testing.T) {
	op := &InsertCharsOperation{
		BlockID:  "block1",
		Position: 5,
		Char:     "X",
	}

	newStartPos, _ := TransformCursorPosition(6, 6, MsgInsertChars, op, "block1", "user1")
	if newStartPos != 7 {
		t.Errorf("Expected cursor start position 7, got %d", newStartPos)
	}
}

func TestTransformCursorPosition_InsertAfter(t *testing.T) {
	op := &InsertCharsOperation{
		BlockID:  "block1",
		Position: 5,
		Char:     "X",
	}

	newStartPos, _ := TransformCursorPosition(3, 3, MsgInsertChars, op, "block1", "user1")
	if newStartPos != 3 {
		t.Errorf("Expected cursor start position 3, got %d", newStartPos)
	}
}

func TestTransformCursorPosition_InsertAtSamePosition(t *testing.T) {
	op := &InsertCharsOperation{
		BlockID:  "block1",
		Position: 5,
		Char:     "X",
	}

	newStartPos, _ := TransformCursorPosition(5, 5, MsgInsertChars, op, "block1", "user1")
	if newStartPos != 5 {
		t.Errorf("Expected cursor start position 5, got %d", newStartPos)
	}
}

func TestTransformCursorPosition_InsertDifferentBlock(t *testing.T) {
	op := &InsertCharsOperation{
		BlockID:  "block2",
		Position: 5,
		Char:     "X",
	}

	newStartPos, _ := TransformCursorPosition(10, 10, MsgInsertChars, op, "block1", "user1")
	if newStartPos != 10 {
		t.Errorf("Expected cursor start position 10 (unchanged), got %d", newStartPos)
	}
}

func TestTransformCursorPosition_DeleteBefore(t *testing.T) {
	op := &DeleteCharsOperation{
		BlockID:       "block1",
		StartPosition: 5,
		EndPosition:   6,
	}

	newStartPos, _ := TransformCursorPosition(7, 7, MsgDeleteChars, op, "block1", "user1")
	if newStartPos != 6 {
		t.Errorf("Expected cursor start position 6, got %d", newStartPos)
	}
}

func TestTransformCursorPosition_DeleteAfter(t *testing.T) {
	op := &DeleteCharsOperation{
		BlockID:       "block1",
		StartPosition: 5,
		EndPosition:   6,
	}

	newStartPos, _ := TransformCursorPosition(3, 3, MsgDeleteChars, op, "block1", "user1")
	if newStartPos != 3 {
		t.Errorf("Expected cursor start position 3, got %d", newStartPos)
	}
}

func TestTransformCursorPosition_DeleteAtCursor(t *testing.T) {
	op := &DeleteCharsOperation{
		BlockID:       "block1",
		StartPosition: 5,
		EndPosition:   6,
	}

	newPos, _ := TransformCursorPosition(6, 6, MsgDeleteChars, op, "block1", "user1")
	if newPos != 5 {
		t.Errorf("Expected cursor position 5, got %d", newPos)
	}
}

func TestTransformCursorPosition_DeleteAllBeforeCursor(t *testing.T) {
	op := &DeleteCharsOperation{
		BlockID:       "block1",
		StartPosition: 0,
		EndPosition:   1,
	}

	newPos, _ := TransformCursorPosition(5, 5, MsgDeleteChars, op, "block1", "user1")
	if newPos != 4 {
		t.Errorf("Expected cursor position 4, got %d", newPos)
	}
}

func TestTransformCursorPosition_NegativeResult(t *testing.T) {
	op := &DeleteCharsOperation{
		BlockID:       "block1",
		StartPosition: 0,
		EndPosition:   1,
	}

	newPos, _ := TransformCursorPosition(0, 0, MsgDeleteChars, op, "block1", "user1")
	if newPos != 0 {
		t.Errorf("Expected cursor position 0 (clamped), got %d", newPos)
	}
}

func TestTransformCursorPosition_UnknownOperation(t *testing.T) {
	newPos, _ := TransformCursorPosition(10, 10, "unknown_type", nil, "block1", "user1")
	if newPos != 10 {
		t.Errorf("Expected cursor position 10 (unchanged), got %d", newPos)
	}
}

func TestTransformCursorPosition_InvalidOperationType(t *testing.T) {
	newPos, _ := TransformCursorPosition(10, 10, "not an operation", nil, "block1", "user1")
	if newPos != 10 {
		t.Errorf("Expected cursor position 10 (unchanged), got %d", newPos)
	}
}

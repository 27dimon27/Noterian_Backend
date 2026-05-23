// file: transform_test.go
package websocket

import "testing"

func TestTransformCursorPosition_InsertBefore(t *testing.T) {
	op := &InsertCharOperation{
		BlockID:  "block1",
		Position: 5,
	}

	// Cursor at position 6, insert at 5 -> cursor should move to 7
	newPos := TransformCursorPosition(6, MsgInsertChar, op, "block1", "user1")
	if newPos != 7 {
		t.Errorf("Expected cursor position 7, got %d", newPos)
	}
}

func TestTransformCursorPosition_InsertAfter(t *testing.T) {
	op := &InsertCharOperation{
		BlockID:  "block1",
		Position: 5,
	}

	// Cursor at position 3, insert at 5 -> cursor should stay at 3
	newPos := TransformCursorPosition(3, MsgInsertChar, op, "block1", "user1")
	if newPos != 3 {
		t.Errorf("Expected cursor position 3, got %d", newPos)
	}
}

func TestTransformCursorPosition_InsertAtSamePosition(t *testing.T) {
	op := &InsertCharOperation{
		BlockID:  "block1",
		Position: 5,
	}

	// Cursor at position 5, insert at 5 -> cursor should move to 6
	newPos := TransformCursorPosition(5, MsgInsertChar, op, "block1", "user1")
	if newPos != 6 {
		t.Errorf("Expected cursor position 6, got %d", newPos)
	}
}

func TestTransformCursorPosition_InsertDifferentBlock(t *testing.T) {
	op := &InsertCharOperation{
		BlockID:  "block2",
		Position: 5,
	}

	// Cursor in a different block should not be affected
	newPos := TransformCursorPosition(10, MsgInsertChar, op, "block1", "user1")
	if newPos != 10 {
		t.Errorf("Expected cursor position 10 (unchanged), got %d", newPos)
	}
}

func TestTransformCursorPosition_DeleteBefore(t *testing.T) {
	op := &DeleteCharOperation{
		BlockID:  "block1",
		Position: 5,
	}

	// Cursor at position 7, delete at 5 -> cursor should move to 6
	newPos := TransformCursorPosition(7, MsgDeleteChar, op, "block1", "user1")
	if newPos != 6 {
		t.Errorf("Expected cursor position 6, got %d", newPos)
	}
}

func TestTransformCursorPosition_DeleteAfter(t *testing.T) {
	op := &DeleteCharOperation{
		BlockID:  "block1",
		Position: 5,
	}

	// Cursor at position 3, delete at 5 -> cursor should stay at 3
	newPos := TransformCursorPosition(3, MsgDeleteChar, op, "block1", "user1")
	if newPos != 3 {
		t.Errorf("Expected cursor position 3, got %d", newPos)
	}
}

func TestTransformCursorPosition_DeleteAtCursor(t *testing.T) {
	op := &DeleteCharOperation{
		BlockID:  "block1",
		Position: 5,
	}

	// Cursor at position 6, delete at 5 -> cursor should move to 5
	newPos := TransformCursorPosition(6, MsgDeleteChar, op, "block1", "user1")
	if newPos != 5 {
		t.Errorf("Expected cursor position 5, got %d", newPos)
	}
}

func TestTransformCursorPosition_DeleteAllBeforeCursor(t *testing.T) {
	op := &DeleteCharOperation{
		BlockID:  "block1",
		Position: 0,
	}

	// Cursor at position 5, delete at 0 -> cursor should move to 4
	newPos := TransformCursorPosition(5, MsgDeleteChar, op, "block1", "user1")
	if newPos != 4 {
		t.Errorf("Expected cursor position 4, got %d", newPos)
	}
}

func TestTransformCursorPosition_NegativeResult(t *testing.T) {
	op := &DeleteCharOperation{
		BlockID:  "block1",
		Position: 0,
	}

	// Cursor at position 0, delete at 0 -> should not go negative
	newPos := TransformCursorPosition(0, MsgDeleteChar, op, "block1", "user1")
	if newPos != 0 {
		t.Errorf("Expected cursor position 0 (clamped), got %d", newPos)
	}
}

func TestTransformCursorPosition_UnknownOperation(t *testing.T) {
	// Unknown operation type should not change cursor
	newPos := TransformCursorPosition(10, "unknown_type", nil, "block1", "user1")
	if newPos != 10 {
		t.Errorf("Expected cursor position 10 (unchanged), got %d", newPos)
	}
}

func TestTransformCursorPosition_InvalidOperationType(t *testing.T) {
	// Passing wrong operation type should not panic
	newPos := TransformCursorPosition(10, MsgInsertChar, "not an operation", "block1", "user1")
	if newPos != 10 {
		t.Errorf("Expected cursor position 10 (unchanged), got %d", newPos)
	}
}

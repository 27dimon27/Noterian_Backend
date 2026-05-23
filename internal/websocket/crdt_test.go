// file: crdt_test.go
package websocket

import (
	"strings"
	"testing"
)

func TestNewCRDTDocument(t *testing.T) {
	doc := NewCRDTDocument("user1")

	if doc == nil {
		t.Fatal("Expected document to be created")
	}

	if doc.userID != "user1" {
		t.Errorf("Expected userID 'user1', got '%s'", doc.userID)
	}

	if len(doc.chars) != 1 {
		t.Errorf("Expected 1 char (root), got %d", len(doc.chars))
	}

	if doc.head != "root:0:0" {
		t.Errorf("Expected head 'root:0:0', got '%s'", doc.head)
	}
}

func TestCRDTDocument_GenerateID(t *testing.T) {
	doc := NewCRDTDocument("user1")

	id1 := doc.GenerateID("user1")
	id2 := doc.GenerateID("user1")

	if id1 == id2 {
		t.Error("Generated IDs should be unique")
	}

	if !strings.Contains(id1, "user1") {
		t.Errorf("ID should contain userID, got '%s'", id1)
	}
}

func TestCRDTDocument_InsertChar(t *testing.T) {
	doc := NewCRDTDocument("user1")

	// Insert first character
	id1 := doc.InsertChar(0, 'a', "user1")
	if id1 == "" {
		t.Error("InsertChar should return a valid ID")
	}

	text := doc.GetText()
	if text != "a" {
		t.Errorf("Expected text 'a', got '%s'", text)
	}

	// Insert second character at position 1
	id2 := doc.InsertChar(1, 'b', "user1")
	text = doc.GetText()
	if text != "ab" {
		t.Errorf("Expected text 'ab', got '%s'", text)
	}

	// Insert character in the middle
	id3 := doc.InsertChar(1, 'c', "user1")
	text = doc.GetText()
	if text != "acb" {
		t.Errorf("Expected text 'acb', got '%s'", text)
	}

	_ = id2
	_ = id3
}

func TestCRDTDocument_DeleteChar(t *testing.T) {
	doc := NewCRDTDocument("user1")

	doc.InsertChar(0, 'a', "user1")
	doc.InsertChar(1, 'b', "user1")
	doc.InsertChar(2, 'c', "user1")

	text := doc.GetText()
	if text != "abc" {
		t.Errorf("Expected 'abc', got '%s'", text)
	}

	// Delete character at position 1 ('b')
	deletedID := doc.DeleteChar(1)
	if deletedID == "" {
		t.Error("DeleteChar should return the deleted ID")
	}

	text = doc.GetText()
	if text != "ac" {
		t.Errorf("Expected 'ac', got '%s'", text)
	}

	// Try to delete at invalid position
	deletedID = doc.DeleteChar(10)
	if deletedID != "" {
		t.Error("DeleteChar should return empty string for invalid position")
	}
}

func TestCRDTDocument_DeleteFirstChar(t *testing.T) {
	doc := NewCRDTDocument("user1")

	doc.InsertChar(0, 'a', "user1")
	doc.InsertChar(1, 'b', "user1")

	deletedID := doc.DeleteChar(0)
	if deletedID == "" {
		t.Error("DeleteChar should return the deleted ID")
	}

	text := doc.GetText()
	if text != "b" {
		t.Errorf("Expected 'b', got '%s'", text)
	}
}

func TestCRDTDocument_DeleteLastChar(t *testing.T) {
	doc := NewCRDTDocument("user1")

	doc.InsertChar(0, 'a', "user1")
	doc.InsertChar(1, 'b', "user1")
	doc.InsertChar(2, 'c', "user1")

	deletedID := doc.DeleteChar(2)
	if deletedID == "" {
		t.Error("DeleteChar should return the deleted ID")
	}

	text := doc.GetText()
	if text != "ab" {
		t.Errorf("Expected 'ab', got '%s'", text)
	}
}

func TestCRDTDocument_ApplyInsert(t *testing.T) {
	doc := NewCRDTDocument("user1")

	// Insert first character
	id1 := doc.GenerateID("user2")
	op := &InsertCharOperation{
		UniqueID: id1,
		Char:     "x",
		UserID:   "user2",
		Lamport:  1,
		PrevID:   doc.head,
	}

	doc.ApplyInsert(op)

	text := doc.GetText()
	if text != "x" {
		t.Errorf("Expected 'x', got '%s'", text)
	}

	// Apply same operation again (should be ignored)
	doc.ApplyInsert(op)
	text = doc.GetText()
	if text != "x" {
		t.Errorf("Expected still 'x', got '%s'", text)
	}
}

func TestCRDTDocument_ApplyDelete(t *testing.T) {
	doc := NewCRDTDocument("user1")

	id1 := doc.InsertChar(0, 'a', "user1")

	op := &DeleteCharOperation{
		UniqueID: id1,
		Lamport:  2,
	}

	doc.ApplyDelete(op)

	text := doc.GetText()
	if text != "" {
		t.Errorf("Expected empty string, got '%s'", text)
	}
}

func TestCRDTDocument_GetCRDTState(t *testing.T) {
	doc := NewCRDTDocument("user1")

	doc.InsertChar(0, 'a', "user1")
	doc.InsertChar(1, 'b', "user1")

	state := doc.GetCRDTState()

	if len(state) != 2 {
		t.Errorf("Expected 2 chars in state, got %d", len(state))
	}

	// Check sorting by Lamport
	if state[0].Lamport > state[1].Lamport {
		t.Error("State should be sorted by Lamport timestamp")
	}
}

func TestCRDTDocument_LoadText(t *testing.T) {
	doc := NewCRDTDocument("user1")

	initialText := "Hello, World!"
	doc.LoadText(initialText, "user1")

	loadedText := doc.GetText()
	if loadedText != initialText {
		t.Errorf("Expected '%s', got '%s'", initialText, loadedText)
	}
}

func TestCRDTDocument_ConcurrentOperations(t *testing.T) {
	doc := NewCRDTDocument("user1")

	// Simulate concurrent inserts from different users
	doc.InsertChar(0, 'a', "user1")
	doc.InsertChar(1, 'c', "user2")
	doc.InsertChar(1, 'b', "user1")

	text := doc.GetText()
	// Order might depend on CRDT logic
	if len(text) != 3 {
		t.Errorf("Expected 3 characters, got %d", len(text))
	}
}

func TestCRDTDocument_FindPositionByID(t *testing.T) {
	doc := NewCRDTDocument("user1")

	id1 := doc.InsertChar(0, 'a', "user1")
	id2 := doc.InsertChar(1, 'b', "user1")
	id3 := doc.InsertChar(2, 'c', "user1")

	pos1 := doc.findPositionByID(id1)
	pos2 := doc.findPositionByID(id2)
	pos3 := doc.findPositionByID(id3)

	if pos1 != 0 {
		t.Errorf("Expected position 0 for id1, got %d", pos1)
	}
	if pos2 != 1 {
		t.Errorf("Expected position 1 for id2, got %d", pos2)
	}
	if pos3 != 2 {
		t.Errorf("Expected position 2 for id3, got %d", pos3)
	}

	// Test with non-existent ID
	pos := doc.findPositionByID("nonexistent")
	if pos != -1 {
		t.Errorf("Expected -1 for non-existent ID, got %d", pos)
	}
}

func TestCRDTDocument_InsertInOrder(t *testing.T) {
	doc := NewCRDTDocument("user1")

	// Insert characters in different orders
	id1 := doc.GenerateID("user1")
	doc.insertInOrder(doc.head, id1)

	id2 := doc.GenerateID("user1")
	doc.insertInOrder(id1, id2)

	id3 := doc.GenerateID("user1")
	doc.insertInOrder(doc.head, id3)

	// Check order
	if len(doc.order) != 4 { // including root
		t.Errorf("Expected 4 items in order, got %d", len(doc.order))
	}
}

package websocket

import (
	"fmt"
	"sort"
)

func NewCRDTDocument() *CRDTDocument {
	return &CRDTDocument{
		characters: make([]*CRDTChar, 0),
	}
}

func (doc *CRDTDocument) GetText() string {
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	var result string
	for _, char := range doc.characters {
		if char.Visible {
			result += char.Char
		}
	}

	return result
}

func (doc *CRDTDocument) Insert(pos int, char string, userID string, lamport int64, uniqueID string) (*InsertCharOperation, error) {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	// if lamport > doc.lamportClock {
	// 	doc.lamportClock = lamport
	// }
	// doc.lamportClock++

	doc.lamportClock = max(doc.lamportClock, lamport) + 1

	insertIndex := doc.findInsertIndex(pos)

	fmt.Println(insertIndex)

	newChar := &CRDTChar{
		ID:      uniqueID,
		Char:    char,
		UserID:  userID,
		Lamport: doc.lamportClock,
		Visible: true,
	}

	doc.characters = append(doc.characters[:insertIndex], append([]*CRDTChar{newChar}, doc.characters[insertIndex:]...)...)

	return &InsertCharOperation{
		BlockID:  "",
		Position: pos,
		Char:     char,
		UserID:   userID,
		Lamport:  doc.lamportClock,
		UniqueID: uniqueID,
	}, nil
}

func (doc *CRDTDocument) Delete(pos int, userID string, lamport int64) (*DeleteCharOperation, error) {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	// if lamport > doc.lamportClock {
	// 	doc.lamportClock = lamport
	// }
	// doc.lamportClock++

	doc.lamportClock = max(doc.lamportClock, lamport) + 1

	visiblePos := 0
	for _, c := range doc.characters {
		if c.Visible {
			if visiblePos == pos {
				c.Visible = false
				return &DeleteCharOperation{
					BlockID:  "",
					Position: pos,
					UserID:   userID,
					Lamport:  doc.lamportClock,
				}, nil
			}
			visiblePos++
		}
	}

	return nil, nil
}

func (doc *CRDTDocument) ApplyInsert(op InsertCharOperation) {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	if op.Lamport > doc.lamportClock {
		doc.lamportClock = op.Lamport
	}

	for _, c := range doc.characters {
		if c.ID == op.UniqueID {
			return
		}
	}

	newChar := &CRDTChar{
		ID:      op.UniqueID,
		Char:    op.Char,
		UserID:  op.UserID,
		Lamport: op.Lamport,
		Visible: true,
	}

	insertIndex := sort.Search(len(doc.characters), func(i int) bool {
		if doc.characters[i].Lamport == op.Lamport {
			return doc.characters[i].ID > op.UniqueID
		}
		return doc.characters[i].Lamport > op.Lamport
	})

	doc.characters = append(doc.characters[:insertIndex], append([]*CRDTChar{newChar}, doc.characters[insertIndex:]...)...)
}

func (doc *CRDTDocument) ApplyDelete(op DeleteCharOperation) {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	if op.Lamport > doc.lamportClock {
		doc.lamportClock = op.Lamport
	}

	visiblePos := 0
	for _, c := range doc.characters {
		if c.Visible {
			if visiblePos == op.Position {
				c.Visible = false
				break
			}
			visiblePos++
		}
	}
}

func (doc *CRDTDocument) GetCRDTState() []CRDTChar {
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	result := make([]CRDTChar, len(doc.characters))
	for i, c := range doc.characters {
		result[i] = *c
	}
	return result
}

// func (doc *CRDTDocument) LoadCRDTState(chars []CRDTChar) {
// 	doc.mu.Lock()
// 	defer doc.mu.Unlock()

// 	doc.characters = make([]*CRDTChar, len(chars))
// 	for i, c := range chars {
// 		doc.characters[i] = &CRDTChar{
// 			ID:      c.ID,
// 			Char:    c.Char,
// 			UserID:  c.UserID,
// 			Lamport: c.Lamport,
// 			Visible: c.Visible,
// 		}

// 		if c.Lamport > doc.lamportClock {
// 			doc.lamportClock = c.Lamport
// 		}
// 	}
// }

func (doc *CRDTDocument) findInsertIndex(pos int) int {
	// doc.mu.RLock()
	// defer doc.mu.RUnlock()
	visibleCount := 0

	for i, c := range doc.characters {
		if c.Visible {
			if visibleCount == pos {
				return i
			}
			visibleCount++
		}
	}

	if pos >= visibleCount {
		return len(doc.characters)
	}

	return len(doc.characters)
}

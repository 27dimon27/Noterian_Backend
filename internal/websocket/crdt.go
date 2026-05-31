package websocket

import (
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CRDTChar struct {
	ID        string `json:"id"`
	Char      rune   `json:"char"`
	UserID    string `json:"userId"`
	Lamport   int64  `json:"lamport"`
	Visible   bool   `json:"visible"`
	PrevID    string `json:"prevId"`
	Timestamp int64  `json:"timestamp"`
}

type CRDTDocument struct {
	mu           sync.RWMutex
	chars        map[string]*CRDTChar
	head         string
	order        []string
	lamportClock int64
	counter      map[string]int64
	userID       string
}

func NewCRDTDocument(userID string) *CRDTDocument {
	doc := &CRDTDocument{
		chars:        make(map[string]*CRDTChar),
		head:         "",
		order:        make([]string, 0),
		lamportClock: 0,
		counter:      make(map[string]int64),
		userID:       userID,
	}

	rootID := "root:0:0"
	doc.chars[rootID] = &CRDTChar{
		ID:        rootID,
		Char:      0,
		UserID:    "system",
		Lamport:   0,
		Visible:   false,
		PrevID:    "",
		Timestamp: time.Now().UnixNano(),
	}
	doc.head = rootID
	doc.order = []string{rootID}

	return doc
}

func (doc *CRDTDocument) GenerateID(userID string) string {
	doc.counter[userID]++
	counter := doc.counter[userID]
	doc.lamportClock++

	return userID + ":" + strconv.FormatInt(counter, 10) + ":" + strconv.FormatInt(doc.lamportClock, 10)
}

func (doc *CRDTDocument) GetLamport() int64 {
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	return doc.lamportClock
}

func (doc *CRDTDocument) InsertChar(pos int, ch rune, userID string) string {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	prevID := doc.findPrevID(pos)

	newID := doc.GenerateID(userID)
	newChar := &CRDTChar{
		ID:        newID,
		Char:      ch,
		UserID:    userID,
		Lamport:   doc.lamportClock,
		Visible:   true,
		PrevID:    prevID,
		Timestamp: time.Now().UnixNano(),
	}

	doc.chars[newID] = newChar
	doc.insertInOrder(prevID, newID)

	return newID
}

func (doc *CRDTDocument) DeleteChar(pos int) string {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	visiblePos := 0
	for _, id := range doc.order {
		ch := doc.chars[id]
		if ch.Visible && ch.ID != "root:0:0" {
			if visiblePos == pos {
				ch.Visible = false
				ch.Timestamp = time.Now().UnixNano()
				return id
			}
			visiblePos++
		}
	}
	return ""
}

func (doc *CRDTDocument) findPrevID(pos int) string {
	visiblePos := 0
	prevID := doc.head

	for _, id := range doc.order {
		ch := doc.chars[id]
		if ch.Visible {
			if visiblePos == pos {
				return prevID
			}
			prevID = id
			visiblePos++
		}
	}
	return prevID
}

func (doc *CRDTDocument) findPositionByID(targetID string) int {
	visiblePos := 0
	for _, id := range doc.order {
		ch := doc.chars[id]
		if ch.Visible && ch.ID != "root:0:0" {
			if id == targetID {
				return visiblePos
			}
			visiblePos++
		}
	}
	return -1
}

func (doc *CRDTDocument) insertInOrder(afterID, newID string) {
	idx := -1
	for i, id := range doc.order {
		if id == afterID {
			idx = i
			break
		}
	}

	if idx == -1 {
		doc.order = append([]string{newID}, doc.order...)
		return
	}

	newOrder := make([]string, 0, len(doc.order)+1)
	newOrder = append(newOrder, doc.order[:idx+1]...)
	newOrder = append(newOrder, newID)
	newOrder = append(newOrder, doc.order[idx+1:]...)
	doc.order = newOrder
}

// func (doc *CRDTDocument) ApplyInsert(op *InsertCharOperation) {
// 	doc.mu.Lock()
// 	defer doc.mu.Unlock()

// 	if _, exists := doc.chars[op.UniqueID]; exists {
// 		return
// 	}

// 	if op.Lamport > doc.lamportClock {
// 		doc.lamportClock = op.Lamport
// 	}

// 	newChar := &CRDTChar{
// 		ID:        op.UniqueID,
// 		Char:      []rune(op.Char)[0],
// 		UserID:    op.UserID,
// 		Lamport:   op.Lamport,
// 		Visible:   true,
// 		PrevID:    op.PrevID,
// 		Timestamp: op.Timestamp,
// 	}

// 	doc.chars[op.UniqueID] = newChar
// 	doc.insertInOrder(op.PrevID, op.UniqueID)
// }

func (doc *CRDTDocument) ApplyInserts(op *InsertCharsOperation) {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	for _, uniqueID := range op.UniqueIDs {
		if _, exists := doc.chars[uniqueID]; exists {
			return
		}

		if op.Lamport > doc.lamportClock {
			doc.lamportClock = op.Lamport
		}

		newChar := &CRDTChar{
			ID:        uniqueID,
			Char:      []rune(op.Char)[0],
			UserID:    op.UserID,
			Lamport:   op.Lamport,
			Visible:   true,
			PrevID:    op.PrevID,
			Timestamp: op.Timestamp,
		}

		doc.chars[uniqueID] = newChar
		doc.insertInOrder(op.PrevID, uniqueID)
	}
}

// func (doc *CRDTDocument) ApplyDelete(op *DeleteCharOperation) {
// 	doc.mu.Lock()
// 	defer doc.mu.Unlock()

// 	if op.Lamport > doc.lamportClock {
// 		doc.lamportClock = op.Lamport
// 	}

// 	if ch, exists := doc.chars[op.UniqueID]; exists {
// 		ch.Visible = false
// 		ch.Timestamp = op.Timestamp
// 	}
// }

func (doc *CRDTDocument) ApplyDeletes(op *DeleteCharsOperation) {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	if op.Lamport > doc.lamportClock {
		doc.lamportClock = op.Lamport
	}

	for _, uniqueID := range op.UniqueIDs {
		if ch, exists := doc.chars[uniqueID]; exists {
			ch.Visible = false
			ch.Timestamp = op.Timestamp
		}
	}
}

func (doc *CRDTDocument) GetText() string {
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	var result strings.Builder
	for _, id := range doc.order {
		ch := doc.chars[id]
		if ch.Visible && ch.ID != "root:0:0" {
			result.WriteRune(ch.Char)
		}
	}
	return result.String()
}

func (doc *CRDTDocument) GetCRDTState() []CRDTChar {
	result := make([]CRDTChar, 0, len(doc.chars))
	for _, ch := range doc.chars {
		if ch.ID != "root:0:0" {
			result = append(result, *ch)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Lamport == result[j].Lamport {
			return result[i].ID < result[j].ID
		}
		return result[i].Lamport < result[j].Lamport
	})

	return result
}

func (doc *CRDTDocument) LoadText(text string, userID string) {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	for i, ch := range text {
		newID := doc.GenerateID(userID)
		prevID := doc.findPrevID(i)

		newChar := &CRDTChar{
			ID:        newID,
			Char:      ch,
			UserID:    userID,
			Lamport:   doc.lamportClock,
			Visible:   true,
			PrevID:    prevID,
			Timestamp: time.Now().UnixNano(),
		}

		doc.chars[newID] = newChar
		doc.insertInOrder(prevID, newID)
	}
}

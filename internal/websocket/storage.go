package websocket

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type BlockContentUpdate struct {
	NoteID    string
	BlockID   string
	UserID    string
	Content   string
	Timestamp int64
}

type BatchStorage struct {
	hub         *Hub
	updates     map[string]*BlockContentUpdate
	mu          sync.RWMutex
	batchTicker *time.Ticker
	saveQueue   chan *BlockContentUpdate
}

func NewBatchStorage(hub *Hub) *BatchStorage {
	bs := &BatchStorage{
		hub:         hub,
		updates:     make(map[string]*BlockContentUpdate),
		batchTicker: time.NewTicker(1 * time.Second),
		saveQueue:   make(chan *BlockContentUpdate, 1000),
	}

	return bs
}

func (bs *BatchStorage) Run(ctx context.Context) {
	go func() {
		for {
			select {
			case update := <-bs.saveQueue:
				bs.addUpdate(update)

			case <-bs.batchTicker.C:
				bs.flush()

			case <-ctx.Done():
				bs.flush()
				return
			}
		}
	}()
}

func (bs *BatchStorage) SaveBlockContent(noteID, blockID, userID, content string) {
	select {
	case bs.saveQueue <- &BlockContentUpdate{
		NoteID:    noteID,
		BlockID:   blockID,
		UserID:    userID,
		Content:   content,
		Timestamp: time.Now().UnixNano(),
	}:
	default:
		log.Printf("Save queue full for block %s", blockID)
		bs.saveSync(noteID, blockID, userID, content)
	}
}

func (bs *BatchStorage) addUpdate(update *BlockContentUpdate) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.updates[update.BlockID] = update
}

func (bs *BatchStorage) flush() {
	bs.mu.Lock()
	updates := make([]*BlockContentUpdate, 0, len(bs.updates))
	for _, update := range bs.updates {
		updates = append(updates, update)
	}
	bs.updates = make(map[string]*BlockContentUpdate)
	bs.mu.Unlock()

	if len(updates) == 0 {
		return
	}

	for _, update := range updates {
		bs.saveSync(update.NoteID, update.BlockID, update.UserID, update.Content)
	}

	log.Printf("Flushed %d block updates", len(updates))
}

func (bs *BatchStorage) saveSync(noteID, blockID, userID, content string) {
	blockUUID, err := uuid.Parse(blockID)
	if err != nil {
		log.Printf("Failed to parse block ID: %v", err)
		return
	}

	noteUUID, err := uuid.Parse(noteID)
	if err != nil {
		log.Printf("Failed to parse note ID: %v", err)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Failed to parse user ID: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = bs.hub.noteUsecase.UpdateBlockContent(ctx, blockUUID, noteUUID, userUUID, content)
	if err != nil {
		log.Printf("Failed to save block content: %v", err)
	}
}

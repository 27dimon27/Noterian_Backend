package websocket

import (
	"context"
	"fmt"
	"log/slog"
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
	logger      *slog.Logger
}

func NewBatchStorage(hub *Hub, logger *slog.Logger) *BatchStorage {
	bs := &BatchStorage{
		hub:         hub,
		updates:     make(map[string]*BlockContentUpdate),
		batchTicker: time.NewTicker(1 * time.Second),
		saveQueue:   make(chan *BlockContentUpdate, 1000),
		logger:      logger,
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
		bs.logger.Error(fmt.Sprintf("[%s:%s] %v", "ws/BatchStorage", "SaveBlockContent", fmt.Errorf("save queue full for block %s", blockID)))
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

	bs.logger.Info(fmt.Sprintf("[%s:%s] %s", "ws/BatchStorage", "flush", fmt.Sprintf("flushed %d block updates", len(updates))))
}

func (bs *BatchStorage) saveSync(noteID, blockID, userID, content string) {
	blockUUID, err := uuid.Parse(blockID)
	if err != nil {
		bs.logger.Error(fmt.Sprintf("[%s:%s] %v", "ws/BatchStorage", "saveSync", err))
		return
	}

	noteUUID, err := uuid.Parse(noteID)
	if err != nil {
		bs.logger.Error(fmt.Sprintf("[%s:%s] %v", "ws/BatchStorage", "saveSync", err))
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		bs.logger.Error(fmt.Sprintf("[%s:%s] %v", "ws/BatchStorage", "saveSync", err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, customErr := bs.hub.noteUsecase.UpdateBlockContent(ctx, blockUUID, noteUUID, userUUID, content)
	if customErr != nil {
		bs.logger.Error(customErr.Error())
	}
}

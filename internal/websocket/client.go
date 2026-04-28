package websocket

import (
	"sync"
	"time"
)

type ClientInfo struct {
	UserID     string
	UserName   string
	NoteID     string
	LastCursor CursorPosition
	Send       chan WebSocketMessage
	LastPing   int64
	mu         sync.RWMutex
}

func (c *ClientInfo) UpdateCursor(cursor CursorPosition) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastCursor = cursor
	c.LastCursor.Timestamp = time.Now().UnixNano()
}

func (c *ClientInfo) GetCursor() CursorPosition {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LastCursor
}

package websocket

import (
	"sync"
	"time"
)

type NoteRoom struct {
	mu            sync.RWMutex
	NoteID        string
	Clients       map[string]*ClientInfo
	CRDTDocuments map[string]*CRDTDocument
	SequenceID    int64
	IsDeleted     bool
	CreatedAt     int64
}

func NewNoteRoom(noteID string) *NoteRoom {
	return &NoteRoom{
		NoteID:        noteID,
		Clients:       make(map[string]*ClientInfo),
		CRDTDocuments: make(map[string]*CRDTDocument),
		SequenceID:    0,
		IsDeleted:     false,
		CreatedAt:     time.Now().Unix(),
	}
}

func (r *NoteRoom) AddClient(client *ClientInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Clients[client.UserID] = client
}

func (r *NoteRoom) RemoveClient(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Clients, userID)
}

func (r *NoteRoom) GetClient(userID string) (*ClientInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	client, ok := r.Clients[userID]
	return client, ok
}

func (r *NoteRoom) GetAllCursors() []UserCursor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cursors := make([]UserCursor, 0, len(r.Clients))
	for _, client := range r.Clients {
		cursor := client.GetCursor()
		cursors = append(cursors, UserCursor{
			UserID:   client.UserID,
			UserName: client.UserName,
			Cursor:   cursor,
		})
	}
	return cursors
}

func (r *NoteRoom) GetCRDTDocument(blockID string) (*CRDTDocument, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	doc, ok := r.CRDTDocuments[blockID]
	return doc, ok
}

func (r *NoteRoom) SetCRDTDocument(blockID string, doc *CRDTDocument) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CRDTDocuments[blockID] = doc
}

func (r *NoteRoom) DeleteCRDTDocument(blockID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.CRDTDocuments, blockID)
}

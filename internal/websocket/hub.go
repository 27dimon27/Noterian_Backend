package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

func NewHub(noteUsecase NoteUsecaseInterface, profileUsecase ProfileUsecaseInterface) *Hub {
	return &Hub{
		rooms:          make(map[string]*NoteRoom),
		register:       make(chan ClientInfo),
		unregister:     make(chan ClientInfo),
		broadcast:      make(chan BroadcastMessage, 256),
		noteUsecase:    noteUsecase,
		profileUsecase: profileUsecase,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case msg := <-h.broadcast:
			h.handleBroadcast(msg)
		}
	}
}

func (h *Hub) handleRegister(client ClientInfo) {
	h.mu.Lock()
	room, exists := h.rooms[client.NoteID]
	h.mu.Unlock()

	if !exists {
		room = &NoteRoom{
			NoteID:            client.NoteID,
			Clients:           make(map[string]ClientInfo),
			CRDTDocuments:     make(map[string]*CRDTDocument),
			PendingOperations: make([]PendingOperation, 0),
		}
		h.rooms[client.NoteID] = room
	}

	room.mu.Lock()
	if room.IsDeleted {
		room.mu.Unlock()
		client.Send <- WebSocketMessage{
			Type:     MsgDeleteNote,
			UserID:   client.UserID,
			UserName: client.UserName,
			NoteID:   client.NoteID,
			Msg: ErrMessage{
				Error: "Note has been deleted",
			},
		}
		return
	}

	room.Clients[client.UserID] = client
	room.mu.Unlock()

	h.sendSyncState(client, room)

	h.broadcastToRoom(client.NoteID, WebSocketMessage{
		Type:     MsgUserJoined,
		UserID:   client.UserID,
		UserName: client.UserName,
	}, client.UserID)
}

func (h *Hub) handleUnregister(client ClientInfo) {
	h.mu.Lock()
	room, exists := h.rooms[client.NoteID]
	h.mu.Unlock()

	if !exists {
		return
	}

	room.mu.Lock()
	delete(room.Clients, client.UserID)
	room.mu.Unlock()

	h.broadcastToRoom(client.NoteID, WebSocketMessage{
		Type:     MsgUserLeft,
		UserID:   client.UserID,
		UserName: client.UserName,
	}, client.UserID)

	room.mu.RLock()
	isEmpty := len(room.Clients) == 0
	room.mu.RUnlock()

	if isEmpty {
		h.mu.Lock()
		delete(h.rooms, client.NoteID)
		h.mu.Unlock()
	}
}

func (h *Hub) handleBroadcast(msg BroadcastMessage) {
	h.mu.RLock()
	room, exists := h.rooms[msg.NoteID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	room.mu.RLock()
	for userID, client := range room.Clients {
		if userID != msg.Exclude {
			select {
			case client.Send <- msg.Message:
			default:
				go func(c ClientInfo) {
					h.unregister <- c
				}(client)
			}
		}
	}
	room.mu.RUnlock()
}

func (h *Hub) broadcastToRoom(noteID string, message WebSocketMessage, excludeUserID string) {
	excludeUserID = "" // убрать после тестов
	h.broadcast <- BroadcastMessage{
		NoteID:  noteID,
		Message: message,
		Exclude: excludeUserID,
	}
}

func (h *Hub) broadcastUpdatedCursors(room *NoteRoom) {
	cursors := room.getAllCursors()

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type: MsgCursorMove,
		Msg:  cursors,
	}, "")
}

func (h *Hub) sendSyncState(client ClientInfo, room *NoteRoom) {
	noteID, _ := uuid.Parse(client.NoteID)
	userID, _ := uuid.Parse(client.UserID)

	note, err := h.noteUsecase.GetNote(context.Background(), noteID, userID)
	if err != nil {
		client.Send <- WebSocketMessage{
			Type:     MsgError,
			UserID:   client.UserID,
			UserName: client.UserName,
			NoteID:   client.NoteID,
			Msg: ErrMessage{
				Error: err.Error(),
			},
		}
		return
	}

	blocks, err := h.noteUsecase.GetBlocks(context.Background(), noteID)
	if err != nil {
		client.Send <- WebSocketMessage{
			Type:     MsgError,
			UserID:   client.UserID,
			UserName: client.UserName,
			NoteID:   client.NoteID,
			Msg: ErrMessage{
				Error: err.Error(),
			},
		}
		return
	}

	room.mu.Lock()
	for _, block := range blocks {
		if block.BlockTypeID == 1 {
			crdt, exists := room.CRDTDocuments[block.ID.String()]
			if !exists {
				crdt = NewCRDTDocument()
				h.loadTextToCRDT(crdt, client.UserID, block.Content, block.ID.String())
				room.CRDTDocuments[block.ID.String()] = crdt
			}
		}
	}
	room.mu.Unlock()

	crdtStates := make(map[string][]CRDTChar)

	room.mu.RLock()
	for blockID, crdt := range room.CRDTDocuments {
		crdtStates[blockID] = crdt.GetCRDTState()
	}
	room.mu.RUnlock()

	cursors := room.getAllCursors()

	client.Send <- WebSocketMessage{
		Type:     MsgSyncState,
		UserID:   client.UserID,
		UserName: client.UserName,
		NoteID:   client.NoteID,
		Msg: map[string]any{
			"note":           note,
			"blocks":         blocks,
			"crdt_states":    crdtStates,
			"cursors":        cursors,
			"owner_id":       note.UserID.String(),
			"is_public":      note.IsPublic,
			"note_title":     note.Title,
			"sync_timestamp": time.Now().Unix(),
		},
	}
}

func (h *Hub) loadTextToCRDT(crdt *CRDTDocument, userID string, text string, blockID string) {
	for i, ch := range text {
		op := InsertCharOperation{
			BlockID:   blockID,
			Position:  i,
			Char:      string(ch),
			Lamport:   int64(i),
			UniqueID:  fmt.Sprintf("%s:%d", blockID, i),
			Timestamp: time.Now().UnixNano(),
		}
		crdt.ApplyInsert(op, userID)
	}
}

func (h *Hub) HandleOperation(noteID string, userID string, msg WebSocketMessage) {
	h.mu.RLock()
	room, exists := h.rooms[noteID]
	h.mu.RUnlock()

	if !exists || room.IsDeleted {
		return
	}

	switch msg.Type {
	case MsgCursorMove:
		var cursor CursorPosition
		if err := mapToStruct(msg.Msg, &cursor); err == nil {
			h.handleCursorMove(room, userID, cursor)
		}
	case MsgInsertChar:
		var op InsertCharOperation
		if err := mapToStruct(msg.Msg, &op); err == nil {
			op.Timestamp = time.Now().UnixNano()
			h.handleInsertChar(room, userID, op, msg.IsLocal)
		}
	case MsgDeleteChar:
		var op DeleteCharOperation
		if err := mapToStruct(msg.Msg, &op); err == nil {
			op.Timestamp = time.Now().UnixNano()
			h.handleDeleteChar(room, userID, op, msg.IsLocal)
		}
	case MsgApplyFormatting:
		var op FormattingOperation
		if err := mapToStruct(msg.Msg, &op); err == nil {
			op.Timestamp = time.Now().UnixNano()
			h.handleApplyFormatting(room, userID, op)
		}
	case MsgCreateBlock:
		var blockData map[string]any
		if err := mapToStruct(msg.Msg, &blockData); err == nil {
			h.handleCreateBlock(room, userID, blockData)
		}
	case MsgDeleteBlock:
		var blockID string
		if err := mapToStruct(msg.Msg, &blockID); err == nil {
			h.handleDeleteBlock(room, userID, blockID)
		}
	case MsgMoveBlock:
		var moveData map[string]any
		if err := mapToStruct(msg.Msg, &moveData); err == nil {
			h.handleMoveBlock(room, userID, moveData)
		}
	case MsgUpdateNoteTitle:
		var title string
		if err := mapToStruct(msg.Msg, &title); err == nil {
			h.handleUpdateNoteTitle(room, userID, title)
		}
	case MsgUpdateNotePublic:
		var isPublic bool
		if err := mapToStruct(msg.Msg, &isPublic); err == nil {
			h.handleUpdateNotePublic(room, userID, isPublic)
		}
	case MsgDeleteNote:
		h.handleDeleteNote(room, userID)
	}
}

func (h *Hub) handleCursorMove(room *NoteRoom, userID string, cursor CursorPosition) {
	room.mu.Lock()
	client, exists := room.Clients[userID]
	if !exists {
		room.mu.Unlock()
		return
	}

	client.LastCursor = cursor
	room.Clients[userID] = client
	room.mu.Unlock()

	h.broadcastUpdatedCursors(room)
}

func (h *Hub) handleInsertChar(room *NoteRoom, userID string, op InsertCharOperation, isLocal bool) {
	room.mu.Lock()
	client, exists := room.Clients[userID]
	if !exists {
		room.mu.Unlock()
		return
	}
	room.mu.Unlock()

	room.mu.Lock()
	crdt, exists := room.CRDTDocuments[op.BlockID]
	if !exists {
		crdt = NewCRDTDocument()
		room.CRDTDocuments[op.BlockID] = crdt
	}
	room.mu.Unlock()

	operationID := fmt.Sprintf("%s:%d:%d", userID, op.Timestamp, op.Position)

	if isLocal {
		result, err := crdt.Insert(op.Position, op.Char, userID, op.Lamport, op.UniqueID)
		if err != nil {
			return
		}

		result.BlockID = op.BlockID
		result.Timestamp = op.Timestamp

		blockID, _ := uuid.Parse(op.BlockID)
		noteID, _ := uuid.Parse(room.NoteID)
		userUUID, _ := uuid.Parse(userID)

		newContent := crdt.GetText()
		_, err = h.noteUsecase.UpdateBlockContent(context.Background(), blockID, noteID, userUUID, newContent)
		if err != nil {
			return
		}

		pendingOp := PendingOperation{
			ID:        operationID,
			UserID:    userID,
			Type:      MsgInsertChar,
			Operation: result,
			Timestamp: op.Timestamp,
		}

		h.addPendingOperationAndUpdateCursors(room, pendingOp)

		h.broadcastToRoom(room.NoteID, WebSocketMessage{
			Type:     MsgInsertChar,
			IsLocal:  false,
			UserID:   userID,
			UserName: client.UserName,
			NoteID:   room.NoteID,
			Msg:      result,
		}, userID)

		h.broadcastUpdatedCursors(room)

		go func() {
			time.Sleep(100 * time.Millisecond)
			h.clearPendingOperations(room, operationID)
		}()
	} else {
		crdt.ApplyInsert(op, userID)

		pendingOp := PendingOperation{
			ID:        operationID,
			UserID:    userID,
			Type:      MsgInsertChar,
			Operation: op,
			Timestamp: op.Timestamp,
		}

		h.addPendingOperationAndUpdateCursors(room, pendingOp)

		h.broadcastUpdatedCursors(room)

		go func() {
			time.Sleep(100 * time.Millisecond)
			h.clearPendingOperations(room, operationID)
		}()
	}
}

func (h *Hub) handleDeleteChar(room *NoteRoom, userID string, op DeleteCharOperation, isLocal bool) {
	room.mu.Lock()
	client, exists := room.Clients[userID]
	if !exists {
		room.mu.Unlock()
		return
	}
	room.mu.Unlock()

	room.mu.RLock()
	crdt, exists := room.CRDTDocuments[op.BlockID]
	room.mu.RUnlock()

	if !exists {
		return
	}

	operationID := fmt.Sprintf("%s:%d:%d", userID, op.Timestamp, op.Position)

	if isLocal {
		result, err := crdt.Delete(op.Position, userID, op.Lamport)
		if err != nil {
			return
		}

		result.BlockID = op.BlockID
		result.Timestamp = op.Timestamp

		blockID, _ := uuid.Parse(op.BlockID)
		noteID, _ := uuid.Parse(room.NoteID)
		userUUID, _ := uuid.Parse(userID)

		newContent := crdt.GetText()
		_, err = h.noteUsecase.UpdateBlockContent(context.Background(), blockID, noteID, userUUID, newContent)
		if err != nil {
			return
		}

		pendingOp := PendingOperation{
			ID:        operationID,
			UserID:    userID,
			Type:      MsgDeleteChar,
			Operation: result,
			Timestamp: op.Timestamp,
		}

		h.addPendingOperationAndUpdateCursors(room, pendingOp)

		h.broadcastToRoom(room.NoteID, WebSocketMessage{
			Type:     MsgDeleteChar,
			IsLocal:  false,
			UserID:   userID,
			UserName: client.UserName,
			NoteID:   room.NoteID,
			Msg:      result,
		}, userID)

		h.broadcastUpdatedCursors(room)

		go func() {
			time.Sleep(100 * time.Millisecond)
			h.clearPendingOperations(room, operationID)
		}()
	} else {
		crdt.ApplyDelete(op)

		pendingOp := PendingOperation{
			ID:        operationID,
			UserID:    userID,
			Type:      MsgDeleteChar,
			Operation: op,
			Timestamp: op.Timestamp,
		}

		h.addPendingOperationAndUpdateCursors(room, pendingOp)

		h.broadcastUpdatedCursors(room)

		go func() {
			time.Sleep(100 * time.Millisecond)
			h.clearPendingOperations(room, operationID)
		}()
	}
}

func (h *Hub) handleApplyFormatting(room *NoteRoom, userID string, op FormattingOperation) {
	room.mu.Lock()
	client, exists := room.Clients[userID]
	if !exists {
		room.mu.Unlock()
		return
	}
	room.mu.Unlock()

	room.mu.Lock()
	op.SequenceID = room.SequenceID
	room.SequenceID++
	room.mu.Unlock()

	blockID, _ := uuid.Parse(op.BlockID)
	noteID, _ := uuid.Parse(room.NoteID)
	userUUID, _ := uuid.Parse(userID)

	formattingRange := models.FormattingRange{
		StartPos:  op.StartPos,
		EndPos:    op.EndPos,
		Bold:      op.Bold,
		Italic:    op.Italic,
		Underline: op.Underline,
		TextAlign: op.TextAlign,
	}

	_, err := h.noteUsecase.UpdateBlockFormatting(context.Background(), blockID, noteID, userUUID, formattingRange)
	if err != nil {
		return
	}

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type:     MsgApplyFormatting,
		UserID:   userID,
		UserName: client.UserName,
		NoteID:   room.NoteID,
		Msg:      op,
	}, userID)
}

func (h *Hub) handleCreateBlock(room *NoteRoom, userID string, blockData map[string]any) {
	room.mu.Lock()
	client, exists := room.Clients[userID]
	if !exists {
		room.mu.Unlock()
		return
	}
	room.mu.Unlock()

	noteID, _ := uuid.Parse(room.NoteID)
	userUUID, _ := uuid.Parse(userID)

	block := models.Block{
		BlockTypeID: int(blockData["blockTypeId"].(float64)),
		Position:    int(blockData["position"].(float64)),
	}

	createdBlock, err := h.noteUsecase.CreateBlock(context.Background(), noteID, userUUID, block)
	if err != nil {
		return
	}

	if createdBlock.BlockTypeID == 1 {
		room.mu.Lock()
		room.CRDTDocuments[createdBlock.ID.String()] = NewCRDTDocument()
		room.mu.Unlock()
	}

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type:     MsgCreateBlock,
		UserID:   userID,
		UserName: client.UserName,
		NoteID:   room.NoteID,
		Msg:      createdBlock,
	}, userID)
}

func (h *Hub) handleDeleteBlock(room *NoteRoom, userID string, blockID string) {
	room.mu.Lock()
	client, exists := room.Clients[userID]
	if !exists {
		room.mu.Unlock()
		return
	}
	room.mu.Unlock()

	noteID, _ := uuid.Parse(room.NoteID)
	userUUID, _ := uuid.Parse(userID)
	blockUUID, _ := uuid.Parse(blockID)

	err := h.noteUsecase.DeleteBlock(context.Background(), blockUUID, noteID, userUUID)
	if err != nil {
		return
	}

	room.mu.Lock()
	delete(room.CRDTDocuments, blockID)
	room.mu.Unlock()

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type:     MsgDeleteBlock,
		UserID:   userID,
		UserName: client.UserName,
		NoteID:   room.NoteID,
		Msg:      blockID,
	}, userID)
}

func (h *Hub) handleMoveBlock(room *NoteRoom, userID string, moveData map[string]any) {
	room.mu.Lock()
	client, exists := room.Clients[userID]
	if !exists {
		room.mu.Unlock()
		return
	}
	room.mu.Unlock()

	noteID, _ := uuid.Parse(room.NoteID)
	userUUID, _ := uuid.Parse(userID)
	blockUUID, _ := uuid.Parse(moveData["blockId"].(string))
	newPosition := int(moveData["newPosition"].(float64))

	_, err := h.noteUsecase.MoveBlock(context.Background(), blockUUID, noteID, userUUID, newPosition)
	if err != nil {
		return
	}

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type:     MsgMoveBlock,
		UserID:   userID,
		UserName: client.UserName,
		NoteID:   room.NoteID,
		Msg:      moveData,
	}, userID)
}

func (h *Hub) handleUpdateNoteTitle(room *NoteRoom, userID string, newTitle string) {
	if !h.isNoteOwner(room.NoteID, userID) {
		return
	}

	room.mu.Lock()
	client, exists := room.Clients[userID]
	if !exists {
		room.mu.Unlock()
		return
	}
	room.mu.Unlock()

	noteID, _ := uuid.Parse(room.NoteID)
	userUUID, _ := uuid.Parse(userID)

	note, err := h.noteUsecase.GetNote(context.Background(), noteID, userUUID)
	if err != nil {
		return
	}

	note.Title = newTitle

	_, err = h.noteUsecase.UpdateNote(context.Background(), noteID, *note, userUUID)
	if err != nil {
		return
	}

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type:     MsgUpdateNoteTitle,
		UserID:   userID,
		UserName: client.UserName,
		NoteID:   room.NoteID,
		Msg:      newTitle,
	}, userID)
}

func (h *Hub) handleUpdateNotePublic(room *NoteRoom, userID string, isPublic bool) {
	if !h.isNoteOwner(room.NoteID, userID) {
		return
	}

	noteID, _ := uuid.Parse(room.NoteID)
	userUUID, _ := uuid.Parse(userID)

	note, err := h.noteUsecase.GetNote(context.Background(), noteID, userUUID)
	if err != nil {
		return
	}

	note.IsPublic = isPublic

	_, err = h.noteUsecase.UpdateNote(context.Background(), noteID, *note, userUUID)
	if err != nil {
		return
	}

	if !isPublic {
		room.mu.Lock()
		clients := make([]ClientInfo, 0, len(room.Clients))
		for _, client := range room.Clients {
			clients = append(clients, client)
		}

		for _, client := range clients {
			client.Send <- WebSocketMessage{
				Type:     MsgNotePrivate,
				UserID:   userID,
				UserName: client.UserName,
				NoteID:   room.NoteID,
				Msg:      "Note is now private. WebSocket connection closed.",
			}
			close(client.Send)
		}

		room.Clients = make(map[string]ClientInfo)

		room.IsDeleted = true

		room.mu.Unlock()

		h.mu.Lock()
		delete(h.rooms, room.NoteID)
		h.mu.Unlock()
	}
}

func (h *Hub) handleDeleteNote(room *NoteRoom, userID string) {
	if !h.isNoteOwner(room.NoteID, userID) {
		return
	}

	noteID, _ := uuid.Parse(room.NoteID)
	userUUID, _ := uuid.Parse(userID)

	err := h.noteUsecase.DeleteNote(context.Background(), noteID, userUUID)
	if err != nil {
		return
	}

	room.mu.Lock()

	clients := make([]ClientInfo, 0, len(room.Clients))
	for _, client := range room.Clients {
		clients = append(clients, client)
	}

	for _, client := range clients {
		client.Send <- WebSocketMessage{
			Type:     MsgNoteDeleted,
			UserID:   userID,
			UserName: client.UserName,
			NoteID:   room.NoteID,
			Msg:      "Note is now deleted. WebSocket connection closed.",
		}
		close(client.Send)
	}

	room.Clients = make(map[string]ClientInfo)

	room.IsDeleted = true

	room.mu.Unlock()

	h.mu.Lock()
	delete(h.rooms, room.NoteID)
	h.mu.Unlock()
}

func (h *Hub) isNoteOwner(noteID string, userID string) bool {
	noteUUID, _ := uuid.Parse(noteID)
	userUUID, _ := uuid.Parse(userID)

	note, err := h.noteUsecase.GetNote(context.Background(), noteUUID, userUUID)
	if err != nil || note == nil {
		return false
	}

	return note.UserID == userUUID
}

func (r *NoteRoom) getAllCursors() []UserCursor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cursors := make([]UserCursor, 0, len(r.Clients))
	for _, client := range r.Clients {
		cursors = append(cursors, UserCursor{
			UserID:   client.UserID,
			UserName: client.UserName,
			Cursor:   client.LastCursor,
		})
	}
	return cursors
}

func (h *Hub) addPendingOperationAndUpdateCursors(room *NoteRoom, op PendingOperation) {
	room.mu.Lock()
	defer room.mu.Unlock()

	room.PendingOperations = append(room.PendingOperations, op)

	h.updateCursorsPositions(room)
}

func (h *Hub) clearPendingOperations(room *NoteRoom, operationID string) {
	room.mu.Lock()
	defer room.mu.Unlock()

	for i, op := range room.PendingOperations {
		if op.ID == operationID {
			room.PendingOperations = append(room.PendingOperations[:i], room.PendingOperations[i+1:]...)
			break
		}
	}
}

func (h *Hub) updateCursorsPositions(room *NoteRoom) {
	allOps := room.PendingOperations

	for userID, client := range room.Clients {
		tempCursor := client.LastCursor

		fmt.Println(tempCursor)

		for _, op := range allOps {
			switch op.Type {
			case MsgInsertChar:
				if insertOp, ok := op.Operation.(*InsertCharOperation); ok {
					if insertOp.BlockID == tempCursor.BlockID {
						if insertOp.Position < tempCursor.Position {
							tempCursor.Position++
						}
					}
				}
			case MsgDeleteChar:
				if deleteOp, ok := op.Operation.(*DeleteCharOperation); ok {
					if deleteOp.BlockID == tempCursor.BlockID {
						if deleteOp.Position < tempCursor.Position {
							tempCursor.Position--
						}
					}
				}
			}
		}

		fmt.Println(tempCursor)

		client.LastCursor = tempCursor
		room.Clients[userID] = client
	}
}

func mapToStruct(data any, target any) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, target)
}

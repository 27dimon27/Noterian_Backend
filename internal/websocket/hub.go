package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

func NewHub(noteUsecase NoteUsecaseInterface, profileUsecase ProfileUsecaseInterface) *Hub {
	return &Hub{
		rooms:          make(map[string]*NoteRoom),
		register:       make(chan *ClientInfo),
		unregister:     make(chan *ClientInfo),
		broadcast:      make(chan *BroadcastMessage, 256),
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

func (h *Hub) handleRegister(client *ClientInfo) {
	h.mu.Lock()
	room, exists := h.rooms[client.NoteID]
	h.mu.Unlock()

	if !exists {
		room = &NoteRoom{
			NoteID:         client.NoteID,
			Clients:        make(map[string]*ClientInfo),
			CRDTDocuments:  make(map[string]*CRDTDocument),
			OperationQueue: make([]Operation, 0),
		}
		h.rooms[client.NoteID] = room
	}

	room.mu.Lock()
	if room.IsDeleted {
		room.mu.Unlock()
		client.Send <- WebSocketMessage{
			Type: MsgNoteDeleted,
			Data: map[string]string{"message": "Note has been deleted"},
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

func (h *Hub) handleUnregister(client *ClientInfo) {
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

func (h *Hub) handleBroadcast(msg *BroadcastMessage) {
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
				go func(c *ClientInfo) {
					h.unregister <- c
				}(client)
			}
		}
	}
	room.mu.RUnlock()
}

func (h *Hub) broadcastToRoom(noteID string, message WebSocketMessage, excludeUserID string) {
	excludeUserID = "" // убрать после тестов
	h.broadcast <- &BroadcastMessage{
		NoteID:  noteID,
		Message: message,
		Exclude: excludeUserID,
	}
}

func (h *Hub) sendSyncState(client *ClientInfo, room *NoteRoom) {
	noteID, _ := uuid.Parse(client.NoteID)
	userID, _ := uuid.Parse(client.UserID)

	note, err := h.noteUsecase.GetNote(context.Background(), noteID, userID)
	if err != nil {
		client.Send <- WebSocketMessage{
			Type: MsgError,
			Data: map[string]string{"error": err.Error()},
		}
		return
	}

	blocks, err := h.noteUsecase.GetBlocks(context.Background(), noteID)
	if err != nil {
		client.Send <- WebSocketMessage{
			Type: MsgError,
			Data: map[string]string{"error": err.Error()},
		}
		return
	}

	room.mu.Lock()
	for _, block := range blocks {
		if block.BlockTypeID == 1 {
			crdt, exists := room.CRDTDocuments[block.ID.String()]
			if !exists {
				crdt = NewCRDTDocument()
				h.loadTextToCRDT(crdt, block.Content, block.ID.String())
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

	client.Send <- WebSocketMessage{
		Type: MsgSyncState,
		Data: map[string]interface{}{
			"note":        note,
			"blocks":      blocks,
			"crdt_states": crdtStates,
		},
	}
}

func (h *Hub) loadTextToCRDT(crdt *CRDTDocument, text string, blockID string) {
	for i, ch := range text {
		op := InsertCharOperation{
			BlockID:  blockID,
			Position: i,
			Char:     string(ch),
			UserID:   "system",
			Lamport:  int64(i),
			UniqueID: fmt.Sprintf("%s:%d", blockID, i),
		}
		crdt.ApplyInsert(op)
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
		if err := mapToStruct(msg.Data, &cursor); err == nil {
			h.handleCursorMove(room, userID, cursor)
		}
	case MsgInsertChar:
		var op InsertCharOperation
		if err := mapToStruct(msg.Data, &op); err == nil {
			h.handleInsertChar(room, userID, op)
		}
	case MsgDeleteChar:
		var op DeleteCharOperation
		if err := mapToStruct(msg.Data, &op); err == nil {
			h.handleDeleteChar(room, userID, op)
		}
	case MsgApplyFormatting:
		var op FormattingOperation
		if err := mapToStruct(msg.Data, &op); err == nil {
			h.handleApplyFormatting(room, userID, op)
		}
	case MsgCreateBlock:
		var blockData map[string]interface{}
		if err := mapToStruct(msg.Data, &blockData); err == nil {
			h.handleCreateBlock(room, userID, blockData)
		}
	case MsgDeleteBlock:
		var blockID string
		if err := mapToStruct(msg.Data, &blockID); err == nil {
			h.handleDeleteBlock(room, userID, blockID)
		}
	case MsgMoveBlock:
		var moveData map[string]interface{}
		if err := mapToStruct(msg.Data, &moveData); err == nil {
			h.handleMoveBlock(room, userID, moveData)
		}
	case MsgUpdateNoteTitle:
		var title string
		if err := mapToStruct(msg.Data, &title); err == nil {
			h.handleUpdateNoteTitle(room, userID, title)
		}
	case MsgUpdateNotePublic:
		var isPublic bool
		if err := mapToStruct(msg.Data, &isPublic); err == nil {
			h.handleUpdateNotePublic(room, userID, isPublic)
		}
	case MsgDeleteNote:
		h.handleDeleteNote(room, userID)
	}
}

func mapToStruct(data interface{}, target interface{}) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, target)
}

func (h *Hub) handleCursorMove(room *NoteRoom, userID string, cursor CursorPosition) {
	// cursor, ok := msg.Data.(CursorPosition)
	// if !ok {
	// 	return
	// }

	room.mu.Lock()
	if client, exists := room.Clients[userID]; exists {
		client.LastCursor = &cursor
	}
	room.mu.Unlock()

	cursors := room.getAllCursors()

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type: MsgCursorsUpdate,
		Data: cursors,
	}, userID)
}

func (h *Hub) handleInsertChar(room *NoteRoom, userID string, op InsertCharOperation) {
	// fmt.Println(1)
	// fmt.Println(msg.Data)
	// op, ok := msg.Data.(InsertCharOperation)
	// fmt.Println(op)
	// if !ok {
	// 	fmt.Println(2)
	// 	return
	// }

	room.mu.Lock()
	crdt, exists := room.CRDTDocuments[op.BlockID]
	if !exists {
		crdt = NewCRDTDocument()
		room.CRDTDocuments[op.BlockID] = crdt
	}
	room.mu.Unlock()

	isLocal := userID == op.UserID

	if isLocal {
		result, err := crdt.Insert(op.Position, op.Char, userID, op.Lamport, op.UniqueID)
		if err != nil {
			return
		}

		result.BlockID = op.BlockID

		blockID, _ := uuid.Parse(op.BlockID)
		noteID, _ := uuid.Parse(room.NoteID)
		userUUID, _ := uuid.Parse(userID)

		newContent := crdt.GetText()
		_, err = h.noteUsecase.UpdateBlockContent(context.Background(), blockID, noteID, userUUID, newContent)
		if err != nil {
			return
		}

		h.broadcastToRoom(room.NoteID, WebSocketMessage{
			Type: MsgOperation,
			Data: result,
		}, userID)
	} else {
		crdt.ApplyInsert(op)
	}
}

func (h *Hub) handleDeleteChar(room *NoteRoom, userID string, op DeleteCharOperation) {
	// op, ok := msg.Data.(DeleteCharOperation)
	// if !ok {
	// 	return
	// }

	room.mu.RLock()
	crdt, exists := room.CRDTDocuments[op.BlockID]
	room.mu.RUnlock()

	if !exists {
		return
	}

	isLocal := userID == op.UserID

	if isLocal {
		result, err := crdt.Delete(op.Position, userID, op.Lamport)
		if err != nil {
			return
		}

		result.BlockID = op.BlockID

		blockID, _ := uuid.Parse(op.BlockID)
		noteID, _ := uuid.Parse(room.NoteID)
		userUUID, _ := uuid.Parse(userID)

		newContent := crdt.GetText()
		_, err = h.noteUsecase.UpdateBlockContent(context.Background(), blockID, noteID, userUUID, newContent)
		if err != nil {
			return
		}

		h.broadcastToRoom(room.NoteID, WebSocketMessage{
			Type: MsgOperation,
			Data: result,
		}, userID)
	} else {
		crdt.ApplyDelete(op)
	}
}

func (h *Hub) handleApplyFormatting(room *NoteRoom, userID string, op FormattingOperation) {
	// op, ok := msg.Data.(FormattingOperation)
	// if !ok {
	// 	return
	// }

	room.mu.Lock()
	op.SequenceID = room.SequenceID
	room.SequenceID++
	room.OperationQueue = append(room.OperationQueue, Operation{ // а где юзается OperationQueue?
		ID:   uuid.New().String(),
		Type: OpApplyFormatting,
		Data: map[string]interface{}{
			"formatting": op,
		},
		SequenceID: op.SequenceID,
		UserID:     userID,
	})
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
		Type: MsgOperation,
		Data: op,
	}, userID)
}

func (h *Hub) handleCreateBlock(room *NoteRoom, userID string, blockData map[string]interface{}) {
	// if !h.isNoteOwner(room.NoteID, userID) {
	// 	return
	// }

	// blockData, ok := msg.Data.(map[string]interface{})
	// if !ok {
	// 	return
	// }

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
		Type: MsgOperation,
		Data: map[string]interface{}{
			"type":  "create_block",
			"block": createdBlock,
		},
	}, userID)
}

func (h *Hub) handleDeleteBlock(room *NoteRoom, userID string, blockID string) {
	// if !h.isNoteOwner(room.NoteID, userID) {
	// 	return
	// }

	// blockID, ok := msg.Data.(string)
	// if !ok {
	// 	return
	// }

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
		Type: MsgOperation,
		Data: map[string]string{
			"type":    "delete_block",
			"blockId": blockID,
		},
	}, userID)
}

func (h *Hub) handleMoveBlock(room *NoteRoom, userID string, moveData map[string]interface{}) {
	// if !h.isNoteOwner(room.NoteID, userID) {
	// 	return
	// }

	// moveData, ok := msg.Data.(map[string]interface{})
	// if !ok {
	// 	return
	// }

	noteID, _ := uuid.Parse(room.NoteID)
	userUUID, _ := uuid.Parse(userID)
	blockUUID, _ := uuid.Parse(moveData["blockId"].(string))
	newPosition := int(moveData["newPosition"].(float64))

	_, err := h.noteUsecase.MoveBlock(context.Background(), blockUUID, noteID, userUUID, newPosition)
	if err != nil {
		return
	}

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type: MsgOperation,
		Data: moveData,
	}, userID)
}

func (h *Hub) handleUpdateNoteTitle(room *NoteRoom, userID string, newTitle string) {
	if !h.isNoteOwner(room.NoteID, userID) {
		return
	}

	// newTitle, ok := msg.Data.(string)
	// if !ok {
	// 	return
	// }

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
		Type: MsgOperation,
		Data: map[string]string{
			"type":  "update_title",
			"title": newTitle,
		},
	}, userID)
}

func (h *Hub) handleUpdateNotePublic(room *NoteRoom, userID string, isPublic bool) {
	if !h.isNoteOwner(room.NoteID, userID) {
		return
	}

	// isPublic, ok := msg.Data.(bool)
	// if !ok {
	// 	return
	// }

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
		clients := make([]*ClientInfo, 0, len(room.Clients))
		for _, client := range room.Clients {
			clients = append(clients, client)
		}

		for _, client := range clients {
			client.Send <- WebSocketMessage{
				Type: MsgNotePrivate,
				Data: map[string]string{
					"message": "Note is now private. WebSocket connection closed.",
				},
			}
			close(client.Send)
		}

		room.Clients = make(map[string]*ClientInfo)

		room.IsDeleted = true

		room.mu.Unlock()

		h.mu.Lock()
		delete(h.rooms, room.NoteID)
		h.mu.Unlock()
	}

	// h.broadcastToRoom(room.NoteID, WebSocketMessage{
	// 	Type: MsgOperation,
	// 	Data: map[string]interface{}{
	// 		"type":      "update_public",
	// 		"is_public": isPublic,
	// 	},
	// }, userID)
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

	clients := make([]*ClientInfo, 0, len(room.Clients))
	for _, client := range room.Clients {
		clients = append(clients, client)
	}

	for _, client := range clients {
		client.Send <- WebSocketMessage{
			Type: MsgNoteDeleted,
			Data: map[string]string{
				"message": "Note is now deleted. WebSocket connection closed.",
			},
		}
		close(client.Send)
	}

	room.Clients = make(map[string]*ClientInfo)

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

func (r *NoteRoom) getAllCursors() []CursorPosition {
	r.mu.RLock()
	cursors := make([]CursorPosition, 0, len(r.Clients))
	for _, client := range r.Clients {
		if client.LastCursor != nil {
			cursors = append(cursors, *client.LastCursor)
		}
	}
	r.mu.RUnlock()

	return cursors
}

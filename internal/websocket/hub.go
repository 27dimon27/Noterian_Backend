package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type Hub struct {
	mu                sync.RWMutex
	rooms             map[string]*NoteRoom
	register          chan *ClientInfo
	unregister        chan *ClientInfo
	broadcast         chan *BroadcastMessage
	noteUsecase       NoteUsecaseInterface
	profileUsecase    ProfileUsecaseInterface
	attachmentUsecase AttachmentUsecaseInterface
	storage           *BatchStorage
}

func NewHub(noteUsecase NoteUsecaseInterface, profileUsecase ProfileUsecaseInterface, attachmentUsecase AttachmentUsecaseInterface) *Hub {
	hub := &Hub{
		rooms:             make(map[string]*NoteRoom),
		register:          make(chan *ClientInfo, 256),
		unregister:        make(chan *ClientInfo, 256),
		broadcast:         make(chan *BroadcastMessage, 512),
		noteUsecase:       noteUsecase,
		profileUsecase:    profileUsecase,
		attachmentUsecase: attachmentUsecase,
	}

	hub.storage = NewBatchStorage(hub)

	return hub
}

func (h *Hub) Run(ctx context.Context) {
	go h.storage.Run(ctx)

	cleanupTicker := time.NewTicker(5 * time.Minute)
	defer cleanupTicker.Stop()

	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case msg := <-h.broadcast:
			h.handleBroadcast(msg)

		case <-cleanupTicker.C:
			h.cleanupInactiveRooms()

		case <-heartbeatTicker.C:
			h.sendHeartbeats()

		case <-ctx.Done():
			h.shutdown()
			return
		}
	}
}

func (h *Hub) handleRegister(client *ClientInfo) {
	h.mu.Lock()
	room, exists := h.rooms[client.NoteID]
	if !exists {
		room = NewNoteRoom(client.NoteID)
		h.rooms[client.NoteID] = room
	}
	h.mu.Unlock()

	room.mu.RLock()
	isDeleted := room.IsDeleted
	room.mu.RUnlock()

	if isDeleted {
		client.Send <- WebSocketMessage{
			Type:   MsgError,
			UserID: client.UserID,
			NoteID: client.NoteID,
			Msg:    map[string]string{"error": "note has been deleted"},
		}
		return
	}

	room.AddClient(client)

	go h.sendSyncState(client, room)

	h.broadcastToRoom(client.NoteID, WebSocketMessage{
		Type:     MsgUserJoined,
		UserID:   client.UserID,
		UserName: client.UserName,
		Msg:      map[string]string{"message": "user joined"},
	}, client.UserID)
}

func (h *Hub) handleUnregister(client *ClientInfo) {
	h.mu.RLock()
	room, exists := h.rooms[client.NoteID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	room.RemoveClient(client.UserID)

	h.broadcastToRoom(client.NoteID, WebSocketMessage{
		Type:     MsgUserLeft,
		UserID:   client.UserID,
		UserName: client.UserName,
		Msg:      map[string]string{"message": "user left"},
	}, client.UserID)

	close(client.Send)

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
	defer room.mu.RUnlock()

	for _, client := range room.Clients {
		if client.UserID != msg.Exclude {
			select {
			case client.Send <- msg.Message:
			default:
				go func(c *ClientInfo, m WebSocketMessage) {
					select {
					case c.Send <- m:
					case <-time.After(5 * time.Second):
						log.Printf("Client %s send channel blocked", c.UserID)
					}
				}(client, msg.Message)
			}
		}
	}
}

func (h *Hub) broadcastToRoom(noteID string, message WebSocketMessage, excludeUserID string) {
	select {
	case h.broadcast <- &BroadcastMessage{
		NoteID:  noteID,
		Message: message,
		Exclude: excludeUserID,
	}:
	default:
		log.Printf("Broadcast channel full for note %s", noteID)
	}
}

func (h *Hub) sendSyncState(client *ClientInfo, room *NoteRoom) {
	noteID, err := uuid.Parse(client.NoteID)
	if err != nil {
		client.Send <- h.errorMessage("Invalid note ID", client)
		return
	}

	userID, err := uuid.Parse(client.UserID)
	if err != nil {
		client.Send <- h.errorMessage("Invalid user ID", client)
		return
	}

	note, blocks, blockFormattings, err := h.noteUsecase.GetNote(context.Background(), noteID, userID)
	if err != nil {
		client.Send <- h.errorMessage(err.Error(), client)
		return
	}

	for _, block := range blocks {
		if block.BlockTypeID == 1 {
			_, exists := room.GetCRDTDocument(block.ID.String())
			if !exists {
				doc := NewCRDTDocument(client.UserID)
				doc.LoadText(block.Content, client.UserID)
				room.SetCRDTDocument(block.ID.String(), doc)
			}
		}
	}

	if len(blocks) > 0 {
		client.UpdateCursor(CursorPosition{
			BlockID:       blocks[0].ID.String(),
			StartPosition: 0,
			EndPosition:   0,
			UserID:        client.UserID,
			UserName:      client.UserName,
		})
	}

	client.Send <- WebSocketMessage{
		Type:   MsgSyncState,
		UserID: client.UserID,
		NoteID: client.NoteID,
		Msg: map[string]any{
			"note":              note,
			"blocks":            blocks,
			"block_formattings": blockFormattings,
			"cursors":           room.GetAllCursors(),
			"sync_timestamp":    time.Now().Unix(),
		},
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

	case MsgInsertChars:
		var op InsertCharsOperation
		if err := mapToStruct(msg.Msg, &op); err == nil {
			h.handleInsertChars(room, userID, msg, &op)
		}

	case MsgDeleteChars:
		var op DeleteCharsOperation
		if err := mapToStruct(msg.Msg, &op); err == nil {
			h.handleDeleteChars(room, userID, msg, &op)
		}

	case MsgApplyFormatting:
		var op FormattingOperation
		if err := mapToStruct(msg.Msg, &op); err == nil {
			h.handleApplyFormatting(room, userID, &op)
		}

	case MsgCreateBlock:
		var op CreateBlockOperation
		if err := mapToStruct(msg.Msg, &op); err == nil {
			h.handleCreateBlock(room, userID, &op)
		}

	case MsgDeleteBlock:
		var op DeleteBlockOperation
		if err := mapToStruct(msg.Msg, &op); err == nil {
			h.handleDeleteBlock(room, userID, &op)
		}

	case MsgMoveBlock:
		var op MoveBlockOperation
		if err := mapToStruct(msg.Msg, &op); err == nil {
			h.handleMoveBlock(room, userID, &op)
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

	case MsgUploadAttachment:
		var op UploadAttachmentOperation
		if err := mapToStruct(msg.Msg, &op); err == nil {
			h.handleUploadAttachment(room, userID, &op)
		}

	case MsgUploadHeader:
		var op UploadHeaderOperation
		if err := mapToStruct(msg.Msg, &op); err == nil {
			h.handleUploadHeader(room, userID, &op)
		}

	case MsgDeleteHeader:
		var op DeleteHeaderOperation
		if err := mapToStruct(msg.Msg, &op); err == nil {
			h.handleDeleteHeader(room, userID, &op)
		}

	case MsgDeleteNote:
		h.handleDeleteNote(room, userID)

	case MsgHeartbeat:
		h.handleHeartbeat(room, userID)
	}
}

func (h *Hub) handleCursorMove(room *NoteRoom, userID string, cursor CursorPosition) {
	client, exists := room.GetClient(userID)
	if !exists {
		return
	}

	cursor.UserID = client.UserID
	cursor.UserName = client.UserName
	client.UpdateCursor(cursor)

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type: MsgCursorMove,
		Msg:  room.GetAllCursors(),
	}, userID)
}

func (h *Hub) handleInsertChars(room *NoteRoom, userID string, msg WebSocketMessage, op *InsertCharsOperation) {
	if !msg.IsLocal {
		doc, exists := room.GetCRDTDocument(op.BlockID)
		if exists {
			doc.ApplyInserts(op)
		}
		return
	}

	client, exists := room.GetClient(userID)
	if !exists {
		return
	}

	doc, exists := room.GetCRDTDocument(op.BlockID)
	if !exists {
		doc = NewCRDTDocument(userID)
		room.SetCRDTDocument(op.BlockID, doc)
	}

	startCursorPos := client.GetCursor().StartPosition
	endCursorPos := client.GetCursor().EndPosition

	if startCursorPos != endCursorPos {
		deletedChars := []string{}
		for pos := endCursorPos - 1; pos > startCursorPos; pos-- {
			deletedChars = append(deletedChars, doc.DeleteChar(pos))
		}

		tempOp := &DeleteCharsOperation{
			ID:            uuid.New().String(),
			BlockID:       op.BlockID,
			StartPosition: startCursorPos,
			EndPosition:   endCursorPos,
			UniqueIDs:     deletedChars,
			Lamport:       doc.GetLamport(),
			UserID:        userID,
			Timestamp:     time.Now().UnixNano(),
		}

		h.broadcastToRoom(room.NoteID, WebSocketMessage{
			Type:     MsgDeleteChars,
			UserID:   userID,
			UserName: client.UserName,
			Msg:      tempOp,
		}, userID)

		h.updateCursorsAfterOperation(room, MsgDeleteChars, tempOp)

		chars := []rune(op.Char)
		newID := ""
		for i, char := range chars {
			tempID := doc.InsertChar(startCursorPos+i, char, userID)
			if newID == "" {
				newID = tempID
			}
		}

		transformedPos := doc.findPositionByID(newID)

		broadcastOp := &InsertCharsOperation{
			ID:        uuid.New().String(),
			BlockID:   op.BlockID,
			Position:  transformedPos,
			Char:      op.Char,
			Lamport:   doc.GetLamport(),
			UniqueIDs: []string{newID},
			PrevID:    doc.findPrevID(startCursorPos),
			UserID:    userID,
			Timestamp: time.Now().UnixNano(),
		}

		h.storage.SaveBlockContent(room.NoteID, op.BlockID, userID, doc.GetText())

		h.broadcastToRoom(room.NoteID, WebSocketMessage{
			Type:     MsgInsertChars,
			UserID:   userID,
			UserName: client.UserName,
			Msg:      broadcastOp,
		}, userID)

		h.updateCursorsAfterOperation(room, MsgInsertChars, broadcastOp)

		h.broadcastToRoom(room.NoteID, WebSocketMessage{
			Type: MsgCursorMove,
			Msg:  room.GetAllCursors(),
		}, "")
	} else {
		chars := []rune(op.Char)
		newID := ""
		for i, char := range chars {
			tempID := doc.InsertChar(startCursorPos+i, char, userID)
			if newID == "" {
				newID = tempID
			}
		}

		transformedPos := doc.findPositionByID(newID)

		broadcastOp := &InsertCharsOperation{
			ID:        uuid.New().String(),
			BlockID:   op.BlockID,
			Position:  transformedPos,
			Char:      op.Char,
			Lamport:   doc.GetLamport(),
			UniqueIDs: []string{newID},
			PrevID:    doc.findPrevID(startCursorPos),
			UserID:    userID,
			Timestamp: time.Now().UnixNano(),
		}

		h.storage.SaveBlockContent(room.NoteID, op.BlockID, userID, doc.GetText())

		h.broadcastToRoom(room.NoteID, WebSocketMessage{
			Type:     MsgInsertChars,
			UserID:   userID,
			UserName: client.UserName,
			Msg:      broadcastOp,
		}, userID)

		h.updateCursorsAfterOperation(room, MsgInsertChars, broadcastOp)

		h.broadcastToRoom(room.NoteID, WebSocketMessage{
			Type: MsgCursorMove,
			Msg:  room.GetAllCursors(),
		}, "")
	}
}

func (h *Hub) handleDeleteChars(room *NoteRoom, userID string, msg WebSocketMessage, op *DeleteCharsOperation) {
	if !msg.IsLocal {
		doc, exists := room.GetCRDTDocument(op.BlockID)
		if exists {
			doc.ApplyDeletes(op)
		}
		return
	}

	client, exists := room.GetClient(userID)
	if !exists {
		return
	}

	doc, exists := room.GetCRDTDocument(op.BlockID)
	if !exists {
		return
	}

	startCursorPos := client.GetCursor().StartPosition
	endCursorPos := client.GetCursor().EndPosition

	if startCursorPos != endCursorPos {
		deletedChars := []string{}
		for pos := endCursorPos - 1; pos >= startCursorPos; pos-- {
			deletedChars = append(deletedChars, doc.DeleteChar(pos))
		}

		broadcastOp := &DeleteCharsOperation{
			ID:            uuid.New().String(),
			BlockID:       op.BlockID,
			StartPosition: startCursorPos,
			EndPosition:   endCursorPos,
			UniqueIDs:     deletedChars,
			Lamport:       doc.GetLamport(),
			UserID:        userID,
			Timestamp:     time.Now().UnixNano(),
		}

		h.storage.SaveBlockContent(room.NoteID, op.BlockID, userID, doc.GetText())

		h.broadcastToRoom(room.NoteID, WebSocketMessage{
			Type:     MsgDeleteChars,
			UserID:   userID,
			UserName: client.UserName,
			Msg:      broadcastOp,
		}, userID)

		h.updateCursorsAfterOperation(room, MsgDeleteChars, broadcastOp)

		h.broadcastToRoom(room.NoteID, WebSocketMessage{
			Type: MsgCursorMove,
			Msg:  room.GetAllCursors(),
		}, "")
	} else {
		deletePos := startCursorPos - 1
		if deletePos < 0 {
			return
		}

		deletedID := doc.DeleteChar(deletePos)
		if deletedID == "" {
			return
		}

		broadcastOp := &DeleteCharsOperation{
			ID:            uuid.New().String(),
			BlockID:       op.BlockID,
			StartPosition: deletePos,
			EndPosition:   deletePos + 1,
			UniqueIDs:     []string{deletedID},
			Lamport:       doc.GetLamport(),
			UserID:        userID,
			Timestamp:     time.Now().UnixNano(),
		}

		h.storage.SaveBlockContent(room.NoteID, op.BlockID, userID, doc.GetText())

		h.broadcastToRoom(room.NoteID, WebSocketMessage{
			Type:     MsgDeleteChars,
			UserID:   userID,
			UserName: client.UserName,
			Msg:      broadcastOp,
		}, userID)

		h.updateCursorsAfterOperation(room, MsgDeleteChars, broadcastOp)

		h.broadcastToRoom(room.NoteID, WebSocketMessage{
			Type: MsgCursorMove,
			Msg:  room.GetAllCursors(),
		}, "")
	}
}

func (h *Hub) handleApplyFormatting(room *NoteRoom, userID string, op *FormattingOperation) {
	client, exists := room.GetClient(userID)
	if !exists {
		return
	}

	room.mu.Lock()
	op.SequenceID = room.SequenceID
	room.SequenceID++
	room.mu.Unlock()

	go func() {
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
			client.Send <- h.errorMessage(err.Error(), client)
		}
	}()

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type:     MsgApplyFormatting,
		UserID:   userID,
		UserName: client.UserName,
		Msg:      op,
	}, userID)
}

func (h *Hub) handleCreateBlock(room *NoteRoom, userID string, op *CreateBlockOperation) {
	client, exists := room.GetClient(userID)
	if !exists {
		return
	}

	noteID, _ := uuid.Parse(room.NoteID)
	userUUID, _ := uuid.Parse(userID)

	block := models.Block{
		BlockTypeID: op.BlockTypeID,
		Position:    op.Position,
	}

	createdBlock, err := h.noteUsecase.CreateBlock(context.Background(), noteID, userUUID, block)
	if err != nil {
		client.Send <- h.errorMessage(err.Error(), client)
		return
	}

	if createdBlock.BlockTypeID == 1 {
		doc := NewCRDTDocument(userID)
		room.SetCRDTDocument(createdBlock.ID.String(), doc)
	}

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type:     MsgCreateBlock,
		UserID:   userID,
		UserName: client.UserName,
		Msg:      createdBlock,
	}, "")

	creationBlockOp := CreateBlockOperation{
		ID:          uuid.New().String(),
		BlockID:     createdBlock.ID.String(),
		BlockTypeID: createdBlock.BlockTypeID,
		Position:    createdBlock.Position,
		Content:     createdBlock.Content,
		UserID:      userID,
		Timestamp:   time.Now().UnixNano(),
	}

	h.updateCursorsAfterBlockOperation(room, MsgCreateBlock, creationBlockOp, userID, nil)

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type: MsgCursorMove,
		Msg:  room.GetAllCursors(),
	}, "")
}

func (h *Hub) handleDeleteBlock(room *NoteRoom, userID string, op *DeleteBlockOperation) {
	client, exists := room.GetClient(userID)
	if !exists {
		return
	}

	noteID, _ := uuid.Parse(room.NoteID)
	userUUID, _ := uuid.Parse(userID)
	blockUUID, _ := uuid.Parse(op.BlockID)

	_, blocks, _, err := h.noteUsecase.GetNote(context.Background(), noteID, userUUID)
	if err != nil {
		client.Send <- h.errorMessage(err.Error(), client)
		return
	}

	prevBlock := &models.Block{}
	for i, block := range blocks {
		if block.ID == blockUUID {
			if i == 0 {
				break
			}
			prevBlock = &blocks[i]
		}
	}

	err = h.noteUsecase.DeleteBlock(context.Background(), blockUUID, noteID, userUUID)
	if err != nil {
		client.Send <- h.errorMessage(err.Error(), client)
		return
	}

	room.DeleteCRDTDocument(op.BlockID)

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type:     MsgDeleteBlock,
		UserID:   userID,
		UserName: client.UserName,
		Msg:      op.BlockID,
	}, userID)

	deleteBlockOp := DeleteBlockOperation{
		ID:        uuid.New().String(),
		BlockID:   op.BlockID,
		UserID:    userID,
		Timestamp: time.Now().UnixNano(),
	}

	h.updateCursorsAfterBlockOperation(room, MsgDeleteBlock, deleteBlockOp, userID, prevBlock)

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type: MsgCursorMove,
		Msg:  room.GetAllCursors(),
	}, "")
}

func (h *Hub) handleMoveBlock(room *NoteRoom, userID string, op *MoveBlockOperation) {
	client, exists := room.GetClient(userID)
	if !exists {
		return
	}

	noteID, _ := uuid.Parse(room.NoteID)
	userUUID, _ := uuid.Parse(userID)
	blockUUID, _ := uuid.Parse(op.BlockID)

	_, err := h.noteUsecase.MoveBlock(context.Background(), blockUUID, noteID, userUUID, op.NewPosition)
	if err != nil {
		client.Send <- h.errorMessage(err.Error(), client)
		return
	}

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type:     MsgMoveBlock,
		UserID:   userID,
		UserName: client.UserName,
		Msg:      op,
	}, userID)
}

func (h *Hub) handleUpdateNoteTitle(room *NoteRoom, userID string, newTitle string) {
	client, exists := room.GetClient(userID)
	if !exists {
		return
	}

	noteID, _ := uuid.Parse(room.NoteID)
	userUUID, _ := uuid.Parse(userID)

	note, _, _, err := h.noteUsecase.GetNote(context.Background(), noteID, userUUID)
	if err != nil {
		client.Send <- h.errorMessage(err.Error(), client)
		return
	}

	note.Title = newTitle

	_, err = h.noteUsecase.UpdateNote(context.Background(), noteID, userUUID, *note)
	if err != nil {
		client.Send <- h.errorMessage(err.Error(), client)
		return
	}

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type:     MsgUpdateNoteTitle,
		UserID:   userID,
		UserName: client.UserName,
		Msg:      map[string]string{"title": newTitle},
	}, userID)
}

func (h *Hub) handleUpdateNotePublic(room *NoteRoom, userID string, isPublic bool) {
	if !h.isNoteOwner(room.NoteID, userID) {
		return
	}

	noteID, _ := uuid.Parse(room.NoteID)
	userUUID, _ := uuid.Parse(userID)

	note, _, _, err := h.noteUsecase.GetNote(context.Background(), noteID, userUUID)
	if err != nil {
		return
	}

	note.IsPublic = isPublic

	_, err = h.noteUsecase.UpdateNote(context.Background(), noteID, userUUID, *note)
	if err != nil {
		return
	}

	if !isPublic {
		room.mu.Lock()
		clients := make([]*ClientInfo, 0, len(room.Clients))
		for _, c := range room.Clients {
			clients = append(clients, c)
		}
		room.mu.Unlock()

		for _, c := range clients {
			c.Send <- WebSocketMessage{
				Type: MsgNotePrivate,
				Msg:  "Note is now private",
			}
			close(c.Send)
			room.RemoveClient(c.UserID)
		}
	}
}

func (h *Hub) handleUploadAttachment(room *NoteRoom, userID string, op *UploadAttachmentOperation) {
	client, exists := room.GetClient(userID)
	if !exists {
		return
	}

	noteID, err := uuid.Parse(room.NoteID)
	if err != nil {
		client.Send <- h.errorMessage("Invalid note ID", client)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		client.Send <- h.errorMessage("Invalid user ID", client)
		return
	}

	buffer := make([]byte, 512)
	copy(buffer, op.FileData)

	mimeType := http.DetectContentType(buffer)

	maxSize, err := getMaxSizeByMimeType(mimeType)
	if err != nil {
		client.Send <- h.errorMessage("Invalid MIME-type of file", client)
		return
	}

	fileSize := int64(len(op.FileData))

	if fileSize > maxSize {
		client.Send <- h.errorMessage("File too large", client)
		return
	}

	attachment, err := h.attachmentUsecase.UploadAttachment(
		context.Background(),
		noteID,
		userUUID,
		op.FileName,
		fileSize,
		mimeType,
		bytes.NewReader(op.FileData),
		op.HasPosition,
		op.Position,
	)
	if err != nil {
		log.Printf("Failed to upload attachment: %v", err)
		client.Send <- h.errorMessage(err.Error(), client)
		return
	}

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type:     MsgUploadAttachment,
		UserID:   userID,
		UserName: client.UserName,
		Msg: map[string]any{
			"id":          attachment.ID,
			"block_id":    attachment.BlockID,
			"minio_key":   attachment.MinioKey,
			"attach_url":  attachment.AttachURL,
			"url_expires": attachment.URLExpiresAt.Unix(),
			"created_at":  attachment.CreatedAt.Unix(),
			"updated_at":  attachment.UpdatedAt.Unix(),
			"note_id":     room.NoteID,
			"position":    op.Position,
			"mime_type":   mimeType,
		},
	}, "")
}

func (h *Hub) handleUploadHeader(room *NoteRoom, userID string, op *UploadHeaderOperation) {
	client, exists := room.GetClient(userID)
	if !exists {
		return
	}

	noteID, err := uuid.Parse(room.NoteID)
	if err != nil {
		client.Send <- h.errorMessage("Invalid note ID", client)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		client.Send <- h.errorMessage("Invalid user ID", client)
		return
	}

	buffer := make([]byte, 512)
	copy(buffer, op.FileData)

	mimeType := http.DetectContentType(buffer)

	if !AllowedMimeTypesForImage[mimeType] {
		client.Send <- h.errorMessage("Invalid MIME-type of file", client)
		return
	}

	fileSize := int64(len(op.FileData))

	if fileSize > MAX_IMAGE_SIZE {
		client.Send <- h.errorMessage("File too large", client)
		return
	}

	header, err := h.attachmentUsecase.UploadHeader(
		context.Background(),
		noteID,
		userUUID,
		op.FileName,
		fileSize,
		mimeType,
		bytes.NewReader(op.FileData),
	)
	if err != nil {
		log.Printf("Failed to upload header: %v", err)
		client.Send <- h.errorMessage(err.Error(), client)
		return
	}

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type:     MsgUploadHeader,
		UserID:   userID,
		UserName: client.UserName,
		Msg: map[string]any{
			"id":          header.ID,
			"minio_key":   header.MinioKey,
			"header_url":  header.HeaderURL,
			"url_expires": header.URLExpiresAt.Unix(),
			"created_at":  header.CreatedAt.Unix(),
			"updated_at":  header.UpdatedAt.Unix(),
			"note_id":     room.NoteID,
			"mime_type":   mimeType,
		},
	}, "")
}

func (h *Hub) handleDeleteHeader(room *NoteRoom, userID string, op *DeleteHeaderOperation) {
	client, exists := room.GetClient(userID)
	if !exists {
		return
	}

	noteID, err := uuid.Parse(room.NoteID)
	if err != nil {
		client.Send <- h.errorMessage("Invalid note ID", client)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		client.Send <- h.errorMessage("Invalid user ID", client)
		return
	}

	err = h.attachmentUsecase.DeleteHeader(
		context.Background(),
		noteID,
		userUUID,
	)
	if err != nil {
		log.Printf("Failed to delete header: %v", err)
		client.Send <- h.errorMessage(err.Error(), client)
		return
	}

	h.broadcastToRoom(room.NoteID, WebSocketMessage{
		Type:     MsgDeleteHeader,
		UserID:   userID,
		UserName: client.UserName,
		Msg: map[string]any{
			"id":        op.ID,
			"fileName":  op.FileName,
			"note_id":   room.NoteID,
			"user_id":   userID,
			"timestamp": op.Timestamp,
		},
	}, "")
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
	for _, c := range room.Clients {
		clients = append(clients, c)
	}

	for _, c := range clients {
		c.Send <- WebSocketMessage{
			Type: MsgNoteDeleted,
			Msg:  "Note has been deleted",
		}
		close(c.Send)
	}
	room.Clients = make(map[string]*ClientInfo)
	room.IsDeleted = true
	room.mu.Unlock()

	h.mu.Lock()
	delete(h.rooms, room.NoteID)
	h.mu.Unlock()
}

func (h *Hub) handleHeartbeat(room *NoteRoom, userID string) {
	client, exists := room.GetClient(userID)
	if exists {
		client.LastPing = time.Now().Unix()
	}
}

func (h *Hub) updateCursorsAfterOperation(room *NoteRoom, opType MessageType, op any) {
	room.mu.RLock()
	defer room.mu.RUnlock()

	for i, client := range room.Clients {
		cursor := client.GetCursor()
		newStartPos, newEndPos := TransformCursorPosition(cursor, opType, op, cursor.BlockID, client.UserID)

		if newStartPos != cursor.StartPosition || newEndPos != cursor.EndPosition {
			cursor.StartPosition = newStartPos
			cursor.EndPosition = newEndPos
			room.Clients[i].UpdateCursor(cursor)
		}
	}
}

func (h *Hub) updateCursorsAfterBlockOperation(room *NoteRoom, opType MessageType, op any, userID string, block *models.Block) {
	room.mu.RLock()
	defer room.mu.RUnlock()

	author, ok := room.GetClient(userID)
	if !ok {
		return
	}

	for i, client := range room.Clients {
		cursor := client.GetCursor()
		newCursor := TransformCursorPositionAfterBlockOperation(cursor, opType, op, author.LastCursor, block)
		room.Clients[i].UpdateCursor(newCursor)
	}
}

func (h *Hub) isNoteOwner(noteID string, userID string) bool {
	noteUUID, _ := uuid.Parse(noteID)
	userUUID, _ := uuid.Parse(userID)

	note, _, _, err := h.noteUsecase.GetNote(context.Background(), noteUUID, userUUID)
	if err != nil || note == nil {
		return false
	}

	return note.UserID == userUUID
}

func (h *Hub) cleanupInactiveRooms() {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now().Unix()
	for noteID, room := range h.rooms {
		room.mu.RLock()
		isEmpty := len(room.Clients) == 0
		isOld := (now - room.CreatedAt) > 600
		room.mu.RUnlock()

		if isEmpty && isOld {
			delete(h.rooms, noteID)
			log.Printf("Cleaned up inactive room %s", noteID)
		}
	}
}

func (h *Hub) sendHeartbeats() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, room := range h.rooms {
		room.mu.RLock()
		for _, client := range room.Clients {
			select {
			case client.Send <- WebSocketMessage{
				Type: MsgHeartbeat,
				Msg:  map[string]int64{"timestamp": time.Now().Unix()},
			}:
			default:
			}
		}
		room.mu.RUnlock()
	}
}

func (h *Hub) shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, room := range h.rooms {
		room.mu.Lock()
		for _, client := range room.Clients {
			close(client.Send)
		}
		room.mu.Unlock()
	}
}

func (h *Hub) errorMessage(errMsg string, client *ClientInfo) WebSocketMessage {
	return WebSocketMessage{
		Type:   MsgError,
		UserID: client.UserID,
		NoteID: client.NoteID,
		Msg:    map[string]string{"error": errMsg},
	}
}

func mapToStruct(data any, target any) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, target)
}

func getMaxSizeByMimeType(mimeType string) (int64, error) {
	if AllowedMimeTypesForImage[mimeType] {
		return MAX_IMAGE_SIZE, nil
	}
	if AllowedMimeTypesForGIF[mimeType] {
		return MAX_GIF_SIZE, nil
	}
	if AllowedMimeTypesForAudio[mimeType] {
		return MAX_AUDIO_SIZE, nil
	}
	if AllowedMimeTypesForVideo[mimeType] {
		return MAX_VIDEO_SIZE, nil
	}
	return 0, ErrInvalidMimeType
}

package websocket

import (
	"context"
	"sync"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"

	"github.com/google/uuid"
)

type MessageType string

const (
	// messages from client
	MsgJoin             MessageType = "join"
	MsgLeave            MessageType = "leave"
	MsgCursorMove       MessageType = "cursor_move"
	MsgInsertChar       MessageType = "insert_char"
	MsgDeleteChar       MessageType = "delete_char"
	MsgApplyFormatting  MessageType = "apply_formatting"
	MsgCreateBlock      MessageType = "create_block"
	MsgDeleteBlock      MessageType = "delete_block"
	MsgMoveBlock        MessageType = "move_block"
	MsgUpdateNoteTitle  MessageType = "update_note_title"
	MsgUpdateNotePublic MessageType = "update_note_public"
	MsgDeleteNote       MessageType = "delete_note"

	// messages from server
	MsgUserJoined    MessageType = "user_joined"
	MsgUserLeft      MessageType = "user_left"
	MsgCursorsUpdate MessageType = "cursors_update"
	MsgSyncState     MessageType = "sync_state"
	MsgOperation     MessageType = "operation"
	MsgError         MessageType = "error"
	MsgNoteDeleted   MessageType = "note_deleted"
	MsgNotePrivate   MessageType = "note_private"
)

type WebSocketMessage struct {
	Type      MessageType `json:"type"`
	UserID    string      `json:"userId,omitempty"`
	UserName  string      `json:"userName,omitempty"`
	NoteID    string      `json:"noteId,omitempty"`
	BlockID   string      `json:"blockId,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
	// Version   int64       `json:"version,omitempty"` // где юзается?
}

type CursorPosition struct {
	BlockID  string `json:"blockId"`
	Position int    `json:"position"`
	UserID   string `json:"userId"`
	UserName string `json:"userName"`
}

type InsertCharOperation struct {
	BlockID  string `json:"blockId"`
	Position int    `json:"position"`
	Char     string `json:"char"`
	UserID   string `json:"userId"`
	Lamport  int64  `json:"lamport"`  // зачем нужно?
	UniqueID string `json:"uniqueId"` // зачем нужно?
}

type DeleteCharOperation struct {
	BlockID  string `json:"blockId"`
	Position int    `json:"position"`
	UserID   string `json:"userId"`
	Lamport  int64  `json:"lamport"` // зачем нужно?
}

type FormattingOperation struct {
	BlockID    string `json:"blockId"`
	StartPos   int    `json:"startPos"`
	EndPos     int    `json:"endPos"`
	Bold       *bool  `json:"bold,omitempty"`
	Italic     *bool  `json:"italic,omitempty"`
	Underline  *bool  `json:"underline,omitempty"`
	TextAlign  *int   `json:"textAlign,omitempty"`
	UserID     string `json:"userId"`
	SequenceID int64  `json:"sequenceId"` // зачем нужно?
}

type OperationType string // зачем нужна эта структура? +всё с ней связанное

const (
	OpCreateBlock     OperationType = "create_block"
	OpDeleteBlock     OperationType = "delete_block"
	OpMoveBlock       OperationType = "move_block"
	OpUpdateTitle     OperationType = "update_title"
	OpUpdatePublic    OperationType = "update_public"
	OpApplyFormatting OperationType = "apply_formatting"
)

type Operation struct { // зачем нужна эта структура? +всё с ней связанное
	ID         string                 `json:"id"`
	Type       OperationType          `json:"type"`
	Data       map[string]interface{} `json:"data"`
	SequenceID int64                  `json:"sequenceId"`
	UserID     string                 `json:"userId"`
	Timestamp  int64                  `json:"timestamp"`
}

type ClientInfo struct {
	UserID     string
	UserName   string
	NoteID     string
	LastCursor *CursorPosition
	Send       chan WebSocketMessage
	// убрал conn, если сломается - вернуть
}

// type SyncState struct {
// 	Note          map[string]interface{}   `json:"note"`
// 	Blocks        []map[string]interface{} `json:"blocks"`
// 	ActiveUsers   []map[string]interface{} `json:"activeUsers"`
// 	Cursors       []CursorPosition         `json:"cursors"`
// 	OwnerID       string                   `json:"ownerId"`
// 	IsPublic      bool                     `json:"isPublic"`
// 	NoteTitle     string                   `json:"noteTitle"`
// 	SyncTimestamp int64                    `json:"syncTimestamp"`
// 	SequenceID    int64                    `json:"sequenceId"`
// }

// type BlockSnapshot struct {
// 	ID          string      `json:"id"`
// 	NoteID      string      `json:"noteId"`
// 	BlockTypeID int         `json:"blockTypeId"`
// 	Position    int         `json:"position"`
// 	Content     string      `json:"content"`
// 	CreatedAt   string      `json:"createdAt"`
// 	UpdatedAt   string      `json:"updatedAt"`
// 	Formatting  interface{} `json:"formatting,omitempty"`
// }

type CRDTDocument struct {
	mu           sync.RWMutex
	characters   []*CRDTChar
	lamportClock int64
}

type CRDTChar struct {
	ID      string // уникальный ID символа (userID:lamport)
	Char    string
	UserID  string
	Lamport int64
	Visible bool
}

type Hub struct {
	mu             sync.RWMutex
	rooms          map[string]*NoteRoom // noteID -> room
	register       chan *ClientInfo
	unregister     chan *ClientInfo
	broadcast      chan *BroadcastMessage
	noteUsecase    NoteUsecaseInterface
	profileUsecase ProfileUsecaseInterface
}

type NoteRoom struct {
	mu             sync.RWMutex
	NoteID         string
	Clients        map[string]*ClientInfo   // userID -> client
	CRDTDocuments  map[string]*CRDTDocument // blockID -> CRDT document
	OperationQueue []Operation              // очередь операций для OT (где используется? зачем нужна?)
	SequenceID     int64                    // текущий sequence ID для OT (зачем нужен?)
	IsDeleted      bool
}

type BroadcastMessage struct {
	NoteID  string
	Message WebSocketMessage
	Exclude string
}

type NoteUsecaseInterface interface {
	GetNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, error)
	UpdateNote(ctx context.Context, noteID uuid.UUID, note models.Note, userID uuid.UUID) (*models.Note, error)
	DeleteNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) error
	GetBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) (*models.Block, error)
	GetBlocks(ctx context.Context, noteID uuid.UUID) ([]models.Block, error)
	CreateBlock(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, block models.Block) (*models.Block, error)
	DeleteBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) error
	MoveBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, newPosition int) (*models.Block, error)
	UpdateBlockContent(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, content string) (*models.Block, error)
	UpdateBlockFormatting(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, formattingRange models.FormattingRange) (*models.BlockFormatting, error)
}

type ProfileUsecaseInterface interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
}

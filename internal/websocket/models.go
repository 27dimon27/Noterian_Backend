package websocket

import (
	"context"
	"sync"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"

	"github.com/google/uuid"
)

type MessageType string

const (
	MsgUserJoined MessageType = "user_joined"
	MsgUserLeft   MessageType = "user_left"
	MsgError      MessageType = "error"
	MsgSyncState  MessageType = "sync_state"

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

	MsgNotePrivate MessageType = "note_private"
	MsgNoteDeleted MessageType = "note_deleted"
)

type WebSocketMessage struct {
	Type     MessageType `json:"type"`
	IsLocal  bool        `json:"is_local"`
	UserID   string      `json:"userId,omitempty"`
	UserName string      `json:"userName,omitempty"`
	NoteID   string      `json:"noteId,omitempty"`
	Msg      any         `json:"msg"`
}

type InfoMessage struct {
	Info any
}

type ErrMessage struct {
	Error any
}

type CursorPosition struct {
	BlockID  string `json:"blockId"`
	Position int    `json:"position"`
}

type InsertCharOperation struct {
	BlockID  string `json:"blockId"`
	Position int    `json:"position"`
	Char     string `json:"char"`
	Lamport  int64  `json:"lamport"`
	UniqueID string `json:"uniqueId"`
}

type DeleteCharOperation struct {
	BlockID  string `json:"blockId"`
	Position int    `json:"position"`
	Lamport  int64  `json:"lamport"`
}

type FormattingOperation struct {
	BlockID    string `json:"blockId"`
	StartPos   int    `json:"startPos"`
	EndPos     int    `json:"endPos"`
	Bold       *bool  `json:"bold,omitempty"`
	Italic     *bool  `json:"italic,omitempty"`
	Underline  *bool  `json:"underline,omitempty"`
	TextAlign  *int   `json:"textAlign,omitempty"`
	SequenceID int64  `json:"sequenceId"`
}

type ClientInfo struct {
	UserID     string
	UserName   string
	NoteID     string
	LastCursor CursorPosition
	Send       chan WebSocketMessage
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

type CRDTDocument struct {
	mu           sync.RWMutex
	characters   []CRDTChar
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
	register       chan ClientInfo
	unregister     chan ClientInfo
	broadcast      chan BroadcastMessage
	noteUsecase    NoteUsecaseInterface
	profileUsecase ProfileUsecaseInterface
}

type NoteRoom struct {
	mu            sync.RWMutex
	NoteID        string
	Clients       map[string]ClientInfo    // userID -> client
	CRDTDocuments map[string]*CRDTDocument // blockID -> CRDT document
	SequenceID    int64                    // текущий sequence ID для OT (зачем нужен?)
	IsDeleted     bool
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

package websocket

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type MessageType string

const (
	MsgUserJoined MessageType = "user_joined"
	MsgUserLeft   MessageType = "user_left"
	MsgError      MessageType = "error"
	MsgSyncState  MessageType = "sync_state"
	MsgHeartbeat  MessageType = "heartbeat"

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
	Type      MessageType `json:"type"`
	IsLocal   bool        `json:"is_local,omitempty"`
	UserID    string      `json:"userId,omitempty"`
	UserName  string      `json:"userName,omitempty"`
	NoteID    string      `json:"noteId,omitempty"`
	BlockID   string      `json:"blockId,omitempty"`
	Msg       any         `json:"msg"`
	Timestamp int64       `json:"timestamp"`
}

type CursorPosition struct {
	BlockID   string `json:"blockId"`
	Position  int    `json:"position"`
	UserID    string `json:"userId"`
	UserName  string `json:"userName"`
	Timestamp int64  `json:"timestamp"`
}

type UserCursor struct {
	UserID   string         `json:"userId"`
	UserName string         `json:"userName"`
	Cursor   CursorPosition `json:"cursor"`
}

type InsertCharOperation struct {
	ID        string `json:"id"`
	BlockID   string `json:"blockId"`
	Position  int    `json:"position"`
	Char      string `json:"char"`
	Lamport   int64  `json:"lamport"`
	UniqueID  string `json:"uniqueId"`
	PrevID    string `json:"prevId"`
	UserID    string `json:"userId"`
	Timestamp int64  `json:"timestamp"`
}

type DeleteCharOperation struct {
	ID        string `json:"id"`
	BlockID   string `json:"blockId"`
	Position  int    `json:"position"`
	UniqueID  string `json:"uniqueId"`
	Lamport   int64  `json:"lamport"`
	UserID    string `json:"userId"`
	Timestamp int64  `json:"timestamp"`
}

type FormattingOperation struct {
	ID         string `json:"id"`
	BlockID    string `json:"blockId"`
	StartPos   int    `json:"startPos"`
	EndPos     int    `json:"endPos"`
	Bold       *bool  `json:"bold,omitempty"`
	Italic     *bool  `json:"italic,omitempty"`
	Underline  *bool  `json:"underline,omitempty"`
	TextAlign  *int   `json:"textAlign,omitempty"`
	SequenceID int64  `json:"sequenceId"`
	Lamport    int64  `json:"lamport"`
	UserID     string `json:"userId"`
	Timestamp  int64  `json:"timestamp"`
}

type CreateBlockOperation struct {
	ID          string `json:"id"`
	BlockID     string `json:"blockId"`
	BlockTypeID int    `json:"block_type_id"`
	Position    int    `json:"position"`
	Content     string `json:"content"`
	UserID      string `json:"userId"`
	Timestamp   int64  `json:"timestamp"`
}

type DeleteBlockOperation struct {
	ID        string `json:"id"`
	BlockID   string `json:"blockId"`
	UserID    string `json:"userId"`
	Timestamp int64  `json:"timestamp"`
}

type MoveBlockOperation struct {
	ID          string `json:"id"`
	BlockID     string `json:"blockId"`
	NewPosition int    `json:"newPosition"`
	UserID      string `json:"userId"`
	Timestamp   int64  `json:"timestamp"`
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

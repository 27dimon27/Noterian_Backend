package websocket

import "errors"

var (
	ErrNoteDeleted      = errors.New("note has been deleted")
	ErrBlockedChannel   = errors.New("send channel blocked")
	ErrFullChannel      = errors.New("broadcast channel full")
	ErrRoomDeleted      = errors.New("room is deleted")
	ErrUnknownOperation = errors.New("unknown operation")
	ErrClientNotFound   = errors.New("client not found")
	ErrDocumentNotFound = errors.New("document not found")
	ErrForbidden        = errors.New("access denied")
	ErrInvalidMimeType  = errors.New("invalid MIME-type of file")
	ErrFileTooLarge     = errors.New("file too large")

	ErrInternalServer = errors.New("internal server error")
)

var (
	PublicMsgUserJoined = "user joined"
	PublicMsgUserLeft   = "user left"
)

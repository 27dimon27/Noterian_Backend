package support

import "errors"

var (
	ErrTicketNotFound      = errors.New("support ticket not found")
	ErrUnauthorizedTicket  = errors.New("unauthorized to access this ticket")
	ErrInvalidTicketID     = errors.New("invalid ticket ID")
	ErrInvalidUserID       = errors.New("invalid user ID")
	ErrInvalidCategoryID   = errors.New("invalid category ID")
	ErrInvalidStatusID     = errors.New("invalid status ID")
	ErrCategoryNotFound    = errors.New("support category not found")
	ErrStatusNotFound      = errors.New("support status not found")
	ErrMessageNotFound     = errors.New("support message not found")
	ErrUnauthorizedMessage = errors.New("unauthorized to access this message")
	ErrInvalidMessageID    = errors.New("invalid message ID")
	ErrRatingNotFound      = errors.New("support rating not found")
	ErrRatingAlreadyExists = errors.New("rating already exists for this ticket")
	ErrInvalidRating       = errors.New("invalid rating value")
	ErrUnauthorized        = errors.New("unauthorized access to support feature")
	ErrAdminOnly           = errors.New("admin only operation")
	ErrUnauthorizedAccess  = errors.New("unauthorized access to this resource")
)

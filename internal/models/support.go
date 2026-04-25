package models

import (
	"time"

	"github.com/google/uuid"
)

type SupportCategory struct {
	ID          int
	Name        string
	Description string
	CreatedAt   time.Time
}

type SupportStatus struct {
	ID          int
	Name        string
	Description string
	CreatedAt   time.Time
}

type UserRole struct {
	ID          int
	Name        string
	Description string
	CreatedAt   time.Time
}

type SupportTicket struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	CategoryID  int
	StatusID    int
	AssignedTo  *uuid.UUID
	Title       string
	Description string
	Priority    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ResolvedAt  *time.Time
}

type SupportMessage struct {
	ID         uuid.UUID
	TicketID   uuid.UUID
	UserID     uuid.UUID
	Message    string
	IsInternal bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type SupportRating struct {
	ID        uuid.UUID
	TicketID  uuid.UUID
	UserID    uuid.UUID
	Rating    int
	Comment   *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

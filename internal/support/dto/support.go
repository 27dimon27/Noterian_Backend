package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type CreateTicketRequest struct {
	Title       string `json:"title" validate:"required,min=5,max=255"`
	Description string `json:"description" validate:"required,min=10"`
	CategoryID  int    `json:"category_id" validate:"required,min=1"`
}

type UpdateTicketRequest struct {
	Title       string `json:"title" validate:"omitempty,min=5,max=255"`
	Description string `json:"description" validate:"omitempty,min=10"`
	StatusID    int    `json:"status_id" validate:"omitempty,min=1"`
	Priority    int    `json:"priority" validate:"omitempty,min=1,max=5"`
}

type CreateMessageRequest struct {
	Message string `json:"message" validate:"required,min=1"`
}

type CreateRatingRequest struct {
	Rating  int     `json:"rating" validate:"required,min=1,max=5"`
	Comment *string `json:"comment" validate:"omitempty,max=1000"`
}

type AssignTicketRequest struct {
	AssignedToID uuid.UUID `json:"assigned_to_id" validate:"required"`
}

type ChangeStatusRequest struct {
	StatusID int `json:"status_id" validate:"required,min=1"`
}

type TicketResponse struct {
	ID           uuid.UUID       `json:"id"`
	UserID       uuid.UUID       `json:"user_id"`
	CategoryID   int             `json:"category_id"`
	CategoryName string          `json:"category_name"`
	StatusID     int             `json:"status_id"`
	StatusName   string          `json:"status_name"`
	AssignedTo   *uuid.UUID      `json:"assigned_to,omitempty"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	Priority     int             `json:"priority"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	ResolvedAt   *time.Time      `json:"resolved_at,omitempty"`
	MessageCount int             `json:"message_count"`
	Rating       *RatingResponse `json:"rating,omitempty"`
}

type MessageResponse struct {
	ID         uuid.UUID `json:"id"`
	TicketID   uuid.UUID `json:"ticket_id"`
	UserID     uuid.UUID `json:"user_id"`
	Message    string    `json:"message"`
	IsInternal bool      `json:"is_internal"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type RatingResponse struct {
	ID        uuid.UUID `json:"id"`
	TicketID  uuid.UUID `json:"ticket_id"`
	Rating    int       `json:"rating"`
	Comment   *string   `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CategoryResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type StatusResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TicketsListResponse struct {
	Tickets []TicketResponse `json:"tickets"`
	Total   int              `json:"total"`
}

type TicketStatsResponse struct {
	Total           int      `json:"total"`
	Open            int      `json:"open"`
	InProgress      int      `json:"in_progress"`
	WaitingUser     int      `json:"waiting_user"`
	Closed          int      `json:"closed"`
	BugCount        int      `json:"bug_count"`
	SuggestionCount int      `json:"suggestion_count"`
	ComplaintCount  int      `json:"complaint_count"`
	AverageRating   *float64 `json:"average_rating,omitempty"`
}

func ToTicketResponse(ticket *models.SupportTicket, categoryName, statusName string, messageCount int, rating *models.SupportRating) *TicketResponse {
	var ratingResp *RatingResponse
	if rating != nil {
		ratingResp = &RatingResponse{
			ID:        rating.ID,
			TicketID:  rating.TicketID,
			Rating:    rating.Rating,
			Comment:   rating.Comment,
			CreatedAt: rating.CreatedAt,
		}
	}

	return &TicketResponse{
		ID:           ticket.ID,
		UserID:       ticket.UserID,
		CategoryID:   ticket.CategoryID,
		CategoryName: categoryName,
		StatusID:     ticket.StatusID,
		StatusName:   statusName,
		AssignedTo:   ticket.AssignedTo,
		Title:        ticket.Title,
		Description:  ticket.Description,
		Priority:     ticket.Priority,
		CreatedAt:    ticket.CreatedAt,
		UpdatedAt:    ticket.UpdatedAt,
		ResolvedAt:   ticket.ResolvedAt,
		MessageCount: messageCount,
		Rating:       ratingResp,
	}
}

func ToMessageResponse(msg *models.SupportMessage) *MessageResponse {
	return &MessageResponse{
		ID:         msg.ID,
		TicketID:   msg.TicketID,
		UserID:     msg.UserID,
		Message:    msg.Message,
		IsInternal: msg.IsInternal,
		CreatedAt:  msg.CreatedAt,
		UpdatedAt:  msg.UpdatedAt,
	}
}

func ToRatingResponse(rating *models.SupportRating) *RatingResponse {
	return &RatingResponse{
		ID:        rating.ID,
		TicketID:  rating.TicketID,
		Rating:    rating.Rating,
		Comment:   rating.Comment,
		CreatedAt: rating.CreatedAt,
	}
}

func ToCategoryResponse(cat *models.SupportCategory) *CategoryResponse {
	return &CategoryResponse{
		ID:          cat.ID,
		Name:        cat.Name,
		Description: cat.Description,
	}
}

func ToStatusResponse(status *models.SupportStatus) *StatusResponse {
	return &StatusResponse{
		ID:          status.ID,
		Name:        status.Name,
		Description: status.Description,
	}
}

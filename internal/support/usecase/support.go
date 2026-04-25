package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/support"
	"github.com/google/uuid"
)

type SupportRepository interface {
	CreateTicket(ctx context.Context, ticket *models.SupportTicket) (*models.SupportTicket, error)
	GetTicketByID(ctx context.Context, ticketID uuid.UUID) (*models.SupportTicket, error)
	GetUserTickets(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.SupportTicket, int, error)
	GetAllTickets(ctx context.Context, limit, offset int) ([]models.SupportTicket, int, error)
	UpdateTicket(ctx context.Context, ticketID uuid.UUID, updates map[string]interface{}) (*models.SupportTicket, error)
	CloseTicket(ctx context.Context, ticketID uuid.UUID) (*models.SupportTicket, error)

	CreateMessage(ctx context.Context, message *models.SupportMessage) (*models.SupportMessage, error)
	GetTicketMessages(ctx context.Context, ticketID uuid.UUID) ([]models.SupportMessage, error)
	GetMessageCount(ctx context.Context, ticketID uuid.UUID) (int, error)
	DeleteMessage(ctx context.Context, messageID uuid.UUID) error

	GetCategories(ctx context.Context) ([]models.SupportCategory, error)
	GetCategoryName(ctx context.Context, categoryID int) (string, error)

	GetStatuses(ctx context.Context) ([]models.SupportStatus, error)
	GetStatusName(ctx context.Context, statusID int) (string, error)

	CreateRating(ctx context.Context, rating *models.SupportRating) (*models.SupportRating, error)
	GetRatingByTicketID(ctx context.Context, ticketID uuid.UUID) (*models.SupportRating, error)
	GetAverageRating(ctx context.Context) (*float64, error)

	GetStats(ctx context.Context) (*map[string]interface{}, error)
	GetUserStats(ctx context.Context, userID uuid.UUID) (*map[string]interface{}, error)
	GetRoleByUserID(ctx context.Context, userID uuid.UUID) (string, error)
}

type SupportUsecase struct {
	repo SupportRepository
}

func NewSupportUsecase(repo SupportRepository) *SupportUsecase {
	return &SupportUsecase{
		repo: repo,
	}
}

func (u *SupportUsecase) CreateTicket(ctx context.Context, userID uuid.UUID, title, description string, categoryID int) (*models.SupportTicket, error) {
	ticket := &models.SupportTicket{
		UserID:      userID,
		Title:       title,
		Description: description,
		CategoryID:  categoryID,
		StatusID:    1,
		Priority:    2,
	}

	return u.repo.CreateTicket(ctx, ticket)
}

func (u *SupportUsecase) GetTicketByID(ctx context.Context, ticketID uuid.UUID, userID uuid.UUID) (*models.SupportTicket, error) {
	return u.repo.GetTicketByID(ctx, ticketID)
}

func (u *SupportUsecase) GetUserTickets(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.SupportTicket, int, error) {
	return u.repo.GetUserTickets(ctx, userID, limit, offset)
}

func (u *SupportUsecase) GetAllTickets(ctx context.Context, limit, offset int) ([]models.SupportTicket, int, error) {
	return u.repo.GetAllTickets(ctx, limit, offset)
}

func (u *SupportUsecase) UpdateTicket(ctx context.Context, ticketID uuid.UUID, userID uuid.UUID, updates map[string]interface{}) (*models.SupportTicket, error) {
	ticket, err := u.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if ticket.UserID != userID {
		return nil, support.ErrUnauthorizedAccess
	}

	return u.repo.UpdateTicket(ctx, ticketID, updates)
}

func (u *SupportUsecase) CloseTicket(ctx context.Context, ticketID uuid.UUID, userID uuid.UUID) (*models.SupportTicket, error) {
	ticket, err := u.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if ticket.UserID != userID {
		return nil, support.ErrUnauthorizedAccess
	}

	return u.repo.CloseTicket(ctx, ticketID)
}

func (u *SupportUsecase) CreateMessage(ctx context.Context, ticketID uuid.UUID, userID uuid.UUID, message string, isInternal bool) (*models.SupportMessage, error) {
	ticket, err := u.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if !isInternal && ticket.UserID != userID {
		return nil, support.ErrUnauthorizedAccess
	}

	msg := &models.SupportMessage{
		TicketID:   ticketID,
		UserID:     userID,
		Message:    message,
		IsInternal: isInternal,
	}

	return u.repo.CreateMessage(ctx, msg)
}

func (u *SupportUsecase) GetTicketMessages(ctx context.Context, ticketID uuid.UUID, userID uuid.UUID) ([]models.SupportMessage, error) {
	ticket, err := u.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if ticket.UserID != userID {
		return nil, support.ErrUnauthorizedAccess
	}

	messages, err := u.repo.GetTicketMessages(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	var filteredMessages []models.SupportMessage
	for _, msg := range messages {
		if !msg.IsInternal || msg.UserID == userID {
			filteredMessages = append(filteredMessages, msg)
		}
	}

	return filteredMessages, nil
}

func (u *SupportUsecase) DeleteMessage(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error { // не прописано
	return u.repo.DeleteMessage(ctx, messageID)
}

func (u *SupportUsecase) GetCategories(ctx context.Context) ([]models.SupportCategory, error) {
	return u.repo.GetCategories(ctx)
}

func (u *SupportUsecase) GetStatuses(ctx context.Context) ([]models.SupportStatus, error) {
	return u.repo.GetStatuses(ctx)
}

func (u *SupportUsecase) CreateRating(ctx context.Context, ticketID uuid.UUID, userID uuid.UUID, rating int, comment *string) (*models.SupportRating, error) {
	ticket, err := u.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if ticket.UserID != userID {
		return nil, support.ErrUnauthorizedAccess
	}

	existingRating, err := u.repo.GetRatingByTicketID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if existingRating != nil {
		return nil, support.ErrRatingAlreadyExists
	}

	ratingModel := &models.SupportRating{
		TicketID: ticketID,
		UserID:   userID,
		Rating:   rating,
		Comment:  comment,
	}

	return u.repo.CreateRating(ctx, ratingModel)
}

func (u *SupportUsecase) GetRatingByTicketID(ctx context.Context, ticketID uuid.UUID) (*models.SupportRating, error) {
	return u.repo.GetRatingByTicketID(ctx, ticketID)
}

func (u *SupportUsecase) GetStats(ctx context.Context, userID uuid.UUID) (*map[string]interface{}, error) {
	role_name, err := u.repo.GetRoleByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if role_name == "user" {
		return u.repo.GetUserStats(ctx, userID)
	}

	return u.repo.GetStats(ctx)
}

func (u *SupportUsecase) GetUserRole(ctx context.Context, userID uuid.UUID) (string, error) {
	return u.repo.GetRoleByUserID(ctx, userID)
}

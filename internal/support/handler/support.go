package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/support"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/support/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/body"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type SupportUsecase interface {
	CreateTicket(ctx context.Context, userID uuid.UUID, title, description string, categoryID int) (*models.SupportTicket, error)
	GetTicketByID(ctx context.Context, ticketID uuid.UUID, userID uuid.UUID) (*models.SupportTicket, error)
	GetUserTickets(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.SupportTicket, int, error)
	GetAllTickets(ctx context.Context, limit, offset int) ([]models.SupportTicket, int, error)
	UpdateTicket(ctx context.Context, ticketID uuid.UUID, userID uuid.UUID, updates map[string]interface{}) (*models.SupportTicket, error)
	CloseTicket(ctx context.Context, ticketID uuid.UUID, userID uuid.UUID) (*models.SupportTicket, error)

	CreateMessage(ctx context.Context, ticketID uuid.UUID, userID uuid.UUID, message string, isInternal bool) (*models.SupportMessage, error)
	GetTicketMessages(ctx context.Context, ticketID uuid.UUID, userID uuid.UUID) ([]models.SupportMessage, error)
	DeleteMessage(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error

	GetCategories(ctx context.Context) ([]models.SupportCategory, error)
	GetStatuses(ctx context.Context) ([]models.SupportStatus, error)

	CreateRating(ctx context.Context, ticketID uuid.UUID, userID uuid.UUID, rating int, comment *string) (*models.SupportRating, error)
	GetRatingByTicketID(ctx context.Context, ticketID uuid.UUID) (*models.SupportRating, error)

	GetStats(ctx context.Context, userID uuid.UUID) (*map[string]interface{}, error)
	// GetUserStats(ctx context.Context, userID uuid.UUID) (*map[string]interface{}, error)
}

type SupportHandler struct {
	usecase   SupportUsecase
	validator *validator.Validate
}

func NewSupportHandler(supportUsecase SupportUsecase) *SupportHandler {
	return &SupportHandler{
		usecase:   supportUsecase,
		validator: validator.New(),
	}
}

func (h *SupportHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, support.ErrInvalidUserID)
		return
	}

	var req dto.CreateTicketRequest
	if err := body.GetBody(r, &req); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	ticket, err := h.usecase.CreateTicket(r.Context(), userID, req.Title, req.Description, req.CategoryID)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToTicketResponse(ticket, "", "", 0, nil)
	write.JSONResponse(w, http.StatusCreated, response)
}

func (h *SupportHandler) GetUserTickets(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, support.ErrInvalidUserID)
		return
	}

	limit := 10
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsedOffset, err := strconv.Atoi(o); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	tickets, total, err := h.usecase.GetUserTickets(r.Context(), userID, limit, offset)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	var responses []dto.TicketResponse
	for _, ticket := range tickets {
		response := dto.ToTicketResponse(&ticket, "", "", 0, nil)
		responses = append(responses, *response)
	}

	result := dto.TicketsListResponse{
		Tickets: responses,
		Total:   total,
	}

	write.JSONResponse(w, http.StatusOK, result)
}

func (h *SupportHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, support.ErrInvalidUserID)
		return
	}

	ticketIDStr := r.PathValue("id")
	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, support.ErrInvalidTicketID)
		return
	}

	ticket, err := h.usecase.GetTicketByID(r.Context(), ticketID, userID)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	if ticket.UserID != userID {
		write.JSONErrorResponse(w, http.StatusForbidden, support.ErrUnauthorizedTicket)
		return
	}

	response := dto.ToTicketResponse(ticket, "", "", 0, nil)
	write.JSONResponse(w, http.StatusOK, response)
}

func (h *SupportHandler) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, support.ErrInvalidUserID)
		return
	}

	ticketIDStr := r.PathValue("id")
	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, support.ErrInvalidTicketID)
		return
	}

	var req dto.UpdateTicketRequest
	if err := body.GetBody(r, &req); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.StatusID != 0 {
		updates["status_id"] = req.StatusID
	}
	if req.Priority != 0 {
		updates["priority"] = req.Priority
	}

	ticket, err := h.usecase.UpdateTicket(r.Context(), ticketID, userID, updates)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToTicketResponse(ticket, "", "", 0, nil)
	write.JSONResponse(w, http.StatusOK, response)
}

func (h *SupportHandler) CloseTicket(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, support.ErrInvalidUserID)
		return
	}

	ticketIDStr := r.PathValue("id")
	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, support.ErrInvalidTicketID)
		return
	}

	ticket, err := h.usecase.CloseTicket(r.Context(), ticketID, userID)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToTicketResponse(ticket, "", "", 0, nil)
	write.JSONResponse(w, http.StatusOK, response)
}

func (h *SupportHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, support.ErrInvalidUserID)
		return
	}

	ticketIDStr := r.PathValue("ticket_id")
	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, support.ErrInvalidTicketID)
		return
	}

	var req dto.CreateMessageRequest
	if err := body.GetBody(r, &req); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	message, err := h.usecase.CreateMessage(r.Context(), ticketID, userID, req.Message, req.IsInternal)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToMessageResponse(message)
	write.JSONResponse(w, http.StatusCreated, response)
}

func (h *SupportHandler) GetTicketMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, support.ErrInvalidUserID)
		return
	}

	ticketIDStr := r.PathValue("ticket_id")
	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, support.ErrInvalidTicketID)
		return
	}

	messages, err := h.usecase.GetTicketMessages(r.Context(), ticketID, userID)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	var responses []dto.MessageResponse
	for _, msg := range messages {
		responses = append(responses, *dto.ToMessageResponse(&msg))
	}

	write.JSONResponse(w, http.StatusOK, responses)
}

func (h *SupportHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.usecase.GetCategories(r.Context())
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	var responses []dto.CategoryResponse
	for _, cat := range categories {
		responses = append(responses, *dto.ToCategoryResponse(&cat))
	}

	write.JSONResponse(w, http.StatusOK, responses)
}

func (h *SupportHandler) GetStatuses(w http.ResponseWriter, r *http.Request) {
	statuses, err := h.usecase.GetStatuses(r.Context())
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	var responses []dto.StatusResponse
	for _, status := range statuses {
		responses = append(responses, *dto.ToStatusResponse(&status))
	}

	write.JSONResponse(w, http.StatusOK, responses)
}

func (h *SupportHandler) CreateRating(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, support.ErrInvalidUserID)
		return
	}

	ticketIDStr := r.PathValue("ticket_id")
	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, support.ErrInvalidTicketID)
		return
	}

	var req dto.CreateRatingRequest
	if err := body.GetBody(r, &req); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	rating, err := h.usecase.CreateRating(r.Context(), ticketID, userID, req.Rating, req.Comment)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToRatingResponse(rating)
	write.JSONResponse(w, http.StatusCreated, response)
}

func (h *SupportHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, support.ErrInvalidUserID)
		return
	}

	stats, err := h.usecase.GetStats(r.Context(), userID)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.TicketStatsResponse{
		Total:           (*stats)["total"].(int),
		Open:            (*stats)["open"].(int),
		InProgress:      (*stats)["in_progress"].(int),
		WaitingUser:     (*stats)["waiting_user"].(int),
		Closed:          (*stats)["closed"].(int),
		BugCount:        (*stats)["bug_count"].(int),
		SuggestionCount: (*stats)["suggestion_count"].(int),
		ComplaintCount:  (*stats)["complaint_count"].(int),
		AverageRating:   (*stats)["average_rating"].(*float64),
	}

	write.JSONResponse(w, http.StatusOK, response)
}

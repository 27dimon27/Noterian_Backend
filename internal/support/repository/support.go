package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/support"
	"github.com/google/uuid"
)

type SupportRepository struct {
	db *sql.DB
}

func NewSupportRepository(db *sql.DB) *SupportRepository {
	return &SupportRepository{db: db}
}

func (r *SupportRepository) CreateTicket(ctx context.Context, ticket *models.SupportTicket) (*models.SupportTicket, error) {
	query := `
		INSERT INTO support_tickets (id, user_id, category_id, status_id, title, description, priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, user_id, category_id, status_id, assigned_to, title, description, priority, created_at, updated_at, resolved_at
	`

	ticket.ID = uuid.New()
	var assignedTo sql.NullString

	err := r.db.QueryRowContext(ctx, query,
		ticket.ID, ticket.UserID, ticket.CategoryID, ticket.StatusID, ticket.Title, ticket.Description, ticket.Priority,
	).Scan(
		&ticket.ID, &ticket.UserID, &ticket.CategoryID, &ticket.StatusID, &assignedTo,
		&ticket.Title, &ticket.Description, &ticket.Priority, &ticket.CreatedAt, &ticket.UpdatedAt, &ticket.ResolvedAt,
	)
	if err != nil {
		return nil, err
	}

	if assignedTo.Valid {
		assignedUUID, err := uuid.Parse(assignedTo.String)
		if err == nil {
			ticket.AssignedTo = &assignedUUID
		}
	}

	return ticket, nil
}

func (r *SupportRepository) GetTicketByID(ctx context.Context, ticketID uuid.UUID) (*models.SupportTicket, error) {
	query := `
		SELECT id, user_id, category_id, status_id, assigned_to, title, description, priority, created_at, updated_at, resolved_at
		FROM support_tickets
		WHERE id = $1
	`

	ticket := &models.SupportTicket{}
	var assignedTo sql.NullString

	err := r.db.QueryRowContext(ctx, query, ticketID).Scan(
		&ticket.ID, &ticket.UserID, &ticket.CategoryID, &ticket.StatusID, &assignedTo,
		&ticket.Title, &ticket.Description, &ticket.Priority, &ticket.CreatedAt, &ticket.UpdatedAt, &ticket.ResolvedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, support.ErrTicketNotFound
		}
		return nil, err
	}

	if assignedTo.Valid {
		assignedUUID, err := uuid.Parse(assignedTo.String)
		if err == nil {
			ticket.AssignedTo = &assignedUUID
		}
	}

	return ticket, nil
}

func (r *SupportRepository) GetUserTickets(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.SupportTicket, int, error) {
	countQuery := `SELECT COUNT(*) FROM support_tickets WHERE user_id = $1`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, category_id, status_id, assigned_to, title, description, priority, created_at, updated_at, resolved_at
		FROM support_tickets
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tickets []models.SupportTicket
	for rows.Next() {
		ticket := models.SupportTicket{}
		var assignedTo sql.NullString

		err := rows.Scan(
			&ticket.ID, &ticket.UserID, &ticket.CategoryID, &ticket.StatusID, &assignedTo,
			&ticket.Title, &ticket.Description, &ticket.Priority, &ticket.CreatedAt, &ticket.UpdatedAt, &ticket.ResolvedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if assignedTo.Valid {
			assignedUUID, err := uuid.Parse(assignedTo.String)
			if err == nil {
				ticket.AssignedTo = &assignedUUID
			}
		}

		tickets = append(tickets, ticket)
	}

	return tickets, total, rows.Err()
}

func (r *SupportRepository) GetAllTickets(ctx context.Context, limit, offset int) ([]models.SupportTicket, int, error) {
	countQuery := `SELECT COUNT(*) FROM support_tickets`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, category_id, status_id, assigned_to, title, description, priority, created_at, updated_at, resolved_at
		FROM support_tickets
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tickets []models.SupportTicket
	for rows.Next() {
		ticket := models.SupportTicket{}
		var assignedTo sql.NullString

		err := rows.Scan(
			&ticket.ID, &ticket.UserID, &ticket.CategoryID, &ticket.StatusID, &assignedTo,
			&ticket.Title, &ticket.Description, &ticket.Priority, &ticket.CreatedAt, &ticket.UpdatedAt, &ticket.ResolvedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if assignedTo.Valid {
			assignedUUID, err := uuid.Parse(assignedTo.String)
			if err == nil {
				ticket.AssignedTo = &assignedUUID
			}
		}

		tickets = append(tickets, ticket)
	}

	return tickets, total, rows.Err()
}

func (r *SupportRepository) UpdateTicket(ctx context.Context, ticketID uuid.UUID, updates map[string]interface{}) (*models.SupportTicket, error) {
	query := `UPDATE support_tickets SET `
	args := []interface{}{}
	argCount := 1

	for key, value := range updates {
		if argCount > 1 {
			query += ", "
		}
		query += key + " = $" + string(rune(argCount))
		args = append(args, value)
		argCount++
	}

	query += ", updated_at = NOW() WHERE id = $" + string(rune(argCount)) + " RETURNING id, user_id, category_id, status_id, assigned_to, title, description, priority, created_at, updated_at, resolved_at"
	args = append(args, ticketID)

	ticket := &models.SupportTicket{}
	var assignedTo sql.NullString

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&ticket.ID, &ticket.UserID, &ticket.CategoryID, &ticket.StatusID, &assignedTo,
		&ticket.Title, &ticket.Description, &ticket.Priority, &ticket.CreatedAt, &ticket.UpdatedAt, &ticket.ResolvedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, support.ErrTicketNotFound
		}
		return nil, err
	}

	if assignedTo.Valid {
		assignedUUID, err := uuid.Parse(assignedTo.String)
		if err == nil {
			ticket.AssignedTo = &assignedUUID
		}
	}

	return ticket, nil
}

func (r *SupportRepository) CloseTicket(ctx context.Context, ticketID uuid.UUID) (*models.SupportTicket, error) {
	query := `
		UPDATE support_tickets 
		SET status_id = (SELECT id FROM support_statuses WHERE name = 'closed'), 
		    resolved_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, category_id, status_id, assigned_to, title, description, priority, created_at, updated_at, resolved_at
	`

	ticket := &models.SupportTicket{}
	var assignedTo sql.NullString

	err := r.db.QueryRowContext(ctx, query, ticketID).Scan(
		&ticket.ID, &ticket.UserID, &ticket.CategoryID, &ticket.StatusID, &assignedTo,
		&ticket.Title, &ticket.Description, &ticket.Priority, &ticket.CreatedAt, &ticket.UpdatedAt, &ticket.ResolvedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, support.ErrTicketNotFound
		}
		return nil, err
	}

	if assignedTo.Valid {
		assignedUUID, err := uuid.Parse(assignedTo.String)
		if err == nil {
			ticket.AssignedTo = &assignedUUID
		}
	}

	return ticket, nil
}

func (r *SupportRepository) CreateMessage(ctx context.Context, message *models.SupportMessage) (*models.SupportMessage, error) {
	query := `
		INSERT INTO support_messages (id, ticket_id, user_id, message, is_internal, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, ticket_id, user_id, message, is_internal, created_at, updated_at
	`

	message.ID = uuid.New()

	err := r.db.QueryRowContext(ctx, query,
		message.ID, message.TicketID, message.UserID, message.Message, message.IsInternal,
	).Scan(
		&message.ID, &message.TicketID, &message.UserID, &message.Message, &message.IsInternal, &message.CreatedAt, &message.UpdatedAt,
	)

	return message, err
}

func (r *SupportRepository) GetTicketMessages(ctx context.Context, ticketID uuid.UUID) ([]models.SupportMessage, error) {
	query := `
		SELECT id, ticket_id, user_id, message, is_internal, created_at, updated_at
		FROM support_messages
		WHERE ticket_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.SupportMessage
	for rows.Next() {
		msg := models.SupportMessage{}
		err := rows.Scan(
			&msg.ID, &msg.TicketID, &msg.UserID, &msg.Message, &msg.IsInternal, &msg.CreatedAt, &msg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

func (r *SupportRepository) GetMessageCount(ctx context.Context, ticketID uuid.UUID) (int, error) { // не прописано
	query := `SELECT COUNT(*) FROM support_messages WHERE ticket_id = $1`
	var count int
	err := r.db.QueryRowContext(ctx, query, ticketID).Scan(&count)
	return count, err
}

func (r *SupportRepository) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	query := `DELETE FROM support_messages WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, messageID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return support.ErrMessageNotFound
	}

	return nil
}

func (r *SupportRepository) GetCategories(ctx context.Context) ([]models.SupportCategory, error) {
	query := `SELECT id, name, description, created_at FROM support_categories ORDER BY id ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.SupportCategory
	for rows.Next() {
		cat := models.SupportCategory{}
		err := rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	return categories, rows.Err()
}

func (r *SupportRepository) GetCategoryName(ctx context.Context, categoryID int) (string, error) { // не прописано
	query := `SELECT name FROM support_categories WHERE id = $1`
	var name string
	err := r.db.QueryRowContext(ctx, query, categoryID).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", support.ErrCategoryNotFound
		}
		return "", err
	}
	return name, nil
}

func (r *SupportRepository) GetStatuses(ctx context.Context) ([]models.SupportStatus, error) {
	query := `SELECT id, name, description, created_at FROM support_statuses ORDER BY id ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []models.SupportStatus
	for rows.Next() {
		status := models.SupportStatus{}
		err := rows.Scan(&status.ID, &status.Name, &status.Description, &status.CreatedAt)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, status)
	}

	return statuses, rows.Err()
}

func (r *SupportRepository) GetStatusName(ctx context.Context, statusID int) (string, error) {
	query := `SELECT name FROM support_statuses WHERE id = $1`
	var name string
	err := r.db.QueryRowContext(ctx, query, statusID).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", support.ErrStatusNotFound
		}
		return "", err
	}
	return name, nil
}

func (r *SupportRepository) CreateRating(ctx context.Context, rating *models.SupportRating) (*models.SupportRating, error) {
	query := `
		INSERT INTO support_ratings (id, ticket_id, user_id, rating, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, ticket_id, user_id, rating, comment, created_at, updated_at
	`

	rating.ID = uuid.New()

	err := r.db.QueryRowContext(ctx, query,
		rating.ID, rating.TicketID, rating.UserID, rating.Rating, rating.Comment,
	).Scan(
		&rating.ID, &rating.TicketID, &rating.UserID, &rating.Rating, &rating.Comment, &rating.CreatedAt, &rating.UpdatedAt,
	)

	return rating, err
}

func (r *SupportRepository) GetRatingByTicketID(ctx context.Context, ticketID uuid.UUID) (*models.SupportRating, error) {
	query := `
		SELECT id, ticket_id, user_id, rating, comment, created_at, updated_at
		FROM support_ratings
		WHERE ticket_id = $1
	`

	rating := &models.SupportRating{}
	err := r.db.QueryRowContext(ctx, query, ticketID).Scan(
		&rating.ID, &rating.TicketID, &rating.UserID, &rating.Rating, &rating.Comment, &rating.CreatedAt, &rating.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return rating, nil
}

func (r *SupportRepository) GetAverageRating(ctx context.Context) (*float64, error) {
	query := `SELECT AVG(rating)::float FROM support_ratings`
	var avg sql.NullFloat64
	err := r.db.QueryRowContext(ctx, query).Scan(&avg)
	if err != nil {
		return nil, err
	}
	if avg.Valid {
		return &avg.Float64, nil
	}
	return nil, nil
}

func (r *SupportRepository) GetStats(ctx context.Context) (*map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(CASE WHEN status_id = (SELECT id FROM support_statuses WHERE name = 'open') THEN 1 END) as open,
			COUNT(CASE WHEN status_id = (SELECT id FROM support_statuses WHERE name = 'in_progress') THEN 1 END) as in_progress,
			COUNT(CASE WHEN status_id = (SELECT id FROM support_statuses WHERE name = 'waiting_user') THEN 1 END) as waiting_user,
			COUNT(CASE WHEN status_id = (SELECT id FROM support_statuses WHERE name = 'closed') THEN 1 END) as closed,
			COUNT(CASE WHEN category_id = (SELECT id FROM support_categories WHERE name = 'bug') THEN 1 END) as bug_count,
			COUNT(CASE WHEN category_id = (SELECT id FROM support_categories WHERE name = 'suggestion') THEN 1 END) as suggestion_count,
			COUNT(CASE WHEN category_id = (SELECT id FROM support_categories WHERE name = 'complaint') THEN 1 END) as complaint_count
		FROM support_tickets
	`

	var total, open, inProgress, waitingUser, closed, bugCount, suggestionCount, complaintCount int

	err := r.db.QueryRowContext(ctx, query).Scan(
		&total, &open, &inProgress, &waitingUser, &closed, &bugCount, &suggestionCount, &complaintCount,
	)
	if err != nil {
		return nil, err
	}

	avgRating, _ := r.GetAverageRating(ctx)

	stats := map[string]interface{}{
		"total":            total,
		"open":             open,
		"in_progress":      inProgress,
		"waiting_user":     waitingUser,
		"closed":           closed,
		"bug_count":        bugCount,
		"suggestion_count": suggestionCount,
		"complaint_count":  complaintCount,
		"average_rating":   avgRating,
	}

	return &stats, nil
}

func (r *SupportRepository) GetUserStats(ctx context.Context, userID uuid.UUID) (*map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(CASE WHEN status_id = (SELECT id FROM support_statuses WHERE name = 'open') THEN 1 END) as open,
			COUNT(CASE WHEN status_id = (SELECT id FROM support_statuses WHERE name = 'in_progress') THEN 1 END) as in_progress,
			COUNT(CASE WHEN status_id = (SELECT id FROM support_statuses WHERE name = 'waiting_user') THEN 1 END) as waiting_user,
			COUNT(CASE WHEN status_id = (SELECT id FROM support_statuses WHERE name = 'closed') THEN 1 END) as closed,
			COUNT(CASE WHEN category_id = (SELECT id FROM support_categories WHERE name = 'bug') THEN 1 END) as bug_count,
			COUNT(CASE WHEN category_id = (SELECT id FROM support_categories WHERE name = 'suggestion') THEN 1 END) as suggestion_count,
			COUNT(CASE WHEN category_id = (SELECT id FROM support_categories WHERE name = 'complaint') THEN 1 END) as complaint_count
		FROM support_tickets
		WHERE user_id = $1
	`

	var total, open, inProgress, waitingUser, closed, bugCount, suggestionCount, complaintCount int

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&total, &open, &inProgress, &waitingUser, &closed, &bugCount, &suggestionCount, &complaintCount,
	)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total":            total,
		"open":             open,
		"in_progress":      inProgress,
		"waiting_user":     waitingUser,
		"closed":           closed,
		"bug_count":        bugCount,
		"suggestion_count": suggestionCount,
		"complaint_count":  complaintCount,
	}

	return &stats, nil
}

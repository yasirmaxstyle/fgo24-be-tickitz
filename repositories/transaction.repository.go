package repositories

import (
	"context"
	"fmt"
	"time"

	"noir-backend/models"
	"noir-backend/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, userID int, showtimeID int, seatNumbers []string, paymentMethodID int, recipientEmail, recipientFullName, recipientPhone string) (*models.Transaction, []models.Ticket, error)
	ProcessPayment(ctx context.Context, transactionCode string) (*models.Transaction, []models.Ticket, error)
	CancelTransaction(ctx context.Context, transactionCode string) (*models.Transaction, []models.Ticket, error)
	GetTransactions(ctx context.Context) ([]models.TransactionJoinRow, error)
	GetTransactionByCode(ctx context.Context, transactionCode string) (*models.Transaction, []models.Ticket, error)
}

type transactionRepository struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) CreateTransaction(ctx context.Context, userID int, showtimeID int, seatNumbers []string, paymentMethodID int, recipientEmail, recipientFullName, recipientPhone string) (*models.Transaction, []models.Ticket, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var showtime models.Showtime
	err = tx.QueryRow(ctx, `
		SELECT showtime_id, movie_id, screen_id, show_datetime, base_price, created_at 
		FROM showtimes 
		WHERE showtime_id = $1`,
		showtimeID).Scan(
		&showtime.ShowtimeID, &showtime.MovieID, &showtime.ScreenID,
		&showtime.ShowDatetime, &showtime.BasePrice, &showtime.CreatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("showtime not found: %w", err)
	}

	var totalSeats int
	err = tx.QueryRow(ctx, `SELECT total_seats FROM screens WHERE screen_id = $1`, showtime.ScreenID).Scan(&totalSeats)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get screen info: %w", err)
	}

	var bookedCount int
	err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM tickets WHERE showtime_id = $1 AND status != 'cancelled'`, showtimeID).Scan(&bookedCount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count booked seats: %w", err)
	}

	availableSeats := totalSeats - bookedCount
	if len(seatNumbers) > availableSeats {
		return nil, nil, fmt.Errorf("not enough available seats")
	}

	bookedSeats := make([]string, 0)
	rows, err := tx.Query(ctx, `
		SELECT seat_number FROM tickets 
		WHERE showtime_id = $1 AND seat_number = ANY($2) AND status != 'cancelled'`,
		showtimeID, seatNumbers)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check seat availability: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var seat string
		if err := rows.Scan(&seat); err != nil {
			return nil, nil, fmt.Errorf("failed to scan booked seat: %w", err)
		}
		bookedSeats = append(bookedSeats, seat)
	}

	if len(bookedSeats) > 0 {
		return nil, nil, fmt.Errorf("seats already booked: %v", bookedSeats)
	}

	var paymentMethod models.PaymentMethod
	err = tx.QueryRow(ctx, `
		SELECT payment_method_id, name, code, is_active
		FROM payment_method
		WHERE payment_method_id = $1 AND is_active = true`,
		paymentMethodID).Scan(
		&paymentMethod.PaymentMethodID, &paymentMethod.Name,
		&paymentMethod.Code, &paymentMethod.IsActive)
	if err != nil {
		return nil, nil, fmt.Errorf("payment method not found or inactive: %w", err)
	}

	transactionCode := utils.GenerateTransactionCode()
	totalAmount := showtime.BasePrice * float64(len(seatNumbers))
	expiresAt := time.Now().Add(5 * time.Minute)

	rows, err = tx.Query(ctx, `
		INSERT INTO transactions (
			transaction_code, recipient_email, recipient_full_name, 
			recipient_phone_number, total_seats, total_amount, status, 
			created_at, expires_at, created_by, payment_method_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING transaction_id, transaction_code, recipient_email, recipient_full_name, 
		        recipient_phone_number, total_seats, total_amount, status, 
		        created_at, expires_at, created_by, payment_method_id`,
		transactionCode, recipientEmail, recipientFullName,
		recipientPhone, len(seatNumbers), totalAmount, "pending",
		time.Now(), expiresAt, userID, paymentMethodID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create transaction: %w", err)
	}
	transaction, err := pgx.CollectOneRow[models.Transaction](rows, pgx.RowToStructByName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to collect transaction row: %w", err)
	}

	tickets := make([]models.Ticket, 0, len(seatNumbers))
	for _, seatNumber := range seatNumbers {
		var ticket models.Ticket
		ticketCode := utils.GenerateTicketCode()
		err = tx.QueryRow(ctx, `
			INSERT INTO tickets (ticket_code, showtime_id, seat_number, price, status, transaction_id, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING ticket_id, ticket_code, showtime_id, seat_number, price, status, transaction_id, created_at`,
			ticketCode, showtimeID, seatNumber, showtime.BasePrice, "booked", transaction.TransactionID, time.Now()).Scan(
			&ticket.TicketID, &ticket.TicketCode, &ticket.ShowtimeID, &ticket.SeatNumber,
			&ticket.Price, &ticket.Status, &ticket.TransactionID, &ticket.CreatedAt)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create ticket for seat %s: %w", seatNumber, err)
		}
		tickets = append(tickets, ticket)
	}

	// available_seats logic removed as it's computed dynamically now

	if err = tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &transaction, tickets, nil
}

func (r *transactionRepository) ProcessPayment(ctx context.Context, transactionCode string) (*models.Transaction, []models.Ticket, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT transaction_id, transaction_code, recipient_email, recipient_full_name, 
		       recipient_phone_number, total_seats, total_amount, status, 
		       created_at, expires_at, paid_at, created_by, payment_method_id
		FROM transactions 
		WHERE transaction_code = $1`,
		transactionCode)
	if err != nil {
		return nil, nil, fmt.Errorf("transaction not found: %w", err)
	}

	transaction, err := pgx.CollectOneRow[models.Transaction](rows, pgx.RowToStructByName)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get transaction data: %w", err)
	}

	if transaction.Status != "pending" {
		return nil, nil, fmt.Errorf("transaction is not pending")
	}

	if time.Now().After(transaction.ExpiresAt) {
		return nil, nil, fmt.Errorf("transaction has expired")
	}

	now := time.Now()
	_, err = tx.Exec(ctx, `
		UPDATE transactions 
		SET status = 'paid', paid_at = $1 
		WHERE transaction_id = $2`,
		now, transaction.TransactionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update transaction status: %w", err)
	}

	transaction.Status = "paid"
	transaction.PaidAt = &now

	tickets := make([]models.Ticket, 0)
	rows, err = tx.Query(ctx, `
		SELECT ticket_id, ticket_code, showtime_id, seat_number, price, status, transaction_id, created_at
		FROM tickets 
		WHERE transaction_id = $1`,
		transaction.TransactionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get tickets: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ticket models.Ticket
		if err := rows.Scan(&ticket.TicketID, &ticket.TicketCode, &ticket.ShowtimeID,
			&ticket.SeatNumber, &ticket.Price, &ticket.Status, &ticket.TransactionID, &ticket.CreatedAt); err != nil {
			return nil, nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &transaction, tickets, nil
}

func (r *transactionRepository) CancelTransaction(ctx context.Context, transactionCode string) (*models.Transaction, []models.Ticket, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT transaction_id, transaction_code, recipient_email, recipient_full_name, 
		       recipient_phone_number, total_seats, total_amount, status, 
		       created_at, expires_at, paid_at, created_by, payment_method_id
		FROM transactions 
		WHERE transaction_code = $1`,
		transactionCode)
	if err != nil {
		return nil, nil, fmt.Errorf("transaction not found: %w", err)
	}

	transaction, err := pgx.CollectOneRow[models.Transaction](rows, pgx.RowToStructByName)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get transaction data: %w", err)
	}

	if transaction.Status == "paid" {
		return nil, nil, fmt.Errorf("cannot cancel paid transaction")
	}

	if transaction.Status == "cancelled" {
		return nil, nil, fmt.Errorf("transaction already cancelled")
	}

	var showtimeID int
	err = tx.QueryRow(ctx, `
		SELECT showtime_id FROM tickets 
		WHERE transaction_id = $1 LIMIT 1`,
		transaction.TransactionID).Scan(&showtimeID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get showtime: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE transactions 
		SET status = 'cancelled' 
		WHERE transaction_id = $1`,
		transaction.TransactionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to cancel transaction: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE tickets 
		SET status = 'cancelled' 
		WHERE transaction_id = $1`,
		transaction.TransactionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to cancel tickets: %w", err)
	}

	transaction.Status = "cancelled"

	tickets := make([]models.Ticket, 0)
	rows, err = tx.Query(ctx, `
		SELECT ticket_id, ticket_code, showtime_id, seat_number, price, status, transaction_id, created_at
		FROM tickets 
		WHERE transaction_id = $1`,
		transaction.TransactionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get tickets: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ticket models.Ticket
		if err := rows.Scan(&ticket.TicketID, &ticket.TicketCode, &ticket.ShowtimeID,
			&ticket.SeatNumber, &ticket.Price, &ticket.Status, &ticket.TransactionID, &ticket.CreatedAt); err != nil {
			return nil, nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &transaction, tickets, nil
}

func (r *transactionRepository) GetTransactions(ctx context.Context) ([]models.TransactionJoinRow, error) {
	rows, err := r.db.Query(ctx, `
		SELECT 
  			t.transaction_id, t.transaction_code, t.status, t.total_amount, t.expires_at, t.created_at,
  			tk.seat_number,
  			s.showtime_id, s.show_datetime, tk.price AS ticket_price,
  			m.movie_id, m.title AS movie_title,
  			c.cinema_id, c.name AS cinema_name, c.location AS cinema_location
		FROM transactions t
		LEFT JOIN tickets tk ON t.transaction_id = tk.transaction_id
		LEFT JOIN showtimes s ON tk.showtime_id = s.showtime_id
		LEFT JOIN movies m ON s.movie_id = m.movie_id
		LEFT JOIN screens sc ON s.screen_id = sc.screen_id
		LEFT JOIN cinemas c ON sc.cinema_id = c.cinema_id
		ORDER BY t.transaction_id DESC`,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[models.TransactionJoinRow])
}

func (r *transactionRepository) GetTransactionByCode(ctx context.Context, transactionCode string) (*models.Transaction, []models.Ticket, error) {
	var transaction models.Transaction
	err := r.db.QueryRow(ctx, `
		SELECT transaction_id, transaction_code, recipient_email, recipient_full_name, 
		       recipient_phone_number, total_seats, total_amount, status, 
		       created_at, expires_at, paid_at, created_by, payment_method_id
		FROM transactions 
		WHERE transaction_code = $1`,
		transactionCode).Scan(
		&transaction.TransactionID, &transaction.TransactionCode, &transaction.RecipientEmail,
		&transaction.RecipientFullName, &transaction.RecipientPhone, &transaction.TotalSeats,
		&transaction.TotalAmount, &transaction.Status, &transaction.CreatedAt,
		&transaction.ExpiresAt, &transaction.PaidAt, &transaction.CreatedBy, &transaction.PaymentMethodID)
	if err != nil {
		return nil, nil, fmt.Errorf("transaction not found: %w", err)
	}

	tickets := make([]models.Ticket, 0)
	rows, err := r.db.Query(ctx, `
		SELECT ticket_id, ticket_code, showtime_id, seat_number, price, status, transaction_id, created_at
		FROM tickets 
		WHERE transaction_id = $1`,
		transaction.TransactionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get tickets: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ticket models.Ticket
		if err := rows.Scan(&ticket.TicketID, &ticket.TicketCode, &ticket.ShowtimeID,
			&ticket.SeatNumber, &ticket.Price, &ticket.Status, &ticket.TransactionID, &ticket.CreatedAt); err != nil {
			return nil, nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}

	return &transaction, tickets, nil
}

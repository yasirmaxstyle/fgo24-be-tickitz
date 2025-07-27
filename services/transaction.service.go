package services

import (
	"context"
	"fmt"
	"noir-backend/dto"
	"noir-backend/models"
	"time"

	"noir-backend/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionService struct {
	db *pgxpool.Pool
}

func NewTransactionService(db *pgxpool.Pool) *TransactionService {
	return &TransactionService{db: db}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, req dto.CreateTransactionRequest, userID int) (*dto.TransactionResult, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var showtime models.Showtime
	err = tx.QueryRow(ctx, `
		SELECT showtime_id, movie_id, cinema_id, show_datetime, price, available_seats, created_at 
		FROM showtimes 
		WHERE showtime_id = $1`,
		req.ShowtimeID).Scan(
		&showtime.ShowtimeID, &showtime.MovieID, &showtime.CinemaID,
		&showtime.ShowDatetime, &showtime.Price, &showtime.AvailableSeats, &showtime.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("showtime not found: %w", err)
	}

	if len(req.SeatNumbers) > showtime.AvailableSeats {
		return nil, fmt.Errorf("not enough available seats")
	}

	bookedSeats := make([]string, 0)
	rows, err := tx.Query(ctx, `
		SELECT seat_number FROM tickets 
		WHERE showtime_id = $1 AND seat_number = ANY($2) AND status != 'cancelled'`,
		req.ShowtimeID, req.SeatNumbers)
	if err != nil {
		return nil, fmt.Errorf("failed to check seat availability: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var seat string
		if err := rows.Scan(&seat); err != nil {
			return nil, fmt.Errorf("failed to scan booked seat: %w", err)
		}
		bookedSeats = append(bookedSeats, seat)
	}

	if len(bookedSeats) > 0 {
		return nil, fmt.Errorf("seats already booked: %v", bookedSeats)
	}

	var paymentMethod models.PaymentMethod
	err = tx.QueryRow(ctx, `
		SELECT payment_method_id, name, code, is_active
		FROM payment_method
		WHERE payment_method_id = $1 AND is_active = true`,
		req.PaymentMethodID).Scan(
		&paymentMethod.PaymentMethodID, &paymentMethod.Name,
		&paymentMethod.Code, &paymentMethod.IsActive)
	if err != nil {
		return nil, fmt.Errorf("payment method not found or inactive: %w", err)
	}

	transactionCode := utils.GenerateTransactionCode()
	totalAmount := showtime.Price * float64(len(req.SeatNumbers))
	expiresAt := time.Now().Add(5 * time.Minute) // 5 minutes to complete payment

	rows, err = tx.Query(ctx, `
		INSERT INTO transactions (
			transaction_code, recipient_email, recipient_full_name, 
			recipient_phone_number, total_seats, total_amount, status, 
			created_at, expires_at, created_by, payment_method_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING transaction_id, transaction_code, recipient_email, recipient_full_name, 
		        recipient_phone_number, total_seats, total_amount, status, 
		        created_at, expires_at, created_by, payment_method_id`,
		transactionCode, req.RecipientEmail, req.RecipientFullName,
		req.RecipientPhone, len(req.SeatNumbers), totalAmount, "pending",
		time.Now(), expiresAt, userID, req.PaymentMethodID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}
	transaction, err := pgx.CollectOneRow[models.Transaction](rows, pgx.RowToStructByName)
	if err != nil {
		return nil, fmt.Errorf("")
	}

	tickets := make([]models.Ticket, 0, len(req.SeatNumbers))
	for _, seatNumber := range req.SeatNumbers {
		var ticket models.Ticket
		ticketCode := utils.GenerateTicketCode()
		err = tx.QueryRow(ctx, `
			INSERT INTO tickets (ticket_code, showtime_id, seat_number, status, transaction_id, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING ticket_id, ticket_code, showtime_id, seat_number, status, transaction_id, created_at`,
			ticketCode, req.ShowtimeID, seatNumber, "booked", transaction.TransactionID, time.Now()).Scan(
			&ticket.TicketID, &ticket.TicketCode, &ticket.ShowtimeID, &ticket.SeatNumber,
			&ticket.Status, &ticket.TransactionID, &ticket.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to create ticket for seat %s: %w", seatNumber, err)
		}
		tickets = append(tickets, ticket)
	}

	_, err = tx.Exec(ctx, `
		UPDATE showtimes 
		SET available_seats = available_seats - $1 
		WHERE showtime_id = $2`,
		len(req.SeatNumbers), req.ShowtimeID)
	if err != nil {
		return nil, fmt.Errorf("failed to update available seats: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	transactionResponse := toTransactionResponse(transaction)
	ticketResponse := toTicketResponse(tickets)

	return &dto.TransactionResult{
		Transaction: transactionResponse,
		Tickets:     ticketResponse,
	}, nil
}

func (s *TransactionService) ProcessPayment(ctx context.Context, req dto.ProcessPaymentRequest) (*dto.TransactionResult, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT transaction_id, transaction_code, recipient_email, recipient_full_name, 
		       recipient_phone_number, total_seats, total_amount, status, 
		       created_at, expires_at, paid_at, created_by, payment_method_id
		FROM transactions 
		WHERE transaction_code = $1`,
		req.TransactionCode)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	transaction, err := pgx.CollectOneRow[models.Transaction](rows, pgx.RowToStructByName)
	if err != nil {
		return nil, fmt.Errorf("unable to get transaction data: %w", err)
	}

	if transaction.Status != "pending" {
		return nil, fmt.Errorf("transaction is not pending")
	}

	if time.Now().After(transaction.ExpiresAt) {
		_, err = s.CancelTransaction(ctx, req.TransactionCode)
		if err != nil {
			return nil, fmt.Errorf("transaction expired and failed to cancel: %w", err)
		}
		return nil, fmt.Errorf("transaction has expired")
	}

	now := time.Now()
	_, err = tx.Exec(ctx, `
		UPDATE transactions 
		SET status = 'paid', paid_at = $1 
		WHERE transaction_id = $2`,
		now, transaction.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction status: %w", err)
	}

	transaction.Status = "paid"
	transaction.PaidAt = &now

	tickets := make([]models.Ticket, 0)
	rows, err = tx.Query(ctx, `
		SELECT ticket_id, ticket_code, showtime_id, seat_number, status, transaction_id, created_at
		FROM tickets 
		WHERE transaction_id = $1`,
		transaction.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tickets: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ticket models.Ticket
		if err := rows.Scan(&ticket.TicketID, &ticket.TicketCode, &ticket.ShowtimeID,
			&ticket.SeatNumber, &ticket.Status, &ticket.TransactionID, &ticket.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	transactionResponse := toTransactionResponse(transaction)
	ticketResponse := toTicketResponse(tickets)

	return &dto.TransactionResult{
		Transaction: transactionResponse,
		Tickets:     ticketResponse,
	}, nil
}

func (s *TransactionService) CancelTransaction(ctx context.Context, transactionCode string) (*dto.TransactionResult, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
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
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	transaction, err := pgx.CollectOneRow[models.Transaction](rows, pgx.RowToStructByName)
	if err != nil {
		return nil, fmt.Errorf("unable to get transaction data: %w", err)
	}

	if transaction.Status == "paid" {
		return nil, fmt.Errorf("cannot cancel paid transaction")
	}

	if transaction.Status == "cancelled" {
		return nil, fmt.Errorf("transaction already cancelled")
	}

	var showtimeID int
	err = tx.QueryRow(ctx, `
		SELECT showtime_id FROM tickets 
		WHERE transaction_id = $1 LIMIT 1`,
		transaction.TransactionID).Scan(&showtimeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get showtime: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE transactions 
		SET status = 'cancelled' 
		WHERE transaction_id = $1`,
		transaction.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel transaction: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE tickets 
		SET status = 'cancelled' 
		WHERE transaction_id = $1`,
		transaction.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel tickets: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE showtimes 
		SET available_seats = available_seats + $1 
		WHERE showtime_id = $2`,
		transaction.TotalSeats, showtimeID)
	if err != nil {
		return nil, fmt.Errorf("failed to release seats: %w", err)
	}

	transaction.Status = "cancelled"

	tickets := make([]models.Ticket, 0)
	rows, err = tx.Query(ctx, `
		SELECT ticket_id, ticket_code, showtime_id, seat_number, status, transaction_id, created_at
		FROM tickets 
		WHERE transaction_id = $1`,
		transaction.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tickets: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ticket models.Ticket
		if err := rows.Scan(&ticket.TicketID, &ticket.TicketCode, &ticket.ShowtimeID,
			&ticket.SeatNumber, &ticket.Status, &ticket.TransactionID, &ticket.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	transactionResponse := toTransactionResponse(transaction)
	ticketResponse := toTicketResponse(tickets)

	return &dto.TransactionResult{
		Transaction: transactionResponse,
		Tickets:     ticketResponse,
	}, nil
}

func (s *TransactionService) GetTransactions(ctx context.Context, transactionCode string) (*[]dto.TransactionListResponse, error) {
	rows, err := s.db.Query(ctx, `
		SELECT 
  			t.transaction_id, t.transaction_code, t.status, t.total_amount, t.expires_at, t.created_at,
  			tk.seat_number,
  			s.showtime_id, s.show_datetime, s.price,
  			m.movie_id, m.title AS movie_title,
  			c.id AS cinema_id, c.name AS cinema_name, c.location AS cinema_location
		FROM transactions t
		LEFT JOIN tickets tk ON t.transaction_id = tk.transaction_id
		LEFT JOIN showtimes s ON tk.showtime_id = s.showtime_id
		LEFT JOIN movies m ON s.movie_id = m.movie_id
		LEFT JOIN cinemas c ON s.cinema_id = c.id
		ORDER BY t.transaction_id DESC`,
	)

	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	joinRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.TransactionJoinRow])
	if err != nil {
		return nil, err
	}

	var (
		results   []dto.TransactionListResponse
		lastTxID  int
		currentTx *dto.TransactionListResponse
	)

	for _, row := range joinRows {
		if row.TransactionID != lastTxID {
			lastTxID = row.TransactionID
			tx := dto.TransactionListResponse{
				TransactionID:   row.TransactionID,
				TransactionCode: row.TransactionCode,
				Status:          row.Status,
				TotalAmount:     row.TotalAmount,
				ExpiresAt:       row.ExpiresAt,
				CreatedAt:       row.CreatedAt,
				Movie: dto.MovieResponse{
					MovieID: *row.MovieID,
					Title:   *row.MovieTitle,
				},
				Showtime: dto.ShowtimeResponse{
					ShowtimeID:   *row.ShowtimeID,
					ShowDatetime: *row.ShowDatetime,
					Price:        *row.Price,
				},
				Cinema: dto.CinemaResponse{
					CinemaID: *row.CinemaID,
					Name:     *row.CinemaName,
					Location: *row.CinemaLocation,
				},
				Seats: []string{},
			}
			results = append(results, tx)
			currentTx = &results[len(results)-1]
		}

		if row.SeatNumber != nil {
			currentTx.Seats = append(currentTx.Seats, *row.SeatNumber)
		}
	}

	return &results, nil

}

func (s *TransactionService) GetTransactionByCode(ctx context.Context, transactionCode string) (*dto.TransactionResult, error) {
	var transaction models.Transaction
	err := s.db.QueryRow(ctx, `
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
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	tickets := make([]models.Ticket, 0)
	rows, err := s.db.Query(ctx, `
		SELECT ticket_id, ticket_code, showtime_id, seat_number, status, transaction_id, created_at
		FROM tickets 
		WHERE transaction_id = $1`,
		transaction.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tickets: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ticket models.Ticket
		if err := rows.Scan(&ticket.TicketID, &ticket.TicketCode, &ticket.ShowtimeID,
			&ticket.SeatNumber, &ticket.Status, &ticket.TransactionID, &ticket.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}

	transactionResponse := toTransactionResponse(transaction)
	ticketResponse := toTicketResponse(tickets)

	return &dto.TransactionResult{
		Transaction: transactionResponse,
		Tickets:     ticketResponse,
	}, nil
}

func toTransactionResponse(t models.Transaction) dto.TransactionResponse {
	return dto.TransactionResponse{
		TransactionID:     t.TransactionID,
		TransactionCode:   t.TransactionCode,
		RecipientEmail:    t.RecipientEmail,
		RecipientFullName: t.RecipientFullName,
		RecipientPhone:    t.RecipientPhone,
		TotalSeats:        t.TotalSeats,
		TotalAmount:       t.TotalAmount,
		Status:            t.Status,
		CreatedAt:         t.CreatedAt,
		ExpiresAt:         t.ExpiresAt,
		PaidAt:            t.PaidAt,
		CreatedBy:         t.CreatedBy,
	}
}

func toTicketResponse(tickets []models.Ticket) []dto.TicketResponse {
	responses := make([]dto.TicketResponse, 0, len(tickets))
	for _, t := range tickets {
		responses = append(responses, dto.TicketResponse{
			TicketID:      t.TicketID,
			TicketCode:    t.TicketCode,
			ShowtimeID:    t.ShowtimeID,
			SeatNumber:    t.SeatNumber,
			Status:        t.Status,
			TransactionID: t.TransactionID,
			CreatedAt:     t.CreatedAt,
		})
	}

	return responses
}

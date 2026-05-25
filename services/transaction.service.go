package services

import (
	"context"
	"fmt"
	"noir-backend/dto"
	"noir-backend/models"
	"noir-backend/repositories"
)

type TransactionService interface {
	CreateTransaction(ctx context.Context, req dto.CreateTransactionRequest, userID int) (*dto.TransactionResult, error)
	ProcessPayment(ctx context.Context, req dto.ProcessPaymentRequest) (*dto.TransactionResult, error)
	CancelTransaction(ctx context.Context, transactionCode string) (*dto.TransactionResult, error)
	GetTransactions(ctx context.Context, transactionCode string) (*[]dto.TransactionListResponse, error)
	GetTransactionByCode(ctx context.Context, transactionCode string) (*dto.TransactionResult, error)
}

type transactionService struct {
	repo repositories.TransactionRepository
}

func NewTransactionService(repo repositories.TransactionRepository) TransactionService {
	return &transactionService{repo: repo}
}

func (s *transactionService) CreateTransaction(ctx context.Context, req dto.CreateTransactionRequest, userID int) (*dto.TransactionResult, error) {
	transaction, tickets, err := s.repo.CreateTransaction(ctx, userID, req.ShowtimeID, req.SeatNumbers, req.PaymentMethodID, req.RecipientEmail, req.RecipientFullName, req.RecipientPhone)
	if err != nil {
		return nil, err
	}

	return &dto.TransactionResult{
		Transaction: toTransactionResponse(*transaction),
		Tickets:     toTicketResponse(tickets),
	}, nil
}

func (s *transactionService) ProcessPayment(ctx context.Context, req dto.ProcessPaymentRequest) (*dto.TransactionResult, error) {
	transaction, tickets, err := s.repo.ProcessPayment(ctx, req.TransactionCode)
	if err != nil {
		if err.Error() == "transaction has expired" {
			_, _, cancelErr := s.repo.CancelTransaction(ctx, req.TransactionCode)
			if cancelErr != nil {
				return nil, fmt.Errorf("transaction expired and failed to cancel: %w", cancelErr)
			}
		}
		return nil, err
	}

	return &dto.TransactionResult{
		Transaction: toTransactionResponse(*transaction),
		Tickets:     toTicketResponse(tickets),
	}, nil
}

func (s *transactionService) CancelTransaction(ctx context.Context, transactionCode string) (*dto.TransactionResult, error) {
	transaction, tickets, err := s.repo.CancelTransaction(ctx, transactionCode)
	if err != nil {
		return nil, err
	}

	return &dto.TransactionResult{
		Transaction: toTransactionResponse(*transaction),
		Tickets:     toTicketResponse(tickets),
	}, nil
}

func (s *transactionService) GetTransactions(ctx context.Context, transactionCode string) (*[]dto.TransactionListResponse, error) {
	joinRows, err := s.repo.GetTransactions(ctx)
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
					BasePrice:    *row.TicketPrice,
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

func (s *transactionService) GetTransactionByCode(ctx context.Context, transactionCode string) (*dto.TransactionResult, error) {
	transaction, tickets, err := s.repo.GetTransactionByCode(ctx, transactionCode)
	if err != nil {
		return nil, err
	}

	return &dto.TransactionResult{
		Transaction: toTransactionResponse(*transaction),
		Tickets:     toTicketResponse(tickets),
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
		PaymentMethodID:   t.PaymentMethodID,
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
			Price:         t.Price,
			Status:        t.Status,
			TransactionID: t.TransactionID,
			CreatedAt:     t.CreatedAt,
		})
	}

	return responses
}

package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func StartTransactionExpiryJob(ctx context.Context, db *pgxpool.Pool) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("Transaction expiration job stopped")
				return
			case <-ticker.C:
				_, err := db.Exec(context.Background(), `
					UPDATE transactions 
					SET status = 'expired' 
					WHERE status = 'pending' AND expires_at < NOW()
				`)
				if err != nil {
					fmt.Println("Failed to expire transactions:", err)
				} else {
					fmt.Println("Expired pending transactions (if any)")
				}
			}
		}
	}()
}

func GenerateTransactionCode() string {
	return fmt.Sprintf("TXN-%d-%s", time.Now().Unix(), uuid.New().String()[:8])
}

func GenerateTicketCode() string {
	return fmt.Sprintf("TKT-%d-%s", time.Now().Unix(), uuid.New().String()[:8])
}

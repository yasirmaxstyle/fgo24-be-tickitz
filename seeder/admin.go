package seeder

import (
	"context"
	"fmt"
	"log"
	"noir-backend/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

func SeedAdminUser(db *pgxpool.Pool) {
	ctx := context.Background()

	tx, err := db.Begin(context.Background())
	if err != nil {
		log.Fatal("failed to begin transaction: ", err)
	}
	defer tx.Rollback(context.Background())

	adminEmail := utils.Load().Admin.Email
	adminPassword := utils.Load().Admin.Password

	var count int
	query := `
		SELECT COUNT(*) FROM users
		WHERE email = $1 AND role = 'admin'`

	err = tx.QueryRow(ctx, query, adminEmail).Scan(&count)
	if err != nil {
		log.Fatal("failed to check admin existence: ", err)
	}

	if count > 0 {
		log.Printf("Admin '%s' already exists", adminEmail)
		return
	}

	hashedPassword, err := utils.HashPassword(adminPassword)
	if err != nil {
		log.Fatal("failed to hash password: ", err)
	}

	var userID int

	err = tx.QueryRow(context.Background(), `
		INSERT INTO users (email, password, role)
		VALUES ($1, $2, 'admin')
		RETURNING user_id`,
		adminEmail, hashedPassword).
		Scan(&userID)
	if err != nil {
		log.Fatal("failed to create admin user: ", err)
	}

	result, err := tx.Exec(context.Background(), `
		INSERT INTO profile (user_id) 
		VALUES ($1)
		RETURNING user_id`,
		userID)
	if err != nil {
		log.Fatal("failed to create admin user: ", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Fatal("no rows were inserted")
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Fatal("failed to commit transaction: ", err)
	}

	log.Printf("Admin '%s' created successfully", adminEmail)
	fmt.Println(userID)
}

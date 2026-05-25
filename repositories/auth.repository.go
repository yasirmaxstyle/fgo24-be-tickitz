package repositories

import (
	"context"
	"fmt"
	"noir-backend/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository interface {
	CheckUserExists(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, userID int) (*models.Profile, error)
	UpdateLastLogin(ctx context.Context, userID int, loginTime time.Time) error
	GetUserIDByEmail(ctx context.Context, email string) (int, error)
	UpdatePassword(ctx context.Context, userID int, hashedPassword string) error
}

type authRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) CheckUserExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	return exists, err
}

func (r *authRepository) CreateUser(ctx context.Context, user *models.User) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create user first to get user_id
	err = tx.QueryRow(ctx, `
		INSERT INTO users (email, password, role)
		VALUES ($1, $2, 'user')
		RETURNING user_id, created_at, updated_at`,
		user.Email, user.Password).
		Scan(&user.UserID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Create profile with user_id
	var profileID int
	err = tx.QueryRow(ctx, `
		INSERT INTO profile (user_id, first_name, last_name, phone_number, avatar) 
		VALUES ($1, '', '', '', '')
		RETURNING profile_id`, user.UserID).Scan(&profileID)
	if err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(ctx,
		`SELECT user_id, email, password, role, created_at, updated_at, last_login
		FROM users WHERE email = $1`,
		email).Scan(&user.UserID, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin)
	
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	} else if err != nil {
		return nil, err
	}
	
	return &user, nil
}

func (r *authRepository) GetUserByID(ctx context.Context, userID int) (*models.Profile, error) {
	var profile models.Profile
	err := r.db.QueryRow(ctx, `
		SELECT p.profile_id, p.user_id, p.first_name, p.last_name, p.phone_number, p.avatar, p.created_at, p.updated_at
		FROM profile p
		JOIN users u ON u.user_id = p.user_id
		WHERE u.user_id = $1`,
		userID).Scan(&profile.ProfileID, &profile.UserID, &profile.FirstName, &profile.LastName, &profile.PhoneNumber, &profile.Avatar, &profile.CreatedAt, &profile.UpdatedAt)
	
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	} else if err != nil {
		return nil, err
	}
	
	return &profile, nil
}

func (r *authRepository) UpdateLastLogin(ctx context.Context, userID int, loginTime time.Time) error {
	_, err := r.db.Exec(ctx, "UPDATE users SET last_login = $1 WHERE user_id = $2", loginTime, userID)
	return err
}

func (r *authRepository) GetUserIDByEmail(ctx context.Context, email string) (int, error) {
	var id int
	err := r.db.QueryRow(ctx, "SELECT user_id FROM users WHERE email = $1", email).Scan(&id)
	if err == pgx.ErrNoRows {
		return 0, fmt.Errorf("user not found")
	}
	return id, err
}

func (r *authRepository) UpdatePassword(ctx context.Context, userID int, hashedPassword string) error {
	_, err := r.db.Exec(ctx, "UPDATE users SET password = $1 WHERE user_id = $2", hashedPassword, userID)
	return err
}

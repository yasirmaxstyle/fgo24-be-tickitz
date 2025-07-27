package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"noir-backend/dto"
	"noir-backend/models"
	"text/template"
	"time"

	"noir-backend/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type AuthService struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewAuthService(db *pgxpool.Pool, redis *redis.Client) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)",
		req.Email).Scan(&exists)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, fmt.Errorf("user already exists")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, 'user')
		RETURNING user_id`,
		user.Email, user.PasswordHash).
		Scan(&user.UserID)
	if err != nil {
		return nil, err
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO profile (user_id) 
		VALUES ($1)
		RETURNING user_id`,
		user.UserID).Scan(&user.UserID)
	if err != nil {
		return nil, err
	}

	userReponse := &dto.UserResponse{
		UserID:    user.UserID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		LastLogin: user.LastLogin,
	}

	return userReponse, tx.Commit(ctx)
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user := &models.User{}
	err := s.db.QueryRow(ctx,
		`SELECT user_id, email, password_hash, role, created_at, updated_at, last_login
		FROM users WHERE email = $1`,
		req.Email).Scan(&user.UserID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("invalid credentials")
	}

	if err := utils.CheckPasswordHash(req.Password, user.PasswordHash); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := utils.GenerateTokens(user.UserID, user.Role)
	if err != nil {
		return nil, err
	}

	if err := s.UpdateLastLogin(ctx, &user.UserID); err != nil {
		return nil, err
	}

	userReponse := &dto.UserResponse{
		UserID:    user.UserID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		LastLogin: user.LastLogin,
	}

	return &dto.AuthResponse{
		User:  userReponse,
		Token: token,
	}, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, userID int) (*models.Profile, error) {
	var user models.Profile
	err := s.db.QueryRow(ctx, `
		SELECT profile_id, first_name, last_name, email, phone_number, p.created_at, p.updated_at, last_login
		FROM profile p
		JOIN users u ON u.user_id = p.user_id
		WHERE p.user_id = $1`,
		userID).Scan(&user.UserID, &user.FirstName, &user.LastName, &user.Email, &user.PhoneNumber, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &user, nil
}

func (r *AuthService) UpdateLastLogin(ctx context.Context, userID *int) error {
	query := `UPDATE users SET last_login = $1 WHERE user_id = $2`
	_, err := r.db.Exec(ctx, query, time.Now(), userID)
	return err
}

func (r *AuthService) Logout(token string) error {
	return r.redis.Set(context.Background(), fmt.Sprintf("blacklist-token:%s", token), "1", 24*time.Hour).Err()
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) (string, error) {
	var userID int
	err := s.db.QueryRow(context.Background(),
		"SELECT user_id FROM users WHERE email = $1", email).Scan(&userID)

	if err == pgx.ErrNoRows {
		return "If the email exists, a reset link has been sent", nil
	} else if err != nil {
		return "", fmt.Errorf("database error: %w", err)
	}
	token, err := utils.GenerateTokens(userID, "user")
	if err != nil {
		return "", fmt.Errorf("failed to generate token reset: %w", err)
	}

	utils.InitRedis().Set(ctx, fmt.Sprintf("reset-pwd:%s", token), "1", 1*time.Hour).Err()

	if err := sendResetEmail(email, token); err != nil {
		log.Printf("Failed to send reset email: %v\n", err)
		return "", fmt.Errorf("failed to send reset email")
	}

	return "If the email exists, a reset link has been sent", nil

}

func (s *AuthService) ResetPassword(ctx context.Context, req dto.ResetPasswordRequest, token string) (int, error) {
	expCmd := utils.InitRedis().Exists(context.Background(), fmt.Sprintf("reset-pwd:%s", token))
	if expCmd.Val() == 0 {
		return http.StatusUnauthorized, fmt.Errorf("invalid or expired reset token")
	}

	claims, err := utils.ValidateToken(token)
	if err != nil {
		return http.StatusUnauthorized, fmt.Errorf("invalid token")
	}

	userID := int(claims["user_id"].(float64))

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to hash password")
	}

	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("database error")
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(),
		"UPDATE users SET password_hash = $1 WHERE user_id = $2",
		hashedPassword, userID)

	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update password")
	}

	if err = tx.Commit(context.Background()); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to commit changes")
	}

	utils.InitRedis().Del(context.Background(), fmt.Sprintf("reset-pwd:%s", token)).Err()

	return http.StatusOK, nil
}

func sendResetEmail(email, token string) error {
	resetURL := fmt.Sprintf("http://localhost:8080/reset-password?token=%s", token)
	subject := "Password Reset Request"
	body, err := buildResetEmailBody(resetURL)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", email, subject, body)

	config := utils.Load().SMTP

	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	return smtp.SendMail(addr, auth, config.From, []string{email}, []byte(msg))
}

func buildResetEmailBody(resetURL string) (string, error) {
	tmpl, err := template.ParseFiles("templates/reset_password_email.txt")
	if err != nil {
		return "", fmt.Errorf("error parsing file: %v", err)
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, struct{ ResetURL string }{ResetURL: resetURL})
	if err != nil {
		return "", fmt.Errorf("error execute file: %v", err)
	}

	return body.String(), nil
}

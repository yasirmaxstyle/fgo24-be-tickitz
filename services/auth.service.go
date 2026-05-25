package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"text/template"
	"time"

	"noir-backend/dto"
	"noir-backend/models"
	"noir-backend/repositories"
	"noir-backend/utils"

	"github.com/redis/go-redis/v9"
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error)
	GetUserByID(ctx context.Context, userID int) (*models.Profile, error)
	UpdateLastLogin(ctx context.Context, userID *int) error
	Logout(token string) error
	ForgotPassword(ctx context.Context, email string) (string, error)
	ResetPassword(ctx context.Context, req dto.ResetPasswordRequest, token string) (int, error)
}

type authService struct {
	repo  repositories.AuthRepository
	redis *redis.Client
}

func NewAuthService(repo repositories.AuthRepository, redis *redis.Client) *authService {
	return &authService{repo: repo, redis: redis}
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error) {
	exists, err := s.repo.CheckUserExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("user already exists")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:    req.Email,
		Password: hashedPassword,
	}

	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	userResponse := &dto.UserResponse{
		UserID:    user.UserID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		LastLogin: user.LastLogin,
	}

	return userResponse, nil
}

func (s *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("invalid credentials")
	}

	if err := utils.CheckPasswordHash(req.Password, user.Password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := utils.GenerateTokens(user.UserID, user.Role)
	if err != nil {
		return nil, err
	}

	if err := s.UpdateLastLogin(ctx, &user.UserID); err != nil {
		return nil, err
	}

	userResponse := &dto.UserResponse{
		UserID:    user.UserID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		LastLogin: user.LastLogin,
	}

	return &dto.AuthResponse{
		User:  userResponse,
		Token: token,
	}, nil
}

func (s *authService) GetUserByID(ctx context.Context, userID int) (*models.Profile, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return user, nil
}

func (r *authService) UpdateLastLogin(ctx context.Context, userID *int) error {
	return r.repo.UpdateLastLogin(ctx, *userID, time.Now())
}

func (r *authService) Logout(token string) error {
	return r.redis.Set(context.Background(), fmt.Sprintf("blacklist-token:%s", token), "1", 24*time.Hour).Err()
}

func (s *authService) ForgotPassword(ctx context.Context, email string) (string, error) {
	userID, err := s.repo.GetUserIDByEmail(context.Background(), email)
	if err != nil {
		if err.Error() == "user not found" {
			return "If the email exists, a reset link has been sent", nil
		}
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

func (s *authService) ResetPassword(ctx context.Context, req dto.ResetPasswordRequest, token string) (int, error) {
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

	err = s.repo.UpdatePassword(context.Background(), userID, hashedPassword)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update password")
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

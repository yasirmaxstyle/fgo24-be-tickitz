package controllers

import (
	"net/http"
	"noir-backend/dto"
	"noir-backend/services"
	"noir-backend/utils"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email, password, and and confirm password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Registration request"
// @Success 201 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBind(&req); err != nil {
		utils.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if req.Password != req.ConfirmPassword {
		utils.SendError(ctx, http.StatusBadRequest, "confirm password must match")
		return
	}

	user, err := c.authService.Register(ctx.Request.Context(), req)
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, "failed to create user")
		return
	}

	utils.SendSuccess(ctx, http.StatusCreated, "user registered successfully", user)
}

// Login godoc
// @Summary Login user
// @Description Login user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login request"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBind(&req); err != nil {
		utils.SendError(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	response, err := c.authService.Login(ctx.Request.Context(), &req)
	if err != nil {
		utils.SendError(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, "login successful", response)
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user profile
// @Tags profile
// @Produce json
// @Security Token
// @Success 200 {object} models.User
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /profile [get]
func (c *AuthController) GetProfile(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.SendError(ctx, http.StatusUnauthorized, "Status Unauthorized")
		return
	}

	user, err := c.authService.GetUserByID(ctx.Request.Context(), userID.(int))
	if err != nil {
		utils.SendError(ctx, http.StatusNotFound, "User not found")
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, "data retrieved successfully", user)
}

// Logout godoc
// @Summary Logout user
// @Description Logout user by blacklisting refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Logout request"
// @Security Token
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	var req dto.LogoutRequest
	if err := ctx.ShouldBind(&req); err != nil {
		utils.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := c.authService.Logout(req.Token); err != nil {
		utils.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, "Logged out successfully", nil)
}

// Forgot Password godoc
// @Summary Request reset password
// @Description Request reset password user with email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.PasswordResetRequest true "Forgot password request"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/forgot-password [post]
func (c *AuthController) ForgotPassword(ctx *gin.Context) {
	var req dto.PasswordResetRequest
	if err := ctx.ShouldBind(&req); err != nil {
		utils.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	message, err := c.authService.ForgotPassword(ctx.Request.Context(), req.Email)
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, message, nil)
}

// Reset Password godoc
// @Summary Reset password in new link provided
// @Description Reset password user with new password
// @Tags auth
// @Accept json
// @Produce json
// @Param token query string true "token request"
// @Param request body models.ResetPasswordRequest true "Reset password request"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/reset-password [post]
func (c *AuthController) ResetPassword(ctx *gin.Context) {
	token := ctx.Query("token")
	var req dto.ResetPasswordRequest
	if err := ctx.ShouldBind(&req); err != nil {
		utils.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	status, err := c.authService.ResetPassword(ctx.Request.Context(), req, token)
	if err != nil {
		utils.SendError(ctx, status, err.Error())
	}

	utils.SendSuccess(ctx, http.StatusOK, "password reset successfully", nil)
}

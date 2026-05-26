package controllers

import (
	"net/http"
	"noir-backend/dto"
	"noir-backend/services"
	"noir-backend/utils"

	"github.com/gin-gonic/gin"
)

type TransactionController struct {
	transactionService services.TransactionService
}

func NewTransactionController(transactionService services.TransactionService) *TransactionController {
	return &TransactionController{transactionService: transactionService}
}

// Create Transaction godoc
// @Summary Create a new transaction
// @Description Create a new movie ticket transaction
// @Tags transaction
// @Accept json
// @Produce json
// @Param request body dto.CreateTransactionRequest true "Transaction details"
// @Security Token
// @Success 201 {object} dto.SuccessResponse{data=dto.TransactionResult} "Transaction created successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Router /transaction/ [post]
func (c *TransactionController) CreateTransaction(ctx *gin.Context) {
	var req dto.CreateTransactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.SendError(ctx, http.StatusUnauthorized, "user not aunthenticated")
		return
	}

	response, err := c.transactionService.CreateTransaction(ctx.Request.Context(), req, userID.(int))
	if err != nil {
		utils.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(ctx, http.StatusCreated, "Transaction created successfully", response)
}

// Process Payment godoc
// @Summary Process payment
// @Description Process a payment for a transaction
// @Tags transaction
// @Accept json
// @Produce json
// @Param request body dto.ProcessPaymentRequest true "Payment details"
// @Security Token
// @Success 200 {object} dto.SuccessResponse{data=dto.TransactionResult} "Payment processed successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Router /transaction/payment [post]
func (c *TransactionController) ProcessPayment(ctx *gin.Context) {
	var req dto.ProcessPaymentRequest
	if err := ctx.ShouldBind(&req); err != nil {
		utils.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response, err := c.transactionService.ProcessPayment(ctx.Request.Context(), req)
	if err != nil {
		utils.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, "Payment processed successfully", response)
}

// Get Transaction godoc
// @Summary Get transaction by code
// @Description Get full details of a transaction using its unique code
// @Tags transaction
// @Produce json
// @Param code path string true "Transaction Code"
// @Security Token
// @Success 200 {object} dto.SuccessResponse{data=dto.TransactionListResponse} "Transaction retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 404 {object} dto.ErrorResponse "Transaction not found"
// @Router /transaction/{code} [get]
func (c *TransactionController) GetTransaction(ctx *gin.Context) {
	transactionCode := ctx.Param("code")
	if transactionCode == "" {
		utils.SendError(ctx, http.StatusBadRequest, "transaction code is required")
		return
	}

	response, err := c.transactionService.GetTransactionByCode(ctx.Request.Context(), transactionCode)
	if err != nil {
		utils.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, "Transaction retrieved successfully", response)
}

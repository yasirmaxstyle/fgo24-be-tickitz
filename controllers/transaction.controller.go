package controllers

import (
	"net/http"
	"noir-backend/dto"
	"noir-backend/services"
	"noir-backend/utils"

	"github.com/gin-gonic/gin"
)

type TransactionController struct {
	transactionService *services.TransactionService
}

func NewTransactionController(transactionService *services.TransactionService) *TransactionController {
	return &TransactionController{transactionService: transactionService}
}

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

package router

import (
	"noir-backend/container"
	"noir-backend/middleware"

	"github.com/gin-gonic/gin"
)

func transactionRouter(r *gin.RouterGroup, c *container.Container) {
	r.Use(middleware.AuthMiddleware())
	r.POST("/", c.TransactionController.CreateTransaction)
	r.POST("/payment", c.TransactionController.ProcessPayment)
	r.GET("/:code", c.TransactionController.GetTransaction)
	r.GET("/", c.TransactionController.GetTransaction)

}

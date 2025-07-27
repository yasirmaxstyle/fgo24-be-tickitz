package router

import (
	"noir-backend/container"
	"noir-backend/middleware"

	"github.com/gin-gonic/gin"
)

func authRouter(r *gin.RouterGroup, c *container.Container) {
	r.POST("/register", c.AuthController.Register)
	r.POST("/login", c.AuthController.Login)
	r.POST("/forgot-password", c.AuthController.ForgotPassword)
	r.POST("/reset-password", c.AuthController.ResetPassword)

	r.Use(middleware.AuthMiddleware())
	r.POST("/logout", c.AuthController.Logout)
}

package router

import (
	"noir-backend/container"
	"noir-backend/middleware"

	"github.com/gin-gonic/gin"
)

func userRouter(r *gin.RouterGroup, c *container.Container) {
	r.Use(middleware.AuthMiddleware())
	r.GET("/", c.AuthController.GetProfile)
	r.PATCH("/") //edit profile
}

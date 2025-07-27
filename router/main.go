package router

import (
	"noir-backend/container"
	docs "noir-backend/docs"
	"noir-backend/middleware"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func CombineRouter(r *gin.Engine, c *container.Container) {
	docs.SwaggerInfo.BasePath = "/"
	r.Use(middleware.CORS())
	r.Use(middleware.ErrorHandler())
	r.Static("/uploads", "./uploads")

	authRouter(r.Group("/auth"), c)
	adminRouter(r.Group("/admin"), c)
	userRouter(r.Group("/profile"), c)
	movieRouter(r.Group("/movie"), c)
	transactionRouter(r.Group("/transaction"), c)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

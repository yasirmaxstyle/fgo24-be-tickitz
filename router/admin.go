package router

import (
	"noir-backend/container"
	"noir-backend/middleware"

	"github.com/gin-gonic/gin"
)

func adminRouter(r *gin.RouterGroup, c *container.Container) {
	r.Use(middleware.AuthMiddleware())
	r.POST("/movie", c.MovieController.AddMovie)          //add movie by admin
	r.PATCH("/movie/:id", c.MovieController.UpdateMovie)  //edit movie by admin
	r.DELETE("/movie/:id", c.MovieController.DeleteMovie) //edit movie by admin
}

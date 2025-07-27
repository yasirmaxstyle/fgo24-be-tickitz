package router

import (
	"noir-backend/container"
	"noir-backend/middleware"

	"github.com/gin-gonic/gin"
)

func movieRouter(r *gin.RouterGroup, c *container.Container) {
	r.Use(middleware.AuthMiddleware())
	r.GET("/upcoming-movies", c.MovieController.GetMoviesUpcoming)
	r.GET("/now-playing-movies", c.MovieController.GetMoviesNowPlaying)
	r.GET("/", c.MovieController.GetMovies)
	r.GET("/:id", c.MovieController.GetMovieByID)
	r.GET("/genres", c.MovieController.GetGenres)
}

package controllers

import (
	"log"
	"net/http"
	"noir-backend/dto"
	"noir-backend/services"
	"noir-backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MovieController struct {
	movieService services.MovieService
}

func NewMovieController(movieService services.MovieService) *MovieController {
	return &MovieController{movieService: movieService}
}

// Add Movie godoc
// @Summary Add new movie
// @Description Add new movie by admin
// @Tags admin
// @Produce json
// @Accept multipart/form-data
// @Param title formData string true "Movie Title"
// @Param overview formData string true "Overview"
// @Param duration formData int true "Duration (in minutes)"
// @Param release_date formData string true "Release Date (YYYY-MM-DD)"
// @Param director_id formData string false "Director Name"
// @Param genre_ids formData []string false "Genre IDs list"
// @Param cast formData []string false "Cast list"
// @Param poster_path formData file false "Poster Image"
// @Param backdrop_path formData file false "Backdrop Image"
// @Security Token
// @Success 200 {object} dto.SuccessResponse "Movie created successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Only accessed by admin"
// @Failure 500 {object} dto.ErrorResponse "Something went wrong"
// @Router /admin/movie [post]
func (c *MovieController) AddMovie(ctx *gin.Context) {
	var req dto.CreateMovieRequest
	if err := ctx.ShouldBind(&req); err != nil {
		utils.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	posterPath, err := utils.SaveUploadedFile(ctx, "poster_path", "uploads/movies/posters")
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	backdropPath, err := utils.SaveUploadedFile(ctx, "backdrop_path", "uploads/movies/backdrops")
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	if posterPath != nil {
		log.Println("Poster uploaded to:", *posterPath)
	}

	if backdropPath != nil {
		log.Println("Backdrop uploaded to:", *backdropPath)
	}

	movie, err := c.movieService.CreateMovie(ctx.Request.Context(), req, posterPath, backdropPath)
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendSuccess(ctx, http.StatusCreated, "Movie created successfully", movie)
}

// Update Movie godoc
// @Summary Update existing movie
// @Description Add existing movie by admin
// @Tags admin
// @Produce json
// @Accept multipart/form-data
// @Param title formData string false "Movie Title"
// @Param overview formData string false "Overview"
// @Param duration formData int false "Duration (in minutes)"
// @Param release_date formData string false "Release Date (YYYY-MM-DD)"
// @Param director_id formData string false "Director Name"
// @Param genre_ids formData []string false "Genre IDs list"
// @Param cast formData []string false "Cast list"
// @Param poster_path formData file false "Poster Image"
// @Param backdrop_path formData file false "Backdrop Image"
// @Param id path integer true "Movie id"
// @Security Token
// @Success 200 {object} dto.SuccessResponse "Movie updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Only accessed by admin"
// @Failure 500 {object} dto.ErrorResponse "Something went wrong"
// @Router /admin/movie/{id} [patch]
func (c *MovieController) UpdateMovie(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.SendError(ctx, http.StatusBadRequest, "Invalid movie ID")
		return
	}

	var req dto.UpdateMovieRequest
	if err := ctx.ShouldBind(&req); err != nil {
		utils.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	posterPath, err := utils.SaveUploadedFile(ctx, "poster_path", "uploads/movies/posters")
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	backdropPath, err := utils.SaveUploadedFile(ctx, "backdrop_path", "uploads/movies/backdrops")
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	if posterPath != nil {
		log.Println("Poster uploaded to:", *posterPath)
	}

	if backdropPath != nil {
		log.Println("Backdrop uploaded to:", *backdropPath)
	}

	err = c.movieService.UpdateMovie(ctx.Request.Context(), id, req, backdropPath, posterPath)
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, "Movie updated successfully", nil)
}

// Delete Movie godoc
// @Summary Delete existing movie
// @Description Delete existing movie by admin
// @Tags admin
// @Produce json
// @Param id path integer true "Movie id"
// @Security Token
// @Success 200 {object} dto.SuccessResponse "Movie deleted successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Only accessed by admin"
// @Failure 500 {object} dto.ErrorResponse "Something went wrong"
// @Router /admin/movie/{id} [delete]
func (c *MovieController) DeleteMovie(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.SendError(ctx, http.StatusBadRequest, "Invalid movie ID")
		return
	}

	err = c.movieService.DeleteMovie(ctx.Request.Context(), id)
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, "Movie deleted successfully", nil)
}

func (c *MovieController) GetMovies(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	movies, total, err := c.movieService.GetMovies(ctx.Request.Context(), limit, offset)
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	pagination := dto.NewPagination(ctx, total, page, limit)
	response := dto.PagedMoviesResponse{
		PageInfo: pagination,
		Result:   movies,
	}
	utils.SendSuccess(ctx, http.StatusOK, "all movies retrieved successfully", response)
}

func (c *MovieController) GetMoviesNowPlaying(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	movies, total, err := c.movieService.GetNowPlayingMovies(ctx.Request.Context(), limit, offset)
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	pagination := dto.NewPagination(ctx, total, page, limit)
	response := dto.PagedMoviesResponse{
		PageInfo: pagination,
		Result:   movies,
	}
	utils.SendSuccess(ctx, http.StatusOK, "now playing movies retrieved successfully", response)
}

func (c *MovieController) GetMoviesUpcoming(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	movies, total, err := c.movieService.GetUpcomingMovies(ctx.Request.Context(), limit, offset)
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	pagination := dto.NewPagination(ctx, total, page, limit)
	response := dto.PagedMoviesResponse{
		PageInfo: pagination,
		Result:   movies,
	}
	utils.SendSuccess(ctx, http.StatusOK, "upcoming movies retrieved successfully", response)
}

func (c *MovieController) GetMovieByID(ctx *gin.Context) {
	idParam := ctx.Param("id")
	movieID, err := strconv.Atoi(idParam)
	if err != nil {
		utils.SendError(ctx, http.StatusBadRequest, "invalid movie ID")
		return
	}

	movie, err := c.movieService.GetMovieByID(ctx.Request.Context(), movieID)
	if err != nil {
		utils.SendError(ctx, http.StatusNotFound, err.Error())
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, "Movie retrieved successfully", movie)
}

func (c *MovieController) GetGenres(ctx *gin.Context) {
	genres, err := c.movieService.GetGenres(ctx.Request.Context())
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SendSuccess(ctx, http.StatusOK, "genres retrieved successfully", genres)
}

package dto

import "time"

type CreateMovieRequest struct {
	Title       string    `json:"title" binding:"required"`
	Overview    string    `json:"overview" binding:"required"`
	Duration    int       `json:"duration" binding:"required,min=1"`
	ReleaseDate time.Time `json:"release_date" binding:"required"`
	Director    string    `json:"director_id"`
	GenreIDs    []int     `json:"genres_ids"`
	Cast        []string  `json:"cast"`
}

type UpdateMovieRequest struct {
	Title       *string    `json:"title"`
	Overview    *string    `json:"overview"`
	Duration    *int       `json:"duration"`
	ReleaseDate *time.Time `json:"release_date"`
	Director    *string    `json:"director"`
	GenreIDs    *[]int     `json:"genre_ids"`
	Cast        *[]string  `json:"cast"`
}

type MovieResponse struct {
	MovieID      int       `json:"movie_id"`
	Title        string    `json:"title"`
	PosterPath   *string   `json:"poster_path"`
	BackdropPath *string   `json:"backdrop_path"`
	Overview     string    `json:"overview"`
	Duration     int       `json:"duration"`
	ReleaseDate  time.Time `json:"release_date"`
	Director     string    `json:"director"`
	Genre        []string  `json:"genre"`
	Cast         []string  `json:"cast"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PagedMoviesResponse struct {
	PageInfo Pagination      `json:"page_info"`
	Result   []MovieResponse `json:"movies"`
}
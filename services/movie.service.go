package services

import (
	"context"
	"noir-backend/dto"
	"noir-backend/models"
	"noir-backend/repositories"
	"time"
)

type MovieService interface {
	CreateMovie(ctx context.Context, req dto.CreateMovieRequest, posterPath, backdropPath *string) (*dto.MovieResponse, error)
	UpdateMovie(ctx context.Context, id int, req dto.UpdateMovieRequest, backdropPath, posterPath *string) error
	DeleteMovie(ctx context.Context, id int) error
	GetUpcomingMovies(ctx context.Context, limit, offset int) ([]dto.MovieResponse, int, error)
	GetNowPlayingMovies(ctx context.Context, limit, offset int) ([]dto.MovieResponse, int, error)
	GetMovies(ctx context.Context, limit, offset int) ([]dto.MovieResponse, int, error)
	GetMovieByID(ctx context.Context, movieID int) (*dto.MovieResponse, error)
	GetGenres(ctx context.Context) ([]models.Genre, error)
}

type movieService struct {
	repo repositories.MovieRepository
}

func NewMovieService(repo repositories.MovieRepository) *movieService {
	return &movieService{repo: repo}
}

func (s *movieService) CreateMovie(ctx context.Context, req dto.CreateMovieRequest, posterPath, backdropPath *string) (*dto.MovieResponse, error) {
	movie := &models.Movie{
		Title:        req.Title,
		PosterPath:   posterPath,
		BackdropPath: backdropPath,
		Overview:     req.Overview,
		Duration:     req.Duration,
		ReleaseDate:  req.ReleaseDate,
	}

	joinRow, err := s.repo.CreateMovie(ctx, movie, req.Director, req.GenreIDs, req.Cast)
	if err != nil {
		return nil, err
	}

	res := mapJoinRowToDto(joinRow)
	return &res, nil
}

func (s *movieService) UpdateMovie(ctx context.Context, id int, req dto.UpdateMovieRequest, backdropPath, posterPath *string) error {
	movie := &models.Movie{
		PosterPath:   posterPath,
		BackdropPath: backdropPath,
	}

	if req.Title != nil {
		movie.Title = *req.Title
	}
	if req.Overview != nil {
		movie.Overview = *req.Overview
	}
	if req.Duration != nil {
		movie.Duration = *req.Duration
	}
	if req.ReleaseDate != nil {
		movie.ReleaseDate = *req.ReleaseDate
	}

	return s.repo.UpdateMovie(ctx, id, movie, req.Director, req.GenreIDs, req.Cast)
}

func (s *movieService) DeleteMovie(ctx context.Context, id int) error {
	_, err := s.repo.DeleteMovie(ctx, id)
	return err
}

func (s *movieService) GetUpcomingMovies(ctx context.Context, limit, offset int) ([]dto.MovieResponse, int, error) {
	now := time.Now().Format("2006-01-02")
	return s.fetchMovies(ctx, "WHERE m.release_date > $1", []any{now}, limit, offset, "m.release_date ASC")
}

func (s *movieService) GetNowPlayingMovies(ctx context.Context, limit, offset int) ([]dto.MovieResponse, int, error) {
	now := time.Now().Format("2006-01-02")
	return s.fetchMovies(ctx, "WHERE m.release_date <= $1", []any{now}, limit, offset, "m.release_date DESC")
}

func (s *movieService) GetMovies(ctx context.Context, limit, offset int) ([]dto.MovieResponse, int, error) {
	return s.fetchMovies(ctx, "", nil, limit, offset, "m.created_at DESC")
}

func (s *movieService) GetMovieByID(ctx context.Context, movieID int) (*dto.MovieResponse, error) {
	joinRow, err := s.repo.GetMovieByID(ctx, movieID)
	if err != nil {
		return nil, err
	}

	res := mapJoinRowToDto(joinRow)
	return &res, nil
}

func (s *movieService) GetGenres(ctx context.Context) ([]models.Genre, error) {
	return s.repo.GetGenres(ctx)
}

func (s *movieService) fetchMovies(ctx context.Context, condition string, args []any, limit, offset int, orderBy string) ([]dto.MovieResponse, int, error) {
	rows, total, err := s.repo.GetMovies(ctx, condition, args, limit, offset, orderBy)
	if err != nil {
		return nil, 0, err
	}

	var movies []dto.MovieResponse
	for _, row := range rows {
		movies = append(movies, mapJoinRowToDto(&row))
	}

	return movies, total, nil
}

func mapJoinRowToDto(row *models.MovieJoinRow) dto.MovieResponse {
	res := dto.MovieResponse{
		MovieID:      row.MovieID,
		Title:        row.Title,
		PosterPath:   row.PosterPath,
		BackdropPath: row.BackdropPath,
		Overview:     row.Overview,
		Duration:     row.Duration,
		ReleaseDate:  row.ReleaseDate,
		CreatedAt:    row.CreatedAt,
	}

	if row.Director != nil {
		res.Director = *row.Director
	}
	if row.Genres != nil {
		res.Genre = *row.Genres
	} else {
		res.Genre = []string{}
	}
	if row.Cast != nil {
		res.Cast = *row.Cast
	} else {
		res.Cast = []string{}
	}

	return res
}

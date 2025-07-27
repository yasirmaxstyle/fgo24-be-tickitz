package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"noir-backend/dto"
	"noir-backend/models"
	"noir-backend/utils"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MovieService struct {
	db *pgxpool.Pool
}

func NewMovieService(db *pgxpool.Pool) *MovieService {
	return &MovieService{db: db}
}

func (s *MovieService) CreateMovie(ctx context.Context, req dto.CreateMovieRequest, posterPath, backdropPath *string) (*dto.MovieResponse, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("database transaction error")
	}
	defer tx.Rollback(ctx)

	directorID, err := getOrCreateDirectorID(ctx, tx, req.Director)
	if err != nil {
		return nil, err
	}

	row, err := tx.Query(ctx,
		`INSERT INTO movies (title, poster_path, backdrop_path, overview, duration, release_date, directors_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		 RETURNING movie_id, title, poster_path, backdrop_path, overview, duration, release_date, directors_id, created_at, updated_at`,
		req.Title, posterPath, backdropPath, req.Overview, req.Duration, req.ReleaseDate, directorID)

	if err != nil {
		return nil, fmt.Errorf("failed to create movie: %w", err)
	}

	movie, err := pgx.CollectOneRow[models.Movie](row, pgx.RowToStructByName)
	if err != nil {
		return nil, fmt.Errorf("failed to create movie: %w", err)
	}

	for _, genre := range req.GenreIDs {
		_, err = tx.Exec(ctx,
			"INSERT INTO movies_genres (movie_id, genre_id) VALUES ($1, $2)",
			movie.MovieID, genre)
		if err != nil {
			return nil, fmt.Errorf("failed to add genre: %w", err)

		}
	}

	for _, actor := range req.Cast {
		actorID, err := getOrCreateActorID(ctx, tx, actor)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx,
			"INSERT INTO movies_cast (movie_id, actor_id) VALUES ($1, $2)",
			movie.MovieID, actorID)
		if err != nil {
			return nil, fmt.Errorf("failed to add movies_cast: %w", err)

		}

		firstName, lastName := utils.SplitFullName(actor)
		_, err = tx.Exec(ctx,
			"INSERT INTO actors (first_name, last_name) VALUES ($1, $2)",
			firstName, lastName)
		if err != nil {
			return nil, fmt.Errorf("failed to add actor: %w", err)

		}

	}

	genreNames, err := getGenreNames(tx, ctx, req.GenreIDs)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction")
	}

	result := dto.MovieResponse{
		MovieID:      movie.MovieID,
		Title:        movie.Title,
		PosterPath:   movie.PosterPath,
		BackdropPath: movie.BackdropPath,
		Overview:     movie.Overview,
		Duration:     movie.Duration,
		ReleaseDate:  movie.ReleaseDate,
		Director:     req.Director,
		Genre:        genreNames,
		Cast:         req.Cast,
		CreatedAt:    movie.CreatedAt,
		UpdatedAt:    movie.UpdatedAt,
	}

	return &result, nil
}

func (s *MovieService) UpdateMovie(ctx context.Context, id int, req dto.UpdateMovieRequest, backdropPath, posterPath *string) (int, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("database transaction error: %v", err)
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM movies WHERE movie_id = $1)", id).Scan(&exists)

	if err != nil || !exists {
		return http.StatusNotFound, fmt.Errorf("movie not found")
	}

	setParts := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIndex := 1

	if req.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, *req.Title)
		argIndex++
	}
	if req.Overview != nil {
		setParts = append(setParts, fmt.Sprintf("overview = $%d", argIndex))
		args = append(args, *req.Overview)
		argIndex++
	}
	if req.Duration != nil {
		setParts = append(setParts, fmt.Sprintf("duration = $%d", argIndex))
		args = append(args, *req.Duration)
		argIndex++
	}
	if req.ReleaseDate != nil {
		setParts = append(setParts, fmt.Sprintf("release_date = $%d", argIndex))
		args = append(args, *req.ReleaseDate)
		argIndex++
	}
	if req.Director != nil {
		setParts = append(setParts, fmt.Sprintf("director = $%d", argIndex))
		args = append(args, *req.Director)
		argIndex++
	}
	if posterPath != nil {
		setParts = append(setParts, fmt.Sprintf("poster_path = $%d", argIndex))
		args = append(args, posterPath)
		argIndex++
	}
	if backdropPath != nil {
		setParts = append(setParts, fmt.Sprintf("backdrop_path = $%d", argIndex))
		args = append(args, backdropPath)
		argIndex++
	}

	args = append(args, id)

	setClause := ""
	if len(setParts) > 0 {
		setClause = setParts[0]
		for i := 1; i < len(setParts); i++ {
			setClause += ", " + setParts[i]
		}
	}

	query := fmt.Sprintf("UPDATE movies SET %s WHERE movie_id = $%d", setClause, argIndex)

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update movie: %v", err)
	}

	if req.GenreIDs != nil {
		_, err = tx.Exec(ctx,
			"DELETE FROM movies_genres WHERE movie_id = $1", id)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to update genre")
		}

		for _, genre := range *req.GenreIDs {
			_, err = tx.Exec(ctx,
				"INSERT INTO movies_genres (movie_id, genre) VALUES ($1, $2)",
				id, genre)
			if err != nil {
				return http.StatusInternalServerError, fmt.Errorf("failed to add genre")
			}
		}
	}

	if req.Cast != nil {
		_, err = tx.Exec(ctx,
			"DELETE FROM movies_cast WHERE movie_id = $1", id)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to update genre")
		}

		for _, cast := range *req.Cast {
			_, err = tx.Exec(ctx,
				"INSERT INTO movies_cast (movie_id, cast) VALUES ($1, $2)",
				id, cast)
			if err != nil {
				return http.StatusInternalServerError, fmt.Errorf("failed to add genre")
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to commit transaction")
	}

	return http.StatusOK, nil
}

func (s *MovieService) DeleteMovie(ctx context.Context, id int) (int, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("database transaction error")
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		"DELETE FROM movies_genres WHERE movie_id = $1", id)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to delete movie genre")
	}

	result, err := tx.Exec(ctx,
		"DELETE FROM movies WHERE movie_id = $1", id)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to delete movie")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return http.StatusNotFound, fmt.Errorf("movie not found")
	}

	if err = tx.Commit(ctx); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to commit transaction")
	}

	return http.StatusOK, nil
}

func (s *MovieService) GetUpcomingMovies(ctx context.Context, limit, offset int) ([]dto.MovieResponse, int, error) {
	now := time.Now().Format("2006-01-02")
	return s.getMoviesByCondition(ctx, "WHERE m.release_date > $1", []any{now}, limit, offset, "m.release_date ASC")
}

func (s *MovieService) GetNowPlayingMovies(ctx context.Context, limit, offset int) ([]dto.MovieResponse, int, error) {
	now := time.Now().Format("2006-01-02")
	return s.getMoviesByCondition(ctx, "WHERE m.release_date <= $1", []any{now}, limit, offset, "m.release_date DESC")
}

func (s *MovieService) GetMovies(ctx context.Context, limit, offset int) ([]dto.MovieResponse, int, error) {
	return s.getMoviesByCondition(ctx, "", nil, limit, offset, "m.created_at DESC")
}

func (s *MovieService) GetMovieByID(ctx context.Context, movieID int) (*dto.MovieResponse, error) {
	movies, _, err := s.getMoviesByCondition(
		ctx,
		"WHERE m.movie_id = $1",
		[]any{movieID},
		1, 0,
		"m.created_at DESC",
	)

	if err != nil {
		return nil, err
	}

	if len(movies) == 0 {
		return nil, fmt.Errorf("movie not found")
	}

	return &movies[0], nil
}

func (s *MovieService) getMoviesByCondition(
	ctx context.Context,
	condition string,
	args []any,
	limit, offset int,
	orderBy string,
) ([]dto.MovieResponse, int, error) {
	query := fmt.Sprintf(`
		SELECT
			m.movie_id,
			m.title,
			m.poster_path,
			m.backdrop_path,
			m.overview,
			m.duration,
			m.release_date,
			m.created_at,
			m.updated_at,
			d.first_name || ' ' || d.last_name AS director,
			ARRAY_AGG(DISTINCT g.name) AS genres,
			ARRAY_AGG(DISTINCT a.first_name || ' ' || a.last_name) AS cast
		FROM movies m
		LEFT JOIN directors d ON d.id = m.director_id
		LEFT JOIN movie_genres mg ON mg.movie_id = m.movie_id
		LEFT JOIN genres g ON g.id = mg.genre_id
		LEFT JOIN movie_casts mc ON mc.movie_id = m.movie_id
		LEFT JOIN actors a ON a.id = mc.actor_id
		%s
		GROUP BY
			m.movie_id, m.title, m.poster_path, m.backdrop_path, m.overview,
			m.duration, m.release_date, m.created_at, m.updated_at,
			d.first_name, d.last_name
		ORDER BY %s
		LIMIT $%d OFFSET $%d`,
		condition, orderBy, len(args)+1, len(args)+2)

	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}

	flatRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.MovieJoinRow])
	if err != nil {
		return nil, 0, err
	}

	var movies []dto.MovieResponse
	for _, row := range flatRows {
		movies = append(movies, dto.MovieResponse{
			MovieID:      row.MovieID,
			Title:        row.Title,
			PosterPath:   row.PosterPath,
			BackdropPath: row.BackdropPath,
			Overview:     row.Overview,
			Duration:     row.Duration,
			ReleaseDate:  row.ReleaseDate,
			Director:     *row.Director,
			Genre:        *row.Genres,
			Cast:         *row.Cast,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		})
	}

	countQuery := "SELECT COUNT(*) FROM movies"
	if condition != "" {
		countQuery += " " + condition
	}
	var total int
	err = s.db.QueryRow(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return movies, total, nil
}

func (s *MovieService) GetGenres(ctx context.Context) (*[]models.Genre, error) {
	rows, err := s.db.Query(ctx, "SELECT id, name FROM genres")
	if err != nil {
		return nil, err
	}

	genres, err := pgx.CollectRows[models.Genre](rows, pgx.RowToStructByName)
	if err != nil {
		return nil, err
	}

	return &genres, nil
}

func (s *MovieService) ParseCreateMovieRequest(form map[string][]string) (*dto.CreateMovieRequest, error) {
	var req dto.CreateMovieRequest

	if title := utils.GetStringField(form, "title"); title != nil {
		req.Title = *title
	}
	if overview := utils.GetStringField(form, "overview"); overview != nil {
		req.Overview = *overview
	}
	if i, err := utils.GetIntField(form, "duration"); err != nil {
		return nil, err
	} else if i != nil {
		req.Duration = *i
	}
	if t, err := utils.GetDateField(form, "release_date"); err != nil {
		return nil, err
	} else if t != nil {
		req.ReleaseDate = *t
	}
	if director := utils.GetStringField(form, "director"); director != nil {
		req.Director = *director
	}
	if ids, err := utils.GetIntArray(form, "genre_ids"); err != nil {
		return nil, err
	} else {
		req.GenreIDs = *ids
	}
	if cast, ok := form["cast"]; ok {
		req.Cast = append(req.Cast, cast...)
	}

	return &req, nil
}

func (s *MovieService) ParseUpdateMovieRequest(form map[string][]string) (*dto.UpdateMovieRequest, error) {
	var req dto.UpdateMovieRequest

	req.Title = utils.GetStringField(form, "title")
	req.Overview = utils.GetStringField(form, "overview")

	if i, err := utils.GetIntField(form, "duration"); err != nil {
		return nil, err
	} else {
		req.Duration = i
	}

	if t, err := utils.GetDateField(form, "release_date"); err != nil {
		return nil, err
	} else {
		req.ReleaseDate = t
	}

	if director := utils.GetStringField(form, "director"); director != nil {
		req.Director = director
	}

	if ids, err := utils.GetIntArray(form, "genre_ids"); err != nil {
		return nil, err
	} else if len(*ids) > 0 {
		req.GenreIDs = ids
	}

	if cast, ok := form["cast"]; ok && len(cast) > 0 {
		req.Cast = &cast
	}

	return &req, nil
}

func getOrCreateActorID(ctx context.Context, db pgx.Tx, fullName string) (int, error) {
	firstName, lastName := utils.SplitFullName(fullName)
	var actorID int

	query := `
		SELECT id
		FROM actors
		WHERE first_name = $1 AND last_name = $2
	`
	err := db.QueryRow(ctx, query, firstName, lastName).Scan(&actorID)
	if err == nil {
		return actorID, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	insertQuery := `
		INSERT INTO actor (first_name, last_name)
		VALUES ($1, $2)
		RETURNING id
	`
	err = db.QueryRow(ctx, insertQuery, firstName, lastName).Scan(&actorID)
	if err != nil {
		return 0, err
	}

	return actorID, nil
}

func getOrCreateDirectorID(ctx context.Context, db pgx.Tx, fullName string) (int, error) {
	firstName, lastName := utils.SplitFullName(fullName)
	var directorID int

	query := `
		SELECT id
		FROM directors
		WHERE first_name = $1 AND last_name = $2
	`
	err := db.QueryRow(ctx, query, firstName, lastName).Scan(&directorID)
	if err == nil {
		return directorID, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	insertQuery := `
		INSERT INTO directors (first_name, last_name)
		VALUES ($1, $2)
		RETURNING id
	`
	err = db.QueryRow(ctx, insertQuery, firstName, lastName).Scan(&directorID)
	if err != nil {
		return 0, err
	}

	return directorID, nil
}

func getGenreNames(tx pgx.Tx, ctx context.Context, genreIDs []int) ([]string, error) {
	rows, err := tx.Query(ctx, "SELECT id, name FROM genres WHERE id = ANY($1)", genreIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch genre names: %w", err)
	}

	genres, err := pgx.CollectRows[models.Genre](rows, pgx.RowToStructByName)
	if err != nil {
		return nil, fmt.Errorf("failed to collect genre names: %w", err)
	}

	genreNames := make([]string, len(genres))
	for i, g := range genres {
		genreNames[i] = g.Name
	}

	return genreNames, nil
}

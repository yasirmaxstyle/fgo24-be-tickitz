package repositories

import (
	"context"
	"errors"
	"fmt"
	"noir-backend/models"
	"noir-backend/utils"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MovieRepository interface {
	CreateMovie(ctx context.Context, movie *models.Movie, directorName string, genreIDs []int, castNames []string) (*models.MovieJoinRow, error)
	UpdateMovie(ctx context.Context, id int, movie *models.Movie, directorName *string, genreIDs *[]int, castNames *[]string) error
	DeleteMovie(ctx context.Context, id int) (int, error)
	GetMovieByID(ctx context.Context, id int) (*models.MovieJoinRow, error)
	GetMovies(ctx context.Context, condition string, args []any, limit, offset int, orderBy string) ([]models.MovieJoinRow, int, error)
	GetGenres(ctx context.Context) ([]models.Genre, error)
}

type movieRepository struct {
	db *pgxpool.Pool
}

func NewMovieRepository(db *pgxpool.Pool) MovieRepository {
	return &movieRepository{db: db}
}

func (r *movieRepository) CreateMovie(ctx context.Context, movie *models.Movie, directorName string, genreIDs []int, castNames []string) (*models.MovieJoinRow, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("database transaction error: %w", err)
	}
	defer tx.Rollback(ctx)

	directorID, err := getOrCreateDirectorID(ctx, tx, directorName)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create director: %w", err)
	}

	row := tx.QueryRow(ctx,
		`INSERT INTO movies (title, poster_path, backdrop_path, overview, duration, release_date, director_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING movie_id, created_at`,
		movie.Title, movie.PosterPath, movie.BackdropPath, movie.Overview, movie.Duration, movie.ReleaseDate, directorID)

	var movieID int
	var createdAt time.Time
	if err := row.Scan(&movieID, &createdAt); err != nil {
		return nil, fmt.Errorf("failed to create movie: %w", err)
	}

	for _, genreID := range genreIDs {
		_, err = tx.Exec(ctx,
			"INSERT INTO movies_genres (movie_id, genre_id) VALUES ($1, $2)",
			movieID, genreID)
		if err != nil {
			return nil, fmt.Errorf("failed to add genre: %w", err)
		}
	}

	for _, actorName := range castNames {
		actorID, err := getOrCreateActorID(ctx, tx, actorName)
		if err != nil {
			return nil, fmt.Errorf("failed to get/create actor: %w", err)
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO movies_cast (movie_id, actor_id, role) VALUES ($1, $2, $3)`,
			movieID, actorID, "")
		if err != nil {
			return nil, fmt.Errorf("failed to add movies_cast: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetMovieByID(ctx, movieID)
}

func (r *movieRepository) UpdateMovie(ctx context.Context, id int, movie *models.Movie, directorName *string, genreIDs *[]int, castNames *[]string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("database transaction error: %w", err)
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM movies WHERE movie_id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		return fmt.Errorf("movie not found")
	}

	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if movie.Title != "" {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, movie.Title)
		argIndex++
	}
	if movie.Overview != "" {
		setParts = append(setParts, fmt.Sprintf("overview = $%d", argIndex))
		args = append(args, movie.Overview)
		argIndex++
	}
	if movie.Duration != 0 {
		setParts = append(setParts, fmt.Sprintf("duration = $%d", argIndex))
		args = append(args, movie.Duration)
		argIndex++
	}
	if !movie.ReleaseDate.IsZero() {
		setParts = append(setParts, fmt.Sprintf("release_date = $%d", argIndex))
		args = append(args, movie.ReleaseDate)
		argIndex++
	}
	if movie.PosterPath != nil {
		setParts = append(setParts, fmt.Sprintf("poster_path = $%d", argIndex))
		args = append(args, movie.PosterPath)
		argIndex++
	}
	if movie.BackdropPath != nil {
		setParts = append(setParts, fmt.Sprintf("backdrop_path = $%d", argIndex))
		args = append(args, movie.BackdropPath)
		argIndex++
	}

	if directorName != nil {
		directorID, err := getOrCreateDirectorID(ctx, tx, *directorName)
		if err != nil {
			return fmt.Errorf("failed to get/create director: %w", err)
		}
		setParts = append(setParts, fmt.Sprintf("director_id = $%d", argIndex))
		args = append(args, directorID)
		argIndex++
	}

	args = append(args, id)
	if len(setParts) > 0 {
		setClause := setParts[0]
		for i := 1; i < len(setParts); i++ {
			setClause += ", " + setParts[i]
		}
		query := fmt.Sprintf("UPDATE movies SET %s WHERE movie_id = $%d", setClause, argIndex)
		if _, err = tx.Exec(ctx, query, args...); err != nil {
			return fmt.Errorf("failed to update movie: %w", err)
		}
	}

	if genreIDs != nil {
		if _, err = tx.Exec(ctx, "DELETE FROM movies_genres WHERE movie_id = $1", id); err != nil {
			return fmt.Errorf("failed to delete related genres: %w", err)
		}

		for _, genreID := range *genreIDs {
			if _, err = tx.Exec(ctx, "INSERT INTO movies_genres (movie_id, genre_id) VALUES ($1, $2)", id, genreID); err != nil {
				return fmt.Errorf("failed to reconstruct genre relationship: %w", err)
			}
		}
	}

	if castNames != nil {
		if _, err = tx.Exec(ctx, "DELETE FROM movies_cast WHERE movie_id = $1", id); err != nil {
			return fmt.Errorf("failed to delete related movie cast: %w", err)
		}

		for _, actor := range *castNames {
			actorID, err := getOrCreateActorID(ctx, tx, actor)
			if err != nil {
				return fmt.Errorf("failed to get/create actor: %w", err)
			}
			if _, err = tx.Exec(ctx, "INSERT INTO movies_cast (movie_id, actor_id) VALUES ($1, $2)", id, actorID); err != nil {
				return fmt.Errorf("failed to reconstruct cast relationship: %w", err)
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *movieRepository) DeleteMovie(ctx context.Context, id int) (int, error) {
	// movies_genres and movies_cast have ON DELETE CASCADE based on the migration! So we just delete the movie.
	result, err := r.db.Exec(ctx, "DELETE FROM movies WHERE movie_id = $1", id)
	if err != nil {
		return 0, fmt.Errorf("failed to delete movie: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return 0, fmt.Errorf("movie not found")
	}

	return int(rowsAffected), nil
}

func (r *movieRepository) GetMovies(ctx context.Context, condition string, args []any, limit, offset int, orderBy string) ([]models.MovieJoinRow, int, error) {
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
			d.first_name || ' ' || d.last_name AS director,
			ARRAY_REMOVE(ARRAY_AGG(DISTINCT g.name), NULL) AS genres,
			ARRAY_REMOVE(ARRAY_AGG(DISTINCT a.first_name || ' ' || a.last_name), NULL) AS cast
		FROM movies m
		LEFT JOIN directors d ON d.director_id = m.director_id
		LEFT JOIN movies_genres mg ON mg.movie_id = m.movie_id
		LEFT JOIN genres g ON g.genre_id = mg.genre_id
		LEFT JOIN movies_cast mc ON mc.movie_id = m.movie_id
		LEFT JOIN actors a ON a.actor_id = mc.actor_id
		%s
		GROUP BY
			m.movie_id, m.title, m.poster_path, m.backdrop_path, m.overview,
			m.duration, m.release_date, m.created_at,
			d.first_name, d.last_name
		ORDER BY %s
		LIMIT $%d OFFSET $%d`,
		condition, orderBy, len(args)+1, len(args)+2)

	queryArgs := append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}

	flatRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.MovieJoinRow])
	if err != nil {
		return nil, 0, err
	}

	countQuery := "SELECT COUNT(*) FROM movies m " + condition
	var total int
	err = r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return flatRows, total, nil
}

func (r *movieRepository) GetMovieByID(ctx context.Context, id int) (*models.MovieJoinRow, error) {
	movies, _, err := r.GetMovies(ctx, "WHERE m.movie_id = $1", []any{id}, 1, 0, "m.created_at DESC")
	if err != nil {
		return nil, err
	}

	if len(movies) == 0 {
		return nil, fmt.Errorf("movie not found")
	}

	return &movies[0], nil
}

func (r *movieRepository) GetGenres(ctx context.Context) ([]models.Genre, error) {
	rows, err := r.db.Query(ctx, "SELECT genre_id, name FROM genres")
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows[models.Genre](rows, pgx.RowToStructByName)
}

// Helpers

func getOrCreateActorID(ctx context.Context, tx pgx.Tx, fullName string) (int, error) {
	firstName, lastName := utils.SplitFullName(fullName)
	var actorID int

	query := `SELECT actor_id FROM actors WHERE first_name = $1 AND last_name = $2`
	err := tx.QueryRow(ctx, query, firstName, lastName).Scan(&actorID)
	if err == nil {
		return actorID, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	insertQuery := `INSERT INTO actors (first_name, last_name) VALUES ($1, $2) RETURNING actor_id`
	err = tx.QueryRow(ctx, insertQuery, firstName, lastName).Scan(&actorID)
	if err != nil {
		return 0, err
	}

	return actorID, nil
}

func getOrCreateDirectorID(ctx context.Context, tx pgx.Tx, fullName string) (int, error) {
	firstName, lastName := utils.SplitFullName(fullName)
	var directorID int

	query := `SELECT director_id FROM directors WHERE first_name = $1 AND last_name = $2`
	err := tx.QueryRow(ctx, query, firstName, lastName).Scan(&directorID)
	if err == nil {
		return directorID, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	insertQuery := `INSERT INTO directors (first_name, last_name) VALUES ($1, $2) RETURNING director_id`
	err = tx.QueryRow(ctx, insertQuery, firstName, lastName).Scan(&directorID)
	if err != nil {
		return 0, err
	}

	return directorID, nil
}

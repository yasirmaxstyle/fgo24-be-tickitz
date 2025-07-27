package seeder

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"noir-backend/utils"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type MovieListResponse struct {
	Results []struct {
		ID int `json:"id"`
	} `json:"results"`
}

type MovieDetails struct {
	Title        string `json:"title"`
	Overview     string `json:"overview"`
	ReleaseDate  string `json:"release_date"`
	Runtime      int    `json:"runtime"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
	Genres       []struct {
		Name string `json:"name"`
	} `json:"genres"`
}

type Credits struct {
	Crew []struct {
		Name string `json:"name"`
		Job  string `json:"job"`
	} `json:"crew"`
	Cast []struct {
		Name string `json:"name"`
		Role string `json:"character"`
	} `json:"cast"`
}

func fetchJSON(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

func SeedTMDBMovies(db *pgxpool.Pool) error {
	if err := seedFromTMDBEndpoint(db, "now_playing"); err != nil {
		return err
	}
	if err := seedFromTMDBEndpoint(db, "upcoming"); err != nil {
		return err
	}
	return nil
}

func seedFromTMDBEndpoint(db *pgxpool.Pool, category string) error {
	godotenv.Load()
	API_KEY := os.Getenv("TMDB_API_KEY")

	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/%s?api_key=%s&language=en-US&page=1", category, API_KEY)

	var result MovieListResponse
	if err := fetchJSON(url, &result); err != nil {
		return fmt.Errorf("failed to fetch %s: %v", category, err)
	}

	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	for _, m := range result.Results {
		var detail MovieDetails
		var credits Credits

		detailURL := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d?api_key=%s", m.ID, API_KEY)
		creditURL := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d/credits?api_key=%s", m.ID, API_KEY)

		if err := fetchJSON(detailURL, &detail); err != nil {
			log.Printf("failed to fetch movie details for ID %d: %v", m.ID, err)
			continue
		}

		if err := fetchJSON(creditURL, &credits); err != nil {
			log.Printf("failed to fetch movie credits for ID %d: %v", m.ID, err)
			continue
		}

		var directorID int
		for _, crew := range credits.Crew {
			if crew.Job == "Director" {
				firstName, lastName := utils.SplitFullName(crew.Name)
				tx.QueryRow(context.Background(), `
						INSERT INTO directors (first_name, last_name) VALUES ($1, $2)
						ON CONFLICT (first_name, last_name) DO NOTHING
						RETURNING id
					`, firstName, lastName).Scan(&directorID)
				break
			}
		}

		releaseDate, err := time.Parse("2006-01-02", detail.ReleaseDate)
		if err != nil {
			continue
		}

		var movieID int
		err = tx.QueryRow(context.Background(), `
			INSERT INTO movies (title, overview, duration, release_date, directors_id, poster_path, backdrop_path, created_at, updated_at)
		 	VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
			RETURNING id
		`, detail.Title, detail.Overview, detail.Runtime, releaseDate, directorID,
			"https://image.tmdb.org/t/p/w500"+detail.PosterPath,
			"https://image.tmdb.org/t/p/original"+detail.BackdropPath,
		).Scan(&movieID)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Insert failed for %s: %v\n", detail.Title, err)
		}

		for _, g := range detail.Genres {
			var genreID int
			tx.QueryRow(context.Background(), `
					INSERT INTO genres (genre_name) VALUES ($1)
					ON CONFLICT (genre_name) DO NOTHING
					RETURNING id
				`, g.Name).Scan(&genreID)

			tx.Exec(context.Background(), `INSERT INTO movie_genres (movie_id, genre_id) VALUES ($1, $2)`, movieID, genreID)
		}
		for i, cast := range credits.Cast {
			if i >= 5 {
				break
			}

			firstName, lastName := utils.SplitFullName(cast.Name)

			var actorID int
			tx.QueryRow(context.Background(), `
					INSERT INTO actors (first_name, last_name) VALUES ($1, $2)
					ON CONFLICT (first_name, last_name) DO NOTHING
					RETURNING id
				`, firstName, lastName).Scan(&actorID)

			tx.Exec(context.Background(), `INSERT INTO movie_casts (movie_id, actor_id, role) VALUES ($1, $2)`, movieID, actorID, cast.Role)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

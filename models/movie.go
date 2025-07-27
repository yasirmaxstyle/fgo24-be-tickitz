package models

import (
	"time"
)

type Movie struct {
	MovieID      int       `json:"movie_id" db:"movie_id"`
	Title        string    `json:"title" db:"title"`
	PosterPath   *string   `json:"poster_path" db:"poster_path"`
	BackdropPath *string   `json:"backdrop_path" db:"backdrop_path"`
	Overview     string    `json:"overview" db:"overview"`
	Duration     int       `json:"duration" db:"duration"`
	ReleaseDate  time.Time `json:"release_date" db:"release_date"`
	DirectorID   *int      `json:"director_id" db:"director_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type MovieCast struct {
	ID      int    `json:"id" db:"id"`
	MovieID int    `json:"movie_id" db:"movie_id"`
	ActorID int    `json:"actor_id" db:"actor_id"`
	Role    string `json:"role" db:"role"`
}

type Actor struct {
	ID        int    `json:"id" db:"id"`
	FirstName string `json:"first_name" db:"first_name"`
	LastName  string `json:"last_name" db:"last_name"`
}

type Director struct {
	DirectorID int    `json:"director_id" db:"id"`
	FirstName  string `json:"first_name" db:"first_name"`
	LastName   string `json:"last_name" db:"last_name"`
}

type MovieGenre struct {
	ID      int    `json:"id" db:"id"`
	MovieID int    `json:"movie_id" db:"movie_id"`
	Genre   string `json:"genre" db:"genre"`
}

type Genre struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type MovieJoinRow struct {
	MovieID      int       `db:"movie_id"`
	Title        string    `db:"title"`
	PosterPath   *string   `db:"poster_path"`
	BackdropPath *string   `db:"backdrop_path"`
	Overview     string    `db:"overview"`
	Duration     int       `db:"duration"`
	ReleaseDate  time.Time `db:"release_date"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	Director     *string   `db:"director"`
	Genres       *[]string `db:"genres"`
	Cast         *[]string `db:"cast"`
}

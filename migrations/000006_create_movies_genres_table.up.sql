CREATE TABLE movies_genres (
    movie_genre_id SERIAL PRIMARY KEY,
    movie_id INTEGER NOT NULL REFERENCES movies (movie_id) ON DELETE CASCADE,
    genre_id INTEGER NOT NULL REFERENCES genres (genre_id) ON DELETE CASCADE,
    UNIQUE (movie_id, genre_id)
);
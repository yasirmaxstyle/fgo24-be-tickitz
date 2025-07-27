CREATE TABLE movies_genres (
    id SERIAL PRIMARY KEY,
    movie_id INTEGER NOT NULL REFERENCES movies (id) ON DELETE CASCADE,
    genre_id INTEGER NOT NULL REFERENCES genres (id) ON DELETE CASCADE,
    UNIQUE (movie_id, genre_id)
);
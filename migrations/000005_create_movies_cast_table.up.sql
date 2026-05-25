CREATE TABLE movies_cast (
    movie_cast_id SERIAL PRIMARY KEY,
    movie_id INTEGER NOT NULL REFERENCES movies (movie_id) ON DELETE CASCADE,
    actor_id INTEGER NOT NULL REFERENCES actors (actor_id) ON DELETE CASCADE,
    role VARCHAR(255)
);
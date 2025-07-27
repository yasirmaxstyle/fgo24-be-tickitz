CREATE TABLE movies (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    poster_path VARCHAR(500),
    backdrop_path VARCHAR(500),
    overview TEXT,
    duration INTEGER NOT NULL,
    release_date DATE NOT NULL,
    director_id INTEGER REFERENCES directors (id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
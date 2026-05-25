CREATE TABLE showtimes (
    showtime_id SERIAL PRIMARY KEY,
    movie_id INTEGER NOT NULL REFERENCES movies (movie_id) ON DELETE CASCADE,
    screen_id INTEGER NOT NULL REFERENCES screens (screen_id) ON DELETE CASCADE,
    show_datetime TIMESTAMP NOT NULL,
    base_price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
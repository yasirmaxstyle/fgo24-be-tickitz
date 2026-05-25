CREATE TABLE tickets (
    ticket_id SERIAL PRIMARY KEY,
    ticket_code VARCHAR(50) UNIQUE NOT NULL,
    showtime_id INTEGER NOT NULL REFERENCES showtimes (showtime_id) ON DELETE CASCADE,
    seat_number VARCHAR(10) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    status VARCHAR(20) DEFAULT 'booked' CHECK (
        status IN ('booked', 'used', 'cancelled')
    ),
    transaction_id INTEGER NOT NULL REFERENCES transactions (transaction_id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (showtime_id, seat_number)
);
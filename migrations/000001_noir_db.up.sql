CREATE TABLE directors (
    director_id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL
);

CREATE TABLE actors (
    actor_id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL
);

CREATE TABLE genres (
    genre_id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL
);

CREATE TABLE movies (
    movie_id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    poster_path VARCHAR(500),
    backdrop_path VARCHAR(500),
    overview TEXT,
    duration INTEGER NOT NULL,
    release_date DATE NOT NULL,
    director_id INTEGER REFERENCES directors (director_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE movies_cast (
    movie_cast_id SERIAL PRIMARY KEY,
    movie_id INTEGER NOT NULL REFERENCES movies (movie_id) ON DELETE CASCADE,
    actor_id INTEGER NOT NULL REFERENCES actors (actor_id) ON DELETE CASCADE,
    role VARCHAR(255)
);

CREATE TABLE movies_genres (
    movie_genre_id SERIAL PRIMARY KEY,
    movie_id INTEGER NOT NULL REFERENCES movies (movie_id) ON DELETE CASCADE,
    genre_id INTEGER NOT NULL REFERENCES genres (genre_id) ON DELETE CASCADE,
    UNIQUE (movie_id, genre_id)
);

CREATE TABLE cinemas (
    cinema_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    image_path VARCHAR(255),
    location VARCHAR(255) NOT NULL,
    total_seats INTEGER NOT NULL,
    address TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE showtimes (
    showtime_id SERIAL PRIMARY KEY,
    movie_id INTEGER NOT NULL REFERENCES movies (movie_id) ON DELETE CASCADE,
    cinema_id INTEGER NOT NULL REFERENCES cinemas (cinema_id) ON DELETE CASCADE,
    show_datetime TIMESTAMP NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    available_seats INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE payment_method (
    payment_method_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE profile (
    profile_id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20),
    avatar VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP,
    profile_id INTEGER UNIQUE REFERENCES profile (profile_id) ON DELETE SET NULL
);

CREATE TABLE transactions (
    transaction_id SERIAL PRIMARY KEY,
    transaction_code VARCHAR(50) UNIQUE NOT NULL,
    recipient_email VARCHAR(255) NOT NULL,
    recipient_full_name VARCHAR(255) NOT NULL,
    recipient_phone_number VARCHAR(20) NOT NULL,
    total_seats INTEGER NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (
        status IN (
            'pending',
            'paid',
            'cancelled'
        )
    ),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    paid_at TIMESTAMP,
    created_by INTEGER NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
    payment_method_id INTEGER REFERENCES payment_method (payment_method_id)
);

CREATE TABLE tickets (
    ticket_id SERIAL PRIMARY KEY,
    ticket_code VARCHAR(50) UNIQUE NOT NULL,
    showtime_id INTEGER NOT NULL REFERENCES showtimes (showtime_id) ON DELETE CASCADE,
    seat_number VARCHAR(10) NOT NULL,
    status VARCHAR(20) DEFAULT 'booked' CHECK (
        status IN ('booked', 'used', 'cancelled')
    ),
    transaction_id INTEGER NOT NULL REFERENCES transactions (transaction_id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (showtime_id, seat_number)
);

CREATE INDEX idx_movies_director_id ON movies (director_id);

CREATE INDEX idx_movies_cast_movie_id ON movies_cast (movie_id);

CREATE INDEX idx_movies_cast_actor_id ON movies_cast (actor_id);

CREATE INDEX idx_movies_genres_movie_id ON movies_genres (movie_id);

CREATE INDEX idx_movies_genres_genre_id ON movies_genres (genre_id);

CREATE INDEX idx_showtimes_movie_id ON showtimes (movie_id);

CREATE INDEX idx_showtimes_cinema_id ON showtimes (cinema_id);

CREATE INDEX idx_showtimes_datetime ON showtimes (show_datetime);

CREATE INDEX idx_transactions_created_by ON transactions (created_by);

CREATE INDEX idx_transactions_status ON transactions (status);

CREATE INDEX idx_tickets_showtime_id ON tickets (showtime_id);

CREATE INDEX idx_tickets_transaction_id ON tickets (transaction_id);

CREATE INDEX idx_users_email ON users (email);

INSERT INTO
    payment_method (name, code)
VALUES ('Google Pay', 'EWALLET'),
    ('Visa', 'CREDIT_CARD'),
    ('Gopay', 'EWALLET'),
    ('Ovo', 'EWALLET'),
    ('Dana', 'EWALLET'),
    ('Paypal', 'EWALLET'),
    ('BRI', 'BANK_TRANSFER'),
    ('BCA', 'BANK_TRANSFER');

INSERT INTO
    genres (name)
VALUES ('Action'),
    ('Comedy'),
    ('Crime'),
    ('Drama'),
    ('Horror'),
    ('Romance'),
    ('Sci-Fi'),
    ('Thriller'),
    ('Adventure'),
    ('Animation'),
    ('Documentary'),
    ('Fantasy'),
    ('Mystery'),
    ('Family'),
    ('History'),
    ('Music'),
    ('Thriller'),
    ('War');

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_profile_updated_at BEFORE UPDATE ON profile
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
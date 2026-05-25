const fs = require('fs');
const path = require('path');

const migrations = [
  {
    name: '000001_create_directors_table',
    up: `CREATE TABLE directors (
    director_id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL
);`,
    down: `DROP TABLE IF EXISTS directors CASCADE;`
  },
  {
    name: '000002_create_actors_table',
    up: `CREATE TABLE actors (
    actor_id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL
);`,
    down: `DROP TABLE IF EXISTS actors CASCADE;`
  },
  {
    name: '000003_create_genres_table',
    up: `CREATE TABLE genres (
    genre_id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL
);`,
    down: `DROP TABLE IF EXISTS genres CASCADE;`
  },
  {
    name: '000004_create_movies_table',
    up: `CREATE TABLE movies (
    movie_id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    poster_path VARCHAR(500),
    backdrop_path VARCHAR(500),
    overview TEXT,
    duration INTEGER NOT NULL,
    release_date DATE NOT NULL,
    director_id INTEGER REFERENCES directors (director_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`,
    down: `DROP TABLE IF EXISTS movies CASCADE;`
  },
  {
    name: '000005_create_movies_cast_table',
    up: `CREATE TABLE movies_cast (
    movie_cast_id SERIAL PRIMARY KEY,
    movie_id INTEGER NOT NULL REFERENCES movies (movie_id) ON DELETE CASCADE,
    actor_id INTEGER NOT NULL REFERENCES actors (actor_id) ON DELETE CASCADE,
    role VARCHAR(255)
);`,
    down: `DROP TABLE IF EXISTS movies_cast CASCADE;`
  },
  {
    name: '000006_create_movies_genres_table',
    up: `CREATE TABLE movies_genres (
    movie_genre_id SERIAL PRIMARY KEY,
    movie_id INTEGER NOT NULL REFERENCES movies (movie_id) ON DELETE CASCADE,
    genre_id INTEGER NOT NULL REFERENCES genres (genre_id) ON DELETE CASCADE,
    UNIQUE (movie_id, genre_id)
);`,
    down: `DROP TABLE IF EXISTS movies_genres CASCADE;`
  },
  {
    name: '000007_create_cinemas_table',
    up: `CREATE TABLE cinemas (
    cinema_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    image_path VARCHAR(255),
    location VARCHAR(255) NOT NULL,
    total_seats INTEGER NOT NULL,
    address TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`,
    down: `DROP TABLE IF EXISTS cinemas CASCADE;`
  },
  {
    name: '000008_create_screens_table',
    up: `CREATE TABLE screens (
    screen_id SERIAL PRIMARY KEY,
    cinema_id INTEGER NOT NULL REFERENCES cinemas (cinema_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    total_seats INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`,
    down: `DROP TABLE IF EXISTS screens CASCADE;`
  },
  {
    name: '000009_create_showtimes_table',
    up: `CREATE TABLE showtimes (
    showtime_id SERIAL PRIMARY KEY,
    movie_id INTEGER NOT NULL REFERENCES movies (movie_id) ON DELETE CASCADE,
    screen_id INTEGER NOT NULL REFERENCES screens (screen_id) ON DELETE CASCADE,
    show_datetime TIMESTAMP NOT NULL,
    base_price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`,
    down: `DROP TABLE IF EXISTS showtimes CASCADE;`
  },
  {
    name: '000010_create_payment_method_table',
    up: `CREATE TABLE payment_method (
    payment_method_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`,
    down: `DROP TABLE IF EXISTS payment_method CASCADE;`
  },
  {
    name: '000011_create_users_table',
    up: `CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP
);`,
    down: `DROP TABLE IF EXISTS users CASCADE;`
  },
  {
    name: '000012_create_profile_table',
    up: `CREATE TABLE profile (
    profile_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20),
    avatar VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`,
    down: `DROP TABLE IF EXISTS profile CASCADE;`
  },
  {
    name: '000013_create_transactions_table',
    up: `CREATE TABLE transactions (
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
);`,
    down: `DROP TABLE IF EXISTS transactions CASCADE;`
  },
  {
    name: '000014_create_tickets_table',
    up: `CREATE TABLE tickets (
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
);`,
    down: `DROP TABLE IF EXISTS tickets CASCADE;`
  },
  {
    name: '000015_insert_payment_method',
    up: `INSERT INTO
    payment_method (name, code)
VALUES ('Google Pay', 'EWALLET'),
    ('Visa', 'CREDIT_CARD'),
    ('Gopay', 'EWALLET'),
    ('Ovo', 'EWALLET'),
    ('Dana', 'EWALLET'),
    ('Paypal', 'EWALLET'),
    ('BRI', 'BANK_TRANSFER'),
    ('BCA', 'BANK_TRANSFER');`,
    down: `TRUNCATE TABLE payment_method RESTART IDENTITY CASCADE;`
  },
  {
    name: '000016_insert_genres',
    up: `INSERT INTO
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
    ('War');`,
    down: `TRUNCATE TABLE genres RESTART IDENTITY CASCADE;`
  }
];

const migrationsDir = path.join(__dirname, 'migrations');
if (!fs.existsSync(migrationsDir)) {
  fs.mkdirSync(migrationsDir);
}

for (const migration of migrations) {
  fs.writeFileSync(path.join(migrationsDir, migration.name + '.up.sql'), migration.up);
  fs.writeFileSync(path.join(migrationsDir, migration.name + '.down.sql'), migration.down);
}

console.log('Migrations generated successfully.');

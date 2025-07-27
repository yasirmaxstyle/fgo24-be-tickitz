CREATE TABLE cinemas (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    image_path VARCHAR(255),
    location VARCHAR(255) NOT NULL,
    total_seats INTEGER NOT NULL,
    address TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
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
    created_by INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    payment_method_id INTEGER REFERENCES payment_method (id)
);
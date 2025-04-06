CREATE TABLE IF NOT EXISTS concerts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    artist VARCHAR(255) NOT NULL,
    venue VARCHAR(255) NOT NULL,
    concert_date TIMESTAMP NOT NULL,
    total_tickets INT NOT NULL,
    available_tickets INT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    booking_start_time TIMESTAMP NOT NULL,
    booking_end_time TIMESTAMP NOT NULL,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

CREATE INDEX idx_concerts_date ON concerts(concert_date);

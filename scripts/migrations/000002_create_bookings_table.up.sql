CREATE TABLE IF NOT EXISTS bookings (
                                        id SERIAL PRIMARY KEY,
                                        concert_id INT NOT NULL REFERENCES concerts(id),
    user_id VARCHAR(255) NOT NULL,
    ticket_count INT NOT NULL,
    booking_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'confirmed',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_ticket_count CHECK (ticket_count > 0)
    );

CREATE INDEX idx_bookings_concert_id ON bookings(concert_id);
CREATE INDEX idx_bookings_user_id ON bookings(user_id);

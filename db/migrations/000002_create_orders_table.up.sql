CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    status VARCHAR(45) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    user_id INTEGER NOT NULL,
    ticket_id INTEGER NOT NULL REFERENCES tickets (id) ON DELETE SET NULL,
    version INTEGER DEFAULT 1
);
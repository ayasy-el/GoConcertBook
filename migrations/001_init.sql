CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS ticket_categories (
    id TEXT PRIMARY KEY,
    event_id TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    total_stock INT NOT NULL,
    price BIGINT NOT NULL,
    UNIQUE(event_id, name)
);

CREATE TABLE IF NOT EXISTS reservations (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    event_id TEXT NOT NULL,
    category TEXT NOT NULL,
    qty INT NOT NULL,
    status TEXT NOT NULL,
    expired_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS bookings (
    id TEXT PRIMARY KEY,
    reservation_id TEXT NOT NULL UNIQUE,
    payment_status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

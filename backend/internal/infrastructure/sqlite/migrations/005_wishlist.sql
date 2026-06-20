CREATE TABLE IF NOT EXISTS wishlist_items (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    url           TEXT NOT NULL DEFAULT '',
    store         TEXT NOT NULL DEFAULT '',
    target_price  REAL NOT NULL DEFAULT 0,
    current_price REAL NOT NULL DEFAULT 0,
    currency      TEXT NOT NULL DEFAULT 'USD',
    priority      TEXT NOT NULL DEFAULT 'medium',
    status        TEXT NOT NULL DEFAULT 'pending',
    notes         TEXT NOT NULL DEFAULT '',
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL
);

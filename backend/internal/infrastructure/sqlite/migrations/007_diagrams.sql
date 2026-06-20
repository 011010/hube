CREATE TABLE IF NOT EXISTS diagrams (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL DEFAULT 'Untitled diagram',
    nodes      TEXT NOT NULL DEFAULT '[]',
    edges      TEXT NOT NULL DEFAULT '[]',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

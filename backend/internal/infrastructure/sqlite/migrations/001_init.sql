CREATE TABLE IF NOT EXISTS tasks (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    priority    TEXT NOT NULL DEFAULT 'medium',
    status      TEXT NOT NULL DEFAULT 'todo',
    due_date    DATETIME,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS events (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    start_at    DATETIME NOT NULL,
    end_at      DATETIME NOT NULL,
    all_day     INTEGER NOT NULL DEFAULT 0,
    color       TEXT NOT NULL DEFAULT '#6366f1',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS apps (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    url         TEXT NOT NULL,
    icon        TEXT NOT NULL DEFAULT '',
    color       TEXT NOT NULL DEFAULT '#6366f1',
    sort_order  INTEGER NOT NULL DEFAULT 0,
    active      INTEGER NOT NULL DEFAULT 1
);

INSERT OR IGNORE INTO apps (id, name, description, url, icon, color, sort_order) VALUES
    ('app-ct',  'PayPinga',      'Card tracker', 'https://ct.husari.dev', '💳', '#10b981', 0),
    ('app-fin', 'FinControlPWA', 'Finance control', 'https://fincontrol.husari.dev', '💰', '#f59e0b', 1);

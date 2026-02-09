-- +goose Up
CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL CHECK (type IN ('income', 'expense')),
    amount DECIMAL(10, 2) NOT NULL CHECK (amount >= 0),
    date TIMESTAMPTZ NOT NULL,
    category VARCHAR(100),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_items_date ON items (date);
CREATE INDEX IF NOT EXISTS idx_items_amount ON items (amount);
CREATE INDEX IF NOT EXISTS idx_items_category ON items (category);

-- +goose Down
DROP INDEX IF EXISTS idx_items_category;
DROP INDEX IF EXISTS idx_items_amount;
DROP INDEX IF EXISTS idx_items_date;
DROP TABLE IF EXISTS items;
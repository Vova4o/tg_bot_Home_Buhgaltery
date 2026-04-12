CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_user_id BIGINT UNIQUE NOT NULL,
    display_name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    parent_id INT REFERENCES categories(id)
);

CREATE TABLE IF NOT EXISTS expenses (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    category_id INT REFERENCES categories(id),
    source_type VARCHAR(50) NOT NULL CHECK (source_type IN ('text', 'receipt')),
    description TEXT,
    amount NUMERIC(15, 2) NOT NULL,
    expense_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS receipts (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    telegram_message_id BIGINT,
    file_path VARCHAR(500),
    ocr_text TEXT,
    merchant_name VARCHAR(255),
    purchase_at TIMESTAMP WITH TIME ZONE,
    total_amount NUMERIC(15, 2),
    currency VARCHAR(10) DEFAULT 'RUB',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS receipt_items (
    id SERIAL PRIMARY KEY,
    receipt_id INT REFERENCES receipts(id) ON DELETE CASCADE,
    item_name VARCHAR(255),
    amount NUMERIC(15, 2),
    category_id INT REFERENCES categories(id)
);

-- Seed predefined categories
INSERT INTO categories (name) VALUES
    ('groceries'),
    ('transport'),
    ('entertainment'),
    ('cigarettes'),
    ('pharmacy'),
    ('home'),
    ('cafe'),
    ('subscriptions'),
    ('clothes'),
    ('other')
ON CONFLICT (name) DO NOTHING;

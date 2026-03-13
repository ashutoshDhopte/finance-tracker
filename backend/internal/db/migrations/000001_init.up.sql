CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enum types
CREATE TYPE transaction_source AS ENUM ('email', 'sms', 'csv', 'plaid', 'manual');
CREATE TYPE transaction_type AS ENUM ('credit', 'debit');
CREATE TYPE account_type AS ENUM ('checking', 'savings', 'credit_card', 'investment');
CREATE TYPE alert_period AS ENUM ('daily', 'weekly', 'biweekly', 'monthly');
CREATE TYPE budget_period AS ENUM ('weekly', 'biweekly', 'monthly', 'yearly');

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Accounts (bank accounts, credit cards)
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    institution VARCHAR(200) NOT NULL,
    account_type account_type NOT NULL,
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_accounts_user_id ON accounts(user_id);

-- Categories
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    icon VARCHAR(50) NOT NULL DEFAULT 'tag',
    color VARCHAR(7) NOT NULL DEFAULT '#6B7280',
    parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_parent_id ON categories(parent_id);

-- Transactions
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    amount DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    merchant_name VARCHAR(300),
    merchant_raw VARCHAR(500),
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    transaction_date DATE NOT NULL,
    posted_date DATE,
    txn_type transaction_type NOT NULL,
    source transaction_source NOT NULL,
    source_hash VARCHAR(64) UNIQUE,
    ai_confidence REAL,
    raw_text TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_date ON transactions(transaction_date);
CREATE INDEX idx_transactions_category ON transactions(category_id);
CREATE INDEX idx_transactions_account ON transactions(account_id);
CREATE INDEX idx_transactions_source ON transactions(source);

-- Budgets
CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID REFERENCES categories(id) ON DELETE CASCADE,
    amount DECIMAL(12, 2) NOT NULL,
    period budget_period NOT NULL DEFAULT 'monthly',
    start_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_budgets_user_id ON budgets(user_id);

-- Alerts
CREATE TABLE alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    category_id UUID REFERENCES categories(id) ON DELETE CASCADE,
    threshold DECIMAL(12, 2) NOT NULL,
    period alert_period NOT NULL DEFAULT 'monthly',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_triggered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alerts_user_id ON alerts(user_id);

-- Email sync state (track last processed email)
CREATE TABLE email_sync_state (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_history_id BIGINT NOT NULL DEFAULT 0,
    last_synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id)
);

-- Seed default categories
INSERT INTO categories (name, icon, color) VALUES
    ('Groceries',       'shopping-cart', '#22C55E'),
    ('Dining',          'utensils',      '#F97316'),
    ('Gas',             'fuel',          '#EAB308'),
    ('Transportation',  'car',           '#3B82F6'),
    ('Shopping',        'bag',           '#8B5CF6'),
    ('Bills & Utilities','file-text',    '#EF4444'),
    ('Rent & Mortgage', 'home',          '#DC2626'),
    ('Healthcare',      'heart-pulse',   '#EC4899'),
    ('Entertainment',   'film',          '#A855F7'),
    ('Subscriptions',   'repeat',        '#6366F1'),
    ('Income',          'dollar-sign',   '#10B981'),
    ('Transfer',        'arrow-right-left','#64748B'),
    ('ATM',             'landmark',      '#0EA5E9'),
    ('Fees',            'alert-triangle','#F43F5E'),
    ('Other',           'tag',           '#6B7280');

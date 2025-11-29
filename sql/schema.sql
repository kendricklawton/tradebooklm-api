-- ============================================================================
-- Section 1: Extensions & Utilities
-- ============================================================================
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Function to auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- ============================================================================
-- Section 2: Enums
-- ============================================================================
CREATE TYPE asset_class AS ENUM ('equities', 'fixed_income', 'commodities', 'etfs', 'forex', 'derivatives', 'crypto');
CREATE TYPE trade_order_type AS ENUM ('market', 'limit', 'stop', 'stop_limit');
CREATE TYPE trade_purchase_type AS ENUM ('cash', 'margin');
CREATE TYPE tradebook_role AS ENUM ('owner', 'editor', 'reader');

-- ============================================================================
-- Section 3: Core Application Tables
-- ============================================================================

-- 1. Users
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY, -- Matches Auth Provider ID
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Tradebooks (The Container)
CREATE TABLE IF NOT EXISTS tradebooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Members (Access Control)
CREATE TABLE IF NOT EXISTS tradebook_members (
    tradebook_id UUID NOT NULL REFERENCES tradebooks(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role tradebook_role NOT NULL DEFAULT 'reader',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tradebook_id, user_id)
);

-- 4. Trades
CREATE TABLE IF NOT EXISTS trades (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tradebook_id UUID NOT NULL REFERENCES tradebooks(id) ON DELETE CASCADE,

    -- Status Flag
    is_open BOOLEAN NOT NULL DEFAULT TRUE,

    -- Classification
    asset_class asset_class NOT NULL,
    purchase_type trade_purchase_type NOT NULL,
    order_type trade_order_type NOT NULL,

    -- Data
    entry_date TIMESTAMPTZ NOT NULL,
    symbol TEXT NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'USD', -- Added based on best practice

    -- Financials
    entry_quantity NUMERIC(19, 8) NOT NULL,
    entry_price NUMERIC(19, 8) NOT NULL,
    entry_fees NUMERIC(19, 8) DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 5. Exit Legs
CREATE TABLE IF NOT EXISTS exit_legs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trade_id UUID NOT NULL REFERENCES trades(id) ON DELETE CASCADE,

    exit_date TIMESTAMPTZ NOT NULL,

    exit_quantity NUMERIC(19, 8) NOT NULL,
    exit_price NUMERIC(19, 8) NOT NULL,
    exit_fees NUMERIC(19, 8) DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- Section 4: Metering
-- ============================================================================
CREATE TABLE IF NOT EXISTS token_usage_log (
    event_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         TEXT NOT NULL,
    timestamp       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    model_name      VARCHAR(50) NOT NULL,
    prompt_tokens   INTEGER NOT NULL,
    completion_tokens INTEGER NOT NULL,
    total_tokens    INTEGER NOT NULL,
    cost            DECIMAL(10, 8) NOT NULL
);

-- ============================================================================
-- Section 5: Indexes & Triggers
-- ============================================================================

-- 1. Standard Performance Indexes
CREATE INDEX IF NOT EXISTS idx_tradebooks_owner ON tradebooks(owner_id);
CREATE INDEX IF NOT EXISTS idx_members_user ON tradebook_members(user_id);
CREATE INDEX IF NOT EXISTS idx_exit_legs_trade ON exit_legs(trade_id);

-- 2. AI & Search Indexes
CREATE INDEX IF NOT EXISTS idx_trades_symbol ON trades(symbol);
CREATE INDEX IF NOT EXISTS idx_trades_is_open ON trades(tradebook_id) WHERE is_open = TRUE;
CREATE INDEX IF NOT EXISTS idx_trades_asset_analysis ON trades(tradebook_id, asset_class, entry_date);
CREATE INDEX IF NOT EXISTS idx_trades_date_lookup ON trades(tradebook_id, entry_date DESC);

-- 3. Triggers (Auto-update updated_at)
CREATE TRIGGER update_users_modtime BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tradebooks_modtime BEFORE UPDATE ON tradebooks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_trades_modtime BEFORE UPDATE ON trades FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_exits_modtime BEFORE UPDATE ON exit_legs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

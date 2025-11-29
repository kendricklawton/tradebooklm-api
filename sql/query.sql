-- ============================================================================
-- 1. USERS
-- ============================================================================

-- name: UpsertUser :one
INSERT INTO users (id)
VALUES (@id)
ON CONFLICT (id) DO UPDATE
SET updated_at = NOW()
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = @id;

-- ============================================================================
-- 2. TRADEBOOKS
-- ============================================================================

-- name: CreateTradebook :one
INSERT INTO tradebooks (owner_id, title)
VALUES (@owner_id, @title)
RETURNING *;

-- name: ListTradebooks :many
SELECT
    tb.id, tb.owner_id, tb.title, tb.created_at, tb.updated_at,
    COALESCE(tm.role, 'owner')::tradebook_role AS user_role
FROM tradebooks tb
LEFT JOIN tradebook_members tm
    ON tb.id = tm.tradebook_id AND tm.user_id = @user_id
WHERE tb.owner_id = @user_id
    OR tm.user_id = @user_id
ORDER BY tb.updated_at DESC
LIMIT @limit_val OFFSET @offset_val;

-- name: GetTradebook :one
SELECT
    tb.*,
    COALESCE(tm.role, 'owner')::tradebook_role AS user_role
FROM tradebooks tb
LEFT JOIN tradebook_members tm
    ON tb.id = tm.tradebook_id AND tm.user_id = @user_id
WHERE tb.id = @tradebook_id
    AND (tb.owner_id = @user_id OR tm.user_id IS NOT NULL);

-- name: UpdateTradebook :one
UPDATE tradebooks
SET
    title = COALESCE(sqlc.narg('title'), title),
    updated_at = NOW()
WHERE id = @tradebook_id
    AND owner_id = @user_id
RETURNING *;

-- name: DeleteTradebook :exec
DELETE FROM tradebooks
WHERE id = @tradebook_id AND owner_id = @user_id;

-- name: DeleteAllTradebooks :exec
DELETE FROM tradebooks
WHERE owner_id = @user_id;

-- ============================================================================
-- 3. MEMBERS (Refactored to use UPSERT)
-- ============================================================================

-- name: UpsertTradebookMember :one
WITH authorized_check AS (
    -- Only proceed if the @owner_id is the actual owner of the tradebook
    SELECT 1
    FROM tradebooks
    WHERE id = @tradebook_id AND owner_id = @owner_id
)
INSERT INTO tradebook_members (tradebook_id, user_id, role)
SELECT @tradebook_id, @new_member_id, @role
FROM authorized_check
ON CONFLICT (tradebook_id, user_id) DO UPDATE
SET
    role = EXCLUDED.role
RETURNING *;

-- name: RemoveTradebookMember :exec
DELETE FROM tradebook_members
WHERE tradebook_id = @tradebook_id
    AND user_id = @target_user_id
    AND EXISTS (
        SELECT 1 FROM tradebooks
        WHERE id = @tradebook_id AND owner_id = @owner_id
    );

-- ============================================================================
-- 4. TRADES
-- ============================================================================

-- name: CreateTrade :one
INSERT INTO trades (
    tradebook_id, asset_class, purchase_type, order_type,
    entry_date, symbol, currency, entry_quantity, entry_price, entry_fees
)
SELECT
    @tradebook_id, @asset_class, @purchase_type, @order_type,
    @entry_date, @symbol, @currency, @entry_quantity, @entry_price, @entry_fees
WHERE EXISTS (
    SELECT 1 FROM tradebooks tb
    LEFT JOIN tradebook_members tm ON tb.id = tm.tradebook_id
    WHERE tb.id = @tradebook_id
    AND (
        tb.owner_id = @user_id
        OR (tm.user_id = @user_id AND tm.role IN ('owner', 'editor'))
    )
)
RETURNING *;

-- name: ListTrades :many
SELECT t.* FROM trades t
JOIN tradebooks tb ON t.tradebook_id = tb.id
LEFT JOIN tradebook_members tm ON tb.id = tm.tradebook_id
WHERE t.tradebook_id = @tradebook_id
    AND (tb.owner_id = @user_id OR tm.user_id = @user_id)
ORDER BY t.entry_date DESC
LIMIT @limit_val OFFSET @offset_val;

-- name: GetTrade :one
SELECT t.* FROM trades t
JOIN tradebooks tb ON t.tradebook_id = tb.id
LEFT JOIN tradebook_members tm ON tb.id = tm.tradebook_id
WHERE t.id = @trade_id
    AND (tb.owner_id = @user_id OR tm.user_id = @user_id);

-- name: UpdateTrade :one
UPDATE trades
SET
    is_open = COALESCE(sqlc.narg('is_open'), is_open),
    updated_at = NOW()
FROM tradebooks tb
LEFT JOIN tradebook_members tm ON tb.id = tm.tradebook_id
WHERE trades.id = @trade_id
    AND trades.tradebook_id = tb.id
    AND (tb.owner_id = @user_id OR (tm.user_id = @user_id AND tm.role IN ('owner', 'editor')))
RETURNING trades.*;

-- name: DeleteTrade :exec
DELETE FROM trades
USING tradebooks tb, tradebook_members tm
WHERE trades.id = @trade_id
    AND trades.tradebook_id = tb.id
    AND tb.id = tm.tradebook_id
    AND (tb.owner_id = @user_id OR (tm.user_id = @user_id AND tm.role IN ('owner', 'editor')));

-- ============================================================================
-- 5. EXIT LEGS
-- ============================================================================

-- name: GetExitLegCount :one
-- Used for Code-Level check before insertion
SELECT COUNT(*) FROM exit_legs WHERE trade_id = @trade_id;

-- name: AddExitLeg :one
-- Security: Verifies ownership AND enforces max 100 exit legs per trade (DB Level Safety Net)
INSERT INTO exit_legs (
    trade_id, exit_date, exit_quantity, exit_price, exit_fees
)
SELECT
    @trade_id, @exit_date, @exit_quantity, @exit_price, @exit_fees
FROM trades t
JOIN tradebooks tb ON t.tradebook_id = tb.id
LEFT JOIN tradebook_members tm ON tb.id = tm.tradebook_id
WHERE t.id = @trade_id
    AND (tb.owner_id = @user_id OR (tm.user_id = @user_id AND tm.role IN ('owner', 'editor')))
    -- LIMIT CHECK: Ensure we haven't hit 100 legs yet
    AND (SELECT COUNT(*) FROM exit_legs WHERE trade_id = @trade_id) < 100
RETURNING *;

-- name: ListExitLegs :many
SELECT el.* FROM exit_legs el
JOIN trades t ON el.trade_id = t.id
JOIN tradebooks tb ON t.tradebook_id = tb.id
LEFT JOIN tradebook_members tm ON tb.id = tm.tradebook_id
WHERE t.id = @trade_id
    AND (tb.owner_id = @user_id OR tm.user_id = @user_id)
ORDER BY el.exit_date ASC;

-- ============================================================================
-- 6. DASHBOARD
-- ============================================================================

-- name: GetOpenPositions :many
SELECT t.*
FROM trades t
JOIN tradebooks tb ON t.tradebook_id = tb.id
LEFT JOIN tradebook_members tm ON tb.id = tm.tradebook_id
WHERE t.tradebook_id = @tradebook_id
    AND t.is_open = TRUE
    AND (tb.owner_id = @user_id OR tm.user_id = @user_id)
ORDER BY t.entry_date DESC;

-- ============================================================================
-- 7. METERING
-- ============================================================================

-- name: LogTokenUsage :exec
INSERT INTO token_usage_log (
    user_id, model_name, prompt_tokens, completion_tokens, total_tokens, cost
) VALUES (
    @user_id, @model_name, @prompt_tokens, @completion_tokens, @total_tokens, @cost
);

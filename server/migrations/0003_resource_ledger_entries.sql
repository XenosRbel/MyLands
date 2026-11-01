-- ML-B01 / DB_DESIGN F-01, R-1: normalize resource_ledger.delta (JSONB) into rows.
-- The header row stays in resource_ledger. One entry row per (ledger row, resource).
CREATE TABLE resource_ledger_entries (
    ledger_id   BIGINT NOT NULL REFERENCES resource_ledger(id) ON DELETE RESTRICT,
    resource_id TEXT   NOT NULL CHECK (resource_id <> ''),
    amount      BIGINT NOT NULL CHECK (amount <> 0),   -- whole units, +grant / -spend
    PRIMARY KEY (ledger_id, resource_id)
);

-- Backfill legacy rows. Zero amounts carry no information and are skipped.
INSERT INTO resource_ledger_entries(ledger_id, resource_id, amount)
SELECT l.id, d.key, d.value::numeric::bigint
FROM resource_ledger l, jsonb_each_text(l.delta) d
WHERE d.value::numeric::bigint <> 0;

-- Per-resource sums by player/city come from the entries joined to the header.
-- Time-ordered access by player and by city (existing indexes order by id only).
CREATE INDEX resource_ledger_player_at ON resource_ledger(player_id, at DESC);
CREATE INDEX resource_ledger_city_at   ON resource_ledger(city_id, at DESC);

-- delta is deprecated: kept only so released code and old rows still load.
-- Current code writes resource_ledger_entries and never reads delta.
ALTER TABLE resource_ledger ALTER COLUMN delta SET DEFAULT '{}'::jsonb;
COMMENT ON COLUMN resource_ledger.delta IS 'DEPRECATED (0003): use resource_ledger_entries. Legacy rows only; remove in a later migration.';

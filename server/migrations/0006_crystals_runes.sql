-- 0006_crystals_runes (ML-B04): crystal stock and ledger, hero artefact slots, runes in artefacts.
-- Append-only. Crystal type and rune ids are Game Pack ids stored as text. Prices and amounts stay in the pack.
-- See docs/modernization/CRYSTAL_ECONOMY.md and DB_DESIGN.md section 6.

-- ---------------------------------------------------------------------------
-- 1. Crystal stock per player and crystal type
-- ---------------------------------------------------------------------------
CREATE TABLE crystal_stocks (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    player_id     BIGINT NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    crystal_type  TEXT NOT NULL CHECK (crystal_type ~ '^[a-z][a-z0-9_.]{0,63}$'),
    quantity      BIGINT NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (player_id, crystal_type)
);
COMMENT ON TABLE crystal_stocks IS 'One row per player and crystal type. quantity changes only through crystal_ledger inserts.';

-- ---------------------------------------------------------------------------
-- 2. Append-only crystal ledger. Every row mirrors one resource_ledger entry (rule: every spend and grant is journaled).
-- The server writes resource_ledger + resource_ledger_entries first, then the crystal_ledger row.
-- ---------------------------------------------------------------------------
CREATE TABLE crystal_ledger (
    id                  BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    stock_id            BIGINT NOT NULL REFERENCES crystal_stocks(id) ON DELETE RESTRICT,
    resource_ledger_id  BIGINT NOT NULL REFERENCES resource_ledger(id) ON DELETE RESTRICT,
    delta               BIGINT NOT NULL CHECK (delta <> 0),
    balance_after       BIGINT NOT NULL CHECK (balance_after >= 0),
    reason              TEXT NOT NULL CHECK (reason IN ('quest_reward','ruin_monster','building_spend','refund','admin_grant','admin_revoke')),
    ref                 TEXT NOT NULL CHECK (length(ref) BETWEEN 1 AND 128),
    at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((reason IN ('quest_reward','ruin_monster','refund','admin_grant') AND delta > 0)
        OR (reason IN ('building_spend','admin_revoke') AND delta < 0)),
    UNIQUE (stock_id, reason, ref),
    UNIQUE (resource_ledger_id, stock_id)
);
CREATE INDEX crystal_ledger_stock ON crystal_ledger(stock_id, id DESC);
COMMENT ON TABLE crystal_ledger IS 'Append-only crystal journal. balance_after is computed by trigger. UPDATE, DELETE and TRUNCATE are rejected.';

CREATE FUNCTION crystal_ledger_apply() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE s crystal_stocks%ROWTYPE;
BEGIN
    SELECT * INTO s FROM crystal_stocks WHERE id = NEW.stock_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'crystal stock % not found', NEW.stock_id USING ERRCODE = 'foreign_key_violation';
    END IF;
    -- The mirrored resource_ledger header must belong to the same player and carry the same amount.
    IF NOT EXISTS (
        SELECT 1 FROM resource_ledger l JOIN resource_ledger_entries e ON e.ledger_id = l.id
        WHERE l.id = NEW.resource_ledger_id AND l.player_id = s.player_id
          AND e.resource_id = s.crystal_type AND e.amount = NEW.delta) THEN
        RAISE EXCEPTION 'crystal ledger row needs a matching resource_ledger entry (ledger %, type %, delta %)',
            NEW.resource_ledger_id, s.crystal_type, NEW.delta USING ERRCODE = 'foreign_key_violation';
    END IF;
    NEW.balance_after := s.quantity + NEW.delta;
    IF NEW.balance_after < 0 THEN
        RAISE EXCEPTION 'insufficient crystals: % + % < 0', s.quantity, NEW.delta USING ERRCODE = 'check_violation';
    END IF;
    UPDATE crystal_stocks SET quantity = NEW.balance_after, updated_at = now() WHERE id = NEW.stock_id;
    RETURN NEW;
END $$;
CREATE TRIGGER crystal_ledger_apply BEFORE INSERT ON crystal_ledger
    FOR EACH ROW EXECUTE FUNCTION crystal_ledger_apply();

CREATE FUNCTION crystal_append_only() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION '% is append-only', TG_TABLE_NAME USING ERRCODE = 'restrict_violation';
END $$;
CREATE TRIGGER crystal_ledger_immutable BEFORE UPDATE OR DELETE ON crystal_ledger
    FOR EACH ROW EXECUTE FUNCTION crystal_append_only();
CREATE TRIGGER crystal_ledger_no_truncate BEFORE TRUNCATE ON crystal_ledger
    FOR EACH STATEMENT EXECUTE FUNCTION crystal_append_only();

CREATE FUNCTION crystal_stock_guard() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.quantity <> 0 THEN
            RAISE EXCEPTION 'a new crystal stock must start at 0' USING ERRCODE = 'check_violation';
        END IF;
    ELSE
        IF NEW.player_id <> OLD.player_id OR NEW.crystal_type <> OLD.crystal_type THEN
            RAISE EXCEPTION 'crystal stock identity is immutable' USING ERRCODE = 'restrict_violation';
        END IF;
        -- pg_trigger_depth() > 1 means the change comes from the crystal_ledger trigger.
        IF NEW.quantity <> OLD.quantity AND pg_trigger_depth() < 2 THEN
            RAISE EXCEPTION 'crystal quantity changes only through crystal_ledger' USING ERRCODE = 'restrict_violation';
        END IF;
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER crystal_stock_guard BEFORE INSERT OR UPDATE ON crystal_stocks
    FOR EACH ROW EXECUTE FUNCTION crystal_stock_guard();

CREATE VIEW crystal_stock_audit AS
    SELECT s.id AS stock_id, s.quantity, COALESCE(SUM(l.delta), 0)::BIGINT AS ledger_sum,
           s.quantity = COALESCE(SUM(l.delta), 0) AS consistent
    FROM crystal_stocks s LEFT JOIN crystal_ledger l ON l.stock_id = s.id
    GROUP BY s.id, s.quantity;
COMMENT ON VIEW crystal_stock_audit IS 'Reconciliation: quantity must equal the sum of crystal_ledger deltas for every stock.';

-- ---------------------------------------------------------------------------
-- 3. Hero artefact slots. items.hero_id/slot already store equipment; this table normalizes it.
-- A trigger on items keeps both in step until released code moves to the new table.
-- ---------------------------------------------------------------------------
CREATE TABLE hero_artefact_slots (
    hero_id  BIGINT NOT NULL REFERENCES heroes(id) ON DELETE CASCADE,
    slot     TEXT NOT NULL CHECK (slot <> ''),  -- pack slot id
    item_id  BIGINT NOT NULL UNIQUE REFERENCES items(id) ON DELETE CASCADE,  -- an item sits in one slot
    PRIMARY KEY (hero_id, slot)  -- slot unique per hero
);
COMMENT ON TABLE hero_artefact_slots IS 'Equipped artefact per hero and slot. Mirrors items.hero_id/slot (DEPRECATED) via trigger items_sync_artefact_slot.';

CREATE FUNCTION items_sync_artefact_slot() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    DELETE FROM hero_artefact_slots WHERE item_id = NEW.id;
    IF NEW.hero_id IS NOT NULL AND NEW.slot IS NOT NULL AND NEW.slot <> '' THEN
        INSERT INTO hero_artefact_slots(hero_id, slot, item_id) VALUES (NEW.hero_id, NEW.slot, NEW.id);
    END IF;
    RETURN NULL;
END $$;
CREATE TRIGGER items_sync_artefact_slot AFTER INSERT OR UPDATE OF hero_id, slot ON items
    FOR EACH ROW EXECUTE FUNCTION items_sync_artefact_slot();

INSERT INTO hero_artefact_slots(hero_id, slot, item_id)
SELECT hero_id, slot, id FROM items WHERE hero_id IS NOT NULL AND slot IS NOT NULL AND slot <> '';

COMMENT ON COLUMN items.hero_id IS 'DEPRECATED (0006): use hero_artefact_slots. Still written by released code; remove after the Go cutover.';
COMMENT ON COLUMN items.slot IS 'DEPRECATED (0006): use hero_artefact_slots.';

-- Runes inserted into an artefact: sockets 1..5, so an artefact holds at most 5 runes.
CREATE TABLE item_runes (
    item_id   BIGINT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    socket    INT NOT NULL CHECK (socket BETWEEN 1 AND 5),
    rune_id   TEXT NOT NULL CHECK (rune_id ~ '^[a-z][a-z0-9_.]{0,63}$'),  -- Game Pack rune id
    added_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (item_id, socket)
);
COMMENT ON TABLE item_runes IS 'Runes in an artefact. Socket 1..5 plus the primary key cap an artefact at 5 runes.';

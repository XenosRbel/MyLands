-- 0002_mylands_foundations: server profiles, premium-currency wallet (CHZH), mock payments, paid services.
-- Design and review: docs/modernization/DB_DESIGN.md. Append-only: never edit this file after release.
-- Pack data (races, buildings, units) stays in Game Pack JSON. These tables store player state and pack ids only.

-- ---------------------------------------------------------------------------
-- 1. Server profile and typed parameters (ML-D08)
-- ---------------------------------------------------------------------------
CREATE TABLE server_profiles (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code        TEXT NOT NULL UNIQUE CHECK (code ~ '^[a-z][a-z0-9_]{1,47}$'),
    name        TEXT NOT NULL CHECK (length(name) BETWEEN 1 AND 100),
    status      TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','retired')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
COMMENT ON TABLE server_profiles IS 'Server profile (for example combat, economy). Values come from configuration, not from engine code.';

CREATE TABLE server_param_defs (
    code           TEXT PRIMARY KEY CHECK (code ~ '^[a-z][a-z0-9_.]{1,63}$'),
    kind           TEXT NOT NULL CHECK (kind IN ('multiplier','integer','duration_seconds','percent')),
    min_value      NUMERIC(14,4) NOT NULL,
    max_value      NUMERIC(14,4) NOT NULL,
    default_value  NUMERIC(14,4) NOT NULL,
    description    TEXT NOT NULL DEFAULT '',
    CHECK (min_value <= max_value),
    CHECK (default_value BETWEEN min_value AND max_value)
);
COMMENT ON TABLE server_param_defs IS 'Registry of allowed profile parameters with type kind and valid range. Formulas stay in the engine.';

CREATE TABLE server_profile_params (
    profile_id  BIGINT NOT NULL REFERENCES server_profiles(id) ON DELETE CASCADE,
    param_code  TEXT   NOT NULL REFERENCES server_param_defs(code) ON DELETE RESTRICT,
    value       NUMERIC(14,4) NOT NULL,
    PRIMARY KEY (profile_id, param_code)
);
CREATE INDEX server_profile_params_code ON server_profile_params(param_code);
COMMENT ON TABLE server_profile_params IS 'One typed value per profile and parameter. Range is enforced against server_param_defs.';

CREATE FUNCTION server_profile_param_check() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE d server_param_defs%ROWTYPE;
BEGIN
    SELECT * INTO d FROM server_param_defs WHERE code = NEW.param_code;
    IF NEW.value < d.min_value OR NEW.value > d.max_value THEN
        RAISE EXCEPTION 'parameter % value % outside [%, %]', NEW.param_code, NEW.value, d.min_value, d.max_value
            USING ERRCODE = 'check_violation';
    END IF;
    IF d.kind = 'integer' AND NEW.value <> trunc(NEW.value) THEN
        RAISE EXCEPTION 'parameter % must be an integer', NEW.param_code USING ERRCODE = 'check_violation';
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER server_profile_params_range BEFORE INSERT OR UPDATE OF value, param_code ON server_profile_params
    FOR EACH ROW EXECUTE FUNCTION server_profile_param_check();

-- World to profile link. Nullable: worlds created before this migration have no profile.
ALTER TABLE worlds ADD COLUMN profile_id BIGINT REFERENCES server_profiles(id) ON DELETE RESTRICT;
CREATE INDEX worlds_profile ON worlds(profile_id) WHERE profile_id IS NOT NULL;
COMMENT ON COLUMN worlds.profile_id IS 'Server profile chosen at world creation. NULL for legacy worlds.';

-- ---------------------------------------------------------------------------
-- 2. Premium-currency wallet and append-only ledger (ML-M15)
-- ---------------------------------------------------------------------------
CREATE TABLE premium_wallets (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    account_id  BIGINT NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    world_id    BIGINT NOT NULL REFERENCES worlds(id) ON DELETE RESTRICT,
    balance     BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (account_id, world_id)
);
CREATE INDEX premium_wallets_world ON premium_wallets(world_id);
COMMENT ON TABLE premium_wallets IS 'CHZH balance per account on one server (world). Balance changes only through premium_ledger inserts.';

CREATE TABLE premium_ledger (
    id                   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    wallet_id            BIGINT NOT NULL REFERENCES premium_wallets(id) ON DELETE RESTRICT,
    delta                BIGINT NOT NULL CHECK (delta <> 0),
    balance_after        BIGINT NOT NULL CHECK (balance_after >= 0),
    reason               TEXT NOT NULL CHECK (reason IN ('purchase','service_spend','refund','admin_grant','admin_revoke')),
    ref                  TEXT NOT NULL CHECK (length(ref) BETWEEN 1 AND 128),
    payment_order_id     BIGINT,
    service_instance_id  BIGINT,
    at                   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((reason IN ('purchase','refund','admin_grant') AND delta > 0)
        OR (reason IN ('service_spend','admin_revoke') AND delta < 0)),
    CHECK (reason <> 'purchase' OR payment_order_id IS NOT NULL),
    UNIQUE (wallet_id, reason, ref)
);
CREATE UNIQUE INDEX premium_ledger_one_purchase_per_order ON premium_ledger(payment_order_id) WHERE reason = 'purchase';
CREATE INDEX premium_ledger_wallet ON premium_ledger(wallet_id, id DESC);
COMMENT ON TABLE premium_ledger IS 'Append-only CHZH journal. balance_after is computed by trigger. UPDATE, DELETE and TRUNCATE are rejected.';

CREATE FUNCTION premium_ledger_apply() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE cur BIGINT;
BEGIN
    SELECT balance INTO cur FROM premium_wallets WHERE id = NEW.wallet_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'wallet % not found', NEW.wallet_id USING ERRCODE = 'foreign_key_violation';
    END IF;
    NEW.balance_after := cur + NEW.delta;
    IF NEW.balance_after < 0 THEN
        RAISE EXCEPTION 'insufficient premium balance: % + % < 0', cur, NEW.delta USING ERRCODE = 'check_violation';
    END IF;
    UPDATE premium_wallets SET balance = NEW.balance_after, updated_at = now() WHERE id = NEW.wallet_id;
    RETURN NEW;
END $$;
CREATE TRIGGER premium_ledger_apply BEFORE INSERT ON premium_ledger
    FOR EACH ROW EXECUTE FUNCTION premium_ledger_apply();

CREATE FUNCTION premium_append_only() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION '% is append-only', TG_TABLE_NAME USING ERRCODE = 'restrict_violation';
END $$;
CREATE TRIGGER premium_ledger_immutable BEFORE UPDATE OR DELETE ON premium_ledger
    FOR EACH ROW EXECUTE FUNCTION premium_append_only();
CREATE TRIGGER premium_ledger_no_truncate BEFORE TRUNCATE ON premium_ledger
    FOR EACH STATEMENT EXECUTE FUNCTION premium_append_only();

CREATE FUNCTION premium_wallet_guard() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.balance <> 0 THEN
            RAISE EXCEPTION 'a new wallet must start with balance 0' USING ERRCODE = 'check_violation';
        END IF;
    ELSE
        IF NEW.account_id <> OLD.account_id OR NEW.world_id <> OLD.world_id THEN
            RAISE EXCEPTION 'wallet owner is immutable' USING ERRCODE = 'restrict_violation';
        END IF;
        -- pg_trigger_depth() > 1 means the change comes from the premium_ledger trigger.
        IF NEW.balance <> OLD.balance AND pg_trigger_depth() < 2 THEN
            RAISE EXCEPTION 'wallet balance changes only through premium_ledger' USING ERRCODE = 'restrict_violation';
        END IF;
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER premium_wallet_guard BEFORE INSERT OR UPDATE ON premium_wallets
    FOR EACH ROW EXECUTE FUNCTION premium_wallet_guard();

CREATE VIEW premium_wallet_audit AS
    SELECT w.id AS wallet_id, w.balance, COALESCE(SUM(l.delta), 0)::BIGINT AS ledger_sum,
           w.balance = COALESCE(SUM(l.delta), 0) AS consistent
    FROM premium_wallets w LEFT JOIN premium_ledger l ON l.wallet_id = w.id
    GROUP BY w.id, w.balance;
COMMENT ON VIEW premium_wallet_audit IS 'Reconciliation: balance must equal the sum of ledger deltas for every wallet.';

-- ---------------------------------------------------------------------------
-- 3. Mock payment orders (ML-M16). Provider is fixed to mock. No card or provider data.
-- ---------------------------------------------------------------------------
CREATE TABLE payment_orders (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    wallet_id        BIGINT NOT NULL REFERENCES premium_wallets(id) ON DELETE RESTRICT,
    idempotency_key  TEXT NOT NULL CHECK (length(idempotency_key) BETWEEN 8 AND 128),
    provider         TEXT NOT NULL DEFAULT 'mock' CHECK (provider = 'mock'),
    package_code     TEXT NOT NULL CHECK (length(package_code) BETWEEN 1 AND 64),
    crystals         BIGINT NOT NULL CHECK (crystals > 0),
    price_minor      BIGINT NOT NULL CHECK (price_minor >= 0),
    currency         TEXT NOT NULL DEFAULT 'MOCK' CHECK (currency = 'MOCK'),
    status           TEXT NOT NULL DEFAULT 'created'
                     CHECK (status IN ('created','pending','confirmed','failed','cancelled','expired')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    confirmed_at     TIMESTAMPTZ,
    UNIQUE (wallet_id, idempotency_key),
    CHECK ((status = 'confirmed') = (confirmed_at IS NOT NULL))
);
CREATE INDEX payment_orders_open ON payment_orders(created_at) WHERE status IN ('created','pending');
COMMENT ON TABLE payment_orders IS 'Mock payment orders. State machine enforced by trigger payment_orders_transition. No real provider data.';

ALTER TABLE premium_ledger ADD CONSTRAINT premium_ledger_payment_order_fk
    FOREIGN KEY (payment_order_id) REFERENCES payment_orders(id) ON DELETE RESTRICT;
CREATE INDEX premium_ledger_payment_order ON premium_ledger(payment_order_id) WHERE payment_order_id IS NOT NULL;

CREATE FUNCTION payment_orders_transition() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.wallet_id <> OLD.wallet_id OR NEW.idempotency_key <> OLD.idempotency_key OR NEW.provider <> OLD.provider
       OR NEW.package_code <> OLD.package_code OR NEW.crystals <> OLD.crystals OR NEW.price_minor <> OLD.price_minor
       OR NEW.currency <> OLD.currency THEN
        RAISE EXCEPTION 'payment order terms are immutable' USING ERRCODE = 'restrict_violation';
    END IF;
    IF NEW.status <> OLD.status THEN
        IF NOT ((OLD.status = 'created' AND NEW.status IN ('pending','cancelled','expired','failed'))
             OR (OLD.status = 'pending' AND NEW.status IN ('confirmed','failed','cancelled','expired'))) THEN
            RAISE EXCEPTION 'invalid payment status transition % -> %', OLD.status, NEW.status
                USING ERRCODE = 'check_violation';
        END IF;
        IF NEW.status = 'confirmed' AND NEW.confirmed_at IS NULL THEN
            NEW.confirmed_at := now();
        END IF;
    ELSIF NEW.confirmed_at IS DISTINCT FROM OLD.confirmed_at THEN
        RAISE EXCEPTION 'confirmed_at changes only with the status' USING ERRCODE = 'restrict_violation';
    END IF;
    NEW.updated_at := now();
    RETURN NEW;
END $$;
CREATE TRIGGER payment_orders_transition BEFORE UPDATE ON payment_orders
    FOR EACH ROW EXECUTE FUNCTION payment_orders_transition();

-- ---------------------------------------------------------------------------
-- 4. Paid services: catalog reference and active instances (ML-D09, ML-M15)
-- ---------------------------------------------------------------------------
-- Prices, durations and effects stay in the pack (status PROPOSED). The server syncs ids into this registry.
CREATE TABLE service_defs (
    code           TEXT PRIMARY KEY CHECK (code ~ '^[a-z][a-z0-9_.]{1,63}$'),
    service_group  TEXT NOT NULL CHECK (service_group ~ '^[a-z][a-z0-9_.]{1,63}$'),
    pack_id        TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (code, service_group)
);
CREATE INDEX service_defs_group ON service_defs(service_group);
COMMENT ON TABLE service_defs IS 'Reference registry of purchasable service ids and their exclusivity group. Prices and effects stay in the Game Pack.';

CREATE TABLE service_instances (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    player_id      BIGINT NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    service_code   TEXT NOT NULL,
    service_group  TEXT NOT NULL,
    level          INT NOT NULL CHECK (level >= 1),
    price_paid     BIGINT NOT NULL CHECK (price_paid >= 0),
    status         TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','expired','cancelled','replaced')),
    started_at     TIMESTAMPTZ NOT NULL,
    expires_at     TIMESTAMPTZ NOT NULL,
    ended_at       TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (service_code, service_group) REFERENCES service_defs(code, service_group) ON DELETE RESTRICT,
    CHECK (expires_at > started_at),
    CHECK ((status = 'active') = (ended_at IS NULL))
);
-- One active service per player and exclusivity group. A weaker level never coexists with a stronger one.
CREATE UNIQUE INDEX service_instances_one_active ON service_instances(player_id, service_group) WHERE status = 'active';
CREATE INDEX service_instances_expiry ON service_instances(expires_at) WHERE status = 'active';
CREATE INDEX service_instances_player ON service_instances(player_id, id DESC);
CREATE INDEX service_instances_def ON service_instances(service_code, service_group);
COMMENT ON TABLE service_instances IS 'Purchased service state per player: level, start, expiry. At most one active row per player and group.';

ALTER TABLE premium_ledger ADD CONSTRAINT premium_ledger_service_instance_fk
    FOREIGN KEY (service_instance_id) REFERENCES service_instances(id) ON DELETE RESTRICT;
CREATE INDEX premium_ledger_service_instance ON premium_ledger(service_instance_id) WHERE service_instance_id IS NOT NULL;

CREATE FUNCTION service_instances_guard() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.player_id <> OLD.player_id OR NEW.service_code <> OLD.service_code OR NEW.service_group <> OLD.service_group
       OR NEW.started_at <> OLD.started_at THEN
        RAISE EXCEPTION 'service instance identity is immutable' USING ERRCODE = 'restrict_violation';
    END IF;
    IF OLD.status <> 'active' THEN
        RAISE EXCEPTION 'closed service instance is immutable' USING ERRCODE = 'restrict_violation';
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER service_instances_guard BEFORE UPDATE ON service_instances
    FOR EACH ROW EXECUTE FUNCTION service_instances_guard();

-- ---------------------------------------------------------------------------
-- 5. Additive fixes to 0001 tables (review findings F-02..F-07 in DB_DESIGN.md)
-- NOT VALID: new writes are checked at once; legacy rows are checked by a later VALIDATE CONSTRAINT.
-- ---------------------------------------------------------------------------
ALTER TABLE resource_ledger ADD CONSTRAINT resource_ledger_world_fk  FOREIGN KEY (world_id)  REFERENCES worlds(id)  NOT VALID;
ALTER TABLE resource_ledger ADD CONSTRAINT resource_ledger_player_fk FOREIGN KEY (player_id) REFERENCES players(id) NOT VALID;
ALTER TABLE resource_ledger ADD CONSTRAINT resource_ledger_city_fk   FOREIGN KEY (city_id)   REFERENCES cities(id)  NOT VALID;
ALTER TABLE idempotency_keys ADD CONSTRAINT idempotency_keys_account_fk FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE NOT VALID;
CREATE INDEX resource_ledger_player ON resource_ledger(player_id, id DESC);

ALTER TABLE cities  ADD CONSTRAINT cities_population_nonneg CHECK (population >= 0) NOT VALID;
ALTER TABLE heroes  ADD CONSTRAINT heroes_level_xp_valid CHECK (level >= 1 AND xp >= 0) NOT VALID;
ALTER TABLE tiles   ADD CONSTRAINT tiles_neutral_level_nonneg CHECK (neutral_level IS NULL OR neutral_level >= 0) NOT VALID;
ALTER TABLE missions ADD CONSTRAINT missions_time_order CHECK (arrive_at >= depart_at) NOT VALID;
ALTER TABLE jobs    ADD CONSTRAINT jobs_level_nonneg CHECK (level >= 0) NOT VALID;
ALTER TABLE players ADD CONSTRAINT players_json_objects CHECK (jsonb_typeof(techs) = 'object' AND jsonb_typeof(stats) = 'object') NOT VALID;
ALTER TABLE cities  ADD CONSTRAINT cities_json_objects CHECK (jsonb_typeof(resources) = 'object' AND jsonb_typeof(buildings) = 'object'
    AND jsonb_typeof(units) = 'object' AND jsonb_typeof(econ_carry) = 'object') NOT VALID;

-- Missing indexes on foreign keys.
CREATE INDEX players_account ON players(account_id);
CREATE INDEX heroes_home_city ON heroes(home_city_id);
CREATE INDEX missions_hero ON missions(hero_id) WHERE hero_id IS NOT NULL;
CREATE INDEX tiles_city ON tiles(city_id) WHERE city_id IS NOT NULL;
CREATE INDEX deliveries_player ON deliveries(player_id, id DESC);
CREATE INDEX deliveries_city ON deliveries(city_id);
CREATE INDEX messages_from_player ON messages(from_player, id DESC);
CREATE INDEX alliance_invites_player ON alliance_invites(player_id);
CREATE INDEX world_pack_history_world ON world_pack_history(world_id, id DESC);
CREATE INDEX alliance_relations_other ON alliance_relations(other_id);
CREATE INDEX market_orders_city ON market_orders(city_id);

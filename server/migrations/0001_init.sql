-- 0001_init: core schema. See docs/spec/DATA_MODEL.md for rationale.
-- Lock order (must be respected by all code): players (asc id) -> cities (asc id) -> missions -> tiles -> market_orders.

CREATE TABLE accounts (
    id             BIGSERIAL PRIMARY KEY,
    login          TEXT UNIQUE,
    password_hash  TEXT,
    display_name   TEXT NOT NULL,
    is_guest       BOOLEAN NOT NULL DEFAULT false,
    is_admin       BOOLEAN NOT NULL DEFAULT false,
    locale         TEXT NOT NULL DEFAULT 'en',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    banned_until   TIMESTAMPTZ,
    deleted_at     TIMESTAMPTZ,
    CONSTRAINT login_format CHECK (login IS NULL OR login ~ '^[a-z0-9_.@-]{3,64}$')
);

CREATE TABLE platform_links (
    platform     TEXT NOT NULL,
    external_id  TEXT NOT NULL,
    account_id   BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (platform, external_id)
);
CREATE INDEX platform_links_account ON platform_links(account_id);

CREATE TABLE sessions (
    token_hash  BYTEA PRIMARY KEY,
    account_id  BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ
);
CREATE INDEX sessions_account ON sessions(account_id);

CREATE TABLE worlds (
    id            BIGSERIAL PRIMARY KEY,
    slug          TEXT NOT NULL UNIQUE,
    name          TEXT NOT NULL,
    pack_id       TEXT NOT NULL,
    pack_version  TEXT NOT NULL,
    pack_hash     TEXT NOT NULL,
    pack_json     JSONB NOT NULL,      -- pinned snapshot; disk changes never affect a running world
    seed          BIGINT NOT NULL,
    speed         DOUBLE PRECISION NOT NULL DEFAULT 1 CHECK (speed > 0 AND speed <= 1000), -- world multiplier on pack worldSpeed
    status        TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','closed','archived')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE world_pack_history (
    id             BIGSERIAL PRIMARY KEY,
    world_id       BIGINT NOT NULL REFERENCES worlds(id),
    from_version   TEXT NOT NULL,
    to_version     TEXT NOT NULL,
    from_hash      TEXT NOT NULL,
    to_hash        TEXT NOT NULL,
    admin_id       BIGINT REFERENCES accounts(id),
    report         JSONB NOT NULL,
    migrated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE alliances (
    id          BIGSERIAL PRIMARY KEY,
    world_id    BIGINT NOT NULL REFERENCES worlds(id),
    name        TEXT NOT NULL,
    tag         TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    disbanded_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX alliances_name ON alliances(world_id, lower(name)) WHERE disbanded_at IS NULL;
CREATE UNIQUE INDEX alliances_tag ON alliances(world_id, lower(tag)) WHERE disbanded_at IS NULL;

CREATE TABLE players (
    id               BIGSERIAL PRIMARY KEY,
    world_id         BIGINT NOT NULL REFERENCES worlds(id),
    account_id       BIGINT NOT NULL REFERENCES accounts(id),
    name             TEXT NOT NULL,
    race_id          TEXT NOT NULL,
    techs            JSONB NOT NULL DEFAULT '{}',   -- techId -> level
    stats            JSONB NOT NULL DEFAULT '{}',   -- counters used by quests/achievements
    score            BIGINT NOT NULL DEFAULT 0,
    protected_until  TIMESTAMPTZ,
    alliance_id      BIGINT REFERENCES alliances(id),
    alliance_role    TEXT CHECK (alliance_role IN ('leader','officer','member')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (world_id, account_id)
);
CREATE UNIQUE INDEX players_name ON players(world_id, lower(name));
CREATE INDEX players_alliance ON players(alliance_id);
CREATE INDEX players_score ON players(world_id, score DESC);

CREATE TABLE cities (
    id               BIGSERIAL PRIMARY KEY,
    world_id         BIGINT NOT NULL REFERENCES worlds(id),
    player_id        BIGINT NOT NULL REFERENCES players(id),
    name             TEXT NOT NULL,
    x                INT NOT NULL,
    y                INT NOT NULL,
    is_capital       BOOLEAN NOT NULL DEFAULT false,
    resources        JSONB NOT NULL,                 -- resourceId -> milli units
    population       BIGINT NOT NULL,
    econ_updated_at  TIMESTAMPTZ NOT NULL,
    buildings        JSONB NOT NULL DEFAULT '{}',    -- buildingId -> level
    units            JSONB NOT NULL DEFAULT '{}',    -- unitId -> count (home garrison)
    econ_carry       JSONB NOT NULL DEFAULT '{}',    -- fractional remainders of tick processes
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (world_id, x, y)
);
CREATE INDEX cities_player ON cities(player_id);

CREATE TABLE tiles (
    world_id         BIGINT NOT NULL REFERENCES worlds(id),
    x                INT NOT NULL,
    y                INT NOT NULL,
    terrain          TEXT NOT NULL,
    neutral_id       TEXT,
    neutral_level    INT,
    cleared_until    TIMESTAMPTZ,       -- neutral inactive until this instant (lazy respawn)
    owner_player_id  BIGINT REFERENCES players(id),
    owner_since      TIMESTAMPTZ,
    yield_collected_at TIMESTAMPTZ,     -- territory yield accrues from here
    city_id          BIGINT REFERENCES cities(id),
    PRIMARY KEY (world_id, x, y)
);
CREATE INDEX tiles_owner ON tiles(owner_player_id) WHERE owner_player_id IS NOT NULL;

-- Construction / research / training queues. Times are precomputed at enqueue.
CREATE TABLE jobs (
    id           BIGSERIAL PRIMARY KEY,
    world_id     BIGINT NOT NULL REFERENCES worlds(id),
    player_id    BIGINT NOT NULL REFERENCES players(id),
    city_id      BIGINT NOT NULL REFERENCES cities(id),
    kind         TEXT NOT NULL CHECK (kind IN ('build','research','train')),
    item_id      TEXT NOT NULL,
    level        INT NOT NULL DEFAULT 0,
    quantity     BIGINT NOT NULL DEFAULT 1 CHECK (quantity > 0),
    cost         JSONB NOT NULL,
    starts_at    TIMESTAMPTZ NOT NULL,
    finishes_at  TIMESTAMPTZ NOT NULL,
    status       TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued','done','cancelled')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (finishes_at > starts_at)
);
CREATE INDEX jobs_city_open ON jobs(city_id, kind) WHERE status = 'queued';
CREATE INDEX jobs_player_research ON jobs(player_id) WHERE status = 'queued' AND kind = 'research';

CREATE TABLE heroes (
    id             BIGSERIAL PRIMARY KEY,
    world_id       BIGINT NOT NULL REFERENCES worlds(id),
    player_id      BIGINT NOT NULL REFERENCES players(id),
    home_city_id   BIGINT NOT NULL REFERENCES cities(id),
    name           TEXT NOT NULL,
    class_id       TEXT NOT NULL,
    level          INT NOT NULL DEFAULT 1,
    xp             BIGINT NOT NULL DEFAULT 0,
    skills         JSONB NOT NULL DEFAULT '{}',   -- skillId -> level
    status         TEXT NOT NULL DEFAULT 'idle' CHECK (status IN ('idle','mission')),
    wounded_until  TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX heroes_player ON heroes(player_id);

CREATE TABLE items (
    id          BIGSERIAL PRIMARY KEY,
    world_id    BIGINT NOT NULL REFERENCES worlds(id),
    player_id   BIGINT NOT NULL REFERENCES players(id),
    item_def    TEXT NOT NULL,
    hero_id     BIGINT REFERENCES heroes(id),
    slot        TEXT,
    source      TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX items_player ON items(player_id);
CREATE UNIQUE INDEX items_hero_slot ON items(hero_id, slot) WHERE hero_id IS NOT NULL;

CREATE TABLE missions (
    id               BIGSERIAL PRIMARY KEY,
    world_id         BIGINT NOT NULL REFERENCES worlds(id),
    player_id        BIGINT NOT NULL REFERENCES players(id),
    from_city_id     BIGINT NOT NULL REFERENCES cities(id),
    kind             TEXT NOT NULL CHECK (kind IN ('attack','scout','transport','reinforce','settle','occupy')),
    target_x         INT NOT NULL,
    target_y         INT NOT NULL,
    target_city_id   BIGINT REFERENCES cities(id),
    units            JSONB NOT NULL DEFAULT '{}',   -- current (surviving) units
    upkeep_units     JSONB NOT NULL DEFAULT '{}',   -- units charged for upkeep until return
    hero_id          BIGINT REFERENCES heroes(id),
    cargo            JSONB NOT NULL DEFAULT '{}',   -- whole units
    options          JSONB NOT NULL DEFAULT '{}',
    depart_at        TIMESTAMPTZ NOT NULL,
    arrive_at        TIMESTAMPTZ NOT NULL,
    return_at        TIMESTAMPTZ,
    status           TEXT NOT NULL CHECK (status IN ('outbound','stationed','returning','done')),
    result           JSONB NOT NULL DEFAULT '{}',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX missions_player_open ON missions(player_id) WHERE status <> 'done';
CREATE INDEX missions_from_city_open ON missions(from_city_id) WHERE status <> 'done';
CREATE INDEX missions_stationed_city ON missions(target_city_id) WHERE status = 'stationed';
CREATE INDEX missions_stationed_tile ON missions(world_id, target_x, target_y) WHERE status = 'stationed';

-- Scheduled events. Processed exactly once inside the owning player's lock.
CREATE TABLE game_events (
    id            BIGSERIAL PRIMARY KEY,
    world_id      BIGINT NOT NULL REFERENCES worlds(id),
    player_id     BIGINT NOT NULL REFERENCES players(id),
    kind          TEXT NOT NULL,
    ref_id        BIGINT NOT NULL,
    due_at        TIMESTAMPTZ NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','done','failed','cancelled')),
    attempts      INT NOT NULL DEFAULT 0,
    last_error    TEXT,
    processed_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (kind, ref_id, due_at)
);
CREATE INDEX game_events_due ON game_events(due_at) WHERE status = 'pending';
CREATE INDEX game_events_player_due ON game_events(player_id, due_at, id) WHERE status = 'pending';

CREATE TABLE reports (
    id          BIGSERIAL PRIMARY KEY,
    world_id    BIGINT NOT NULL REFERENCES worlds(id),
    kind        TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    data        JSONB NOT NULL
);
CREATE TABLE report_recipients (
    report_id  BIGINT NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    player_id  BIGINT NOT NULL REFERENCES players(id),
    side       TEXT NOT NULL,
    read_at    TIMESTAMPTZ,
    PRIMARY KEY (report_id, player_id)
);
CREATE INDEX report_recipients_player ON report_recipients(player_id, report_id DESC);

CREATE TABLE notifications (
    id          BIGSERIAL PRIMARY KEY,
    player_id   BIGINT NOT NULL REFERENCES players(id),
    kind        TEXT NOT NULL,
    data        JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    read_at     TIMESTAMPTZ
);
CREATE INDEX notifications_player ON notifications(player_id, id DESC);

CREATE TABLE market_orders (
    id           BIGSERIAL PRIMARY KEY,
    world_id     BIGINT NOT NULL REFERENCES worlds(id),
    player_id    BIGINT NOT NULL REFERENCES players(id),
    city_id      BIGINT NOT NULL REFERENCES cities(id),
    sell_res     TEXT NOT NULL,
    sell_amount  BIGINT NOT NULL CHECK (sell_amount > 0),
    buy_res      TEXT NOT NULL,
    buy_amount   BIGINT NOT NULL CHECK (buy_amount > 0),
    status       TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','filled','cancelled')),
    buyer_player_id BIGINT REFERENCES players(id),
    buyer_city_id   BIGINT REFERENCES cities(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    closed_at    TIMESTAMPTZ,
    CHECK (sell_res <> buy_res)
);
CREATE INDEX market_open ON market_orders(world_id, sell_res, buy_res) WHERE status = 'open';
CREATE INDEX market_player ON market_orders(player_id, status);

-- Resources delivered asynchronously to a city (market proceeds, admin grants).
CREATE TABLE deliveries (
    id          BIGSERIAL PRIMARY KEY,
    world_id    BIGINT NOT NULL REFERENCES worlds(id),
    player_id   BIGINT NOT NULL REFERENCES players(id),
    city_id     BIGINT NOT NULL REFERENCES cities(id),
    amounts     JSONB NOT NULL,
    reason      TEXT NOT NULL,
    ref         TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE alliance_invites (
    id           BIGSERIAL PRIMARY KEY,
    alliance_id  BIGINT NOT NULL REFERENCES alliances(id),
    player_id    BIGINT NOT NULL REFERENCES players(id),
    invited_by   BIGINT NOT NULL REFERENCES players(id),
    status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','accepted','declined','revoked')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX alliance_invites_pending ON alliance_invites(alliance_id, player_id) WHERE status = 'pending';

CREATE TABLE alliance_relations (
    alliance_id  BIGINT NOT NULL REFERENCES alliances(id),
    other_id     BIGINT NOT NULL REFERENCES alliances(id),
    status       TEXT NOT NULL CHECK (status IN ('ally','nap','war')),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (alliance_id, other_id),
    CHECK (alliance_id <> other_id)
);

CREATE TABLE messages (
    id            BIGSERIAL PRIMARY KEY,
    world_id      BIGINT NOT NULL REFERENCES worlds(id),
    from_player   BIGINT NOT NULL REFERENCES players(id),
    to_player     BIGINT REFERENCES players(id),
    alliance_id   BIGINT REFERENCES alliances(id),
    subject       TEXT NOT NULL,
    body          TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    read_at       TIMESTAMPTZ,
    reported_at   TIMESTAMPTZ,
    CHECK ((to_player IS NULL) <> (alliance_id IS NULL))
);
CREATE INDEX messages_to ON messages(to_player, id DESC);
CREATE INDEX messages_alliance ON messages(alliance_id, id DESC);

CREATE TABLE player_blocks (
    player_id   BIGINT NOT NULL REFERENCES players(id),
    blocked_id  BIGINT NOT NULL REFERENCES players(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (player_id, blocked_id)
);

CREATE TABLE quest_progress (
    player_id   BIGINT NOT NULL REFERENCES players(id),
    quest_id    TEXT NOT NULL,
    claimed_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (player_id, quest_id)   -- one claim per quest: duplicate rewards impossible
);

-- Append-only audit of every non-production resource change.
CREATE TABLE resource_ledger (
    id          BIGSERIAL PRIMARY KEY,
    world_id    BIGINT NOT NULL,
    player_id   BIGINT NOT NULL,
    city_id     BIGINT NOT NULL,
    reason      TEXT NOT NULL,
    ref         TEXT NOT NULL,
    delta       JSONB NOT NULL,        -- resourceId -> whole units (+/-)
    at          TIMESTAMPTZ NOT NULL
);
CREATE INDEX resource_ledger_city ON resource_ledger(city_id, id DESC);

CREATE TABLE idempotency_keys (
    account_id    BIGINT NOT NULL,
    key           TEXT NOT NULL,
    request_hash  TEXT NOT NULL,
    status_code   INT NOT NULL DEFAULT 0,
    response      BYTEA,                 -- exact bytes replayed to duplicates
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (account_id, key)
);
CREATE INDEX idempotency_created ON idempotency_keys(created_at);

CREATE TABLE admin_audit (
    id          BIGSERIAL PRIMARY KEY,
    admin_id    BIGINT NOT NULL REFERENCES accounts(id),
    action      TEXT NOT NULL,
    target      TEXT NOT NULL,
    payload     JSONB NOT NULL,
    at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

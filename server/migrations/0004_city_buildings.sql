-- 0004_city_buildings (ML-B02): normalized city cells, building instances and the build queue.
-- Append-only: never edit after release. Pack ids (building id, obstacle kind) are text; pack numbers are never copied.
-- Legacy cities.buildings (jsonb) and jobs (kind='build') stay in place and are copied by the backfill below.
-- Deprecation plan: docs/modernization/DB_DESIGN.md section 6.

-- Composite-FK targets so child rows must carry the same owner as the parent city.
ALTER TABLE cities ADD CONSTRAINT cities_id_player_uniq UNIQUE (id, player_id);

CREATE TABLE city_cells (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    city_id        BIGINT NOT NULL REFERENCES cities(id) ON DELETE CASCADE,
    cell_index     INT NOT NULL CHECK (cell_index >= 0),  -- upper bound (pack cell count) is checked by the server
    obstacle_kind  TEXT CHECK (obstacle_kind IS NULL OR obstacle_kind ~ '^[a-z][a-z0-9_.]{0,63}$'),  -- NULL = free cell
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (city_id, cell_index),
    UNIQUE (id, city_id)
);
COMMENT ON TABLE city_cells IS 'Cell of a city plan. cell_index is unique per city. obstacle_kind is a pack terrain-obstacle id or NULL.';

CREATE TABLE city_buildings (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    city_id      BIGINT NOT NULL,
    cell_id      BIGINT NOT NULL UNIQUE,  -- one building per cell
    building_id  TEXT NOT NULL CHECK (building_id ~ '^[a-z][a-z0-9_.]{0,63}$'),  -- Game Pack building id
    level        INT NOT NULL CHECK (level BETWEEN 1 AND 9),
    status       TEXT NOT NULL DEFAULT 'idle' CHECK (status IN ('idle','constructing','upgrading')),
    finished_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (cell_id, city_id) REFERENCES city_cells(id, city_id) ON DELETE CASCADE,
    CHECK ((status = 'idle') = (finished_at IS NULL)),
    UNIQUE (id, city_id)
);
CREATE INDEX city_buildings_city ON city_buildings(city_id, building_id);
COMMENT ON TABLE city_buildings IS 'Building instance on a cell. level = built level (constructing: the level under construction; upgrading: level before the upgrade). Costs and times stay in the pack.';

CREATE TABLE construction_queue (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    player_id     BIGINT NOT NULL,
    city_id       BIGINT NOT NULL,
    building_id   BIGINT NOT NULL,
    target_level  INT NOT NULL CHECK (target_level BETWEEN 1 AND 9),
    position      INT NOT NULL CHECK (position BETWEEN 1 AND 4),
    status        TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued','done','cancelled')),
    started_at    TIMESTAMPTZ NOT NULL,
    finishes_at   TIMESTAMPTZ NOT NULL,
    job_id        BIGINT REFERENCES jobs(id) ON DELETE SET NULL,  -- legacy job this row mirrors
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (city_id, player_id) REFERENCES cities(id, player_id) ON DELETE CASCADE,
    FOREIGN KEY (building_id, city_id) REFERENCES city_buildings(id, city_id) ON DELETE CASCADE,
    CHECK (finishes_at > started_at)
);
-- Four queue positions per player. Closed rows free their position.
CREATE UNIQUE INDEX construction_queue_position ON construction_queue(player_id, position) WHERE status = 'queued';
CREATE UNIQUE INDEX construction_queue_job ON construction_queue(job_id) WHERE job_id IS NOT NULL;
CREATE INDEX construction_queue_building ON construction_queue(building_id) WHERE status = 'queued';
CREATE INDEX construction_queue_city ON construction_queue(city_id);
CREATE INDEX construction_queue_due ON construction_queue(finishes_at) WHERE status = 'queued';
COMMENT ON TABLE construction_queue IS 'Construction and upgrade queue: positions 1..4 per player while status=queued. Spending goes through resource_ledger.';

-- ---------------------------------------------------------------------------
-- Backfill from legacy data. Legacy cities have no layout, so cells are numbered 0.. by building id.
-- Only levels 1..9 are copied. Level 0 means "not built". Higher levels cannot be stored and are skipped.
-- ---------------------------------------------------------------------------
WITH legacy AS (
    SELECT c.id AS city_id, d.key AS building_id, d.value::int AS level,
           row_number() OVER (PARTITION BY c.id ORDER BY d.key) - 1 AS idx
    FROM cities c, jsonb_each_text(c.buildings) d
    WHERE d.value ~ '^[0-9]{1,9}$' AND d.value::int BETWEEN 1 AND 9 AND d.key ~ '^[a-z][a-z0-9_.]{0,63}$'
), ins_cells AS (
    INSERT INTO city_cells(city_id, cell_index) SELECT city_id, idx FROM legacy RETURNING id, city_id, cell_index
)
INSERT INTO city_buildings(city_id, cell_id, building_id, level)
SELECT l.city_id, k.id, l.building_id, l.level FROM legacy l JOIN ins_cells k ON k.city_id = l.city_id AND k.cell_index = l.idx;

-- First four queued build jobs per player become queue rows.
CREATE TEMP TABLE _bq ON COMMIT DROP AS
SELECT * FROM (
    SELECT j.id AS job_id, j.player_id, j.city_id, j.item_id, j.level, j.starts_at, j.finishes_at,
           row_number() OVER (PARTITION BY j.player_id ORDER BY j.finishes_at, j.id) AS position
    FROM jobs j
    WHERE j.kind = 'build' AND j.status = 'queued' AND j.level BETWEEN 1 AND 9
      AND j.item_id ~ '^[a-z][a-z0-9_.]{0,63}$' AND j.finishes_at > j.starts_at
) q WHERE position <= 4;

-- Buildings that exist only as a queued job get a new cell and the status "constructing".
WITH missing AS (
    SELECT DISTINCT ON (q.city_id, q.item_id) q.city_id, q.item_id, q.level,
           (SELECT min(f.finishes_at) FROM _bq f WHERE f.city_id = q.city_id AND f.item_id = q.item_id) AS fin
    FROM _bq q
    WHERE NOT EXISTS (SELECT 1 FROM city_buildings b WHERE b.city_id = q.city_id AND b.building_id = q.item_id)
    ORDER BY q.city_id, q.item_id, q.level
), numbered AS (
    SELECT m.*, COALESCE((SELECT max(cell_index) FROM city_cells k WHERE k.city_id = m.city_id), -1)
           + row_number() OVER (PARTITION BY m.city_id ORDER BY m.item_id) AS idx
    FROM missing m
), ins_cells AS (
    INSERT INTO city_cells(city_id, cell_index) SELECT city_id, idx FROM numbered RETURNING id, city_id, cell_index
)
INSERT INTO city_buildings(city_id, cell_id, building_id, level, status, finished_at)
SELECT n.city_id, k.id, n.item_id, n.level, 'constructing', n.fin
FROM numbered n JOIN ins_cells k ON k.city_id = n.city_id AND k.cell_index = n.idx;

UPDATE city_buildings b SET status = 'upgrading',
       finished_at = (SELECT min(q.finishes_at) FROM _bq q WHERE q.city_id = b.city_id AND q.item_id = b.building_id)
WHERE b.status = 'idle' AND EXISTS (SELECT 1 FROM _bq q WHERE q.city_id = b.city_id AND q.item_id = b.building_id);

INSERT INTO construction_queue(player_id, city_id, building_id, target_level, position, started_at, finishes_at, job_id)
SELECT q.player_id, q.city_id, b.id, q.level, q.position, q.starts_at, q.finishes_at, q.job_id
FROM _bq q JOIN city_buildings b ON b.city_id = q.city_id AND b.building_id = q.item_id;

-- Deprecated, kept for released code: cities.buildings and jobs rows with kind='build'.
COMMENT ON COLUMN cities.buildings IS 'DEPRECATED (0004): use city_buildings. Still written by released code; remove after the Go cutover.';

-- 0005_army (ML-B03): unit groups per city slot and hero placement.
-- Append-only. Unit and category ids are Game Pack ids stored as text. No pack numbers are copied.
-- Legacy cities.units (jsonb) and missions.units stay in place. See DB_DESIGN.md section 6.

-- Composite-FK targets so army rows carry the same owner as the parent rows.
ALTER TABLE heroes   ADD CONSTRAINT heroes_id_player_uniq   UNIQUE (id, player_id);
ALTER TABLE missions ADD CONSTRAINT missions_id_player_uniq UNIQUE (id, player_id);

CREATE TABLE unit_groups (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    player_id   BIGINT NOT NULL,
    city_id     BIGINT NOT NULL,  -- owner city (home city for mission groups)
    mission_id  BIGINT,
    slot_kind   TEXT NOT NULL CHECK (slot_kind IN ('garrison','shelter','mission')),
    slot_index  INT NOT NULL CHECK (slot_index >= 0),
    category    TEXT NOT NULL CHECK (category ~ '^[a-z][a-z0-9_.]{0,63}$'),  -- pack unit category id
    unit_id     TEXT NOT NULL CHECK (unit_id ~ '^[a-z][a-z0-9_.]{0,63}$'),   -- pack unit id
    level       INT NOT NULL CHECK (level >= 1),                              -- upper bound comes from the pack
    count       BIGINT NOT NULL CHECK (count >= 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (city_id, player_id)    REFERENCES cities(id, player_id)   ON DELETE CASCADE,
    FOREIGN KEY (mission_id, player_id) REFERENCES missions(id, player_id) ON DELETE CASCADE,
    CHECK ((slot_kind = 'mission') = (mission_id IS NOT NULL))
);
-- One group per slot. Garrison and shelter slots are per city. Mission slots are per mission.
CREATE UNIQUE INDEX unit_groups_garrison_slot ON unit_groups(city_id, slot_index) WHERE slot_kind = 'garrison';
CREATE UNIQUE INDEX unit_groups_shelter_slot  ON unit_groups(city_id, slot_index) WHERE slot_kind = 'shelter';
CREATE UNIQUE INDEX unit_groups_mission_slot  ON unit_groups(mission_id, slot_index) WHERE slot_kind = 'mission';
CREATE INDEX unit_groups_city ON unit_groups(city_id, slot_kind);
CREATE INDEX unit_groups_player ON unit_groups(player_id);
COMMENT ON TABLE unit_groups IS 'Unit group in one slot: garrison and shelter slots belong to a city, mission slots to a mission. count may be 0 (empty slot).';

CREATE TABLE hero_placements (
    hero_id     BIGINT PRIMARY KEY,  -- a hero is in exactly one place
    player_id   BIGINT NOT NULL,
    placement   TEXT NOT NULL CHECK (placement IN ('garrison','shelter','mission')),
    city_id     BIGINT,
    mission_id  BIGINT,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (hero_id, player_id)    REFERENCES heroes(id, player_id)   ON DELETE CASCADE,
    FOREIGN KEY (city_id, player_id)    REFERENCES cities(id, player_id)   ON DELETE CASCADE,
    FOREIGN KEY (mission_id, player_id) REFERENCES missions(id, player_id) ON DELETE CASCADE,
    CHECK ((placement IN ('garrison','shelter') AND city_id IS NOT NULL AND mission_id IS NULL)
        OR (placement = 'mission' AND mission_id IS NOT NULL AND city_id IS NULL))
);
CREATE INDEX hero_placements_city ON hero_placements(city_id) WHERE city_id IS NOT NULL;
CREATE INDEX hero_placements_mission ON hero_placements(mission_id) WHERE mission_id IS NOT NULL;
CREATE INDEX hero_placements_player ON hero_placements(player_id);
COMMENT ON TABLE hero_placements IS 'Where a hero is: city garrison, city shelter or an open mission. References the heroes table.';

-- Backfill hero placement. Heroes on a mission with an open mission row go to it. All others sit in the home city garrison.
INSERT INTO hero_placements(hero_id, player_id, placement, city_id, mission_id)
SELECT h.id, h.player_id,
       CASE WHEN m.id IS NULL THEN 'garrison' ELSE 'mission' END,
       CASE WHEN m.id IS NULL THEN h.home_city_id END,
       m.id
FROM heroes h
LEFT JOIN LATERAL (
    SELECT id FROM missions WHERE hero_id = h.id AND status <> 'done' AND player_id = h.player_id ORDER BY id DESC LIMIT 1
) m ON h.status = 'mission'
WHERE EXISTS (SELECT 1 FROM cities c WHERE c.id = h.home_city_id AND c.player_id = h.player_id);

-- No unit_groups backfill: legacy cities.units holds only unit ids. Category and level come from the pack,
-- so the server fills unit_groups from the pack when the army code moves to the new tables.
COMMENT ON COLUMN cities.units IS 'DEPRECATED (0005): use unit_groups (garrison slots). Still written by released code; remove after the Go cutover.';

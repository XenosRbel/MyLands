-- 0008_mission_participants (ML-B06): mission participants as rows, sender-city ownership, report-to-mission link.
-- Append-only. missions.* columns stay. Nothing from the Game Pack is copied: unit lists remain in missions.units
-- and unit_groups (0005). See DB_DESIGN.md section 8.

-- Composite-FK targets.
ALTER TABLE missions ADD CONSTRAINT missions_id_world_uniq UNIQUE (id, world_id);

-- The sending city must belong to the sending player. NOT VALID: new rows are checked, legacy rows are not rescanned.
ALTER TABLE missions ADD CONSTRAINT missions_from_city_owner_fk
    FOREIGN KEY (from_city_id, player_id) REFERENCES cities(id, player_id) NOT VALID;

CREATE TABLE mission_participants (
    mission_id  BIGINT NOT NULL,
    player_id   BIGINT NOT NULL,
    world_id    BIGINT NOT NULL,
    role        TEXT NOT NULL CHECK (role IN ('initiator','joined','support')),
    city_id     BIGINT NOT NULL,  -- the participant's own sending city
    depart_at   TIMESTAMPTZ NOT NULL,
    arrive_at   TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (mission_id, player_id),  -- a player takes part once per mission
    CHECK (arrive_at >= depart_at),
    FOREIGN KEY (mission_id, world_id) REFERENCES missions(id, world_id)  ON DELETE CASCADE,
    FOREIGN KEY (player_id, world_id)  REFERENCES players(id, world_id)   ON DELETE CASCADE,
    FOREIGN KEY (city_id, player_id)   REFERENCES cities(id, player_id)   ON DELETE CASCADE
);
CREATE UNIQUE INDEX mission_participants_one_initiator ON mission_participants(mission_id) WHERE role = 'initiator';
CREATE INDEX mission_participants_player ON mission_participants(player_id, mission_id DESC);
COMMENT ON TABLE mission_participants IS 'One row per player in a mission: initiator (the sender in missions.player_id), joined (joint attack) or support (reinforcement). arrive_at >= depart_at.';

CREATE FUNCTION mission_participants_initiator_guard() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE m missions%ROWTYPE;
BEGIN
    IF NEW.role = 'initiator' THEN
        SELECT * INTO m FROM missions WHERE id = NEW.mission_id;
        IF m.player_id <> NEW.player_id THEN
            RAISE EXCEPTION 'initiator of mission % must be its sender %', NEW.mission_id, m.player_id USING ERRCODE = 'check_violation';
        END IF;
    ELSIF TG_OP = 'UPDATE' AND OLD.role = 'initiator' THEN
        RAISE EXCEPTION 'the initiator row of mission % cannot change role', NEW.mission_id USING ERRCODE = 'restrict_violation';
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER mission_participants_initiator_guard BEFORE INSERT OR UPDATE OF role, player_id ON mission_participants
    FOR EACH ROW EXECUTE FUNCTION mission_participants_initiator_guard();

-- Backfill and keep in step: every mission has its initiator row. Released code inserts missions only.
INSERT INTO mission_participants(mission_id, player_id, world_id, role, city_id, depart_at, arrive_at)
SELECT id, player_id, world_id, 'initiator', from_city_id, depart_at, arrive_at FROM missions;

CREATE FUNCTION missions_sync_initiator() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    INSERT INTO mission_participants(mission_id, player_id, world_id, role, city_id, depart_at, arrive_at)
    VALUES (NEW.id, NEW.player_id, NEW.world_id, 'initiator', NEW.from_city_id, NEW.depart_at, NEW.arrive_at);
    RETURN NULL;
END $$;
CREATE TRIGGER missions_sync_initiator AFTER INSERT ON missions
    FOR EACH ROW EXECUTE FUNCTION missions_sync_initiator();

-- ---------------------------------------------------------------------------
-- Reports linked to missions. A report may cover several missions (attacker, supporters), a mission may have several reports.
-- reports.data does not carry a mission id, so there is nothing to backfill.
-- ---------------------------------------------------------------------------
CREATE TABLE mission_reports (
    report_id   BIGINT NOT NULL REFERENCES reports(id)  ON DELETE CASCADE,
    mission_id  BIGINT NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
    PRIMARY KEY (report_id, mission_id)
);
CREATE INDEX mission_reports_mission ON mission_reports(mission_id);
COMMENT ON TABLE mission_reports IS 'Link between a report and the missions it describes. Report text stays in reports.data.';

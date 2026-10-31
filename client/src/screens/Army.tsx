import { useState } from "react";
import { command } from "../api/client";
import { useGame } from "../game/context";
import { Action, CostList, duration, fmt, Icon, Reqs, Empty } from "../ui/kit";
import { QueuePanel, blockedLabel } from "./City";

export function ArmyScreen() {
  const { city, idx, t, refresh, toast, world } = useGame();
  const [qty, setQty] = useState<Record<string, number>>({});
  if (!city) return null;
  const maxAfford = (cost: Record<string, number> = {}) =>
    Math.max(0, Math.min(...Object.entries(cost).filter(([, n]) => n > 0).map(([r, n]) => Math.floor((city.resources[r] ?? 0) / n)), 100000));
  return (
    <div className="screen">
      <h1>{t("army")}</h1>
      <QueuePanel jobs={city.queues.train} title={t("queue")} />
      <section className="panel">
        <div className="units">
          {city.units.map((uv) => {
            const u = idx.unit(uv.id)!;
            const q = qty[uv.id] ?? 1;
            const m = maxAfford(u.cost);
            return (
              <article key={uv.id} className="ucard">
                <header>
                  {idx.art(uv.id) && <img className="portrait" src={idx.art(uv.id)} alt="" loading="lazy" draggable={false} />}
                  <div className="grow">
                    <strong>{idx.name(uv.id)}</strong>
                    <div className="muted small">{u.class}</div>
                  </div>
                </header>
                <p className="small muted">{idx.desc(uv.id)}</p>
                <div className="stats small" aria-label="Stats">
                  <span title={t("attack")}><Icon name="sword" size={14} /> {u.attack}</span>
                  <span title={t("defense")}><Icon name="shield" size={14} /> {u.defense}%</span>
                  <span title={t("hp")}><Icon name="heart" size={14} /> {u.hp}</span>
                  <span title={t("move")}><Icon name="boot" size={14} /> {u.speed}</span>
                  <span title={t("carry")}><Icon name="bag" size={14} /> {u.carry}</span>
                  {u.scout ? <span title="Scout"><Icon name="eye" size={14} /> {u.scout}</span> : null}
                </div>
                {u.bonuses?.length ? <div className="small">⚔ {u.bonuses.map((b) => `+${b.pct}% vs ${b.vsClass}`).join(", ")}</div> : null}
                <div className="small">{t("upkeep")}: {u.upkeep}{t("perHour")}</div>
                <div className="row small">
                  <span>{t("home")}: <strong>{fmt(city.city.units[uv.id] ?? 0)}</strong></span>
                  {city.unitsAway[uv.id] ? <span>{t("away")}: {fmt(city.unitsAway[uv.id])}</span> : null}
                </div>
                <div className="row small"><CostList cost={u.cost ?? {}} have={city.resources} /> <span>{duration(uv.secondsEach)}</span></div>
                <Reqs reqs={uv.unmet} />
                <div className="row">
                  <input type="number" min={1} max={100000} value={q} aria-label={t("quantity")}
                    onChange={(e) => setQty({ ...qty, [uv.id]: Math.max(1, Math.min(100000, Number(e.target.value) || 1)) })} />
                  <button className="btn small ghost" onClick={() => setQty({ ...qty, [uv.id]: Math.max(1, m) })}>{t("max")} ({m})</button>
                  <Action className="primary small" disabled={!uv.trainable || m < q} title={uv.blocked ? blockedLabel(uv.blocked, t) : undefined} run={async () => {
                    await command(`/api/v1/cities/${city.city.id}/train`, { unit: uv.id, quantity: q });
                    toast(`${idx.name(uv.id)} ×${q}`, "ok");
                    await refresh();
                  }}>{t("train")}</Action>
                </div>
              </article>
            );
          })}
        </div>
      </section>
      <section className="panel">
        <h3>{t("foreignTroops")}</h3>
        {!city.foreignTroops.length ? <Empty /> : (
          <ul className="list">
            {city.foreignTroops.map((g) => (
              <li key={g.missionId}><strong>{g.player}</strong> {Object.entries(g.units).map(([u, n]) => `${idx.name(u)} ${n}`).join(", ")}</li>
            ))}
          </ul>
        )}
        <p className="small muted">{world.name}</p>
      </section>
    </div>
  );
}

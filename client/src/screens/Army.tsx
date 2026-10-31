import { useState } from "react";
import { command } from "../api/client";
import { useGame } from "../game/context";
import { Action, CostList, duration, fmt, Icon, Reqs, Empty } from "../ui/kit";
import { QueuePanel, blockedLabel } from "./City";

export function ArmyScreen() {
  const { city, idx, t, refresh, toast, world } = useGame();
  const [qty, setQty] = useState<Record<string, number>>({});
  const [selected, setSelected] = useState<string | null>(null);
  if (!city) return null;
  const maxAfford = (cost: Record<string, number> = {}) =>
    Math.max(0, Math.min(...Object.entries(cost).filter(([, n]) => n > 0)
      .map(([r, n]) => Math.floor((city.resources[r] ?? 0) / n)), 100000));
  const current = city.units.find((u) => u.id === selected) ?? city.units[0];
  const details = current ? idx.unit(current.id) : undefined;
  const quantity = current ? qty[current.id] ?? 1 : 1;
  const available = maxAfford(details?.cost);
  return (
    <div className="screen army-game-screen">
      <h1>{t("army")}</h1>
      <div className="army-game-layout">
        <section className="panel army-game-roster">
          <h3><Icon name="shield" size={19} /> {t("army")}</h3>
          <div className="army-game-unit-list" role="group" aria-label={t("army")}>
            {city.units.map((view) => (
              <button key={view.id} className={"army-game-unit-row" + (current?.id === view.id ? " selected" : "")}
                aria-pressed={current?.id === view.id} onClick={() => setSelected(view.id)}>
                <span className="army-game-unit-mini">
                  {idx.art(view.id) ? <img src={idx.art(view.id)} alt="" loading="lazy" /> : <Icon name="shield" size={24} />}
                </span>
                <span className="army-game-unit-name">
                  <strong>{idx.name(view.id)}</strong><small>{idx.unit(view.id)?.class}</small>
                </span>
                <span className="army-game-unit-number">{fmt(city.city.units[view.id] ?? 0)}</span>
              </button>
            ))}
          </div>
        </section>
        <div className="army-game-details">
          {current && details ? (
            <section className="panel army-game-selected">
              <header className="army-game-selected-title">
                <Icon name="sword" size={20} /> <h2>{idx.name(current.id)}</h2>
                <span className="tag">{details.class}</span>
              </header>
              <div className="army-game-portrait-row">
                <div className="army-game-main-art">
                  {idx.art(current.id) ? <img src={idx.art(current.id)} alt={idx.name(current.id)} draggable={false} /> : <Icon name="shield" size={62} />}
                </div>
                <div className="army-game-description">
                  <p>{idx.desc(current.id)}</p>
                  <div className="army-game-availability">
                    <div><small>{t("home")}</small><strong>{fmt(city.city.units[current.id] ?? 0)}</strong></div>
                    <div><small>{t("away")}</small><strong>{fmt(city.unitsAway[current.id] ?? 0)}</strong></div>
                    <div><small>{t("upkeep")}</small><strong>{details.upkeep}{t("perHour")}</strong></div>
                  </div>
                </div>
              </div>
              <div className="army-game-stat-grid">
                {([
                  ["sword",t("attack"), details.attack],
                  ["shield",t("defense"), String(details.defense)+"%"],
                  ["heart",t("hp"), details.hp],
                  ["boot",t("move"), details.speed],
                  ["bag",t("carry"), details.carry],
                ] as const).map(([icon,label,value]) => (
                  <div key={label} className="army-game-stat">
                    <Icon name={icon} size={22} /><small>{label}</small><strong>{value}</strong>
                  </div>
                ))}
                {details.scout ? <div className="army-game-stat"><Icon name="eye" size={22} /><small>{t("scout")}</small><strong>{details.scout}</strong></div> : null}
              </div>
              {details.bonuses?.length ? (
                <div className="army-game-bonuses">
                  {details.bonuses.map((bonus,i) => <span key={i} className="tag"><Icon name="sword" size={13} /> +{bonus.pct}% {bonus.vsClass}</span>)}
                </div>
              ) : null}
              <div className="army-game-train">
                <h3>{t("train")}</h3>
                <div className="row wrap">
                  <CostList cost={details.cost ?? {}} have={city.resources} />
                  <span className="small"><Icon name="clock" size={15} /> {duration(current.secondsEach)}</span>
                </div>
                <Reqs reqs={current.unmet} />
                <div className="row wrap">
                  <input type="number" min={1} max={100000} value={quantity} aria-label={t("quantity")}
                    onChange={(event) => setQty({ ...qty, [current.id]: Math.max(1, Math.min(100000, Number(event.target.value) || 1)) })} />
                  <button className="btn small ghost" onClick={() => setQty({ ...qty, [current.id]: Math.max(1,available) })}>
                    {t("max")} ({available})
                  </button>
                  <Action className="primary" disabled={!current.trainable || available < quantity}
                    title={current.blocked ? blockedLabel(current.blocked, t) : undefined}
                    run={async () => {
                      await command(`/api/v1/cities/${city.city.id}/train`, { unit: current.id, quantity });
                      toast(`${idx.name(current.id)} ×${quantity}`, "ok");
                      await refresh();
                    }}>{t("train")}</Action>
                </div>
              </div>
            </section>
          ) : <Empty />}
        </div>
        <aside className="army-game-side">
          <QueuePanel jobs={city.queues.train} title={t("queue")} />
          <section className="panel">
            <h3>{t("foreignTroops")}</h3>
            {!city.foreignTroops.length ? <Empty /> : (
              <ul className="list">{city.foreignTroops.map((group) => (
                <li key={group.missionId}>
                  <strong>{group.player}</strong>
                  {Object.entries(group.units).map(([id,n]) => `${idx.name(id)} ${n}`).join(", ")}
                </li>
              ))}</ul>
            )}
            <p className="small muted">{world.name}</p>
          </section>
        </aside>
      </div>
    </div>
  );
}

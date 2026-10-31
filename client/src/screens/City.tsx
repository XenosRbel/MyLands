import { useState } from "react";
import { command } from "../api/client";
import type { BuildingView, Job } from "../api/types";
import { useGame } from "../game/context";
import { describeModifier } from "../game/pack";
import { Action, CostList, duration, Icon, Modal, Reqs, Timer, fmt } from "../ui/kit";

const catIcon: Record<string, string> = { core: "castle", economy: "wheat", storage: "bag", military: "sword", science: "book", special: "star" };

export function QueuePanel({ jobs, title }: { jobs: Job[]; title: string }) {
  const { idx, t, city, refresh, toast } = useGame();
  return (
    <section className="panel">
      <h3>{title}</h3>
      {!jobs.length ? (
        <p className="muted small">{t("queueEmpty")}</p>
      ) : (
        <ol className="queue">
          {jobs.map((j, i) => (
            <li key={j.id}>
              <span>
                {idx.name(j.item)} {j.kind === "train" ? `×${j.quantity}` : `→ ${j.level}`}
              </span>
              {Date.parse(j.startsAt) <= Date.now() + 1000 ? <Timer until={j.finishesAt} onDone={refresh} /> : <span className="muted small">{t("queue")}</span>}
              {i === jobs.length - 1 && (
                <Action className="small ghost" confirm={t("cancel") + "?"} run={async () => {
                  const r = await command<{ refund: Record<string, number> }>(`/api/v1/cities/${city!.city.id}/jobs/cancel`, { jobId: j.id });
                  toast(t("cancel") + " ✓", "ok");
                  await refresh();
                  return r;
                }}>{t("cancel")}</Action>
              )}
            </li>
          ))}
        </ol>
      )}
    </section>
  );
}

// Isometric slots around the centre: Chebyshev rings of a 5×5 grid (24 + centre).
const SLOTS: [number, number][] = (() => {
  const out: [number, number][] = [];
  for (let r = 1; r <= 2; r++)
    for (let i = -r; i <= r; i++)
      for (let j = -r; j <= r; j++) if (Math.max(Math.abs(i), Math.abs(j)) === r) out.push([i, j]);
  // front slots first so the most-used buildings sit nearest the viewer
  return out.sort((a, b) => b[0] + b[1] - (a[0] + a[1]) || a[0] - b[0]);
})();

/** Isometric city scene: the keep in the centre, every building type on its plot. */
export function CityScene({ onOpen }: { onOpen: (b: BuildingView) => void }) {
  const { city, idx, t, refresh } = useGame();
  if (!city) return null;
  const layout = idx.pack.ui.cityLayout;
  const core = city.buildings.find((b) => idx.building(b.id)?.category === "core");
  const rest = city.buildings.filter((b) => b !== core);
  const place = (b: BuildingView, slot: [number, number] | null): { x: number; y: number } => {
    if (layout?.[b.id]) return { x: layout[b.id][0], y: layout[b.id][1] };
    const [i, j] = slot ?? [0, 0];
    return { x: 50 + (i - j) * 10.5, y: 50 + (i + j) * 7.4 };
  };
  const items = [
    ...(core ? [{ b: core, pos: place(core, null), main: true }] : []),
    ...rest.map((b, k) => ({ b, pos: place(b, SLOTS[k % SLOTS.length]), main: false })),
  ].sort((a, b) => a.pos.y - b.pos.y);
  const busy = new Map(city.queues.build.map((j) => [j.item, j]));
  const walls = Object.entries(city.city.buildings).some(([id, l]) => l > 0 && (idx.building(id)?.effects ?? []).some((e) => e.modifier?.target === "wall"));
  const ground = idx.art(idx.pack.terrains.find((x) => x.passable)?.id ?? "");
  return (
    <section className={"city-scene" + (walls ? " walled" : "")} aria-label={city.city.name}>
      <div className="ground" style={ground ? { backgroundImage: `url(${ground})` } : undefined} />
      {items.map(({ b, pos, main }) => {
        const art = idx.art(b.id);
        const job = busy.get(b.id);
        const built = b.level > 0;
        const scale = main ? 1.25 : 0.82 + Math.min(b.level, 20) * 0.012;
        return (
          <button key={b.id} className={"plot" + (built ? " built" : " empty") + (b.next && !b.next.blocked ? " can" : "")}
            style={{ left: pos.x + "%", top: pos.y + "%", "--s": scale } as React.CSSProperties}
            onClick={() => onOpen(b)} title={`${idx.name(b.id)} — ${built ? t("level") + " " + b.level : t("build")}`}>
            {built && art ? <img src={art} alt={idx.name(b.id)} draggable={false} /> : <span className="plot-mark">{built ? idx.name(b.id) : "+"}</span>}
            <span className="plot-label">{idx.name(b.id)}{built ? <b>{b.level}</b> : null}</span>
            {job && <span className="plot-busy"><Timer until={job.finishesAt} onDone={refresh} /></span>}
          </button>
        );
      })}
    </section>
  );
}

export function CityScreen() {
  const { city, idx, t } = useGame();
  const [open, setOpen] = useState<BuildingView | null>(null);
  if (!city) return null;
  const cats = Array.from(new Set(idx.pack.buildings.map((b) => b.category)));
  return (
    <div className="screen">
      <h1>{city.city.name} <span className="muted small">({city.city.x}:{city.city.y})</span></h1>
      <CityScene onOpen={setOpen} />
      <div className="grid-2">
        <div>
          {cats.map((cat) => {
            const list = city.buildings.filter((b) => idx.building(b.id)?.category === cat);
            if (!list.length) return null;
            return (
              <section key={cat} className="panel">
                <h3><Icon name={catIcon[cat] ?? "star"} /> {cat[0].toUpperCase() + cat.slice(1)}</h3>
                <div className="cards">
                  {list.map((b) => (
                    <button key={b.id} className={"bcard" + (b.level === 0 ? " unbuilt" : "") + (b.next?.blocked ? "" : " ready")} onClick={() => setOpen(b)}>
                      {idx.art(b.id) && <img className="bart" src={idx.art(b.id)} alt="" loading="lazy" draggable={false} />}
                      <div className="bname">{idx.name(b.id)}</div>
                      <div className="blevel">{b.level === 0 ? "—" : `${t("level")} ${b.level}`}{b.targetLevel > b.level ? ` → ${b.targetLevel}` : ""}</div>
                      {b.next && <div className="bstate">{b.next.blocked ? <span className={"tag " + b.next.blocked}>{blockedLabel(b.next.blocked, t)}</span> : <span className="tag ok">{b.level === 0 ? t("build") : t("upgrade")}</span>}</div>}
                    </button>
                  ))}
                </div>
              </section>
            );
          })}
        </div>
        <aside>
          <QueuePanel jobs={city.queues.build} title={t("queue") + " · " + t("build")} />
          <QueuePanel jobs={city.queues.research} title={t("queue") + " · " + t("research")} />
          <QueuePanel jobs={city.queues.train} title={t("queue") + " · " + t("train")} />
        </aside>
      </div>
      {open && <BuildingModal b={city.buildings.find((x) => x.id === open.id) ?? open} onClose={() => setOpen(null)} />}
    </div>
  );
}

export function blockedLabel(b: string, t: (k: any) => string): string {
  return ({ resources: t("cost"), requirements: t("requires"), queue_full: t("queue"), race: t("locked"), specialisation_taken: t("locked") } as Record<string, string>)[b] ?? b;
}

function BuildingModal({ b, onClose }: { b: BuildingView; onClose: () => void }) {
  const { idx, t, city, refresh, toast } = useGame();
  const def = idx.building(b.id)!;
  return (
    <Modal title={idx.name(b.id)} onClose={onClose}>
      {idx.art(b.id) && <img className="modal-art" src={idx.art(b.id)} alt="" draggable={false} />}
      <p className="muted">{idx.desc(b.id)}</p>
      <p>{t("level")}: <strong>{b.level}</strong> / {b.maxLevel}</p>
      <ul className="effects">
        {def.effects.map((e, i) => (
          <li key={i}>{effectText(e, idx, b.level || 1)}</li>
        ))}
      </ul>
      {b.next ? (
        <div className="stack">
          <div className="row"><span>{t("cost")}:</span><CostList cost={b.next.cost} have={city?.resources} /></div>
          <div className="row"><span>{t("time")}:</span><span>{duration(b.next.seconds)}</span></div>
          <Reqs reqs={b.next.unmet} />
          <Action className="primary" disabled={!!b.next.blocked} run={async () => {
            await command(`/api/v1/cities/${city!.city.id}/build`, { building: b.id });
            toast(`${idx.name(b.id)} → ${b.next!.level}`, "ok");
            await refresh();
          }}>{b.level === 0 ? t("build") : t("upgrade")} → {b.next.level}</Action>
        </div>
      ) : (
        <p className="muted">{t("maxLevel")}</p>
      )}
    </Modal>
  );
}

export function effectText(e: any, idx: any, level: number): string {
  const v = e.perLevel * level * Math.pow(e.growth || 1, level - 1);
  switch (e.type) {
    case "produce": return `${idx.name(e.resource)} +${fmt(v)}/h (${e.workers ?? 0}×${level} workers)`;
    case "storage": return `Storage ${e.resource === "*" ? "" : idx.name(e.resource)} +${fmt(v)}`;
    case "protect": return `Hidden ${e.resource === "*" ? "" : idx.name(e.resource)} ${fmt(v)}`;
    case "housing": return `Housing +${fmt(v)}`;
    case "idle_tax": return `${idx.name(e.resource)} +${e.perLevel * level}/h per idle inhabitant`;
    case "modifier": return describeModifier({ ...e.modifier, value: e.modifier.value * level }, idx);
  }
  return e.type;
}

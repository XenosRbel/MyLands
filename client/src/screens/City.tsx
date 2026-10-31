import { useEffect, useState } from "react";
import { command } from "../api/client";
import type { BuildingView, Job } from "../api/types";
import { useGame } from "../game/context";
import { describeModifier } from "../game/pack";
import { Action, CostList, duration, Icon, Modal, Reqs, Timer, fmt } from "../ui/kit";

const catIcon: Record<string, string> = { core: "castle", economy: "wheat", storage: "bag", military: "sword", science: "book", special: "star" };

export function QueuePanel({ jobs, title }: { jobs: Job[]; title: string }) {
  const { idx, t, city, refresh, toast } = useGame();
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    if (!jobs.length) return;
    const timer = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, [jobs.length]);
  const progress = (job: Job) => {
    const start = Date.parse(job.startsAt);
    const end = Date.parse(job.finishesAt);
    if (!Number.isFinite(start) || !Number.isFinite(end) || end <= start) return 0;
    return Math.min(100, Math.max(0, (now - start) * 100 / (end - start)));
  };
  return (
    <section className="panel game-queue-panel">
      <h3>{title}{jobs.length > 0 && <span className="game-queue-count">{jobs.length}</span>}</h3>
      {!jobs.length ? (
        <p className="muted small game-queue-empty">{t("queueEmpty")}</p>
      ) : (
        <ol className="queue game-queue-list">
          {jobs.map((j, i) => (
            <li className={"game-queue-entry" + (i === 0 ? " active" : "")} key={j.id}>
              <span className="game-queue-art" aria-hidden="true">
                {idx.art(j.item) ? <img src={idx.art(j.item)} alt="" loading="lazy" /> : <Icon name="clock" size={28} />}
              </span>
              <span className="game-queue-information">
                <strong>{idx.name(j.item)} {j.kind === "train" ? `×${j.quantity}` : `→ ${j.level}`}</strong>
                <span className="game-queue-time">
                  {Date.parse(j.startsAt) <= now + 1000 ?
                    <Timer until={j.finishesAt} onDone={refresh} /> : t("queue")}
                </span>
                <span className="game-queue-progress" role="progressbar" aria-label={idx.name(j.item)}
                  aria-valuenow={Math.round(progress(j))} aria-valuemin={0} aria-valuemax={100}>
                  <span style={{ width: progress(j) + "%" }} />
                </span>
              </span>
              {i === jobs.length - 1 && (
                <Action className="small ghost game-queue-cancel" confirm={t("cancel") + "?"} run={async () => {
                  const r = await command<{ refund: Record<string, number> }>(`/api/v1/cities/${city!.city.id}/jobs/cancel`, { jobId: j.id });
                  toast(t("cancel") + " ✓", "ok");
                  await refresh();
                  return r;
                }}><Icon name="close" size={16} /></Action>
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
      {items.filter(({ b }) => b.level > 0 || busy.has(b.id)).map(({ b, pos, main }) => {
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
  const { city, idx, t, go } = useGame();
  const [open, setOpen] = useState<BuildingView | null>(null);
  const [category, setCategory] = useState<string>("all");
  const [showCatalogue, setShowCatalogue] = useState(false);

  if (!city) return null;
  const cats = Array.from(new Set(city.buildings.map((b) => idx.building(b.id)?.category).filter((x): x is string => !!x)));
  const filtered = category === "all" ? city.buildings : city.buildings.filter((b) => idx.building(b.id)?.category === category);
  const activeBuildings = city.buildings.filter((b) => b.level > 0).length;
  return (
    <div className="screen city-game-screen">
      <div className="city-game-heading">
        <div className="city-game-crest"><Icon name="castle" size={30} /></div>
        <div className="city-game-identity">
          <h1>{city.city.name}</h1>
          <span className="small">{city.city.x}:{city.city.y} · {activeBuildings}/{city.buildings.length} {t("build")}</span>
        </div>
        <div className="city-game-quick-actions">
          <button className="btn small" onClick={() => go("map")}><Icon name="map" size={16} /> {t("map")}</button>
          <button className="btn small" onClick={() => go("research")}><Icon name="book" size={16} /> {t("research")}</button>
          <button className="btn small" onClick={() => go("army")}><Icon name="sword" size={16} /> {t("army")}</button>
        </div>
      </div>
      <div className="city-game-layout">
        <div className="city-game-stage">
          <div className="city-game-stage-title"><Icon name="castle" size={18} /> {city.city.name}<span className="city-game-stage-coordinates">{city.city.x}:{city.city.y}</span><button type="button" className="btn small city-build-toggle" aria-expanded={showCatalogue} onClick={() => setShowCatalogue(!showCatalogue)}><Icon name="bag" size={16}/>{t("build")}</button></div>
          <CityScene onOpen={setOpen} />
          <section className={"city-game-buildings" + (showCatalogue ? " expanded" : "")} aria-label={t("build")}>
            <div className="city-game-section-title"><Icon name="bag" size={18} /> {t("build")} <span className="small muted">{filtered.length}</span><button type="button" className="city-catalogue-close" aria-label={t("close")} onClick={() => setShowCatalogue(false)}><Icon name="close" size={18}/></button></div>
            <div className="city-game-categories" role="group" aria-label={t("build")}>
              <button className={category === "all" ? "selected" : ""} aria-pressed={category === "all"} onClick={() => setCategory("all")}>{t("all")}</button>
              {cats.map((cat) => (
                <button key={cat} className={category === cat ? "selected" : ""} aria-pressed={category === cat}
                  onClick={() => setCategory(cat)}>
                  <Icon name={catIcon[cat] ?? "star"} size={16} /> {cat}
                </button>
              ))}
            </div>
            <div className="city-game-building-list">
              {filtered.map((b) => (
                <button key={b.id} className={"city-game-building" + (b.level === 0 ? " empty-building" : "")}
                  onClick={() => setOpen(b)}>
                  <span className="city-game-building-art">
                    {idx.art(b.id) ? <img src={idx.art(b.id)} alt="" loading="lazy" /> : <Icon name="castle" size={28} />}
                  </span>
                  <span className="city-game-building-info">
                    <strong>{idx.name(b.id)}</strong>
                    <small>{b.level === 0 ? t("build") : `${t("level")} ${b.level}`}{b.targetLevel > b.level ? ` → ${b.targetLevel}` : ""}</small>
                  </span>
                  <span className="city-game-building-action">
                    {b.next && !b.next.blocked ? <Icon name="star" size={16} /> : <Icon name="book" size={16} />}
                  </span>
                </button>
              ))}
            </div>
          </section>
        </div>
        <aside className="city-game-rail">
          <div className="city-game-rail-title"><Icon name="castle" size={18} /> {t("constructionQueue")}</div>
          <QueuePanel jobs={city.queues.build} title={t("build")} />
          <div className="city-game-rail-title training"><Icon name="sword" size={18} /> {t("trainingQueue")}</div>
          <QueuePanel jobs={city.queues.train} title={t("train")} />
          <div className="city-game-rail-links">
            <button className="btn small" onClick={() => go("economy")}><Icon name="chart" size={17} /> {t("economy")}</button>
            <button className="btn small" onClick={() => go("heroes")}><Icon name="user" size={17} /> {t("heroes")}</button>
          </div>
        </aside>
      </div>
      {open && <BuildingModal b={city.buildings.find((x) => x.id === open.id) ?? open} onClose={() => setOpen(null)} />}
    </div>
  );
}

export function blockedLabel(b: string, t: (k: any) => string): string {
  return ({ resources: t("cost"), requirements: t("requires"), queue_full: t("queue"), race: t("locked"), specialisation_taken: t("locked") } as Record<string, string>)[b] ?? b;
}

function BuildingModal({ b: initialBuilding, onClose }: { b: BuildingView; onClose: () => void }) {
  const { idx, t, city, refresh, toast } = useGame();
  const [selected, setSelected] = useState(initialBuilding.id);
  const [activeCategory, setActiveCategory] = useState(() => idx.building(initialBuilding.id)?.category ?? "core");
  const b = city?.buildings.find((item) => item.id === selected) ?? initialBuilding;
  const def = idx.building(b.id)!;
  const categories = Array.from(new Set((city?.buildings ?? []).map((item) => idx.building(item.id)?.category).filter((x): x is string => !!x)));
  const collection = city?.buildings.filter((item) => idx.building(item.id)?.category === activeCategory) ?? [];
  return (
    <Modal title={t("build")} onClose={onClose} wide>
      <div className="building-catalog">
        <nav className="building-catalog-tabs" aria-label={t("build")}>
          {categories.map((cat) => (
            <button key={cat} className={activeCategory === cat ? "active" : ""}
              aria-current={activeCategory === cat ? "page" : undefined}
              onClick={() => {
                setActiveCategory(cat);
                const first = city?.buildings.find((item) => idx.building(item.id)?.category === cat);
                if (first) setSelected(first.id);
              }}>
              <Icon name={catIcon[cat] ?? "star"} size={17} />
              {cat}
            </button>
          ))}
        </nav>
        <div className="building-catalog-layout">
          <div className="building-catalog-gallery" role="group" aria-label={t("build")}>
            {collection.map((item) => (
              <button type="button" key={item.id}
                className={"building-catalog-choice" + (b.id === item.id ? " active" : "")}
                aria-pressed={b.id === item.id} title={idx.name(item.id)}
                onClick={() => setSelected(item.id)}>
                {idx.art(item.id) ?
                  <img src={idx.art(item.id)} alt="" loading="lazy" /> :
                  <Icon name="castle" size={38} />}
                <small>{idx.name(item.id)}</small>
                {item.level > 0 && <span className="building-catalog-level">{item.level}</span>}
              </button>
            ))}
          </div>
          <section className="building-catalog-detail">
            <div className="building-catalog-detail-title">
              <h3>{idx.name(b.id)}</h3>
              <span className="tag">{t("level")} {b.level}/{b.maxLevel}</span>
            </div>
            <div className="building-catalog-detail-art">
              {idx.art(b.id) ? <img src={idx.art(b.id)} alt="" draggable={false} /> : <Icon name="castle" size={64} />}
            </div>
            <p>{idx.desc(b.id)}</p>
            <ul className="effects">
              {def.effects.map((effect, i) => <li key={i}>{effectText(effect, idx, b.level || 1)}</li>)}
            </ul>
            {b.next ? (
              <div className="building-catalog-build">
                <div className="row wrap">
                  <strong>{t("cost")}</strong>
                  <CostList cost={b.next.cost} have={city?.resources} />
                </div>
                <div className="row"><Icon name="clock" size={15} /> {duration(b.next.seconds)}</div>
                <Reqs reqs={b.next.unmet} />
                <Action className="primary" disabled={!!b.next.blocked}
                  title={b.next.blocked ? blockedLabel(b.next.blocked, t) : undefined}
                  run={async () => {
                    await command(`/api/v1/cities/${city!.city.id}/build`, { building: b.id });
                    toast(`${idx.name(b.id)} → ${b.next!.level}`, "ok");
                    await refresh();
                  }}>
                  {b.level === 0 ? t("build") : t("upgrade")} → {b.next.level}
                </Action>
              </div>
            ) : <p className="muted">{t("maxLevel")}</p>}
          </section>
        </div>
      </div>
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

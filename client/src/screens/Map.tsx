import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { command, get } from "../api/client";
import type { HeroesView, MapTile } from "../api/types";
import { useGame } from "../game/context";
import { Action, ErrorBox, Icon, Modal, Res, fmt, duration } from "../ui/kit";

const VIEW = 15;

export function MapScreen() {
  const g = useGame();
  const { idx, t, world, city, mapFocus, overview } = g;
  const W = idx.pack.params.mapWidth as number, H = idx.pack.params.mapHeight as number;
  const [center, setCenter] = useState(() => mapFocus ?? { x: city?.city.x ?? 0, y: city?.city.y ?? 0 });
  const [tiles, setTiles] = useState<MapTile[]>([]);
  const [err, setErr] = useState<unknown>();
  const [sel, setSel] = useState<MapTile | null>(null);
  const [mission, setMission] = useState<string | null>(null);
  const canvas = useRef<HTMLCanvasElement>(null);
  const drag = useRef<{ x: number; y: number; cx: number; cy: number; moved: boolean } | null>(null);
  const [cell, setCell] = useState(36);
  // Terrain textures from the pack (theme.art); redraw once each image loads.
  const [texReady, setTexReady] = useState(0);
  const textures = useMemo(() => {
    const m = new Map<string, HTMLImageElement>();
    for (const ter of [...idx.pack.terrains, ...idx.pack.neutrals]) {
      const url = idx.art(ter.id);
      if (!url) continue;
      const img = new Image();
      img.onload = () => setTexReady((n) => n + 1);
      img.src = url;
      m.set(ter.id, img);
    }
    return m;
  }, [idx]);

  const origin = useMemo(() => ({
    x: Math.max(0, Math.min(W - VIEW, center.x - Math.floor(VIEW / 2))),
    y: Math.max(0, Math.min(H - VIEW, center.y - Math.floor(VIEW / 2))),
  }), [center, W, H]);

  const load = useCallback(() => {
    get<MapTile[]>(`/api/v1/worlds/${world.id}/map?x=${origin.x}&y=${origin.y}&w=${VIEW}&h=${VIEW}`).then((ts) => { setTiles(ts); setErr(undefined); }, setErr);
  }, [world.id, origin.x, origin.y]);
  useEffect(load, [load]);

  const wrap = useRef<HTMLDivElement>(null);
  useEffect(() => {
    // Size cells from the container (not the window) so the map never overflows.
    const el = wrap.current;
    if (!el) return;
    const fit = () => setCell(Math.max(16, Math.floor(Math.min(el.clientWidth, 640) / VIEW)));
    fit();
    const ro = new ResizeObserver(fit);
    ro.observe(el);
    return () => ro.disconnect();
  }, []);

  // draw
  useEffect(() => {
    const c = canvas.current;
    if (!c) return;
    const dpr = window.devicePixelRatio || 1;
    c.width = VIEW * cell * dpr;
    c.height = VIEW * cell * dpr;
    const ctx = c.getContext("2d")!;
    ctx.scale(dpr, dpr);
    const css = getComputedStyle(document.documentElement);
    const accent = css.getPropertyValue("--accent").trim() || "#e0a44a";
    const textC = css.getPropertyValue("--text").trim() || "#fff";
    for (const tile of tiles) {
      const px = (tile.x - origin.x) * cell, py = (tile.y - origin.y) * cell;
      const ter = idx.terrain(tile.terrain);
      const tex = textures.get(tile.terrain);
      if (tex?.complete && tex.naturalWidth) ctx.drawImage(tex, px, py, cell, cell);
      else {
        ctx.fillStyle = ter?.color ?? "#444";
        ctx.fillRect(px, py, cell, cell);
      }
      ctx.strokeStyle = "rgba(0,0,0,.18)";
      ctx.strokeRect(px + 0.5, py + 0.5, cell - 1, cell - 1);
      if (tile.ownerId && !tile.cityId) {
        const race = idx.pack.races.find((r) => r.id === tile.race);
        ctx.strokeStyle = tile.ownerId === overview?.player.id ? accent : race?.color ?? "#fff";
        ctx.lineWidth = 2;
        ctx.setLineDash([4, 3]);
        ctx.strokeRect(px + 2, py + 2, cell - 4, cell - 4);
        ctx.setLineDash([]);
        ctx.lineWidth = 1;
      }
      if (tile.cityId) {
        const own = tile.ownerId === overview?.player.id;
        const race = idx.pack.races.find((r) => r.id === tile.race);
        ctx.fillStyle = own ? accent : race?.color ?? "#ccc";
        const s = cell * 0.62, o = (cell - s) / 2;
        ctx.beginPath();
        ctx.moveTo(px + o, py + o + s);
        ctx.lineTo(px + o, py + o + s * 0.35);
        ctx.lineTo(px + cell / 2, py + o);
        ctx.lineTo(px + o + s, py + o + s * 0.35);
        ctx.lineTo(px + o + s, py + o + s);
        ctx.closePath();
        ctx.fill();
        ctx.strokeStyle = "rgba(0,0,0,.5)";
        ctx.stroke();
      } else if (tile.neutral) {
        ctx.globalAlpha = tile.cleared ? 0.35 : 1;
        const siteArt = textures.get(tile.neutral);
        if (siteArt?.complete && siteArt.naturalWidth) {
          const inset = cell * 0.04;
          ctx.drawImage(siteArt, px + inset, py + inset, cell - 2 * inset, cell - 2 * inset);
        } else {
        ctx.fillStyle = "#1b1410";
        ctx.beginPath();
        ctx.arc(px + cell / 2, py + cell / 2, cell * 0.3, 0, Math.PI * 2);
        ctx.fill();
        ctx.fillStyle = textC;
        ctx.font = `600 ${Math.floor(cell * 0.36)}px system-ui`;
        ctx.textAlign = "center";
        ctx.textBaseline = "middle";
        ctx.fillText(String(tile.neutralLevel ?? ""), px + cell / 2, py + cell / 2 + 1);
        }
        ctx.globalAlpha = 1;
      } else if (ter && !ter.passable) {
        ctx.fillStyle = "rgba(0,0,0,.25)";
        ctx.fillRect(px, py, cell, cell);
      }
      if (sel && sel.x === tile.x && sel.y === tile.y) {
        ctx.strokeStyle = textC;
        ctx.lineWidth = 2;
        ctx.strokeRect(px + 1, py + 1, cell - 2, cell - 2);
        ctx.lineWidth = 1;
      }
    }
  }, [tiles, cell, origin, sel, idx, overview?.player.id, textures, texReady]);

  const tileAt = (clientX: number, clientY: number) => {
    const r = canvas.current!.getBoundingClientRect();
    const x = origin.x + Math.floor((clientX - r.left) / cell), y = origin.y + Math.floor((clientY - r.top) / cell);
    return tiles.find((tt) => tt.x === x && tt.y === y) ?? null;
  };
  const onDown = (e: React.PointerEvent) => {
    drag.current = { x: e.clientX, y: e.clientY, cx: center.x, cy: center.y, moved: false };
    (e.target as HTMLElement).setPointerCapture(e.pointerId);
  };
  const onMove = (e: React.PointerEvent) => {
    const d = drag.current;
    if (!d) return;
    const dx = Math.round((d.x - e.clientX) / cell), dy = Math.round((d.y - e.clientY) / cell);
    if (dx || dy) {
      d.moved = true;
      setCenter({ x: clamp(d.cx + dx, 0, W - 1), y: clamp(d.cy + dy, 0, H - 1) });
    }
  };
  const onUp = (e: React.PointerEvent) => {
    const d = drag.current;
    drag.current = null;
    if (d && !d.moved) setSel(tileAt(e.clientX, e.clientY));
  };
  const onKey = (e: React.KeyboardEvent) => {
    const step: Record<string, [number, number]> = { ArrowLeft: [-1, 0], ArrowRight: [1, 0], ArrowUp: [0, -1], ArrowDown: [0, 1] };
    const s = step[e.key];
    if (s) {
      e.preventDefault();
      setCenter((c) => ({ x: clamp(c.x + s[0] * 3, 0, W - 1), y: clamp(c.y + s[1] * 3, 0, H - 1) }));
    }
  };

  return (
    <div className="screen">
      <h1>{t("map")} <span className="muted small">{origin.x}:{origin.y} — {origin.x + VIEW - 1}:{origin.y + VIEW - 1}</span></h1>
      {err ? <ErrorBox error={err} retry={load} /> : null}
      <div className="map-layout">
        <div className="map-wrap" ref={wrap}>
          <canvas ref={canvas} className="map" style={{ width: VIEW * cell, height: VIEW * cell, touchAction: "none" }}
            tabIndex={0} aria-label={t("map") + " — arrow keys to pan, click a tile"}
            onPointerDown={onDown} onPointerMove={onMove} onPointerUp={onUp} onKeyDown={onKey} />
          <div className="row map-controls">
            <button className="btn small" onClick={() => city && setCenter({ x: city.city.x, y: city.city.y })}><Icon name="castle" size={14} /> {t("yourCity")}</button>
            <button className="btn small ghost" aria-label="Pan left" onClick={() => setCenter((c) => ({ ...c, x: clamp(c.x - 5, 0, W - 1) }))}>◀</button>
            <button className="btn small ghost" aria-label="Pan up" onClick={() => setCenter((c) => ({ ...c, y: clamp(c.y - 5, 0, H - 1) }))}>▲</button>
            <button className="btn small ghost" aria-label="Pan down" onClick={() => setCenter((c) => ({ ...c, y: clamp(c.y + 5, 0, H - 1) }))}>▼</button>
            <button className="btn small ghost" aria-label="Pan right" onClick={() => setCenter((c) => ({ ...c, x: clamp(c.x + 5, 0, W - 1) }))}>▶</button>
          </div>
        </div>
        <TilePanel tile={sel} onMission={setMission} />
      </div>
      {mission && sel && <MissionDialog kind={mission} tile={sel} onClose={() => setMission(null)} onSent={() => { setMission(null); load(); }} />}
    </div>
  );
}

const clamp = (v: number, a: number, b: number) => Math.max(a, Math.min(b, v));

function TilePanel({ tile, onMission }: { tile: MapTile | null; onMission: (k: string) => void }) {
  const { idx, t, overview, city } = useGame();
  if (!tile) return <aside className="panel tile-panel muted">{t("tile")}: —</aside>;
  const ter = idx.terrain(tile.terrain);
  const mine = tile.ownerId === overview?.player.id;
  const n = tile.neutral ? idx.neutral(tile.neutral) : undefined;
  const lvl = n?.levels[(tile.neutralLevel ?? 1) - 1];
  const isHome = city && tile.x === city.city.x && tile.y === city.city.y;
  const actions: string[] = [];
  if (ter?.passable && !isHome) {
    if (tile.cityId) {
      actions.push("scout", "transport");
      if (mine) actions.push("reinforce");
      else actions.push("attack", "reinforce");
    } else if (tile.neutral && !tile.cleared) actions.push("attack", "scout");
    else if (tile.ownerId && !mine) actions.push("attack", "scout");
    else actions.push("occupy", "settle", "scout");
  }
  return (
    <aside className="panel tile-panel">
      <h3>{t("tile")} {tile.x}:{tile.y}</h3>
      <dl className="kv">
        <dt>{t("terrain")}</dt><dd>{idx.name(tile.terrain)}{ter?.resource ? <> · <Res id={ter.resource} n={idx.pack.params.territoryYieldPerHour ?? 0} />{t("perHour")}</> : null}{!ter?.passable && ` (${t("impassable")})`}</dd>
        {tile.cityName && (<><dt>{t("city")}</dt><dd>{tile.cityName}</dd></>)}
        {tile.owner && (<><dt>{t("owner")}</dt><dd>{tile.owner}{tile.alliance ? ` [${tile.alliance}]` : ""}{tile.race ? ` · ${idx.name(tile.race)}` : ""}</dd></>)}
      </dl>
      {n && lvl && (
        <div className="neutral">
          <strong>{idx.name(n.id)} · {t("level")} {lvl.level}</strong>
          {tile.cleared ? <p className="muted small">{t("cleared")}</p> : null}
          <p className="small muted">{idx.desc(n.id)}</p>
          <div className="small">{t("guards")}: {Object.entries(lvl.guards).map(([u, c]) => `${idx.name(u)} ×${c}`).join(", ")}</div>
          <div className="small">{t("rewards")}: {Object.entries(lvl.rewards).map(([r, a]) => <Res key={r} id={r} n={a} />)}</div>
        </div>
      )}
      {!tile.cityId && !tile.ownerId && !tile.neutral && ter?.passable && <p className="muted small">{t("freeLand")}</p>}
      <div className="row wrap">
        {actions.map((k) => (
          <button key={k} className={"btn small" + (k === "attack" ? " danger" : "")} onClick={() => onMission(k)}>{t(("kind_" + k) as any)}</button>
        ))}
      </div>
    </aside>
  );
}

function MissionDialog({ kind, tile, onClose, onSent }: { kind: string; tile: MapTile; onClose: () => void; onSent: () => void }) {
  const { idx, t, city, refresh, toast, world } = useGame();
  const [units, setUnits] = useState<Record<string, number>>({});
  const [cargo, setCargo] = useState<Record<string, number>>({});
  const [heroId, setHeroId] = useState<number | "">("");
  const [capture, setCapture] = useState(true);
  const [retreat, setRetreat] = useState(100);
  const [heroes, setHeroes] = useState<HeroesView>();
  useEffect(() => {
    get<HeroesView>(`/api/v1/worlds/${world.id}/heroes`).then(setHeroes, () => undefined);
  }, [world.id]);
  if (!city) return null;
  const avail = Object.entries(city.city.units).filter(([u, n]) => n > 0 && (kind !== "scout" || idx.unit(u)?.class === "scout"));
  const chosen = Object.entries(units).filter(([, n]) => n > 0);
  const slowest = Math.min(...chosen.map(([u]) => idx.unit(u)?.speed ?? Infinity));
  const dist = Math.hypot(tile.x - city.city.x, tile.y - city.city.y);
  const ws = (idx.pack.params.worldSpeed ?? 1) * (world.speed ?? 1);
  const eta = chosen.length ? Math.max((idx.pack.params.minTravelSeconds ?? 60) / ws, (dist / slowest) * 3600 / ws) : 0;
  const carry = chosen.reduce((s, [u, n]) => s + (idx.unit(u)?.carry ?? 0) * n, 0);
  const cargoTotal = Object.values(cargo).reduce((a, b) => a + (b || 0), 0);
  const needsCargo = kind === "transport" || kind === "settle";
  const idleHeroes = heroes?.heroes.filter((h) => h.status === "idle" && h.homeCityId === city.city.id && !(h.woundedUntil && Date.parse(h.woundedUntil) > Date.now())) ?? [];
  return (
    <Modal title={`${t(("kind_" + kind) as any)} → ${tile.x}:${tile.y}`} onClose={onClose} wide>
      {!avail.length ? <p className="muted">{t("empty")}</p> : (
        <table className="table">
          <tbody>
            {avail.map(([u, n]) => (
              <tr key={u}>
                <th scope="row">{idx.name(u)}</th>
                <td className="muted small">{fmt(n)}</td>
                <td><input type="number" min={0} max={n} value={units[u] ?? 0} aria-label={idx.name(u)}
                  onChange={(e) => setUnits({ ...units, [u]: clamp(Number(e.target.value) || 0, 0, n) })} /></td>
                <td><button className="btn small ghost" onClick={() => setUnits({ ...units, [u]: n })}>{t("max")}</button></td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      {needsCargo && (
        <fieldset className="stack">
          <legend>{t("cargo")} ({fmt(cargoTotal)} / {fmt(carry)})</legend>
          <div className="row wrap">
            {idx.pack.resources.filter((r) => r.kind !== "premium").map((r) => (
              <label key={r.id} className="small">{idx.name(r.id)}
                <input type="number" min={0} max={city.resources[r.id] ?? 0} value={cargo[r.id] ?? 0}
                  onChange={(e) => setCargo({ ...cargo, [r.id]: clamp(Number(e.target.value) || 0, 0, city.resources[r.id] ?? 0) })} />
              </label>
            ))}
          </div>
        </fieldset>
      )}
      {(kind === "attack" || kind === "settle" || kind === "reinforce") && idleHeroes.length > 0 && (
        <label>{t("hero")}
          <select value={heroId} onChange={(e) => setHeroId(e.target.value ? Number(e.target.value) : "")}>
            <option value="">{t("none")}</option>
            {idleHeroes.map((h) => <option key={h.id} value={h.id}>{h.name} ({idx.name(h.class)} {h.level})</option>)}
          </select>
        </label>
      )}
      {kind === "attack" && (
        <div className="row wrap">
          {tile.ownerId && !tile.cityId && <label className="check"><input type="checkbox" checked={capture} onChange={(e) => setCapture(e.target.checked)} /> {t("capture")}</label>}
          <label className="small">{t("retreatAt")} <input type="number" min={1} max={100} value={retreat} onChange={(e) => setRetreat(clamp(Number(e.target.value) || 100, 1, 100))} /></label>
        </div>
      )}
      <p className="small">{t("time")}: {chosen.length ? duration(eta) : "—"} · {t("carry")}: {fmt(carry)}</p>
      <div className="row end">
        <button className="btn ghost" onClick={onClose}>{t("cancel")}</button>
        <Action className={kind === "attack" ? "danger" : "primary"} disabled={!chosen.length || (needsCargo && cargoTotal > carry)}
          confirm={kind === "attack" ? t("dangerAttack") : undefined}
          run={async () => {
            await command(`/api/v1/cities/${city.city.id}/missions`, {
              kind, x: tile.x, y: tile.y, units: Object.fromEntries(chosen),
              ...(heroId ? { heroId } : {}),
              ...(needsCargo ? { cargo: Object.fromEntries(Object.entries(cargo).filter(([, v]) => v > 0)) } : {}),
              ...(kind === "attack" ? { capture, retreatPct: retreat } : {}),
            });
            toast(t("sendMission") + " ✓", "ok");
            await refresh();
            onSent();
          }}>{t("send")}</Action>
      </div>
    </Modal>
  );
}

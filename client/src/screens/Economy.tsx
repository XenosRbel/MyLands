import { command, get } from "../api/client";
import type { Territory } from "../api/types";
import { useGame } from "../game/context";
import { Action, Empty, ErrorBox, fmt, Icon, Loading, Progress, Res, useLoad } from "../ui/kit";

export function EconomyScreen() {
  const { city, idx, t, world, refresh, toast, go } = useGame();
  const terr = useLoad((s) => get<Territory[]>(`/api/v1/worlds/${world.id}/territories`, s), [world.id]);
  if (!city) return null;
  const r = city.rates;
  return (
    <div className="screen">
      <h1>{t("economy")}</h1>
      <div className="grid-2">
        <section className="panel">
          <h3>{t("production")}</h3>
          <table className="table">
            <thead><tr><th></th><th>{t("production")}</th><th>{t("consumption")}</th><th>{t("net")}</th><th>{t("storage")}</th></tr></thead>
            <tbody>
              {idx.pack.resources.map((res) => {
                const n = city.resources[res.id] ?? 0;
                const cap = r.storage[res.id];
                return (
                  <tr key={res.id}>
                    <th scope="row"><Icon name={idx.icon(res.id) || "star"} /> {idx.name(res.id)}</th>
                    <td>{fmt(r.gross[res.id] ?? 0)}{t("perHour")}</td>
                    <td>{fmt(r.consumption[res.id] ?? 0)}{t("perHour")}</td>
                    <td className={(r.net[res.id] ?? 0) < 0 ? "neg" : "pos"}>{(r.net[res.id] ?? 0) >= 0 ? "+" : ""}{fmt(r.net[res.id] ?? 0)}</td>
                    <td>
                      {cap !== undefined ? (<><Progress value={n} max={cap} label={idx.name(res.id)} /><span className="small muted">{fmt(n)} / {fmt(cap)}</span></>) : <span className="small">{fmt(n)}</span>}
                      {r.protected[res.id] ? <div className="small muted">{t("protected")}: {fmt(r.protected[res.id])}</div> : null}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </section>
        <section className="panel">
          <h3>{t("population")}</h3>
          <dl className="kv">
            <dt>{t("population")}</dt><dd>{fmt(city.city.population)} / {fmt(r.housing)}</dd>
            <dt>{t("workers")}</dt><dd>{fmt(r.workersNeeded)}</dd>
            <dt>{t("employment")}</dt><dd>{Math.round(r.employment * 100)}%</dd>
            <dt>{t("idle")}</dt><dd>{fmt(r.idle)}</dd>
            <dt>{t("upkeep")}</dt><dd><Res id={idx.pack.params.upkeepResource || idx.pack.params.foodResource} n={r.upkeep} />{t("perHour")}</dd>
          </dl>
          <Progress value={city.city.population} max={r.housing} label={t("housing")} />
          {r.employment < 1 && <p className="warn small">⚠ {t("employment")} {Math.round(r.employment * 100)}% — {t("housing")} ↑</p>}
          {Object.entries(r.net).some(([, v]) => v < 0) && <p className="warn small">⚠ {t("net")} &lt; 0</p>}
        </section>
      </div>
      <section className="panel">
        <h3>{t("territories")}</h3>
        {terr.error ? <ErrorBox error={terr.error} retry={terr.reload} /> : !terr.data ? <Loading /> : !terr.data.length ? (
          <Empty><button className="link" onClick={() => go("map")}>{t("map")} →</button></Empty>
        ) : (
          <ul className="list">
            {terr.data.map((tr) => (
              <li key={`${tr.x}:${tr.y}`}>
                <span>{idx.name(tr.terrain)} ({tr.x}:{tr.y})</span>
                <span className="muted small">{t("garrison")}: {Object.entries(tr.garrison).map(([u, n]) => `${idx.name(u)} ${n}`).join(", ") || "—"}</span>
                <span>{Object.entries(tr.pending).map(([res, n]) => <Res key={res} id={res} n={n} />)}</span>
                <Action className="small" disabled={!Object.keys(tr.pending).length} run={async () => {
                  await command(`/api/v1/cities/${city.city.id}/collect`, { x: tr.x, y: tr.y });
                  toast(t("collect") + " ✓", "ok");
                  terr.reload();
                  await refresh();
                }}>{t("collect")}</Action>
                <Action className="small ghost" confirm={t("abandon") + "?"} run={async () => {
                  await command(`/api/v1/worlds/${world.id}/territories/abandon`, { x: tr.x, y: tr.y });
                  terr.reload();
                }}>{t("abandon")}</Action>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}

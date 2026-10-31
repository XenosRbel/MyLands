import { command, get } from "../api/client";
import type { Mission, Report } from "../api/types";
import { useGame } from "../game/context";
import { Action, Empty, ErrorBox, Icon, Loading, Modal, Res, Tabs, Timer, useLoad } from "../ui/kit";
import { useState } from "react";

export function MissionsScreen() {
  const { world, t, idx, refresh, go } = useGame();
  const ms = useLoad((s) => get<Mission[]>(`/api/v1/worlds/${world.id}/missions`, s), [world.id]);
  if (ms.error) return <ErrorBox error={ms.error} retry={ms.reload} />;
  if (!ms.data) return <Loading />;
  const done = async () => {
    ms.reload();
    await refresh();
  };
  return (
    <div className="screen">
      <h1>{t("missions")}</h1>
      {!ms.data.length ? <Empty /> : (
        <ul className="list missions">
          {ms.data.map((m) => (
            <li key={m.id} className={m.direction}>
              <Icon name={m.direction === "incoming" ? "skull" : m.kind === "scout" ? "eye" : m.kind === "attack" ? "sword" : "flag"} />
              <span>
                <strong>{t(("kind_" + m.kind) as any)}</strong> → <button className="link" onClick={() => go("map", { x: m.targetX, y: m.targetY })}>{m.targetX}:{m.targetY}</button>
                {m.direction !== "own" && <span className="tag danger"> {t("incoming")} · {m.owner}</span>}
              </span>
              <span className="muted small">{m.units ? Object.entries(m.units).map(([u, n]) => `${idx.name(u)} ${n}`).join(", ") : "?"}</span>
              <span>
                {m.status === "outbound" && <>{t("arrives")} <Timer until={m.arriveAt} onDone={done} /></>}
                {m.status === "returning" && m.returnAt && <>{t("returns")} <Timer until={m.returnAt} onDone={done} /></>}
                {m.status === "stationed" && t("stationed")}
              </span>
              {m.cargo && Object.keys(m.cargo).length > 0 && <span>{Object.entries(m.cargo).map(([r, n]) => <Res key={r} id={r} n={n} />)}</span>}
              {m.direction === "own" && (m.status === "outbound" || m.status === "stationed") && (
                <Action className="small ghost" run={async () => { await command(`/api/v1/worlds/${world.id}/missions/${m.id}/recall`); await done(); }}>{t("recall")}</Action>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

export function ReportsScreen() {
  const { world, t, refresh } = useGame();
  const [open, setOpen] = useState<number | null>(null);
  const rs = useLoad((s) => get<Report[]>(`/api/v1/worlds/${world.id}/reports?limit=50`, s), [world.id]);
  if (rs.error) return <ErrorBox error={rs.error} retry={rs.reload} />;
  if (!rs.data) return <Loading />;
  return (
    <div className="screen">
      <h1>{t("reports")}</h1>
      {!rs.data.length ? <Empty /> : (
        <ul className="list">
          {rs.data.map((r) => {
            const b = r.data?.battle;
            const won = r.kind === "battle" ? (r.side === "attacker" ? b?.attackerWon : !b?.attackerWon) : r.data?.success;
            return (
              <li key={r.id} className={r.readAt ? "" : "unread"}>
                <Icon name={r.kind === "scout" ? "eye" : "sword"} />
                <button className="link" onClick={() => { setOpen(r.id); }}>
                  {r.kind === "battle" ? t("battle") : t("scout")} · {r.data?.target ? `${r.data.target.x}:${r.data.target.y}` : ""} · {r.data?.attacker} {r.data?.defender ? `vs ${r.data.defender}` : r.data?.neutral ? "vs " + r.data.neutral.id : ""}
                </button>
                <span className={"tag " + (won ? "ok" : "danger")}>{won ? t("victory") : t("defeat")}</span>
                <span className="muted small">{new Date(r.occurredAt).toLocaleString()}</span>
              </li>
            );
          })}
        </ul>
      )}
      {open !== null && <ReportModal id={open} onClose={() => { setOpen(null); rs.reload(); refresh(); }} />}
    </div>
  );
}

function ReportModal({ id, onClose }: { id: number; onClose: () => void }) {
  const { world, t, idx } = useGame();
  const r = useLoad((s) => get<Report>(`/api/v1/worlds/${world.id}/reports/${id}`, s), [id]);
  const d = r.data?.data;
  const units = (m?: Record<string, number>) => (m && Object.keys(m).length ? Object.entries(m).map(([u, n]) => `${idx.name(u)} ${n}`).join(", ") : "—");
  return (
    <Modal title={t("reports")} onClose={onClose} wide>
      {!d ? <Loading /> : r.data!.kind === "battle" ? (
        <div className="report">
          <p><strong>{d.attacker}</strong> {t("attacker")} vs <strong>{d.defender ?? (d.neutral ? `${idx.name(d.neutral.id)} ${d.neutral.level}` : "?")}</strong></p>
          <p className={"tag " + (d.battle.attackerWon ? "ok" : "danger")}>{d.battle.attackerWon ? t("victory") : t("defeat")} ({t("attacker")})</p>
          <table className="table">
            <thead><tr><th></th><th>{t("army")}</th><th>{t("losses")}</th><th>{t("survivors")}</th></tr></thead>
            <tbody>
              <tr><th>{t("attacker")}</th><td>{units(d.battle.attackerStart)}</td><td>{units(d.battle.attackerLosses)}</td><td>{units(d.battle.attackerLeft)}</td></tr>
              <tr><th>{t("defender")}</th><td>{units(d.battle.defenderStart)}</td><td>{units(d.battle.defenderLosses)}</td><td>{units(d.battle.defenderLeft)}</td></tr>
            </tbody>
          </table>
          <p className="small">{t("rounds")}: {d.battle.rounds?.length ?? 0}{d.battle.attackerRetreated ? " · retreat" : ""}{d.wall ? ` · wall +${d.wall}%` : ""}</p>
          {d.loot && Object.keys(d.loot).length > 0 && <p>{t("loot")}: {Object.entries(d.loot as Record<string, number>).map(([res, n]) => <Res key={res} id={res} n={n} />)}</p>}
          {d.rewards && <p>{t("rewards")}: {Object.entries(d.rewards as Record<string, number>).map(([res, n]) => <Res key={res} id={res} n={n} />)}</p>}
          {d.items?.length ? <p>⚜ {d.items.map((i: string) => idx.name(i)).join(", ")}</p> : null}
          {d.captured && <p className="tag ok">{t("territories")} ✓</p>}
        </div>
      ) : (
        <div className="report">
          <p><strong>{d.attacker}</strong> → {d.defender ?? `${d.target?.x}:${d.target?.y}`}</p>
          <p className={"tag " + (d.success ? "ok" : "danger")}>{d.success ? t("victory") : t("defeat")}{d.detected ? " · detected" : ""}</p>
          {d.resources && <p>{Object.entries(d.resources as Record<string, number>).map(([res, n]) => <Res key={res} id={res} n={n} />)}</p>}
          {d.units && <p>{t("garrison")}: {units(d.units)}</p>}
          {d.buildings && <p className="small">{Object.entries(d.buildings as Record<string, number>).map(([b, l]) => `${idx.name(b)} ${l}`).join(", ")}</p>}
          {d.neutral && <p>{idx.name(d.neutral.id)} {d.neutral.level}: {units(d.neutral.guards)}</p>}
          {d.garrison && <p>{t("garrison")}: {units(d.garrison)}</p>}
        </div>
      )}
    </Modal>
  );
}

export function QuestsScreen() {
  const { world, t, idx, city, refresh, toast, go } = useGame();
  const qs = useLoad((s) => get<import("../api/types").Quest[]>(`/api/v1/worlds/${world.id}/quests`, s), [world.id, city?.serverTime]);
  const [tab, setTab] = useState<"active" | "done">("active");
  if (qs.error) return <ErrorBox error={qs.error} retry={qs.reload} />;
  if (!qs.data) return <Loading />;
  const list = qs.data.filter((q) => (tab === "done" ? q.status === "claimed" : q.status !== "claimed" && q.status !== "locked"));
  return (
    <div className="screen">
      <h1>{t("quests")}</h1>
      <Tabs value={tab} onChange={setTab} items={[["active", t("inProgress")], ["done", t("claimed")]]} />
      {!list.length ? <Empty /> : (
        <ul className="list quests">
          {list.map((q) => {
            const def = idx.pack.quests.find((x) => x.id === q.id);
            return (
              <li key={q.id} className={q.status}>
                <div>
                  <strong>{idx.name(q.id)}</strong> {q.tutorial && <span className="tag">{t("tutorial")}</span>}
                  <div className="small muted">{idx.text(def?.desc)}</div>
                  {def?.hint && q.status === "active" && <div className="small">💡 {idx.text(def.hint)}</div>}
                </div>
                <span className="small">{q.progress}/{q.goal}</span>
                <span>{Object.entries(q.rewards).map(([r, n]) => <Res key={r} id={r} n={n} />)}</span>
                {q.status === "complete" ? (
                  <Action className="primary small" run={async () => {
                    await command(`/api/v1/worlds/${world.id}/quests/${q.id}/claim`, { cityId: city?.city.id ?? 0 });
                    toast(t("claimed") + " ✓", "ok");
                    qs.reload();
                    await refresh();
                  }}>{t("claim")}</Action>
                ) : q.status === "active" && q.screen ? (
                  <button className="btn small ghost" onClick={() => go(q.screen as any)}>{t("goTo")}</button>
                ) : null}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}

import { command } from "../api/client";
import { useGame } from "../game/context";
import { Action, CostList, duration, Reqs } from "../ui/kit";
import { QueuePanel, blockedLabel, effectText } from "./City";

export function ResearchScreen() {
  const { city, idx, t, refresh, toast } = useGame();
  if (!city) return null;
  const branches = Array.from(new Set(idx.pack.techs.map((x) => x.branch)));
  return (
    <div className="screen">
      <h1>{t("research")}</h1>
      <QueuePanel jobs={city.queues.research} title={t("queue")} />
      {branches.map((br) => (
        <section className="panel" key={br}>
          <h3>{br[0].toUpperCase() + br.slice(1)}</h3>
          <div className="cards wide">
            {city.techs.filter((x) => idx.pack.techs.find((d) => d.id === x.id)?.branch === br).map((tv) => {
              const def = idx.pack.techs.find((d) => d.id === tv.id)!;
              return (
                <article key={tv.id} className="tcard">
                  <header>
                    <strong>{idx.name(tv.id)}</strong>
                    <span className="muted small">{tv.level}/{tv.maxLevel}{tv.targetLevel > tv.level ? ` → ${tv.targetLevel}` : ""}</span>
                  </header>
                  <p className="small muted">{idx.desc(tv.id)}</p>
                  {def.exclusiveGroup && <span className="tag">★ {def.exclusiveGroup}</span>}
                  <ul className="effects small">{def.effects.map((e, i) => <li key={i}>{effectText(e, idx, Math.max(1, tv.level))}</li>)}</ul>
                  {tv.next ? (
                    <>
                      <div className="row small"><CostList cost={tv.next.cost} have={city.resources} /> <span>{duration(tv.next.seconds)}</span></div>
                      <Reqs reqs={tv.next.unmet} />
                      <Action className="primary small" disabled={!!tv.next.blocked} title={tv.next.blocked ? blockedLabel(tv.next.blocked, t) : undefined} run={async () => {
                        await command(`/api/v1/cities/${city.city.id}/research`, { tech: tv.id });
                        toast(`${idx.name(tv.id)} → ${tv.next!.level}`, "ok");
                        await refresh();
                      }}>{t("research")} → {tv.next.level}</Action>
                    </>
                  ) : <span className="muted small">{t("maxLevel")}</span>}
                </article>
              );
            })}
          </div>
        </section>
      ))}
    </div>
  );
}

import { useState } from "react";
import { command } from "../api/client";
import { useGame } from "../game/context";
import { Action, CostList, duration, Icon, Reqs } from "../ui/kit";
import { QueuePanel, blockedLabel, effectText } from "./City";
import { ResearchSprite } from "../game/ResearchSprite";

const branchNames = {
  economy: "branchEconomy",
  civic: "branchCivic",
  military: "branchMilitary",
  arcane: "branchArcane",
  calling: "branchCalling",
} as const;

const branchSymbols: Record<string, string> = {
  economy: "wheat", civic: "castle", military: "sword", arcane: "crystal", calling: "star",
};
const roman = (n: number) => {
  const numerals: [number, string][] = [[10, "X"], [9, "IX"], [5, "V"], [4, "IV"], [1, "I"]];
  let value = n;
  let result = "";
  for (const [number, symbol] of numerals) {
    while (value >= number) { result += symbol; value -= number; }
  }
  return result;
};

export function ResearchScreen() {
  const { city, idx, t, refresh, toast } = useGame();
  const [selected, setSelected] = useState<string | null>(null);
  const [activeBranch, setActiveBranch] = useState<string | null>(null);
  if (!city) return null;
  const branches = Array.from(new Set(idx.pack.techs.map((x) => x.branch)));
  const branch = activeBranch && branches.includes(activeBranch) ? activeBranch : branches[0];
  const branchTechs = city.techs.filter((view) =>
    idx.pack.techs.find((def) => def.id === view.id)?.branch === branch
  );
  const activeId = selected && branchTechs.some((view) => view.id === selected)
    ? selected : branchTechs[0]?.id;
  const active = city.techs.find((x) => x.id === activeId);
  const definition = active && idx.pack.techs.find((x) => x.id === active.id);
  const currentSpecialization = idx.pack.techs.find((tech) =>
    tech.exclusiveGroup === "calling" && (city.techs.find((item) => item.id === tech.id)?.level ?? 0) > 0
  );
  const switchingSpecialization = Boolean(
    definition?.exclusiveGroup === "calling" && currentSpecialization && currentSpecialization.id !== active?.id
  );

  return (
    <div className="screen research-screen">
      <div className="research-heading">
        <Icon name="book" size={28} />
        <div><h1>{t("research")}</h1><p className="muted small">{t("researchHint")}</p></div>
      </div>
      <nav className="research-branch-tabs" aria-label={t("researchTree")}>
        {branches.map((br) => (
          <button type="button" key={br} className={branch === br ? "active" : ""}
            aria-current={branch === br ? "page" : undefined}
            onClick={() => { setActiveBranch(br); setSelected(null); }}>
            <Icon name={branchSymbols[br] ?? "book"} size={19} />
            {br in branchNames ? t(branchNames[br as keyof typeof branchNames]) : br}
          </button>
        ))}
      </nav>
      <div className="research-layout">
        <div className="research-tree" role="region" aria-label={t("researchTree")}>
          {branches.filter((br) => br === branch).map((br) => (
            <section key={br} className="research-branch">
              <div className="research-branch-title">
                <Icon name={branchSymbols[br] ?? "book"} size={24} />
                <strong>{br in branchNames ? t(branchNames[br as keyof typeof branchNames]) : br}</strong>
              </div>
              <div className="research-paths">
                {city.techs.filter((x) => idx.pack.techs.find((d) => d.id === x.id)?.branch === br).map((tv) => {
                  const milestones = Array.from({ length: tv.maxLevel }, (_, i) => i + 1);
                  return (
                    <div className={"research-path" + (activeId === tv.id ? " selected" : "")} key={tv.id}>
                      <button type="button" className="research-name" onClick={() => setSelected(tv.id)}
                        aria-pressed={activeId === tv.id}>
                        <span className="research-name-icon"><ResearchSprite tech={tv.id} level={tv.level} size={42} /></span>
                        <span><strong>{idx.name(tv.id)}</strong><small>{tv.level}/{tv.maxLevel}</small></span>
                      </button>
                      <div className="research-nodes">
                        {milestones.map((level) => {
                          const complete = tv.level >= level;
                          const pending = tv.targetLevel === level && tv.level < level;
                          const unlocked = level === tv.level + 1;
                          return (
                            <button type="button" key={level}
                              className={["research-node", complete && "completed", pending && "pending", unlocked && "available"].filter(Boolean).join(" ")}
                              onClick={() => setSelected(tv.id)}
                              aria-pressed={activeId === tv.id}
                              title={idx.name(tv.id) + " " + roman(level)}>
                              <ResearchSprite tech={tv.id} level={level} size={36} className="research-tier-art" />
                              <span>{roman(level)}</span>
                              {complete && <b aria-hidden="true">✓</b>}
                            </button>
                          );
                        })}
                      </div>
                    </div>
                  );
                })}
              </div>
            </section>
          ))}
        </div>
        <aside className="research-aside">
          {active && definition && (
            <section className="research-detail panel">
              <div className="research-detail-art"><ResearchSprite tech={active.id} level={Math.max(active.level,1)} size={96} /></div>
              <h2>{idx.name(active.id)}</h2>
              <div className="muted small">{t("level")} {active.level} / {active.maxLevel}</div>
              <p>{idx.desc(active.id)}</p>
              {definition.exclusiveGroup && <span className="tag">{definition.exclusiveGroup}</span>}
              {definition.exclusiveGroup === "calling" && (
                <div className="research-specialization">
                  <h3>{t("specializationBonuses")}</h3>
                  {definition.effects.filter((e) => e.type === "modifier" && e.modifier).map((effect, i) => (
                    <div key={i} className="specialization-bonus">
                      <Icon name="star" size={15} />
                      <strong>{effectText(effect, idx, Math.max(1, active.level || 1))}</strong>
                    </div>
                  ))}
                  {switchingSpecialization && <p className="warn small">{t("switchSpecializationWarning")}</p>}
                </div>
              )}
              <h3>{t("progress")}</h3>
              <div className="research-level-progress" role="progressbar" aria-valuemin={0} aria-valuemax={active.maxLevel} aria-valuenow={active.level}>
                <span style={{ width: (100 * active.level / active.maxLevel) + "%" }} />
              </div>
              <ul className="effects">
                {definition.effects.map((effect, i) => <li key={i}>{effectText(effect, idx, Math.max(1, active.level))}</li>)}
              </ul>
              {active.next ? (
                <div className="stack">
                  <h3>{t("cost")}</h3>
                  <CostList cost={active.next.cost} have={city.resources} />
                  <div className="row"><Icon name="clock" size={15} /> {duration(active.next.seconds)}</div>
                  <Reqs reqs={active.next.unmet} />
                  <Action className="primary" disabled={!!active.next.blocked}
                    confirm={switchingSpecialization ? t("switchSpecializationWarning") : undefined}
                    title={active.next.blocked ? blockedLabel(active.next.blocked, t) : undefined}
                    run={async () => {
                      await command(`/api/v1/cities/${city.city.id}/research`, { tech: active.id });
                      toast(`${idx.name(active.id)} → ${active.next!.level}`, "ok");
                      await refresh();
                    }}>{t("research")} → {active.next.level}</Action>
                </div>
              ) : <p className="muted small">{t("maxLevel")}</p>}
            </section>
          )}
          <QueuePanel jobs={city.queues.research} title={t("queue")} />
        </aside>
      </div>
    </div>
  );
}

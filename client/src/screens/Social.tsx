import { useState } from "react";
import { command, get } from "../api/client";
import type { Alliance, HeroesView, Invite, Message, Notification, Order, RankRow } from "../api/types";
import { useGame } from "../game/context";
import { describeModifier } from "../game/pack";
import { Action, CostList, Empty, ErrorBox, Icon, Loading, Modal, Progress, Res, Tabs, Timer, fmt, useLoad } from "../ui/kit";

// ---------------- Heroes ----------------
export function HeroesScreen() {
  const { world, t, idx, city, refresh, toast } = useGame();
  const hv = useLoad((signal) => get<HeroesView>(`/api/v1/worlds/${world.id}/heroes`, signal), [world.id]);
  const [cls, setCls] = useState(idx.pack.heroes.classes[0]?.id ?? "");
  const [name, setName] = useState("");
  const [selectedHeroId, setSelectedHeroId] = useState<number | null>(null);
  if (hv.error) return <ErrorBox error={hv.error} retry={hv.reload} />;
  if (!hv.data || !city) return <Loading />;
  const data = hv.data;
  const free = data.inventory.filter((item) => !item.heroId);
  const hero = data.heroes.find((item) => item.id === selectedHeroId) ?? data.heroes[0];
  return (
    <div className="screen hero-game-screen">
      <h1>{t("heroes")} <span className="muted small">{data.heroes.length}/{data.maxHeroes}</span></h1>
      <section className="panel hero-recruit-panel">
        <h3><Icon name="star" size={19} /> {t("hire")}</h3>
        <div className="hero-recruit-options">
          <div className="hero-recruit-portrait">
            {idx.art(cls) ? <img src={idx.art(cls)} alt="" draggable={false} /> : <Icon name="user" size={34} />}
          </div>
          <div className="hero-recruit-controls">
            <div className="row wrap">
              <select value={cls} onChange={(e) => setCls(e.target.value)} aria-label={t("hero")}>
                {idx.pack.heroes.classes.map((item) => <option key={item.id} value={item.id}>{idx.name(item.id)}</option>)}
              </select>
              <input placeholder={t("name")} aria-label={t("name")} value={name} maxLength={24} onChange={(e) => setName(e.target.value)} />
            </div>
            <p className="small muted">{idx.desc(cls)}</p>
            <div className="row wrap">
              <CostList cost={data.hireCost} have={city.resources} />
              <Action className="primary" disabled={data.heroes.length >= data.maxHeroes || name.trim().length < 3} run={async () => {
                await command(`/api/v1/cities/${city.city.id}/heroes`, { class: cls, name: name.trim() });
                setName("");
                toast(t("hire") + " ✓", "ok");
                hv.reload();
                await refresh();
              }}>{t("hire")}</Action>
            </div>
          </div>
        </div>
      </section>
      {!hero ? <Empty /> : (
        <div className="hero-game-layout">
          <div className="hero-game-main">
            <div className="hero-game-roster" role="group" aria-label={t("heroes")}>
              {data.heroes.map((h) => (
                <button key={h.id} className={"hero-game-roster-item" + (hero.id === h.id ? " selected" : "")}
                  onClick={() => setSelectedHeroId(h.id)} aria-pressed={hero.id === h.id}>
                  {idx.art(h.class) ? <img src={idx.art(h.class)} alt="" /> : <Icon name="user" size={28} />}
                  <strong>{h.name}</strong>
                  <small>{idx.name(h.class)} · {t("level")} {h.level}</small>
                </button>
              ))}
            </div>
            <section className="panel hero-game-character">
              <div className="hero-game-profile">
                <div className="hero-game-portrait">
                  {idx.art(hero.class) ? <img src={idx.art(hero.class)} alt="" /> : <Icon name="user" size={60} />}
                </div>
                <div className="hero-game-overview">
                  <h2>{hero.name}</h2>
                  <div className="hero-game-class">{idx.name(hero.class)} · {t("level")} {hero.level}</div>
                  <p>{idx.desc(hero.class)}</p>
                  <span className="tag">{hero.status === "mission" ? t("away") : t("home")}</span>
                  {hero.woundedUntil && Date.parse(hero.woundedUntil) > Date.now() &&
                    <span className="tag danger">{t("wounded")} <Timer until={hero.woundedUntil} /></span>}
                  {hero.nextLevelXp > 0 && <Progress value={hero.xp} max={hero.nextLevelXp} label="XP" />}
                  <div className="small muted">XP {fmt(hero.xp)} / {fmt(hero.nextLevelXp)} · {t("skillPoints")}: {hero.skillPoints}</div>
                </div>
              </div>
              <div className="hero-game-stats">
                {Object.entries(hero.modifiers).map(([key, value]) => (
                  <span key={key} className="tag">{key} {value > 0 ? "+" : ""}{value}</span>
                ))}
              </div>
              <h3>{t("inventory")}</h3>
              <div className="hero-game-equipment">
                {idx.pack.heroes.slots.map((slot) => {
                  const item = hero.items.find((entry) => entry.slot === slot);
                  return (
                    <div key={slot} className="hero-game-slot">
                      <div className="hero-game-slot-icon">
                        {item ? <Icon name={idx.icon(item.def) || "star"} size={31} /> : <Icon name="shield" size={27} />}
                      </div>
                      <div className="hero-game-slot-input">
                        <small>{slot}</small>
                        {item ? (
                          <Action className="small ghost" disabled={hero.status !== "idle"} run={async () => {
                            await command(`/api/v1/worlds/${world.id}/heroes/${hero.id}/unequip`, { slot });
                            hv.reload();
                          }}>{idx.name(item.def)} ✕</Action>
                        ) : (
                          <select disabled={hero.status !== "idle"} value="" aria-label={slot} onChange={async (event) => {
                            if (!event.target.value) return;
                            try {
                              await command(`/api/v1/worlds/${world.id}/heroes/${hero.id}/equip`, { itemId: Number(event.target.value) });
                            } catch (error) {
                              toast((error as Error).message, "error");
                            }
                            hv.reload();
                          }}>
                            <option value="">—</option>
                            {free.filter((item) => idx.get<any>(item.def)?.slot === slot).map((item) => (
                              <option key={item.id} value={item.id}>{idx.name(item.def)}</option>
                            ))}
                          </select>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            </section>
          </div>
          <aside className="hero-game-side">
            <section className="panel">
              <h3>{t("skills")}</h3>
              <div className="hero-game-skills">
                {idx.pack.heroes.skills.filter((skill) => !skill.classes?.length || skill.classes.includes(hero.class)).map((skill) => (
                  <div className="hero-game-skill" key={skill.id}>
                    <Icon name="star" size={26} />
                    <span><strong>{idx.name(skill.id)}</strong><small>{hero.skills[skill.id] ?? 0}/{skill.maxLevel}</small></span>
                    <Action className="small" disabled={hero.skillPoints < 1 || (hero.skills[skill.id] ?? 0) >= skill.maxLevel}
                      title={skill.modifiers.map((modifier) => describeModifier(modifier, idx)).join(", ")}
                      run={async () => { await command(`/api/v1/worlds/${world.id}/heroes/${hero.id}/skills`, { skill: skill.id }); hv.reload(); }}>
                      +
                    </Action>
                  </div>
                ))}
              </div>
            </section>
            {free.length > 0 && (
              <section className="panel">
                <h3>{t("inventory")}</h3>
                <ul className="list">
                  {free.map((item) => (
                    <li key={item.id}>
                      <Icon name={idx.icon(item.def) || "star"} size={32} />
                      <strong>{idx.name(item.def)}</strong>
                      <span className="small muted">{idx.desc(item.def)}</span>
                    </li>
                  ))}
                </ul>
              </section>
            )}
          </aside>
        </div>
      )}
    </div>
  );
}

// ---------------- Market ----------------
export function MarketScreen() {
  const { world, t, idx, city, refresh, toast, overview } = useGame();
  const tradable = idx.pack.resources.filter((r) => r.tradable && r.kind !== "premium");
  const [sell, setSell] = useState(tradable[0]?.id ?? "");
  const [buy, setBuy] = useState(tradable[1]?.id ?? "");
  const [sellAmt, setSellAmt] = useState(100);
  const [buyAmt, setBuyAmt] = useState(100);
  const [exAmt, setExAmt] = useState(100);
  const [filter, setFilter] = useState("");
  const os = useLoad((s) => get<Order[]>(`/api/v1/worlds/${world.id}/market${filter ? `?sell=${filter}` : ""}`, s), [world.id, filter]);
  if (!city) return null;
  const rates = idx.pack.params.exchangeRates ?? {};
  const fee = idx.pack.params.exchangeFeePct ?? 0;
  const exOut = rates[sell] && rates[buy] ? Math.floor((exAmt * rates[sell]) / rates[buy] * (1 - fee / 100)) : 0;
  return (
    <div className="screen">
      <h1>{t("market")}</h1>
      <div className="grid-2">
        <section className="panel">
          <h3>{t("placeOrder")}</h3>
          <div className="stack">
            <div className="row wrap">
              <label>{t("sell")}<select value={sell} onChange={(e) => setSell(e.target.value)}>{tradable.map((r) => <option key={r.id} value={r.id}>{idx.name(r.id)}</option>)}</select></label>
              <label>{t("quantity")}<input type="number" min={1} value={sellAmt} onChange={(e) => setSellAmt(Math.max(1, Number(e.target.value) || 1))} /></label>
            </div>
            <div className="row wrap">
              <label>{t("buy")}<select value={buy} onChange={(e) => setBuy(e.target.value)}>{tradable.map((r) => <option key={r.id} value={r.id}>{idx.name(r.id)}</option>)}</select></label>
              <label>{t("quantity")}<input type="number" min={1} value={buyAmt} onChange={(e) => setBuyAmt(Math.max(1, Number(e.target.value) || 1))} /></label>
            </div>
            <Action className="primary" disabled={sell === buy || (city.resources[sell] ?? 0) < sellAmt} run={async () => {
              await command(`/api/v1/cities/${city.city.id}/market/orders`, { sell, sellAmount: sellAmt, buy, buyAmount: buyAmt });
              toast(t("placeOrder") + " ✓", "ok");
              os.reload();
              await refresh();
            }}>{t("placeOrder")}</Action>
          </div>
        </section>
        <section className="panel">
          <h3>{t("exchange")}</h3>
          <p className="small muted">{t("exchangeHint")} ({fee}%)</p>
          <div className="row wrap">
            <label>{t("sell")} {idx.name(sell)}<input type="number" min={1} value={exAmt} onChange={(e) => setExAmt(Math.max(1, Number(e.target.value) || 1))} /></label>
            <span>→ <Res id={buy} n={exOut} /></span>
            <Action disabled={sell === buy || exOut < 1 || (city.resources[sell] ?? 0) < exAmt} run={async () => {
              await command(`/api/v1/cities/${city.city.id}/exchange`, { from: sell, to: buy, amount: exAmt });
              toast(t("exchange") + " ✓", "ok");
              await refresh();
            }}>{t("exchange")}</Action>
          </div>
        </section>
      </div>
      <section className="panel">
        <div className="row wrap">
          <h3>{t("offer")}</h3>
          <select value={filter} onChange={(e) => setFilter(e.target.value)} aria-label="Filter">
            <option value="">*</option>
            {tradable.map((r) => <option key={r.id} value={r.id}>{idx.name(r.id)}</option>)}
          </select>
        </div>
        {os.error ? <ErrorBox error={os.error} retry={os.reload} /> : !os.data ? <Loading /> : !os.data.length ? <Empty /> : (
          <table className="table">
            <thead><tr><th>{t("sell")}</th><th>{t("buy")}</th><th></th><th></th></tr></thead>
            <tbody>
              {os.data.map((o) => {
                const mine = o.playerId === overview?.player.id;
                return (
                  <tr key={o.id}>
                    <td><Res id={o.sellRes} n={o.sellAmount} /></td>
                    <td><Res id={o.buyRes} n={o.buyAmount} have={city.resources[o.buyRes] ?? 0} /></td>
                    <td className="small muted">{o.player}</td>
                    <td>
                      {mine ? (
                        <Action className="small ghost" run={async () => { await command(`/api/v1/worlds/${world.id}/market/${o.id}/cancel`); os.reload(); await refresh(); }}>{t("cancel")}</Action>
                      ) : (
                        <Action className="small primary" disabled={(city.resources[o.buyRes] ?? 0) < o.buyAmount} run={async () => {
                          await command(`/api/v1/cities/${city.city.id}/market/accept`, { orderId: o.id });
                          toast(t("accept") + " ✓", "ok");
                          os.reload();
                          await refresh();
                        }}>{t("accept")}</Action>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </section>
    </div>
  );
}

// ---------------- Alliance ----------------
export function AllianceScreen() {
  const { world, t, overview, refresh, toast, idx } = useGame();
  const [tab, setTab] = useState<"mine" | "list" | "rank">("mine");
  const aid = overview?.player.allianceId;
  const mine = useLoad((s) => (aid ? get<Alliance>(`/api/v1/worlds/${world.id}/alliances/${aid}`, s) : Promise.resolve(undefined)), [world.id, aid]);
  const list = useLoad((s) => get<Alliance[]>(`/api/v1/worlds/${world.id}/alliances`, s), [world.id, tab]);
  const invites = useLoad((s) => get<Invite[]>(`/api/v1/worlds/${world.id}/alliance-invites`, s), [world.id, aid]);
  const ranks = useLoad((s) => (tab === "rank" ? get<RankRow[]>(`/api/v1/worlds/${world.id}/rankings`, s) : Promise.resolve([] as RankRow[])), [world.id, tab]);
  const [name, setName] = useState("");
  const [tag, setTag] = useState("");
  const [inv, setInv] = useState("");
  const role = overview?.player.allianceRole;
  const p = (path: string, body: unknown = {}) => command(`/api/v1/worlds/${world.id}${path}`, body);
  const after = async () => { mine.reload(); invites.reload(); list.reload(); await refresh(); };
  return (
    <div className="screen">
      <h1>{t("alliance")}</h1>
      <Tabs value={tab} onChange={setTab} items={[["mine", t("alliance")], ["list", t("relations")], ["rank", t("rankings")]]} />
      {tab === "mine" && (
        <>
          {invites.data && invites.data.filter((i) => i.playerId === overview?.player.id).map((i) => (
            <div className="panel row" key={i.id}>
              <span>{t("invites")}: <strong>[{i.allianceTag}]</strong></span>
              <Action className="primary small" run={async () => { await p(`/alliance-invites/${i.id}/respond`, { accept: true }); await after(); }}>{t("accept")}</Action>
              <Action className="small ghost" run={async () => { await p(`/alliance-invites/${i.id}/respond`, { accept: false }); await after(); }}>{t("decline")}</Action>
            </div>
          ))}
          {!aid ? (
            <section className="panel stack">
              <h3>{t("createAlliance")}</h3>
              <label>{t("name")}<input value={name} maxLength={24} onChange={(e) => setName(e.target.value)} /></label>
              <label>{t("tag")}<input value={tag} maxLength={5} onChange={(e) => setTag(e.target.value)} /></label>
              <CostList cost={(idx.pack.params.allianceCreateCost ?? {}) as Record<string, number>} />
              <Action className="primary" disabled={name.length < 3 || tag.length < 2} run={async () => { await p("/alliances", { name, tag, description: "" }); toast("✓", "ok"); await after(); }}>{t("createAlliance")}</Action>
            </section>
          ) : mine.error ? <ErrorBox error={mine.error} retry={mine.reload} /> : !mine.data ? <Loading /> : (
            <section className="panel">
              <h3>[{mine.data.tag}] {mine.data.name} <span className="muted small">{t("score")} {fmt(mine.data.score)}</span></h3>
              <table className="table">
                <thead><tr><th>{t("name")}</th><th>{t("role")}</th><th>{t("score")}</th><th></th></tr></thead>
                <tbody>
                  {mine.data.members?.map((m) => (
                    <tr key={m.playerId}>
                      <td>{m.name}</td><td>{m.role}</td><td>{fmt(m.score)}</td>
                      <td className="row">
                        {role === "leader" && m.playerId !== overview?.player.id && (<>
                          <Action className="small ghost" run={async () => { await p(`/alliance/members/${m.playerId}/role`, { role: m.role === "officer" ? "member" : "officer" }); await after(); }}>{m.role === "officer" ? t("demote") : t("promote")}</Action>
                          <Action className="small ghost" confirm={t("makeLeader") + "?"} run={async () => { await p(`/alliance/members/${m.playerId}/role`, { role: "leader" }); await after(); }}>{t("makeLeader")}</Action>
                        </>)}
                        {(role === "leader" || (role === "officer" && m.role === "member")) && m.playerId !== overview?.player.id && (
                          <Action className="small danger" confirm={t("kick") + "?"} run={async () => { await p(`/alliance/members/${m.playerId}/role`, { role: "kick" }); await after(); }}>{t("kick")}</Action>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {(role === "leader" || role === "officer") && (
                <div className="row wrap">
                  <input placeholder={t("invite")} value={inv} onChange={(e) => setInv(e.target.value)} />
                  <Action className="small" disabled={inv.length < 3} run={async () => { await p("/alliance/invites", { player: inv }); setInv(""); toast(t("invite") + " ✓", "ok"); invites.reload(); }}>{t("invite")}</Action>
                </div>
              )}
              <h4>{t("relations")}</h4>
              <ul className="list">{mine.data.relations?.map((r) => <li key={r.otherId}>[{r.otherTag}] — {r.status}</li>)}</ul>
              <Action className="danger small" confirm={t("leave") + "?"} run={async () => { await p("/alliance/leave"); await after(); }}>{t("leave")}</Action>
            </section>
          )}
        </>
      )}
      {tab === "list" && (list.error ? <ErrorBox error={list.error} /> : !list.data ? <Loading /> : !list.data.length ? <Empty /> : (
        <table className="table">
          <thead><tr><th>{t("tag")}</th><th>{t("name")}</th><th>{t("members")}</th><th>{t("score")}</th><th></th></tr></thead>
          <tbody>
            {list.data.map((a) => (
              <tr key={a.id}>
                <td>[{a.tag}]</td><td>{a.name}</td><td>{a.memberCount}</td><td>{fmt(a.score)}</td>
                <td className="row">
                  {role === "leader" && a.id !== aid && (["ally", "nap", "war", "neutral"] as const).map((st) => (
                    <Action key={st} className="small ghost" run={async () => { await p("/alliance/relations", { allianceId: a.id, status: st }); mine.reload(); }}>{st}</Action>
                  ))}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      ))}
      {tab === "rank" && (!ranks.data ? <Loading /> : (
        <table className="table">
          <thead><tr><th>#</th><th>{t("name")}</th><th>{t("race")}</th><th>{t("alliance")}</th><th>{t("score")}</th></tr></thead>
          <tbody>{ranks.data.map((r) => <tr key={r.playerId} className={r.playerId === overview?.player.id ? "me" : ""}><td>{r.rank}</td><td>{r.name}</td><td>{idx.name(r.race)}</td><td>{r.alliance ?? "—"}</td><td>{fmt(r.score)}</td></tr>)}</tbody>
        </table>
      ))}
    </div>
  );
}

// ---------------- Mail & notifications ----------------
export function MailScreen() {
  const { world, t, refresh, toast, idx } = useGame();
  const [tab, setTab] = useState<"notifications" | "inbox" | "sent" | "alliance">("notifications");
  const [compose, setCompose] = useState<{ to: string; alliance: boolean } | null>(null);
  const ns = useLoad((s) => get<{ items: Notification[]; unread: number }>(`/api/v1/worlds/${world.id}/notifications`, s), [world.id, tab]);
  const ms = useLoad((s) => (tab === "notifications" ? Promise.resolve([] as Message[]) : get<Message[]>(`/api/v1/worlds/${world.id}/messages?box=${tab}`, s)), [world.id, tab]);
  const describe = (n: Notification) => {
    const d = n.data;
    switch (n.kind) {
      case "build_done": case "research_done": return `${idx.name(d.item)} → ${d.level}`;
      case "train_done": return `${idx.name(d.item)} ×${d.quantity}`;
      case "report": return `${d.kind === "battle" ? t("battle") : t("scout")} (${d.side})`;
      case "incoming_attack": return `⚠ ${t("incoming")}: ${d.from} → ${d.x}:${d.y}`;
      case "message": return `${t("message")}: ${d.from}`;
      case "mission_returned": return `${t(("kind_" + d.kind) as any)} — ${t("returning")} ✓`;
      default: return n.kind.replace(/_/g, " ");
    }
  };
  return (
    <div className="screen">
      <h1>{t("mail")}</h1>
      <div className="row wrap">
        <Tabs value={tab} onChange={setTab} items={[["notifications", t("notifications")], ["inbox", t("inbox")], ["sent", t("sent")], ["alliance", t("allianceChat")]]} />
        <button className="btn small primary" onClick={() => setCompose({ to: "", alliance: tab === "alliance" })}>{t("compose")}</button>
      </div>
      {tab === "notifications" ? (
        ns.error ? <ErrorBox error={ns.error} retry={ns.reload} /> : !ns.data ? <Loading /> : !ns.data.items.length ? <Empty /> : (
          <>
            <Action className="small ghost" disabled={!ns.data.unread} run={async () => {
              await command(`/api/v1/worlds/${world.id}/notifications/read`, { upTo: ns.data!.items[0].id });
              ns.reload();
              await refresh();
            }}>{t("markAllRead")}</Action>
            <ul className="list">
              {ns.data.items.map((n) => (
                <li key={n.id} className={n.readAt ? "" : "unread"}><span>{describe(n)}</span><span className="muted small">{new Date(n.createdAt).toLocaleString()}</span></li>
              ))}
            </ul>
          </>
        )
      ) : ms.error ? <ErrorBox error={ms.error} retry={ms.reload} /> : !ms.data ? <Loading /> : !ms.data.length ? <Empty /> : (
        <ul className="list messages">
          {ms.data.map((m) => (
            <li key={m.id} className={!m.readAt && tab === "inbox" ? "unread" : ""}>
              <details onToggle={(e) => { if ((e.target as HTMLDetailsElement).open && !m.readAt && tab === "inbox") command(`/api/v1/worlds/${world.id}/messages/${m.id}/read`).catch(() => {}); }}>
                <summary><strong>{tab === "sent" ? `→ ${m.to}` : m.from}</strong> · {m.subject || "—"} <span className="muted small">{new Date(m.createdAt).toLocaleString()}</span></summary>
                <p className="body">{m.body}</p>
                {tab !== "sent" && (
                  <div className="row">
                    <button className="btn small" onClick={() => setCompose({ to: m.from, alliance: tab === "alliance" })}>↩</button>
                    <Action className="small ghost" run={async () => { await command(`/api/v1/worlds/${world.id}/messages/${m.id}/report`); toast(t("reportAbuse") + " ✓", "ok"); }}>{t("reportAbuse")}</Action>
                    <Action className="small ghost" confirm={t("block") + "?"} run={async () => { await command(`/api/v1/worlds/${world.id}/blocks`, { playerId: m.fromId, block: true }); ms.reload(); }}>{t("block")}</Action>
                  </div>
                )}
              </details>
            </li>
          ))}
        </ul>
      )}
      {compose && <Compose init={compose} onClose={() => { setCompose(null); ms.reload(); }} />}
    </div>
  );
}

function Compose({ init, onClose }: { init: { to: string; alliance: boolean }; onClose: () => void }) {
  const { world, t, toast } = useGame();
  const [to, setTo] = useState(init.to);
  const [subject, setSubject] = useState("");
  const [body, setBody] = useState("");
  return (
    <Modal title={t("compose")} onClose={onClose}>
      <div className="stack">
        {!init.alliance && <label>{t("to")}<input value={to} onChange={(e) => setTo(e.target.value)} /></label>}
        <label>{t("subject")}<input value={subject} maxLength={120} onChange={(e) => setSubject(e.target.value)} /></label>
        <label>{t("message")}<textarea rows={6} value={body} maxLength={4000} onChange={(e) => setBody(e.target.value)} /></label>
        <Action className="primary" disabled={!body.trim() || (!init.alliance && to.length < 3)} run={async () => {
          await command(`/api/v1/worlds/${world.id}/messages`, { to, alliance: init.alliance, subject, body });
          toast(t("send") + " ✓", "ok");
          onClose();
        }}>{t("send")}</Action>
      </div>
    </Modal>
  );
}

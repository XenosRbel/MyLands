import { useState, type ReactNode } from "react";
import { useGame, type Screen } from "../game/context";
import { gameAudioEnabled, playGameCue, setGameAudioEnabled } from "../game/audio";
import { Icon, fmt, Loading, ErrorBox, Timer } from "../ui/kit";
import { CityScreen } from "./City";
import { EconomyScreen } from "./Economy";
import { ResearchScreen } from "./Research";
import { ArmyScreen } from "./Army";
import { MapScreen } from "./Map";
import { MissionsScreen, ReportsScreen, QuestsScreen } from "./Missions";
import { HeroesScreen, MarketScreen, AllianceScreen, MailScreen } from "./Social";
import { ProfileScreen } from "./Profile";

const icons: Record<Screen, string> = {
  city: "castle", economy: "chart", research: "book", army: "sword", heroes: "star", map: "map", missions: "flag",
  reports: "scroll", market: "scale", alliance: "shield", quests: "star", mail: "mail", profile: "user", settings: "gear",
};
const primary: Screen[] = ["city", "map", "army", "missions"];

export function Shell() {
  const g = useGame();
  const { t, idx, overview, city, screen, go } = g;
  const [more, setMore] = useState(false);
  const [sound, setSound] = useState(() => {
    try { return localStorage.getItem("rf.sound") === "on"; } catch { return false; }
  });
  setGameAudioEnabled(sound);
  const navigate = (s: Screen) => { playGameCue("navigate"); go(s); };
  const enabled: Screen[] = (idx.pack.ui.screens as Screen[]).filter((s) => s !== "settings");
  const badge = (s: Screen) =>
    s === "reports" ? overview?.unreadReports : s === "quests" ? overview?.questsReady : s === "mail" ? overview?.unreadNotifications : 0;

  let body: ReactNode;
  if (!city) body = g.cityError ? <ErrorBox error={g.cityError} retry={g.refresh} /> : <Loading />;
  else
    body = {
      city: <CityScreen />, economy: <EconomyScreen />, research: <ResearchScreen />, army: <ArmyScreen />, heroes: <HeroesScreen />,
      map: <MapScreen />, missions: <MissionsScreen />, reports: <ReportsScreen />, market: <MarketScreen />, alliance: <AllianceScreen />,
      quests: <QuestsScreen />, mail: <MailScreen />, profile: <ProfileScreen />, settings: <ProfileScreen />,
    }[screen];

  const nextTutorial = overview && city ? <TutorialBar /> : null;

  return (
    <div className="shell">
      <header className="topbar">
        <div className="title">
          <span className="game-title">{idx.title}</span>
          {overview && overview.cities.length > 1 ? (
            <select aria-label="City" value={g.cityId} onChange={(e) => g.setCityId(Number(e.target.value))}>
              {overview.cities.map((c) => (
                <option key={c.id} value={c.id}>{c.name} ({c.x}:{c.y})</option>
              ))}
            </select>
          ) : (
            <span className="city-name">{city?.city.name}</span>
          )}
        </div>
        {city && (
          <div className="resbar" aria-label="Resources">
            {idx.pack.resources.map((r) => {
              const n = city.resources[r.id] ?? 0;
              const cap = city.rates.storage[r.id];
              const net = city.rates.net[r.id] ?? 0;
              const full = cap !== undefined && n >= cap;
              return (
                <div key={r.id} className={"resbar-item" + (full ? " full" : "") + (net < 0 ? " neg" : "")}
                  title={`${idx.name(r.id)}: ${fmt(n)}${cap !== undefined ? " / " + fmt(cap) : ""} (${net >= 0 ? "+" : ""}${fmt(net)}${t("perHour")})`}>
                  <Icon name={idx.icon(r.id) || "star"} size={16} />
                  <span className="amt">{fmt(n)}</span>
                  <span className="rate">{net >= 0 ? "+" : ""}{fmt(net)}</span>
                </div>
              );
            })}
            <div className="resbar-item" title={t("population")}><Icon name="people" size={16} /><span className="amt">{fmt(city.city.population)}</span><span className="rate">/{fmt(city.rates.housing)}</span></div>
          </div>
        )}
        <div className="experience-controls">
          <button type="button" className="experience-sound" aria-label={sound ? t("soundOff") : t("soundOn")} title={sound ? t("soundOff") : t("soundOn")} aria-pressed={sound} onClick={() => {
            const next = !sound;
            setSound(next);
            setGameAudioEnabled(next);
            try { localStorage.setItem("rf.sound", next ? "on" : "off"); } catch { /* storage unavailable */ }
            if (next) playGameCue("confirm");
          }}>{sound ? "♫" : "♪"} <span>{sound ? t("soundOn") : t("soundOff")}</span></button>
        <button className="icon-btn bell" aria-label={t("notifications")} onClick={() => navigate("mail")}>
          <Icon name="bell" />
          {!!overview?.unreadNotifications && <span className="badge">{overview.unreadNotifications}</span>}
        </button>
        </div>
      </header>
      <nav className="sidenav" aria-label="Main">
        {enabled.map((s) => (
          <button key={s} className={screen === s ? "on" : ""} onClick={() => navigate(s)} aria-current={screen === s ? "page" : undefined}>
            <Icon name={icons[s]} /> <span>{t(s as Exclude<Screen, "settings">)}</span>
            {!!badge(s) && <span className="badge">{badge(s)}</span>}
          </button>
        ))}
      </nav>
      <main className="content" id="main">
        {nextTutorial}
        {body}
      </main>
      <nav className="bottomnav" aria-label="Main">
        {primary.filter((s) => enabled.includes(s)).map((s) => (
          <button key={s} className={screen === s ? "on" : ""} onClick={() => go(s)}>
            <Icon name={icons[s]} /> <span>{t(s)}</span>
          </button>
        ))}
        <button className={more ? "on" : ""} onClick={() => setMore(!more)} aria-expanded={more}>
          <Icon name="menu" /> <span>{t("more")}</span>
          {!!(overview && overview.questsReady + overview.unreadReports) && <span className="badge dot" />}
        </button>
      </nav>
      {more && (
        <div className="sheet" onClick={() => setMore(false)}>
          <div className="sheet-body" onClick={(e) => e.stopPropagation()}>
            {enabled.filter((s) => !primary.includes(s)).map((s) => (
              <button key={s} onClick={() => { navigate(s); setMore(false); }}>
                <Icon name={icons[s]} size={22} /> <span>{t(s)}</span>
                {!!badge(s) && <span className="badge">{badge(s)}</span>}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

/** Shows the next tutorial step until the tutorial chain is finished. */
function TutorialBar() {
  const { overview, t, idx, go } = useGame();
  const [hidden, setHidden] = useState(false);
  const prot = overview?.player.protectedUntil;
  const q = idx.pack.quests.find((x: any) => x.tutorial && !(overview?.player.stats?.["tutorial_done." + x.id]));
  if (hidden) return null;
  return (
    <div className="tutorial-bar">
      {prot && Date.parse(prot) > Date.now() && (
        <span className="pill"><Icon name="shield" size={14} /> {t("protection")} <Timer until={prot} /></span>
      )}
      {q && overview!.questsReady > 0 && (
        <button className="pill accent" onClick={() => go("quests")}><Icon name="star" size={14} /> {t("quests")}: {overview!.questsReady}</button>
      )}
      <button className="icon-btn small" aria-label={t("close")} onClick={() => setHidden(true)}><Icon name="close" size={14} /></button>
    </div>
  );
}

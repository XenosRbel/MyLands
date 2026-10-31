import { useEffect, useState, type ReactNode } from "react";
import { get } from "../api/client";
import type { HeroesView } from "../api/types";
import { useGame, type Screen } from "../game/context";
import { playGameCue, setGameAudioEnabled } from "../game/audio";
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
const illustratedNav: Partial<Record<Screen, string>> = {
  city: "keep", map: "forest", army: "knight", research: "academy",
  heroes: "marshal", economy: "mint", market: "market",
  missions: "watchtower", reports: "scriptorium",
};


export function Shell() {
  const g = useGame();
  const { t, idx, overview, city, screen, go } = g;
  const [more, setMore] = useState(false);
  const [heroClass, setHeroClass] = useState<string | null>(null);
  const [heroLevel, setHeroLevel] = useState<number | null>(null);
  const [clockTime, setClockTime] = useState(() => Date.now());
  useEffect(() => {
    const timer = window.setInterval(() => setClockTime(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, []);
  const clock = new Intl.DateTimeFormat(g.locale === "ru" ? "ru-RU" : "en-GB", {
    hour: "2-digit", minute: "2-digit", second: "2-digit", hour12: false,
  }).format(clockTime);
  useEffect(() => {
    let alive = true;
    get<HeroesView>(`/api/v1/worlds/${g.world.id}/heroes`)
      .then((data) => { if (alive) { setHeroClass(data.heroes[0]?.class ?? null); setHeroLevel(data.heroes[0]?.level ?? null); } })
      .catch(() => { if (alive) { setHeroClass(null); setHeroLevel(null); } });
    return () => { alive = false; };
  }, [g.world.id, screen]);

  const [sound, setSound] = useState(() => {
    try { return localStorage.getItem("rf.sound") === "on"; } catch { return false; }
  });
  setGameAudioEnabled(sound);
  const navigate = (s: Screen) => { playGameCue("navigate"); go(s); };
  const specialization = idx.pack.techs
    .filter((tech) => tech.exclusiveGroup === "calling")
    .find((tech) => (overview?.player.techs?.[tech.id] ?? 0) > 0);
  const specializationLevel = specialization ? overview?.player.techs?.[specialization.id] ?? 0 : 0;
  const specializationArt: Record<string, string> = {
    path_architect: "keep",
    path_steward: "farmstead",
    path_warlord: "warhall",
  };
  const specializationImage = specialization
    ? idx.art(specializationArt[specialization.id] ?? "academy")
    : "";

  const specializationBonuses = specialization?.effects
    .filter((effect) => effect.type === "modifier" && effect.modifier)
    .map((effect) => {
      const modifier = effect.modifier!;
      return `${modifier.target}: ${modifier.value * specializationLevel > 0 ? "+" : ""}${modifier.value * specializationLevel}${modifier.op === "pct" ? "%" : ""}`;
    }).join("; ") ?? t("chooseSpecialization");
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
        <div className="game-clock" aria-label={t("gameTime")} role="timer">{clock}</div>
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
        <div className="kingdom-shortcuts">
          <button type="button" className="kingdom-token" onClick={() => navigate("research")}
            title={specialization ? `${idx.name(specialization.id)} (${specializationLevel}): ${specializationBonuses}` : t("chooseSpecialization")}
            aria-label={specialization ? idx.name(specialization.id) : t("chooseSpecialization")}>
            {specializationImage ?
              <img src={specializationImage} alt="" /> : <Icon name="shield" size={28} />}
            {specializationLevel > 0 && <span className="kingdom-token-level">{specializationLevel}</span>}
            <span className="token-label">{specialization ? idx.name(specialization.id) : t("chooseSpecialization")}</span>
          </button>
          <button type="button" className="kingdom-token hero-token" onClick={() => navigate("heroes")}
            title={t("heroes")} aria-label={t("heroes")}>
            {heroClass && idx.art(heroClass) ?
              <img src={idx.art(heroClass)} alt="" /> :
              <Icon name="user" size={28} />}
            {heroLevel !== null && <span className="kingdom-token-level hero-level">{heroLevel}</span>}
            <span className="token-label">{t("heroes")}</span>
          </button>
        </div>
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
            {illustratedNav[s] && idx.art(illustratedNav[s]) ?
              <img className="sidenav-illustrated-icon" src={idx.art(illustratedNav[s])} alt="" loading="lazy" /> :
              <Icon name={icons[s]} />}
            <span>{t(s as Exclude<Screen, "settings">)}</span>
            {!!badge(s) && <span className="badge">{badge(s)}</span>}
          </button>
        ))}
      </nav>
      <main className="content" id="main">
        {nextTutorial}
        {body}
      </main>
      <nav className="game-dock" aria-label={t("army")}>
        <div className="game-dock-actions">
          {(["city", "map", "army", "research", "heroes", "reports"] as Screen[])
            .filter((destination) => enabled.includes(destination))
            .map((destination) => (
              <button key={destination} type="button" className={screen === destination ? "active" : ""}
                aria-current={screen === destination ? "page" : undefined}
                onClick={() => navigate(destination)} title={t(destination as Exclude<Screen, "settings">)}>
                {illustratedNav[destination] && idx.art(illustratedNav[destination]) ?
                  <img className="game-dock-navigation-art" src={idx.art(illustratedNav[destination])} alt="" loading="lazy" /> :
                  <Icon name={icons[destination]} size={23} />}
                <span>{t(destination as Exclude<Screen, "settings">)}</span>
              </button>
            ))}
        </div>
        <div className="game-dock-city">
          <strong>{city?.city.name}</strong>
          <small>{city ? `${city.city.x}:${city.city.y}` : ""}</small>
        </div>
        <div className="game-dock-army">
          {city?.units.slice().sort((a, b) =>
            (city.city.units[b.id] ?? 0) - (city.city.units[a.id] ?? 0)
          ).slice(0, 13).map((view) => (
            <button type="button" key={view.id}
              className={"game-dock-unit" + ((city.city.units[view.id] ?? 0) === 0 ? " inactive" : "")}
              title={`${idx.name(view.id)}: ${fmt(city.city.units[view.id] ?? 0)}`}
              onClick={() => navigate("army")}>
              {idx.art(view.id) ?
                <img src={idx.art(view.id)} alt="" loading="lazy" /> :
                <Icon name="shield" size={23} />}
              <span>{fmt(city.city.units[view.id] ?? 0)}</span>
            </button>
          ))}
        </div>
      </nav>
      <nav className="bottomnav" aria-label="Main">
        {primary.filter((s) => enabled.includes(s)).map((s) => (
          <button key={s} className={screen === s ? "on" : ""} onClick={() => navigate(s)}>
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

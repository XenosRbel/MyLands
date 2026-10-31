import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import { api, get, syncServerTime } from "../api/client";
import type { Account, CityView, Overview, World } from "../api/types";
import { makeT, type UIKey } from "../i18n/ui";
import { applyTheme, PackIndex, type Pack } from "./pack";
import type { PlatformAdapter } from "../platform/adapter";

export type Screen =
  | "city" | "economy" | "research" | "army" | "heroes" | "map" | "missions" | "reports"
  | "market" | "alliance" | "quests" | "mail" | "profile" | "settings";

interface Toast {
  id: number;
  text: string;
  kind: "info" | "error" | "ok";
}

export interface GameCtx {
  account: Account;
  world: World;
  idx: PackIndex;
  locale: string;
  setLocale: (l: string) => void;
  t: (k: UIKey) => string;
  overview?: Overview;
  cityId?: number;
  setCityId: (id: number) => void;
  city?: CityView;
  cityError?: unknown;
  refresh: () => Promise<void>;
  screen: Screen;
  go: (s: Screen, focus?: { x: number; y: number }) => void;
  mapFocus?: { x: number; y: number };
  toast: (text: string, kind?: Toast["kind"]) => void;
  platform: PlatformAdapter;
  leaveWorld: () => void;
  base: string;
}

const Ctx = createContext<GameCtx | null>(null);

export function useGame(): GameCtx {
  const c = useContext(Ctx);
  if (!c) throw new Error("useGame outside GameProvider");
  return c;
}

const SCREEN_KEY = "rf.screen";
const CITY_KEY = "rf.city";

function load(key: string) {
  try {
    return localStorage.getItem(key);
  } catch {
    return null;
  }
}
function save(key: string, v: string) {
  try {
    localStorage.setItem(key, v);
  } catch {
    /* ignore */
  }
}

export function GameProvider(props: {
  account: Account;
  world: World;
  pack: Pack;
  locale: string;
  setLocale: (l: string) => void;
  platform: PlatformAdapter;
  leaveWorld: () => void;
  children: ReactNode;
}) {
  const { pack, locale } = props;
  const idx = useMemo(() => new PackIndex(pack, locale), [pack, locale]);
  const t = useMemo(() => makeT(idx.locale), [idx.locale]);
  const [overview, setOverview] = useState<Overview>();
  const [cityId, setCityIdState] = useState<number | undefined>(() => Number(load(CITY_KEY)) || undefined);
  const [city, setCity] = useState<CityView>();
  const [cityError, setCityError] = useState<unknown>();
  const [screen, setScreen] = useState<Screen>(() => (load(SCREEN_KEY) as Screen) || "city");
  const [mapFocus, setMapFocus] = useState<{ x: number; y: number }>();
  const [toasts, setToasts] = useState<Toast[]>([]);
  const paused = useRef(false);
  const wid = props.world.id;

  useEffect(() => applyTheme(pack), [pack]);

  const toast = useCallback((text: string, kind: Toast["kind"] = "info") => {
    const id = Date.now() + Math.random();
    setToasts((ts) => [...ts.slice(-3), { id, text, kind }]);
    setTimeout(() => setToasts((ts) => ts.filter((x) => x.id !== id)), kind === "error" ? 6000 : 3500);
  }, []);

  const refresh = useCallback(async () => {
    try {
      const ov = await get<Overview>(`/api/v1/worlds/${wid}/overview`);
      syncServerTime(ov.serverTime);
      setOverview(ov);
      let cid = cityId;
      if (!cid || !ov.cities.some((c) => c.id === cid)) {
        cid = ov.cities[0]?.id;
        setCityIdState(cid);
      }
      if (cid) {
        const cv = await get<CityView>(`/api/v1/cities/${cid}`);
        syncServerTime(cv.serverTime);
        setCity(cv);
        setCityError(undefined);
      }
    } catch (e) {
      setCityError(e);
    }
  }, [wid, cityId]);

  // Initial load + polling. REST is authoritative; polling keeps queues and
  // notifications fresh without a socket. Paused while the game is hidden.
  useEffect(() => {
    refresh();
    const id = setInterval(() => !paused.current && refresh(), 15000);
    return () => clearInterval(id);
  }, [refresh]);

  useEffect(() => {
    props.platform.onPause(() => {
      paused.current = true;
      props.platform.gameplayStop();
    });
    props.platform.onResume(() => {
      paused.current = false;
      props.platform.gameplayStart();
      refresh();
    });
    props.platform.gameplayStart();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const go = useCallback((s: Screen, focus?: { x: number; y: number }) => {
    setScreen(s);
    save(SCREEN_KEY, s);
    if (focus) setMapFocus(focus);
    window.scrollTo?.({ top: 0 });
  }, []);

  const setCityId = useCallback((id: number) => {
    setCityIdState(id);
    save(CITY_KEY, String(id));
  }, []);

  const value: GameCtx = {
    account: props.account, world: props.world, idx, locale: idx.locale, setLocale: props.setLocale, t,
    overview, cityId, setCityId, city, cityError, refresh, screen, go, mapFocus, toast,
    platform: props.platform, leaveWorld: props.leaveWorld, base: api.base,
  };
  return (
    <Ctx.Provider value={value}>
      {props.children}
      <div className="toasts" aria-live="polite">
        {toasts.map((x) => (
          <div key={x.id} className={"toast " + x.kind}>{x.text}</div>
        ))}
      </div>
    </Ctx.Provider>
  );
}

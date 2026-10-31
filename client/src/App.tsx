import { useEffect, useState } from "react";
import { api, get } from "./api/client";
import type { Account, Session, World } from "./api/types";
import { GameProvider } from "./game/context";
import type { Pack } from "./game/pack";
import { AuthScreen, WorldScreen } from "./screens/Auth";
import { Shell } from "./screens/Shell";
import type { PlatformAdapter } from "./platform/adapter";
import { supportedLocales } from "./i18n/ui";

const WORLD_KEY = "rf.world";
const LOCALE_KEY = "rf.locale";

function initialLocale(platform: PlatformAdapter): string {
  let l: string | null = null;
  try {
    l = localStorage.getItem(LOCALE_KEY);
  } catch {
    /* ignore */
  }
  const cand = l || platform.locale() || "en";
  return supportedLocales.includes(cand) ? cand : "en";
}

export function App({ platform }: { platform: PlatformAdapter }) {
  const [account, setAccount] = useState<Account | null>(null);
  const [booting, setBooting] = useState(true);
  const [entered, setEntered] = useState<{ world: World; pack: Pack } | null>(null);
  const [locale, setLocaleState] = useState(() => initialLocale(platform));

  const setLocale = (l: string) => {
    setLocaleState(l);
    try {
      localStorage.setItem(LOCALE_KEY, l);
    } catch {
      /* ignore */
    }
  };

  // Restore session and last world after a page reload.
  useEffect(() => {
    const restore = async () => {
      if (!api.token) return;
      try {
        const me = await get<Account>("/api/v1/me");
        setAccount(me);
        let wid: number | null = null;
        try {
          wid = Number(localStorage.getItem(WORLD_KEY)) || null;
        } catch {
          /* ignore */
        }
        if (wid) {
          const worlds = await get<World[]>("/api/v1/worlds");
          const w = worlds.find((x) => x.id === wid && x.joined);
          if (w) setEntered({ world: w, pack: await get<Pack>(`/api/v1/worlds/${w.id}/pack`) });
        }
      } catch {
        /* token invalid → login */
      }
    };
    restore().finally(() => {
      setBooting(false);
      platform.signalReady();
    });
    return api.onAuthChange(() => {
      if (!api.token) {
        setAccount(null);
        setEntered(null);
      }
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (booting) return <div className="boot"><span className="spinner" /></div>;
  if (!account) return <AuthScreen locale={locale} onSession={(s: Session) => setAccount(s.account)} />;
  if (!entered)
    return (
      <WorldScreen account={account} locale={locale} onLogout={() => api.setToken(null)}
        onEnter={(world, pack) => {
          try {
            localStorage.setItem(WORLD_KEY, String(world.id));
          } catch {
            /* ignore */
          }
          setEntered({ world, pack });
        }} />
    );
  return (
    <GameProvider account={account} world={entered.world} pack={entered.pack} locale={locale} setLocale={setLocale}
      platform={platform} leaveWorld={() => setEntered(null)}>
      <Shell />
    </GameProvider>
  );
}

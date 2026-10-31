import { useEffect, useState } from "react";
import { api, ApiError, get, request } from "../api/client";
import type { Account, Session, World } from "../api/types";
import { errorMessage, makeT } from "../i18n/ui";
import type { Pack } from "../game/pack";
import { PackIndex } from "../game/pack";

export function AuthScreen({ locale, onSession }: { locale: string; onSession: (s: Session) => void }) {
  const t = makeT(locale);
  const [mode, setMode] = useState<"login" | "register">("register");
  const [login, setLogin] = useState("");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState("");

  const run = async (fn: () => Promise<Session>) => {
    setBusy(true);
    setErr("");
    try {
      const s = await fn();
      api.setToken(s.token);
      onSession(s);
    } catch (e) {
      const x = e as ApiError;
      setErr(errorMessage(locale, x.code, x.message));
    } finally {
      setBusy(false);
    }
  };
  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    run(() => request<Session>(`/api/v1/auth/${mode}`, { method: "POST", body: { login, password } }));
  };
  return (
    <main className="auth">
      <div className="auth-card">
        <div className="crest" aria-hidden>
          <svg viewBox="0 0 64 64" width="72" height="72"><path d="M32 4l24 10v16c0 16-11 26-24 30C19 56 8 46 8 30V14z" fill="var(--surface2)" stroke="var(--accent)" strokeWidth="2.5" /><path d="M20 38l6-14 6 8 6-8 6 14z" fill="var(--accent)" /></svg>
        </div>
        <h1 className="brand">Realmforge</h1>
        <form onSubmit={submit} className="stack">
          <label>{t("login")}<input autoComplete="username" value={login} onChange={(e) => setLogin(e.target.value)} required minLength={3} maxLength={64} /></label>
          <label>{t("password")}<input type="password" autoComplete={mode === "login" ? "current-password" : "new-password"} value={password} onChange={(e) => setPassword(e.target.value)} required minLength={8} maxLength={128} /></label>
          {err && <div className="error-box" role="alert">{err}</div>}
          <button className="btn primary" disabled={busy}>{mode === "login" ? t("signIn") : t("register")}</button>
        </form>
        <button className="link" onClick={() => setMode(mode === "login" ? "register" : "login")}>
          {mode === "login" ? t("newAccount") : t("haveAccount")}
        </button>
        <div className="divider" />
        <button className="btn" disabled={busy} onClick={() => run(() => request<Session>("/api/v1/auth/guest", { method: "POST", body: {} }))}>
          {t("playAsGuest")}
        </button>
      </div>
    </main>
  );
}

export function WorldScreen({ account, locale, onEnter, onLogout }: {
  account: Account;
  locale: string;
  onEnter: (w: World, pack: Pack) => void;
  onLogout: () => void;
}) {
  const t = makeT(locale);
  const [worlds, setWorlds] = useState<World[]>();
  const [err, setErr] = useState("");
  const [joining, setJoining] = useState<{ world: World; pack: Pack } | null>(null);
  const [name, setName] = useState(account.displayName.replace(/[^\p{L}\p{N}_ .'-]/gu, "").slice(0, 24));
  const [race, setRace] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    get<World[]>("/api/v1/worlds").then(setWorlds, (e) => setErr(String((e as Error).message)));
  }, []);

  const loadPack = (w: World) => get<Pack>(`/api/v1/worlds/${w.id}/pack`);

  const enter = async (w: World) => {
    setBusy(true);
    try {
      const p = await loadPack(w);
      if (w.joined) onEnter(w, p);
      else {
        setJoining({ world: w, pack: p });
        setRace(p.races[0]?.id ?? "");
      }
    } catch (e) {
      setErr((e as Error).message);
    } finally {
      setBusy(false);
    }
  };
  const join = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!joining) return;
    setBusy(true);
    setErr("");
    try {
      await request(`/api/v1/worlds/${joining.world.id}/join`, { method: "POST", body: { name, race } });
      onEnter({ ...joining.world, joined: true }, joining.pack);
    } catch (x) {
      const ae = x as ApiError;
      setErr(errorMessage(locale, ae.code, ae.message));
    } finally {
      setBusy(false);
    }
  };

  if (joining) {
    const idx = new PackIndex(joining.pack, locale);
    return (
      <main className="auth">
        <form className="auth-card wide stack" onSubmit={join}>
          <h1 className="brand">{idx.title}</h1>
          <p className="muted">{idx.text(joining.pack.meta.description)}</p>
          <label>{t("rulerName")}<input value={name} onChange={(e) => setName(e.target.value)} required minLength={3} maxLength={24} /></label>
          <fieldset className="races">
            <legend>{t("chooseRace")}</legend>
            {joining.pack.races.map((r) => (
              <label key={r.id} className={"race" + (race === r.id ? " on" : "")} style={{ borderColor: race === r.id ? r.color : undefined }}>
                <input type="radio" name="race" value={r.id} checked={race === r.id} onChange={() => setRace(r.id)} />
                <strong style={{ color: r.color }}>{idx.text(r.name)}</strong>
                <span className="muted small">{idx.text(r.desc)}</span>
              </label>
            ))}
          </fieldset>
          {err && <div className="error-box" role="alert">{err}</div>}
          <div className="row end">
            <button type="button" className="btn ghost" onClick={() => setJoining(null)}>{t("cancel")}</button>
            <button className="btn primary" disabled={busy || !race}>{t("joinWorld")}</button>
          </div>
        </form>
      </main>
    );
  }
  return (
    <main className="auth">
      <div className="auth-card wide">
        <h1 className="brand">{t("chooseWorld")}</h1>
        {err && <div className="error-box" role="alert">{err}</div>}
        {!worlds ? (
          <div className="loading"><span className="spinner" /> {t("appLoading")}</div>
        ) : (
          <ul className="worlds">
            {worlds.map((w) => (
              <li key={w.id}>
                <div>
                  <strong>{w.name}</strong>
                  <div className="muted small">{w.packId} {w.packVersion} · {w.players} {t("players")} · {t("speed")} ×{w.speed}</div>
                </div>
                <button className="btn primary" disabled={busy} onClick={() => enter(w)}>{w.joined ? t("enterWorld") : t("joinWorld")}</button>
              </li>
            ))}
          </ul>
        )}
        <button className="link" onClick={onLogout}>{t("logout")}</button>
      </div>
    </main>
  );
}

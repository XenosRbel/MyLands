import { useState } from "react";
import { api, request } from "../api/client";
import { useGame } from "../game/context";
import { supportedLocales } from "../i18n/ui";
import { Action, fmt, Timer } from "../ui/kit";

export function ProfileScreen() {
  const { account, overview, t, idx, locale, setLocale, leaveWorld, world, toast } = useGame();
  const [login, setLogin] = useState("");
  const [password, setPassword] = useState("");
  const [motion, setMotion] = useState(() => document.documentElement.dataset.motion === "reduce");
  const p = overview?.player;
  const locales = idx.pack.locales.filter((l) => supportedLocales.includes(l));
  return (
    <div className="screen">
      <h1>{t("profile")}</h1>
      <div className="grid-2">
        <section className="panel">
          <h3>{p?.name}</h3>
          <dl className="kv">
            <dt>{t("race")}</dt><dd>{p && idx.name(p.race)}</dd>
            <dt>{t("score")}</dt><dd>{fmt(p?.score ?? 0)}</dd>
            <dt>{t("alliance")}</dt><dd>{overview?.allianceTag ? `[${overview.allianceTag}]` : "—"}</dd>
            <dt>{t("city")}</dt><dd>{overview?.cities.length} / {overview?.cityCap}</dd>
            <dt>{t("territories")}</dt><dd>{overview?.territories} / {overview?.territoryCap}</dd>
            {p?.protectedUntil && Date.parse(p.protectedUntil) > api.now() && (<><dt>{t("protection")}</dt><dd><Timer until={p.protectedUntil} /></dd></>)}
            <dt>{t("serverTime")}</dt><dd>{new Date(api.now()).toUTCString()}</dd>
          </dl>
          <p className="small muted">{world.name} · {world.packId} {overview?.packVersion}</p>
          <button className="btn" onClick={leaveWorld}>{t("chooseWorld")}</button>
        </section>
        <section className="panel stack">
          <h3>{t("settings")}</h3>
          <label>{t("language")}
            <select value={locale} onChange={(e) => { setLocale(e.target.value); request("/api/v1/me/locale", { method: "PUT", body: { locale: e.target.value } }).catch(() => {}); }}>
              {locales.map((l) => <option key={l} value={l}>{l.toUpperCase()}</option>)}
            </select>
          </label>
          <label className="check"><input type="checkbox" checked={motion} onChange={(e) => {
            setMotion(e.target.checked);
            document.documentElement.dataset.motion = e.target.checked ? "reduce" : "";
            try { localStorage.setItem("rf.motion", e.target.checked ? "reduce" : ""); } catch { /* ignore */ }
          }} /> {t("reducedMotion")}</label>
          <h3>{t("account")}</h3>
          <p className="small">{account.login ?? account.displayName}</p>
          {account.isGuest && (
            <form className="stack" onSubmit={(e) => e.preventDefault()}>
              <p className="warn small">{t("guestWarning")}</p>
              <label>{t("login")}<input value={login} onChange={(e) => setLogin(e.target.value)} autoComplete="username" /></label>
              <label>{t("password")}<input type="password" value={password} onChange={(e) => setPassword(e.target.value)} autoComplete="new-password" /></label>
              <Action className="primary" disabled={login.length < 3 || password.length < 8} run={async () => {
                await request("/api/v1/auth/upgrade", { method: "POST", body: { login, password } });
                toast("✓", "ok");
              }}>{t("saveAccount")}</Action>
            </form>
          )}
          <button className="btn ghost" onClick={async () => { await request("/api/v1/auth/logout", { method: "POST", body: {} }).catch(() => {}); api.setToken(null); }}>{t("logout")}</button>
          <Action className="danger" confirm={t("deleteConfirm")} run={async () => {
            await request("/api/v1/me", { method: "DELETE" });
            api.setToken(null);
          }}>{t("deleteAccount")}</Action>
        </section>
      </div>
    </div>
  );
}

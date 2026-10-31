import { useEffect, useRef, useState, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { api, ApiError } from "../api/client";
import { useGame } from "../game/context";
import type { Amounts, Requirement } from "../api/types";
import { errorMessage } from "../i18n/ui";

// ---------- Rendered WebP glyphs, not inline vectors ----------
const utilityNames = new Set([
  "log", "wheat", "stone", "ingot", "coin", "scroll", "crystal", "leaf",
  "bolt", "chip", "sword", "shield", "heart", "boot", "bag", "people",
  "castle", "clock", "bell", "map", "flag", "star", "book", "mail",
  "chart", "user", "gear", "scale", "menu", "eye", "skull", "close",
]);

export function Icon({ name, size = 18, title }: { name: string; size?: number; title?: string }) {
  const glyph = name.replace(/^glyph:/, "");
  const asset = name.includes("/") ? name :
    `/ui/glyphs/${utilityNames.has(glyph) ? glyph : "star"}.webp`;
  return <img className="icon img" src={asset} width={size} height={size}
    alt={title ?? ""} title={title} decoding="async" draggable={false}
    aria-hidden={title ? undefined : true} />;
}

export function fmt(n: number): string {
  if (!Number.isFinite(n)) return "∞";
  const a = Math.abs(n);
  if (a >= 1e9) return (n / 1e9).toFixed(1) + "B";
  if (a >= 1e6) return (n / 1e6).toFixed(1) + "M";
  if (a >= 1e4) return (n / 1e3).toFixed(1) + "k";
  return Math.floor(n).toLocaleString();
}

export function duration(sec: number): string {
  sec = Math.max(0, Math.ceil(sec));
  const d = Math.floor(sec / 86400), h = Math.floor((sec % 86400) / 3600), m = Math.floor((sec % 3600) / 60), s = sec % 60;
  if (d) return `${d}d ${h}h`;
  if (h) return `${h}h ${String(m).padStart(2, "0")}m`;
  if (m) return `${m}m ${String(s).padStart(2, "0")}s`;
  return `${s}s`;
}

/** Resource amount with icon; red when missing. */
export function Res({ id, n, have, signed }: { id: string; n: number; have?: number; signed?: boolean }) {
  const { idx } = useGame();
  const short = have !== undefined && have < n;
  return (
    <span className={"res" + (short ? " short" : "")} title={idx.name(id)}>
      <Icon name={idx.icon(id) || "star"} size={15} />
      {signed && n > 0 ? "+" : ""}
      {fmt(n)}
    </span>
  );
}

export function CostList({ cost, have }: { cost: Amounts; have?: Amounts }) {
  const entries = Object.entries(cost).filter(([, n]) => n > 0);
  if (!entries.length) return <span className="muted">—</span>;
  return (
    <span className="cost">
      {entries.map(([r, n]) => (
        <Res key={r} id={r} n={n} have={have ? have[r] ?? 0 : undefined} />
      ))}
    </span>
  );
}

export function Reqs({ reqs }: { reqs: Requirement[] }) {
  const { idx, t } = useGame();
  if (!reqs.length) return null;
  return (
    <div className="reqs">
      {t("requires")}: {reqs.map((r, i) => (
        <span key={i} className="req">{idx.name(r.id)}{r.level ? ` ${r.level}` : ""}</span>
      ))}
    </div>
  );
}

/** Countdown to an ISO instant on server time; calls onDone once at zero. */
export function Timer({ until, onDone }: { until: string; onDone?: () => void }) {
  const target = Date.parse(until);
  const [now, setNow] = useState(api.now());
  const fired = useRef(false);
  useEffect(() => {
    const id = setInterval(() => setNow(api.now()), 1000);
    return () => clearInterval(id);
  }, []);
  const left = (target - now) / 1000;
  useEffect(() => {
    if (left <= 0 && !fired.current) {
      fired.current = true;
      // small grace so the server has certainly passed the due time
      setTimeout(() => onDone?.(), 600);
    }
  }, [left, onDone]);
  return <span className="timer"><Icon name="clock" size={14} /> {duration(left)}</span>;
}

export function Progress({ value, max, label }: { value: number; max: number; label?: string }) {
  const pct = max > 0 ? Math.min(100, (value / max) * 100) : 0;
  return (
    <div className="progress" role="progressbar" aria-valuenow={value} aria-valuemin={0} aria-valuemax={max} aria-label={label}>
      <div style={{ width: pct + "%" }} />
    </div>
  );
}

/**
 * Button that runs an async action. It is disabled while pending, preventing
 * accidental double submission; errors surface as a toast.
 */
export function Action({ run, children, className = "", disabled, confirm, title }: {
  run: () => Promise<unknown>;
  children: ReactNode;
  className?: string;
  disabled?: boolean;
  confirm?: string;
  title?: string;
}) {
  const [busy, setBusy] = useState(false);
  const { toast, locale } = useGame();
  const [asking, setAsking] = useState(false);
  const go = async () => {
    setAsking(false);
    setBusy(true);
    try {
      await run();
    } catch (e) {
      const err = e as ApiError;
      toast(errorMessage(locale, err.code ?? "", err.message ?? String(e)), "error");
    } finally {
      setBusy(false);
    }
  };
  return (
    <>
      <button className={"btn " + className} disabled={disabled || busy} aria-busy={busy} title={title}
        onClick={() => (confirm ? setAsking(true) : go())}>
        {busy ? <span className="spinner" aria-hidden /> : null}
        {children}
      </button>
      {asking && <Confirm text={confirm!} onYes={go} onNo={() => setAsking(false)} danger />}
    </>
  );
}

export function Modal({ title, onClose, children, wide }: { title: string; onClose: () => void; children: ReactNode; wide?: boolean }) {
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const prev = document.activeElement as HTMLElement | null;
    ref.current?.focus();
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && onClose();
    window.addEventListener("keydown", onKey);
    return () => {
      window.removeEventListener("keydown", onKey);
      prev?.focus?.();
    };
  }, [onClose]);
  // Render outside the animated game content stacking context.
  // Otherwise the fixed overlay appears below the resource bar and bottom HUD.
  return createPortal(
    <div className="overlay" onMouseDown={(e) => e.target === e.currentTarget && onClose()}>
      <div className={"modal" + (wide ? " wide" : "")} role="dialog" aria-modal="true" aria-label={title} tabIndex={-1} ref={ref}>
        <header>
          <h2>{title}</h2>
          <button className="icon-btn" onClick={onClose} aria-label="Close"><Icon name="close" /></button>
        </header>
        <div className="modal-body">{children}</div>
      </div>
    </div>,
    document.body,
  );
}

export function Confirm({ text, onYes, onNo, danger }: { text: string; onYes: () => void; onNo: () => void; danger?: boolean }) {
  const { t } = useGame();
  return (
    <Modal title={t("confirm")} onClose={onNo}>
      <p>{text}</p>
      <div className="row end">
        <button className="btn ghost" onClick={onNo}>{t("cancel")}</button>
        <button className={"btn " + (danger ? "danger" : "primary")} onClick={onYes} autoFocus>{t("confirm")}</button>
      </div>
    </Modal>
  );
}

export function Empty({ children }: { children?: ReactNode }) {
  const { t } = useGame();
  return <div className="empty">{children ?? t("empty")}</div>;
}

export function Loading() {
  const { t } = useGame();
  return <div className="loading" aria-live="polite"><span className="spinner" /> {t("appLoading")}</div>;
}

export function ErrorBox({ error, retry }: { error: unknown; retry?: () => void }) {
  const { t, locale } = useGame();
  const e = error as ApiError;
  return (
    <div className="error-box" role="alert">
      {errorMessage(locale, e?.code ?? "", e?.message ?? String(error))}
      {retry && <button className="btn small" onClick={retry}>{t("retry")}</button>}
    </div>
  );
}

export function Tabs<T extends string>({ value, onChange, items }: { value: T; onChange: (v: T) => void; items: [T, string][] }) {
  return (
    <div className="tabs" role="tablist">
      {items.map(([k, label]) => (
        <button key={k} role="tab" aria-selected={value === k} className={value === k ? "on" : ""} onClick={() => onChange(k)}>
          {label}
        </button>
      ))}
    </div>
  );
}

/** Generic loader hook for screen data with refresh. */
export function useLoad<T>(fn: (signal: AbortSignal) => Promise<T>, deps: unknown[]) {
  const [data, setData] = useState<T | undefined>();
  const [error, setError] = useState<unknown>();
  const [tick, setTick] = useState(0);
  useEffect(() => {
    const ac = new AbortController();
    setError(undefined);
    fn(ac.signal).then(setData, (e) => {
      if ((e as Error).name !== "AbortError") setError(e);
    });
    return () => ac.abort();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...deps, tick]);
  return { data, error, reload: () => setTick((x) => x + 1), setData };
}

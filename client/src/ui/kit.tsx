import { useEffect, useRef, useState, type ReactNode } from "react";
import { api, ApiError } from "../api/client";
import { useGame } from "../game/context";
import type { Amounts, Requirement } from "../api/types";
import { errorMessage } from "../i18n/ui";

// ---------- Icons: original glyphs drawn in SVG (no third-party art) ----------
const glyphs: Record<string, ReactNode> = {
  log: <><rect x="3" y="8" width="18" height="8" rx="4" /><circle cx="19" cy="12" r="2.2" fill="var(--surface)" /></>,
  wheat: <><path d="M12 21V8" /><path d="M12 9c-3-1-4-4-4-6 3 1 4 3 4 6zm0 0c3-1 4-4 4-6-3 1-4 3-4 6zm0 5c-3-1-4-4-4-6 3 1 4 3 4 6zm0 0c3-1 4-4 4-6-3 1-4 3-4 6z" /></>,
  stone: <path d="M4 17l3-8 6-3 6 4 1 7-7 3z" />,
  ingot: <path d="M3 16l3-7h12l3 7zM6 9l2-3h8l2 3" />,
  coin: <><circle cx="12" cy="12" r="8" /><path d="M12 8v8M9.5 10h4a1.5 1.5 0 010 3h-3a1.5 1.5 0 000 3h4" /></>,
  scroll: <path d="M7 4h11v14a2 2 0 01-2 2H6a2 2 0 01-2-2v-1h10v1a2 2 0 002 2M7 4a2 2 0 00-2 2v11M10 8h5M10 11h5" />,
  crystal: <path d="M12 2l5 6-5 14-5-14zM7 8h10" />,
  leaf: <path d="M5 19c0-8 6-14 15-14 0 9-6 15-14 15M5 19l8-8" />,
  bolt: <path d="M13 2L5 14h6l-1 8 8-12h-6z" />,
  chip: <><rect x="6" y="6" width="12" height="12" rx="2" /><path d="M9 2v4M15 2v4M9 18v4M15 18v4M2 9h4M2 15h4M18 9h4M18 15h4" /></>,
  sword: <path d="M14 4h6v6L9 21l-3-3zM5 15l4 4M3 21l3-3" />,
  shield: <path d="M12 3l8 3v6c0 5-4 8-8 9-4-1-8-4-8-9V6z" />,
  heart: <path d="M12 20s-7-4.5-7-10a4 4 0 017-2.5A4 4 0 0119 10c0 5.5-7 10-7 10z" />,
  boot: <path d="M8 3h5v9l6 3v4H5v-4l3-2z" />,
  bag: <path d="M6 8h12l1 12H5zM9 8a3 3 0 016 0" />,
  people: <><circle cx="9" cy="8" r="3" /><circle cx="17" cy="9" r="2.5" /><path d="M3 20c0-4 3-6 6-6s6 2 6 6M15 14c3 0 6 2 6 5" /></>,
  castle: <path d="M4 21V9h3V6h2v3h2V6h2v3h2V6h2v3h3v12h-6v-5a2 2 0 00-4 0v5z" />,
  clock: <><circle cx="12" cy="12" r="9" /><path d="M12 7v5l3 2" /></>,
  bell: <path d="M6 16V11a6 6 0 0112 0v5l2 2H4zM10 20a2 2 0 004 0" />,
  map: <path d="M3 6l6-2 6 2 6-2v14l-6 2-6-2-6 2zM9 4v14M15 6v14" />,
  flag: <path d="M5 21V4h11l-2 4 2 4H5" />,
  star: <path d="M12 3l2.6 5.6 6 .7-4.5 4.1 1.2 6L12 16.4 6.7 19.4l1.2-6L3.4 9.3l6-.7z" />,
  book: <path d="M4 4h7a3 3 0 013 3v13a2 2 0 00-2-2H4zM20 4h-6M20 4v14h-6" />,
  mail: <><rect x="3" y="6" width="18" height="12" rx="2" /><path d="M3 7l9 6 9-6" /></>,
  chart: <path d="M4 20V10M10 20V4M16 20v-8M22 20H2" />,
  user: <><circle cx="12" cy="8" r="4" /><path d="M4 21c0-4 4-7 8-7s8 3 8 7" /></>,
  gear: <><circle cx="12" cy="12" r="3" /><path d="M12 2v3M12 19v3M2 12h3M19 12h3M5 5l2 2M17 17l2 2M5 19l2-2M17 7l2-2" /></>,
  scale: <path d="M12 3v18M5 7h14M5 7l-3 7a3 3 0 006 0zM19 7l-3 7a3 3 0 006 0zM8 21h8" />,
  menu: <path d="M4 7h16M4 12h16M4 17h16" />,
  eye: <><path d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7S2 12 2 12z" /><circle cx="12" cy="12" r="3" /></>,
  skull: <><path d="M5 11a7 7 0 0114 0v4h-3v4H8v-4H5z" /><circle cx="9.5" cy="11" r="1.5" /><circle cx="14.5" cy="11" r="1.5" /></>,
  close: <path d="M6 6l12 12M18 6L6 18" />,
};

export function Icon({ name, size = 18, title }: { name: string; size?: number; title?: string }) {
  const g = glyphs[name.replace(/^glyph:/, "")] ?? glyphs.star;
  return (
    <svg className="icon" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.7}
      strokeLinecap="round" strokeLinejoin="round" role={title ? "img" : undefined} aria-hidden={title ? undefined : true}>
      {title && <title>{title}</title>}
      {g}
    </svg>
  );
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
  return (
    <div className="overlay" onMouseDown={(e) => e.target === e.currentTarget && onClose()}>
      <div className={"modal" + (wide ? " wide" : "")} role="dialog" aria-modal="true" aria-label={title} tabIndex={-1} ref={ref}>
        <header>
          <h2>{title}</h2>
          <button className="icon-btn" onClick={onClose} aria-label="Close"><Icon name="close" /></button>
        </header>
        <div className="modal-body">{children}</div>
      </div>
    </div>
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

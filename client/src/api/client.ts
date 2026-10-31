// HTTP client for the Realmforge API.
//
// - The server is authoritative: the client only renders state and sends commands.
// - Every command carries an Idempotency-Key that is generated once per user
//   action and reused for automatic retries after network failures, so a
//   retried request can never apply twice.
// - The server clock offset is tracked from each response so timers count down
//   on server time, not on the (possibly wrong) device clock.

export class ApiError extends Error {
  status: number;
  code: string;
  details?: unknown;
  constructor(status: number, code: string, message: string, details?: unknown) {
    super(message);
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

const TOKEN_KEY = "rf.token";

function safeGet(key: string): string | null {
  try {
    return localStorage.getItem(key);
  } catch {
    return null;
  }
}
function safeSet(key: string, value: string | null) {
  try {
    if (value === null) localStorage.removeItem(key);
    else localStorage.setItem(key, value);
  } catch {
    /* storage unavailable (private mode): session lasts until reload */
  }
}

let token: string | null = safeGet(TOKEN_KEY);
let serverOffsetMs = 0;
let baseUrl = (import.meta.env.VITE_API_BASE as string | undefined) ?? "";
const listeners = new Set<() => void>();

export const api = {
  setBase(url: string) {
    baseUrl = url.replace(/\/$/, "");
  },
  get base() {
    return baseUrl;
  },
  get token() {
    return token;
  },
  setToken(t: string | null) {
    token = t;
    safeSet(TOKEN_KEY, t);
    listeners.forEach((l) => l());
  },
  onAuthChange(fn: () => void): () => void {
    listeners.add(fn);
    return () => {
      listeners.delete(fn);
    };
  },
  /** Current server time estimate in ms. */
  now(): number {
    return Date.now() + serverOffsetMs;
  },
  get offset() {
    return serverOffsetMs;
  },
};

export function newIdempotencyKey(): string {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
}

function trackServerTime(res: Response, sentAt: number) {
  const d = res.headers.get("Date");
  if (!d) return;
  const server = Date.parse(d);
  if (Number.isNaN(server)) return;
  const rtt = Date.now() - sentAt;
  // HTTP Date has 1 s resolution; smooth to avoid jitter.
  const sample = server + rtt / 2 - Date.now();
  serverOffsetMs = Math.abs(serverOffsetMs) < 1 ? sample : serverOffsetMs * 0.7 + sample * 0.3;
}

/** Precise offset from an explicit serverTime field in a payload. */
export function syncServerTime(iso: string | undefined) {
  if (!iso) return;
  const t = Date.parse(iso);
  if (!Number.isNaN(t)) serverOffsetMs = t - Date.now();
}

type Opts = { method?: string; body?: unknown; idempotencyKey?: string; retries?: number; signal?: AbortSignal };

export async function request<T>(path: string, opts: Opts = {}): Promise<T> {
  const method = opts.method ?? "GET";
  const headers: Record<string, string> = { Accept: "application/json" };
  if (opts.body !== undefined) headers["Content-Type"] = "application/json";
  if (token) headers.Authorization = `Bearer ${token}`;
  if (opts.idempotencyKey) headers["Idempotency-Key"] = opts.idempotencyKey;
  const retries = opts.retries ?? (method === "GET" || opts.idempotencyKey ? 2 : 0);
  let lastErr: unknown;
  for (let attempt = 0; attempt <= retries; attempt++) {
    if (attempt > 0) await new Promise((r) => setTimeout(r, 300 * attempt * attempt));
    const sentAt = Date.now();
    let res: Response;
    try {
      res = await fetch(baseUrl + path, {
        method,
        headers,
        body: opts.body === undefined ? undefined : JSON.stringify(opts.body),
        signal: opts.signal,
      });
    } catch (e) {
      if ((e as Error).name === "AbortError") throw e;
      lastErr = new ApiError(0, "network", "Network error — check your connection");
      continue; // safe: same Idempotency-Key on retry
    }
    trackServerTime(res, sentAt);
    if (res.status === 304) return undefined as T;
    let payload: any = null;
    const text = await res.text();
    if (text) {
      try {
        payload = JSON.parse(text);
      } catch {
        payload = null;
      }
    }
    if (res.ok) return (payload?.data ?? payload) as T;
    const e = payload?.error ?? {};
    const err = new ApiError(res.status, e.code ?? "http_" + res.status, e.message ?? res.statusText, e.details);
    if (res.status === 401) api.setToken(null);
    if (res.status >= 500 || res.status === 429) {
      lastErr = err;
      continue;
    }
    throw err;
  }
  throw lastErr;
}

export const get = <T>(path: string, signal?: AbortSignal) => request<T>(path, { signal });

/** Sends a game command with a fresh idempotency key (reused on retries). */
export const command = <T>(path: string, body: unknown = {}, key = newIdempotencyKey()) =>
  request<T>(path, { method: "POST", body, idempotencyKey: key });

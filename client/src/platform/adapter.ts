// Platform adapters isolate store/catalogue SDKs from the game client.
// The game never calls a platform SDK directly; it talks to this interface.
// See docs/spec/PLATFORM_REQUIREMENTS.md for the requirements each adapter serves.

export interface PlatformIdentity {
  platform: string;
  externalId: string;
  signature?: string; // signed payload the server must verify (not trusted client-side)
}

export interface PlatformAdapter {
  readonly name: "web" | "yandex" | "android";
  init(): Promise<void>;
  /** Signals the platform that the game finished loading (Yandex LoadingAPI.ready). */
  signalReady(): void;
  /** Gameplay markers (Yandex GameplayAPI.start/stop) — optional. */
  gameplayStart(): void;
  gameplayStop(): void;
  /** Platform locale hint (e.g. Yandex i18n.lang). */
  locale(): string | undefined;
  /** Platform identity for sign-in, when available. */
  identity(): Promise<PlatformIdentity | null>;
  /** Pause/resume hooks (focus loss, ads, app backgrounding). */
  onPause(cb: () => void): void;
  onResume(cb: () => void): void;
  /** Monetisation is disabled in the MVP; adapters report capability only. */
  readonly supportsAds: boolean;
  readonly supportsPayments: boolean;
}

abstract class Base implements PlatformAdapter {
  abstract readonly name: "web" | "yandex" | "android";
  protected pauses: (() => void)[] = [];
  protected resumes: (() => void)[] = [];
  readonly supportsAds: boolean = false;
  readonly supportsPayments: boolean = false;
  async init() {
    document.addEventListener("visibilitychange", () => {
      (document.hidden ? this.pauses : this.resumes).forEach((f) => f());
    });
  }
  signalReady() {}
  gameplayStart() {}
  gameplayStop() {}
  locale(): string | undefined {
    return navigator.language?.slice(0, 2);
  }
  async identity(): Promise<PlatformIdentity | null> {
    return null;
  }
  onPause(cb: () => void) {
    this.pauses.push(cb);
  }
  onResume(cb: () => void) {
    this.resumes.push(cb);
  }
}

export class WebAdapter extends Base {
  readonly name = "web" as const;
}

/**
 * Yandex Games adapter. Loads the SDK from the platform-relative path
 * (/sdk.js), required inside the Yandex iframe. Sign-in through Yandex needs
 * server-side signature verification, which is NOT implemented yet
 * (server returns 501 for platform "yandex"); the adapter therefore falls back
 * to guest play, which the platform requires anyway.
 */
export class YandexAdapter extends Base {
  readonly name = "yandex" as const;
  private ysdk: any = null;
  async init() {
    await super.init();
    await new Promise<void>((resolve) => {
      const s = document.createElement("script");
      s.src = "/sdk.js";
      s.async = true;
      s.onload = () => resolve();
      s.onerror = () => resolve(); // outside Yandex: continue as plain web
      document.head.appendChild(s);
    });
    const w = window as any;
    if (w.YaGames) {
      this.ysdk = await w.YaGames.init();
      this.ysdk.on?.("game_api_pause", () => this.pauses.forEach((f) => f()));
      this.ysdk.on?.("game_api_resume", () => this.resumes.forEach((f) => f()));
    }
  }
  signalReady() {
    this.ysdk?.features?.LoadingAPI?.ready?.();
  }
  gameplayStart() {
    this.ysdk?.features?.GameplayAPI?.start?.();
  }
  gameplayStop() {
    this.ysdk?.features?.GameplayAPI?.stop?.();
  }
  locale() {
    return this.ysdk?.environment?.i18n?.lang ?? super.locale();
  }
}

/** Android (Capacitor) adapter: the WebView reports pause/resume via visibility. */
export class AndroidAdapter extends Base {
  readonly name = "android" as const;
}

export function detectPlatform(): PlatformAdapter {
  const mode = import.meta.env.MODE;
  if (mode === "yandex") return new YandexAdapter();
  if ((window as any).Capacitor?.isNativePlatform?.()) return new AndroidAdapter();
  return new WebAdapter();
}

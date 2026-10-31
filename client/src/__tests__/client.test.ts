import { afterEach, describe, expect, it, vi } from "vitest";
import { api, command, request, ApiError, newIdempotencyKey } from "../api/client";
import { PackIndex, type Pack } from "../game/pack";
import { duration, fmt } from "../ui/kit";
import { errorMessage, makeT } from "../i18n/ui";

const ok = (data: unknown) => new Response(JSON.stringify({ data }), { status: 200, headers: { Date: new Date().toUTCString() } });

afterEach(() => vi.restoreAllMocks());

describe("api client", () => {
  it("reuses the same Idempotency-Key when retrying after a network failure", async () => {
    const keys: string[] = [];
    let calls = 0;
    vi.stubGlobal("fetch", vi.fn(async (_url: string, init: RequestInit) => {
      keys.push((init.headers as Record<string, string>)["Idempotency-Key"]);
      if (++calls === 1) throw new TypeError("network down");
      return ok({ id: 7 });
    }));
    const res = await command<{ id: number }>("/api/v1/cities/1/build", { building: "keep" }, "fixed-key-123");
    expect(res.id).toBe(7);
    expect(keys).toEqual(["fixed-key-123", "fixed-key-123"]);
  });

  it("does not retry domain errors and surfaces the error code", async () => {
    const f = vi.fn(async () => new Response(JSON.stringify({ error: { code: "insufficient_resources", message: "no" } }), { status: 422 }));
    vi.stubGlobal("fetch", f);
    await expect(command("/x", {})).rejects.toMatchObject({ code: "insufficient_resources", status: 422 });
    expect(f).toHaveBeenCalledTimes(1);
  });

  it("clears the session token on 401", async () => {
    api.setToken("tok");
    vi.stubGlobal("fetch", vi.fn(async () => new Response(JSON.stringify({ error: { code: "unauthenticated", message: "x" } }), { status: 401 })));
    await expect(request("/api/v1/me")).rejects.toBeInstanceOf(ApiError);
    expect(api.token).toBeNull();
  });

  it("generates unique, header-safe idempotency keys", () => {
    const a = newIdempotencyKey(), b = newIdempotencyKey();
    expect(a).not.toBe(b);
    expect(a).toMatch(/^[A-Za-z0-9_-]{8,80}$/);
  });
});

describe("pack index", () => {
  const pack = {
    id: "p", version: "1.0.0", locales: ["en", "ru"], defaultLocale: "en",
    meta: { title: { en: "Title", ru: "Заголовок" }, description: { en: "d" } },
    theme: { palette: {} }, params: {}, resources: [{ id: "gold", name: { en: "Gold" }, kind: "rare", capped: false, tradable: true, lootable: true }],
    races: [], buildings: [], techs: [], units: [], heroes: { classes: [], skills: [], items: [], slots: [], maxLevel: 1 },
    terrains: [], neutrals: [], quests: [], ui: { screens: ["city"] },
  } as unknown as Pack;
  it("localises with fallback to the default locale", () => {
    const ru = new PackIndex(pack, "ru");
    expect(ru.title).toBe("Заголовок");
    expect(ru.name("gold")).toBe("Gold");
    expect(new PackIndex(pack, "de").locale).toBe("en");
    expect(ru.name("unknown")).toBe("unknown");
  });
});

describe("formatting and i18n", () => {
  it("formats numbers and durations", () => {
    expect(fmt(12345)).toBe("12.3k");
    expect(fmt(2_500_000)).toBe("2.5M");
    expect(duration(59)).toBe("59s");
    expect(duration(3725)).toBe("1h 02m");
    expect(duration(90000)).toBe("1d 1h");
  });
  it("translates UI strings and error codes", () => {
    expect(makeT("ru")("city")).toBe("Город");
    expect(makeT("xx")("city")).toBe("City");
    expect(errorMessage("ru", "queue_full", "fallback")).toBe("Очередь заполнена");
    expect(errorMessage("en", "weird", "fallback")).toBe("fallback");
  });
});

// Client view of a Game Pack. The client renders names, descriptions, theme
// and icons from the pack, so a different pack yields a different game with
// no client code changes.
import type { Amounts, Loc, Requirement } from "../api/types";
import { api } from "../api/client";

export interface Modifier {
  target: string;
  op: "pct" | "flat";
  value: number;
}
export interface Effect {
  type: string;
  resource?: string;
  perLevel?: number;
  growth?: number;
  workers?: number;
  modifier?: Modifier;
}
interface Named {
  id: string;
  name: Loc;
  desc?: Loc;
}
export interface PackResource extends Named {
  kind: string;
  capped: boolean;
  tradable: boolean;
  lootable: boolean;
}
export interface PackRace extends Named {
  modifiers: Modifier[];
  color?: string;
  faction?: string;
}
export interface PackBuilding extends Named {
  category: string;
  maxLevel: number;
  effects: Effect[];
  requires?: Requirement[];
}
export interface PackTech extends Named {
  branch: string;
  maxLevel: number;
  effects: Effect[];
  requires?: Requirement[];
  exclusiveGroup?: string;
}
export interface PackUnit extends Named {
  class: string;
  attack: number;
  defense: number;
  hp: number;
  speed: number;
  carry: number;
  upkeep: number;
  scout?: number;
  cost?: Amounts;
  trainSeconds: number;
  trainedIn?: string;
  requires?: Requirement[];
  bonuses?: { vsClass: string; pct: number }[];
  races?: string[];
}
export interface PackNeutral extends Named {
  kind: string;
  levels: { level: number; guards: Record<string, number>; rewards: Amounts; xp: number }[];
  respawnSeconds: number;
  capturable: boolean;
}
export interface PackQuest extends Named {
  hint?: Loc;
  goal: { type: string; id?: string; amount: number };
}
export interface Pack {
  id: string;
  version: string;
  meta: { title: Loc; subtitle?: Loc; description: Loc };
  locales: string[];
  defaultLocale: string;
  theme: { palette: Record<string, string>; fonts?: Record<string, string>; icons?: Record<string, string>; art?: Record<string, string> };
  params: Record<string, any>;
  resources: PackResource[];
  races: PackRace[];
  buildings: PackBuilding[];
  techs: PackTech[];
  units: PackUnit[];
  heroes: {
    classes: (Named & { modifiers: Modifier[] })[];
    skills: (Named & { maxLevel: number; classes?: string[]; modifiers: Modifier[] })[];
    items: (Named & { slot: string; rarity: string; modifiers: Modifier[] })[];
    slots: string[];
    maxLevel: number;
  };
  terrains: (Named & { color: string; passable: boolean; resource?: string })[];
  neutrals: PackNeutral[];
  quests: PackQuest[];
  ui: { screens: string[]; cityLayout?: Record<string, [number, number]> };
  strings?: Record<string, Loc>;
}

export class PackIndex {
  readonly pack: Pack;
  readonly locale: string;
  private byId = new Map<string, Named & Record<string, any>>();

  constructor(pack: Pack, locale: string) {
    this.pack = pack;
    this.locale = pack.locales.includes(locale) ? locale : pack.defaultLocale;
    const all: (Named & Record<string, any>)[] = [
      ...pack.resources,
      ...pack.races,
      ...pack.buildings,
      ...pack.techs,
      ...pack.units,
      ...pack.heroes.classes,
      ...pack.heroes.skills,
      ...pack.heroes.items,
      ...pack.terrains,
      ...pack.neutrals,
      ...pack.quests,
    ];
    for (const e of all) this.byId.set(e.id, e);
  }

  text(t: Loc | undefined): string {
    if (!t) return "";
    return t[this.locale] || t[this.pack.defaultLocale] || Object.values(t)[0] || "";
  }
  name(id: string): string {
    const e = this.byId.get(id);
    return e ? this.text(e.name) : id;
  }
  desc(id: string): string {
    const e = this.byId.get(id);
    return e ? this.text(e.desc) : "";
  }
  get<T = any>(id: string): T | undefined {
    return this.byId.get(id) as T | undefined;
  }
  unit(id: string) {
    return this.pack.units.find((u) => u.id === id);
  }
  building(id: string) {
    return this.pack.buildings.find((b) => b.id === id);
  }
  resource(id: string) {
    return this.pack.resources.find((r) => r.id === id);
  }
  terrain(id: string) {
    return this.pack.terrains.find((t) => t.id === id);
  }
  neutral(id: string) {
    return this.pack.neutrals.find((n) => n.id === id);
  }
  /** URL of a file in the pack's assets directory (served by the game server). */
  assetUrl(rel: string): string {
    return `${api.base}/packs/${this.pack.id}/assets/${rel}`;
  }
  /** Small icon: a resolved image URL, a "glyph:<name>" for built-in SVG, or "". */
  icon(id: string): string {
    const v = this.pack.theme.icons?.[id] ?? "";
    return v && !v.startsWith("glyph:") ? this.assetUrl(v) : v;
  }
  /** Large illustration URL (building sprite, portrait, terrain texture) or "". */
  art(id: string): string {
    const v = this.pack.theme.art?.[id];
    return v ? this.assetUrl(v) : "";
  }
  string(key: string): string {
    return this.text(this.pack.strings?.[key]);
  }
  get title() {
    return this.text(this.pack.meta.title);
  }
}

/** Applies a pack palette as CSS custom properties (the pack owns the look). */
export function applyTheme(pack: Pack) {
  const root = document.documentElement;
  for (const [k, v] of Object.entries(pack.theme.palette)) root.style.setProperty(`--${k}`, v);
  const fonts = pack.theme.fonts ?? {};
  if (fonts.display) root.style.setProperty("--font-display", `"${fonts.display}", Georgia, serif`);
  if (fonts.body) root.style.setProperty("--font-body", `"${fonts.body}", system-ui, sans-serif`);
  const families = [fonts.display, fonts.body].filter(Boolean) as string[];
  if (families.length && !document.getElementById("pack-fonts")) {
    const link = document.createElement("link");
    link.id = "pack-fonts";
    link.rel = "stylesheet";
    link.href = `https://fonts.googleapis.com/css2?${families.map((f) => `family=${encodeURIComponent(f)}:wght@400;600;700`).join("&")}&display=swap`;
    document.head.appendChild(link);
  }
}

/** Formats a modifier like "+10% attack (cavalry)". */
export function describeModifier(m: Modifier, idx: PackIndex, per?: string): string {
  const [head, tail] = m.target.split(".");
  const sign = m.value > 0 ? "+" : "";
  const val = m.op === "pct" ? `${sign}${round(m.value)}%` : `${sign}${round(m.value)}`;
  const what = tail ? `${head} (${tail === "*" ? "all" : idx.name(tail)})` : head.replace(/_/g, " ");
  return `${val} ${what}${per ? " " + per : ""}`;
}

export function round(n: number, d = 1): number {
  const f = 10 ** d;
  return Math.round(n * f) / f;
}

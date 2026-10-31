// Wire types mirroring the server's JSON (see server/internal/game).

export type Amounts = Record<string, number>;
export type Loc = Record<string, string>;

export interface Account {
  id: number;
  login: string | null;
  displayName: string;
  isGuest: boolean;
  isAdmin: boolean;
  locale: string;
}
export interface Session {
  token: string;
  expiresAt: string;
  account: Account;
}
export interface World {
  id: number;
  slug: string;
  name: string;
  packId: string;
  packVersion: string;
  speed: number;
  status: string;
  players: number;
  joined: boolean;
  playerId?: number;
}
export interface Player {
  id: number;
  worldId: number;
  name: string;
  race: string;
  techs: Record<string, number>;
  stats: Record<string, number>;
  score: number;
  protectedUntil: string | null;
  allianceId: number | null;
  allianceRole: string | null;
}
export interface CitySummary {
  id: number;
  name: string;
  x: number;
  y: number;
  isCapital: boolean;
  resources: Amounts;
}
export interface Overview {
  player: Player;
  cities: CitySummary[];
  unreadNotifications: number;
  unreadReports: number;
  questsReady: number;
  serverTime: string;
  allianceTag: string | null;
  cityCap: number;
  territoryCap: number;
  territories: number;
  packVersion: string;
}
export interface Requirement {
  type: "building" | "tech" | "race";
  id: string;
  level?: number;
}
export interface NextStep {
  level: number;
  cost: Amounts;
  seconds: number;
  unmet: Requirement[];
  affordable: boolean;
  blocked?: string;
}
export interface BuildingView {
  id: string;
  level: number;
  targetLevel: number;
  maxLevel: number;
  next?: NextStep;
}
export type TechView = BuildingView;
export interface UnitView {
  id: string;
  count: number;
  trainable: boolean;
  unmet: Requirement[];
  secondsEach: number;
  blocked?: string;
}
export interface Job {
  id: number;
  kind: "build" | "research" | "train";
  item: string;
  level: number;
  quantity: number;
  cost: Amounts;
  startsAt: string;
  finishesAt: string;
  cityId: number;
}
export interface Rates {
  gross: Record<string, number>;
  consumption: Record<string, number>;
  net: Record<string, number>;
  storage: Record<string, number>;
  protected: Record<string, number>;
  housing: number;
  workersNeeded: number;
  employment: number;
  idle: number;
  upkeep: number;
}
export interface StationedGroup {
  missionId: number;
  playerId: number;
  player: string;
  units: Record<string, number>;
}
export interface City {
  id: number;
  playerId: number;
  name: string;
  x: number;
  y: number;
  isCapital: boolean;
  population: number;
  buildings: Record<string, number>;
  units: Record<string, number>;
}
export interface CityView {
  city: City;
  resources: Amounts;
  rates: Rates;
  serverTime: string;
  buildings: BuildingView[];
  techs: TechView[];
  units: UnitView[];
  queues: { build: Job[]; research: Job[]; train: Job[] };
  unitsAway: Record<string, number>;
  foreignTroops: StationedGroup[];
  modifiers: Record<string, number>;
}
export interface Mission {
  id: number;
  playerId: number;
  fromCityId: number;
  kind: string;
  targetX: number;
  targetY: number;
  targetCityId: number | null;
  units: Record<string, number> | null;
  heroId: number | null;
  cargo: Amounts | null;
  departAt: string;
  arriveAt: string;
  returnAt: string | null;
  status: "outbound" | "stationed" | "returning" | "done";
  result: Record<string, any> | null;
  direction: "own" | "incoming" | "stationed_here";
  owner: string;
}
export interface MapTile {
  x: number;
  y: number;
  terrain: string;
  neutral?: string;
  neutralLevel?: number;
  cleared?: boolean;
  ownerId?: number;
  owner?: string;
  alliance?: string;
  cityId?: number;
  cityName?: string;
  race?: string;
}
export interface Report {
  id: number;
  kind: string;
  side: string;
  occurredAt: string;
  readAt: string | null;
  data?: Record<string, any>;
}
export interface Notification {
  id: number;
  kind: string;
  data: Record<string, any>;
  createdAt: string;
  readAt: string | null;
}
export interface Quest {
  id: string;
  status: "locked" | "active" | "complete" | "claimed";
  progress: number;
  goal: number;
  rewards: Amounts;
  tutorial: boolean;
  screen?: string;
}
export interface Hero {
  id: number;
  name: string;
  class: string;
  level: number;
  xp: number;
  skills: Record<string, number>;
  status: string;
  woundedUntil: string | null;
  items: Item[];
  skillPoints: number;
  nextLevelXp: number;
  modifiers: Record<string, number>;
  homeCityId: number;
}
export interface Item {
  id: number;
  def: string;
  heroId: number | null;
  slot: string | null;
  source: string;
}
export interface HeroesView {
  heroes: Hero[];
  inventory: Item[];
  maxHeroes: number;
  hireCost: Amounts;
}
export interface Order {
  id: number;
  playerId: number;
  player: string;
  cityId: number;
  sellRes: string;
  sellAmount: number;
  buyRes: string;
  buyAmount: number;
  status: string;
  createdAt: string;
}
export interface Alliance {
  id: number;
  name: string;
  tag: string;
  description: string;
  memberCount: number;
  score: number;
  members?: { playerId: number; name: string; role: string; score: number }[];
  relations?: { otherId: number; otherTag: string; status: string }[];
}
export interface Invite {
  id: number;
  allianceId: number;
  allianceTag: string;
  playerId: number;
  player: string;
  createdAt: string;
}
export interface Message {
  id: number;
  from: string;
  fromId: number;
  to: string | null;
  allianceId: number | null;
  subject: string;
  body: string;
  createdAt: string;
  readAt: string | null;
}
export interface Territory {
  x: number;
  y: number;
  terrain: string;
  resource: string;
  pending: Amounts;
  since: string | null;
  garrison: Record<string, number>;
}
export interface RankRow {
  rank: number;
  playerId: number;
  name: string;
  race: string;
  score: number;
  alliance: string | null;
  cities: number;
}

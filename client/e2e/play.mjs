// Browser E2E: plays a real session against a running server + built client.
// Usage: node e2e/play.mjs http://127.0.0.1:8080 <screenshotDir>
// Uses the pre-installed Playwright (require('playwright')).
import { createRequire } from "node:module";
const require = createRequire(import.meta.url);
const { chromium } = require("playwright");

const base = process.argv[2] ?? "http://127.0.0.1:8080";
const out = process.argv[3] ?? ".";
const log = (...a) => console.log("[e2e]", ...a);
const fail = (m) => { console.error("[e2e] FAIL:", m); process.exit(1); };

const browser = await chromium.launch();
const ctx = await browser.newContext({ viewport: { width: 1360, height: 860 } });
const page = await ctx.newPage();
const errors = [];
page.on("pageerror", (e) => errors.push(String(e)));
page.on("console", (m) => m.type() === "error" && errors.push(m.text()));

await page.goto(base);
await page.getByRole("button", { name: "Play as guest" }).click();
await page.getByText("Ember Vale").waitFor();
log("guest session created, world list visible");
await page.screenshot({ path: `${out}/01-worlds.png` });
await page.locator("li", { hasText: "Ember Vale" }).getByRole("button").click();
await page.getByLabel("Ruler name").fill("Cinder " + String(Date.now()).slice(-6));
await page.locator("label.race", { hasText: "Ashborn" }).click();
await page.screenshot({ path: `${out}/02-join.png` });
await page.getByRole("button", { name: "Join" }).click();
await page.locator(".city-game-building", { hasText: "Lumberyard" }).waitFor();
log("joined world; city screen rendered");
await page.waitForLoadState("networkidle");
await page.screenshot({ path: `${out}/03-city.png` });

await page.locator(".city-game-building", { hasText: "Lumberyard" }).click();
await page.getByRole("dialog").getByRole("button", { name: /Upgrade/ }).click();
await page.locator(".toast.ok").waitFor();
log("upgrade queued");
await page.keyboard.press("Escape");
// world speed x50: level 2 takes ~2 s
await page.waitForTimeout(3500);
await page.reload();
await page.locator(".city-game-building", { hasText: "Lumberyard" }).waitFor();
const lumber = await page.locator(".city-game-building", { hasText: "Lumberyard" }).innerText();
if (!/Level 2/.test(lumber)) fail("after reload lumberyard should be level 2, got: " + lumber);
log("after reload: lumberyard level 2 persisted");

for (const [name, file] of [["Economy", "04-economy"], ["Research", "05-research"], ["Army", "06-army"], ["Map", "07-map"], ["Quests", "08-quests"], ["Market", "09-market"], ["Alliance", "10-alliance"], ["Heroes", "11-heroes"], ["Mail", "12-mail"], ["Profile", "13-profile"], ["Missions", "17-missions"], ["Reports", "18-reports"]]) {
  await page.locator("nav.sidenav").getByRole("button", { name }).click();
  await page.locator("main.content h1").first().waitFor();
  await page.waitForLoadState("networkidle");
  await page.screenshot({ path: `${out}/${file}.png`, fullPage: true });
  const overflow = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
  if (overflow > 2) fail(`${name} horizontal overflow on desktop: ${overflow}px`);
}
// claim the first tutorial quest through the UI
await page.locator("nav.sidenav").getByRole("button", { name: "Quests" }).click();
await page.locator(".list.quests, .empty").first().waitFor();
const claim = page.getByRole("button", { name: "Claim reward" }).first();
if (await claim.count()) {
  await claim.click();
  await page.locator(".toast.ok").waitFor();
  log("tutorial quest claimed");
}
// map: select own city tile area and open a scout dialog on a neutral if visible
await page.locator("nav.sidenav").getByRole("button", { name: "Map" }).click();
const canvas = page.locator("canvas.map");
await canvas.waitFor();
const box = await canvas.boundingBox();
await page.mouse.click(box.x + box.width / 2 + 3, box.y + box.height / 2 + 3);
await page.waitForTimeout(300);
await page.screenshot({ path: `${out}/14-map-selected.png` });

// Render all formerly inconsistent art assets at portrait and UI-icon sizes.
const artPages = [
  ...["griffin", "ballista", "wagon", "settler", "wolf", "brigand", "troll", "wraith", "golem", "drake"].map((id) => ({id, dir:"units"})),
  ...["wolf_den", "brigand_camp", "barrow", "crater"].map((id) => ({id, dir:"neutrals"})),
];
const artHtml = `<!doctype html><html><head><meta charset="utf-8"><style>
body{margin:0;padding:24px;background:#302820;color:#e9d7b6;font:14px Georgia,serif;}
h1{margin:0 0 17px;font-size:23px;color:#e2ca9f;}
section{display:grid;grid-template-columns:repeat(7,1fr);gap:12px;}
article{border:1px solid #8c7655;border-radius:6px;padding:8px;background:linear-gradient(#5e4b35,#2c241d);text-align:center;}
img{width:120px;height:120px;max-width:100%;object-fit:contain;background:radial-gradient(ellipse,#b5a07a55,transparent 70%);}
small{display:block;overflow-wrap:anywhere;font-weight:600;margin:4px 0 6px;}
.mini{width:48px;height:48px;}
</style></head><body><h1>Ashen Crowns · art consistency audit</h1><section>${artPages.map((a) =>
  `<article><small>${a.id}</small><img src="${base}/packs/ashen-crowns/assets/${a.dir}/${a.id}.svg" alt="${a.id}"><div><img class="mini" src="${base}/packs/ashen-crowns/assets/${a.dir}/${a.id}.svg" alt=""></div></article>`
).join("")}</section></body></html>`;
const artPage = await ctx.newPage();
await artPage.setViewportSize({ width: 1320, height: 680 });
await artPage.goto(base);
await artPage.setContent(artHtml);
await artPage.waitForFunction(() =>
  [...document.querySelectorAll("article img")].every((img) => img.complete && img.naturalWidth > 0),
  { timeout: 15000 }
);
await artPage.screenshot({path: `${out}/20-art-consistency.png`, fullPage: true});
await artPage.close();

// mobile layout
const m = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, isMobile: true, hasTouch: true,
  storageState: await ctx.storageState() });
const mp = await m.newPage();
await mp.goto(base);
await mp.locator("main.content h1").first().waitFor(); // restores the last screen (map)
await mp.screenshot({ path: `${out}/15a-mobile-map.png` });
await mp.locator("nav.bottomnav").getByRole("button", { name: "City" }).click();
await mp.locator(".city-game-building").first().waitFor();
await mp.waitForLoadState("networkidle");
await mp.screenshot({ path: `${out}/15-mobile-city.png` });
await mp.locator("nav.bottomnav").getByRole("button", { name: "More" }).click();
await mp.screenshot({ path: `${out}/16-mobile-more.png` });
await mp.locator(".sheet-body").getByRole("button", { name: "Research" }).click();
await mp.locator("main.content h1").first().waitFor();
await mp.screenshot({ path: `${out}/19-mobile-secondary-screen.png`, fullPage: true });
const overflow = await mp.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
if (overflow > 1) fail("horizontal overflow on mobile: " + overflow + "px");
log("mobile layout has no horizontal overflow");

if (errors.length) fail("browser errors: " + errors.join(" | "));
log("OK");
await browser.close();

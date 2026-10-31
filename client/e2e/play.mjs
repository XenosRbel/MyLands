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

// Ensure original painted UI atlas assets survived the production Docker build.
// The approved frame requires both a scenic city and an illustrated world-globe shortcut.
const globeAsset = await fetch(`${base}/ui/world-globe.webp`);
if (!globeAsset.ok || !globeAsset.headers.get("content-type")?.includes("image/")) {
  fail("The game's original globe command artwork is missing.");
}
// The city backdrop is an independently illustrated landscape, not repeated tile art.
const scenery = await fetch(`${base}/ui/city-landscape.svg`);
if (!scenery.ok || !(await scenery.text()).includes('id="water"')) {
  fail("City landscape with water and fortification art is not present in the client build.");
}
for (const asset of ["panel_main.png", "panel_wide.png", "portrait_frame.png", "dialog.png"]) {
  const result = await fetch(`${base}/ui/painted/${asset}`);
  if (!result.ok || !result.headers.get("content-type")?.includes("image/")) {
    fail(`Missing painted game UI asset: ${asset} (${result.status})`);
  }
}
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
await page.locator(".city-game-stage .city-scene").waitFor();
log("joined world; city screen rendered");
await page.waitForLoadState("networkidle");
await page.screenshot({ path: `${out}/03-city.png` });
await page.locator(".game-clock").waitFor();
await page.locator(".game-dock-globe").waitFor();
const cityArtEnabled = await page.locator(".city-game-stage .city-scene").evaluate((element) =>
  getComputedStyle(element).backgroundImage.includes("city-landscape.svg"));
if (!cityArtEnabled) fail("City stage is still displaying a repeated terrain background.");
if (await page.locator(".city-game-rail .game-queue-panel").count() < 2)
  fail("The permanent build and train queues are missing from the city window.");

const cityMetrics = await page.evaluate(() => {
  const stage = document.querySelector(".city-game-stage .city-scene").getBoundingClientRect();
  const shell = document.querySelector(".shell").getBoundingClientRect();
  const dock = document.querySelector(".game-dock").getBoundingClientRect();
  return { sceneWidth: stage.width / window.innerWidth, sceneHeight: stage.height / window.innerHeight,
    shellWidth: shell.width / window.innerWidth, dockBottom: window.innerHeight - dock.bottom };
});
if (cityMetrics.shellWidth < .98 || cityMetrics.sceneWidth < .55 || cityMetrics.sceneHeight < .58 ||
  Math.abs(cityMetrics.dockBottom) > 4) fail("city viewport is not full screen: " + JSON.stringify(cityMetrics));
await page.setViewportSize({ width: 1680, height: 900 });
await page.screenshot({ path: `${out}/23-wide-city-hud.png` });
const wideLayout = await page.evaluate(() => {
  const world = document.querySelector(".city-game-stage").getBoundingClientRect();
  const dock = document.querySelector(".game-dock").getBoundingClientRect();
  const queue = document.querySelector(".city-game-rail");
  const troopCount = document.querySelectorAll(".game-dock-unit").length;
  const menu = document.querySelector(".sidenav").getBoundingClientRect();
  const rail = queue.getBoundingClientRect();
  return { rightGap: window.innerWidth - rail.right, widthRatio: world.width / window.innerWidth,
    dockBottom: window.innerHeight - dock.bottom, queueVisible: !queue.hidden,
    menuWidth: menu.width, troopCount };
});
if (wideLayout.rightGap > 35 || wideLayout.widthRatio < .54 || Math.abs(wideLayout.dockBottom) > 4 ||
    !wideLayout.queueVisible || wideLayout.menuWidth < 170 || wideLayout.troopCount < 4)
  fail("wide reference HUD is incomplete: " + JSON.stringify(wideLayout));
await page.screenshot({ path: `${out}/24-wide-city-queue.png` });
await page.setViewportSize({ width: 1360, height: 860 });
await page.locator(".city-build-toggle").click();
await page.locator(".city-game-buildings.expanded").waitFor();
await page.screenshot({ path: `${out}/22-city-catalogue-drawer.png` });

await page.locator(".city-game-building", { hasText: "Lumberyard" }).click();
await page.getByRole("dialog").locator(".building-catalog-tabs button").first().waitFor();
const dialogAboveHud = await page.evaluate(() => {
  const rect = document.querySelector(".topbar").getBoundingClientRect();
  const topElement = document.elementFromPoint(rect.left + rect.width / 2, rect.top + rect.height / 2);
  return !!document.querySelector(".overlay")?.contains(topElement);
});
if (!dialogAboveHud) fail("game topbar overlays the construction dialog backdrop");
await page.screenshot({ path: `${out}/21-building-catalog.png`, fullPage: true });
const constructionTabs = page.getByRole("dialog").locator(".building-catalog-tabs button");
if (await constructionTabs.count() > 1) {
  await constructionTabs.nth(1).click();
  if (!await page.getByRole("dialog").locator(".building-catalog-choice").count()) fail("building category tab is empty");
  await page.getByRole("dialog").locator(".building-catalog-tabs button").filter({hasText:"economy"}).click();
  await page.getByRole("dialog").locator(".building-catalog-choice").filter({hasText:"Lumberyard"}).click();
}
await page.getByRole("dialog").getByRole("button", { name: /Upgrade/ }).click();
await page.locator(".toast.ok").waitFor();
const firstQueue = page.locator(".city-game-rail .game-queue-entry").first();
await firstQueue.waitFor();
if (!await firstQueue.locator(".game-queue-progress").count()) fail("Queued building is missing visual progress bar.");
await page.keyboard.press("Escape");
await page.screenshot({ path: `${out}/25-active-building-queue.png`, fullPage: true });
log("upgrade queued");
// world speed x50: level 2 takes ~2 s
await page.waitForTimeout(3500);
await page.reload();
await page.locator(".city-build-toggle").click();
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

// Raster technology art QA: inspect all 18 research subjects at three visible levels.
const techIds = [
 "forestry","agriculture","masonry","metallurgy","exchequer","husbandry",
 "tactics","armour","logistics","espionage","beastlore","engineering",
 "arcana","plunder","heroics","path_architect","path_steward","path_warlord"
];
const researchPage = await ctx.newPage();
await researchPage.setViewportSize({width:1380,height:1100});
await researchPage.goto(base);
await researchPage.setContent(`<!doctype html><html><head><meta charset="utf-8"><style>
body{margin:0;padding:20px;background:#29221d;color:#f0dbba;font:13px Georgia,serif;}
h1{font-size:22px;}
main{display:grid;grid-template-columns:repeat(6,1fr);gap:12px;}
article{display:flex;flex-direction:column;align-items:center;border:1px solid #927650;background:linear-gradient(#53422e,#221c17);padding:8px;}
label{margin:4px;color:#ddc8a4;}
.row{display:flex;gap:4px;}
img{width:65px;height:65px;object-fit:contain;}
</style></head><body><h1>Research WebP art — novice / adept / master</h1><main>
${techIds.map(id=>`<article><label>${id}</label><div class="row">${[1,2,3].map(rank=>
 `<img alt="${id} ${rank}" src="${base}/ui/research/${id}-${rank}.webp"/>`).join("")}</div></article>`).join("")}
</main></body></html>`);
await researchPage.waitForFunction(()=>
 [...document.querySelectorAll("article img")].every(img=>img.complete && img.naturalWidth>0),
 {timeout:20000}
);
await researchPage.screenshot({path:`${out}/26-raster-research-art.png`,fullPage:true});
await researchPage.close();

// mobile layout
const m = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, isMobile: true, hasTouch: true,
  storageState: await ctx.storageState() });
const mp = await m.newPage();
await mp.goto(base);
await mp.locator("main.content h1").first().waitFor(); // restores the last screen (map)
await mp.screenshot({ path: `${out}/15a-mobile-map.png` });
await mp.locator("nav.bottomnav").getByRole("button", { name: "City" }).click();
await mp.locator(".city-game-stage .city-scene").waitFor();
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

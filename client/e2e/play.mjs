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
await page.locator(".bcard", { hasText: "Lumberyard" }).waitFor();
log("joined world; city screen rendered");
await page.waitForLoadState("networkidle");
await page.screenshot({ path: `${out}/03-city.png` });

await page.locator(".bcard", { hasText: "Lumberyard" }).click();
await page.getByRole("dialog").getByRole("button", { name: /Upgrade/ }).click();
await page.locator(".toast.ok").waitFor();
log("upgrade queued");
await page.keyboard.press("Escape");
// world speed x50: level 2 takes ~2 s
await page.waitForTimeout(3500);
await page.reload();
await page.locator(".bcard", { hasText: "Lumberyard" }).waitFor();
const lumber = await page.locator(".bcard", { hasText: "Lumberyard" }).innerText();
if (!/Level 2/.test(lumber)) fail("after reload lumberyard should be level 2, got: " + lumber);
log("after reload: lumberyard level 2 persisted");

for (const [name, file] of [["Economy", "04-economy"], ["Research", "05-research"], ["Army", "06-army"], ["Map", "07-map"], ["Quests", "08-quests"], ["Market", "09-market"], ["Alliance", "10-alliance"], ["Heroes", "11-heroes"], ["Mail", "12-mail"], ["Profile", "13-profile"]]) {
  await page.locator("nav.sidenav").getByRole("button", { name }).click();
  await page.locator("main.content h1").first().waitFor();
  await page.waitForLoadState("networkidle");
  await page.screenshot({ path: `${out}/${file}.png` });
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

// mobile layout
const m = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, isMobile: true, hasTouch: true,
  storageState: await ctx.storageState() });
const mp = await m.newPage();
await mp.goto(base);
await mp.locator("main.content h1").first().waitFor(); // restores the last screen (map)
await mp.screenshot({ path: `${out}/15a-mobile-map.png` });
await mp.locator("nav.bottomnav").getByRole("button", { name: "City" }).click();
await mp.locator(".bcard").first().waitFor();
await mp.waitForLoadState("networkidle");
await mp.screenshot({ path: `${out}/15-mobile-city.png` });
await mp.locator("nav.bottomnav").getByRole("button", { name: "More" }).click();
await mp.screenshot({ path: `${out}/16-mobile-more.png` });
const overflow = await mp.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
if (overflow > 1) fail("horizontal overflow on mobile: " + overflow + "px");
log("mobile layout has no horizontal overflow");

if (errors.length) fail("browser errors: " + errors.join(" | "));
log("OK");
await browser.close();

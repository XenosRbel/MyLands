import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";

const root = resolve(import.meta.dirname, "../..");
const ids = ["resources", "buildings", "units", "terrains", "neutrals"];
let failures = 0;
for (const name of ["ashen-crowns"]) {
  const folder = resolve(root, "packs", name);
  const pack = JSON.parse(readFileSync(resolve(folder, "pack.json"), "utf8"));
  const entries = [
    ...pack.resources, ...pack.buildings, ...pack.units,
    ...pack.terrains, ...pack.neutrals,
    ...pack.heroes.classes, ...pack.heroes.items,
    ...pack.races,
  ];
  for (const entry of entries) {
    const id = entry.id;
    const art = pack.theme.art?.[id];
    const icon = pack.theme.icons?.[id];
    if (!art && !icon) {
      console.error(`Missing artwork: ${name}/${id}`);
      failures++;
    }
    // Guard the original art redraw against reverting to generic framed glyphs.
    // This is a structural check, not a substitute for human art approval.
    const illustrated = new Set([
      "griffin", "ballista", "wagon", "settler", "wolf", "brigand",
      "troll", "wraith", "golem", "drake",
      "wolf_den", "brigand_camp", "barrow", "crater",
    ]);
    if (illustrated.has(id) && art?.endsWith(".svg")) {
      const location = resolve(folder, "assets", art);
      if (existsSync(location)) {
        const xml = readFileSync(location, "utf8");
        if (!xml.includes('viewBox="0 0 320 320"') || !xml.includes('filter="url(#') ||
            xml.includes('<rect width="320" height="320"') || xml.length < 1200) {
          console.error(`Flat placeholder art detected: ${name}/${id}`);
          failures++;
        }
      }
    }
    for (const file of [art, icon]) {
      if (!file || file.startsWith("glyph:")) continue;
      if (file.includes("..") || file.startsWith("/")) {
        console.error(`Unsafe asset path: ${name}/${id} -> ${file}`);
        failures++;
      } else if (!existsSync(resolve(folder, "assets", file))) {
        console.error(`Missing asset file: ${name}/${id} -> ${file}`);
        failures++;
      }
    }
  }
  for (const category of ids) {
    console.log(`${name}: checked ${category}`);
  }
}
if (failures) {
  console.error(`Asset validation failed: ${failures} issue(s)`);
  process.exitCode = 1;
} else {
  console.log("Asset integrity check passed.");
}

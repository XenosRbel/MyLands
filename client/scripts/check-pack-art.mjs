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

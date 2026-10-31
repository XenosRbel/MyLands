// Project art policy: no player-visible SVG in new work.
// Existing legacy SVG files are inventoried. Strict mode rejects all of them.
import { readdirSync, readFileSync } from "node:fs";
import { resolve, relative } from "node:path";

const root = resolve(import.meta.dirname, "../..");
const legacy = new Set([
  "client/public/ui/city-landscape.svg",
  "client/public/ui/dark-wood.svg",
  "client/public/ui/medieval-corners.svg",
  "client/public/ui/parchment-grain.svg",
  "client/public/ui/realm-landscape.svg",
  "client/public/ui/world-globe.svg",
  ...["ashborn","dawnpact","duskbound","gravewarden","ironvow","thornkin"].map(x=>`packs/ashen-crowns/assets/emblems/${x}.svg`),
  ...["barrow","brigand_camp","crater","wolf_den"].map(x=>`packs/ashen-crowns/assets/neutrals/${x}.svg`),
  ...["ballista","brigand","drake","golem","griffin","settler","troll","wagon","wolf","wraith"].map(x=>`packs/ashen-crowns/assets/units/${x}.svg`)
]);
const scan = (directory) => {
  const files = [];
  for (const item of readdirSync(directory,{withFileTypes:true})) {
    if (item.name.startsWith(".") || ["node_modules","dist","painted","research","glyphs"].includes(item.name)) continue;
    const path = resolve(directory,item.name);
    if (item.isDirectory()) files.push(...scan(path)); else files.push(path);
  }
  return files;
};
let errors = 0;
const all=[...scan(resolve(root,"client/src")),...scan(resolve(root,"client/public")),...scan(resolve(root,"packs/ashen-crowns/assets"))];
for(const file of all) {
  const path=relative(root,file).replaceAll("\\","/");
  if (path.endsWith(".svg")) {
    if (!legacy.has(path) || process.argv.includes("--strict")) {
      console.error(`Disallowed SVG artwork: ${path}`);
      errors++;
    }
  } else if (/\.(tsx|jsx)$/.test(path)) {
    const source=readFileSync(file,"utf8");
    if (/<svg(?:\s|>)/i.test(source)) {
      console.error(`Inline SVG in game component: ${path}`);
      errors++;
    }
  }
}
const remaining=all.filter(x=>x.endsWith(".svg")).length;
console.log(`Raster policy: ${remaining} inventoried legacy SVG files; ${errors} new or disallowed violations.`);
if (errors) process.exitCode=1;

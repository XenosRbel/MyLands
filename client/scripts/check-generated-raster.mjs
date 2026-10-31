// Verify generated raster artwork against the authoritative Game Pack.
// Run after prebuild sprite generation. Reject missing or corrupt assets.
import { readFileSync, statSync, existsSync } from "node:fs";
import { resolve } from "node:path";

const client = resolve(import.meta.dirname, "..");
const pack = JSON.parse(readFileSync(resolve(client,"../packs/ashen-crowns/pack.json"),"utf8"));
let count=0;
let failures=0;
const hashes=new Set();
for(const tech of pack.techs) {
  for(let level=1;level<=tech.maxLevel;level++) {
    const path=resolve(client,"public/ui/research",`${tech.id}-${level}.webp`);
    if(!existsSync(path)) {
      console.error(`Missing research artwork: ${tech.id} level ${level}`);
      failures++;
      continue;
    }
    const image=readFileSync(path);
    // RIFF + WEBP identifier at byte offset eight.
    if(image.length<300 || image.toString("ascii",0,4)!=="RIFF" ||
       image.toString("ascii",8,12)!=="WEBP") {
      console.error(`Invalid WebP sprite: ${tech.id} level ${level}`);
      failures++;
    }
    const hash=createHash("sha256").update(image).digest("hex");
    if (hashes.has(hash)) {
      console.error(`Duplicate research sprite artwork: ${tech.id} level ${level}`);
      failures++;
    }
    hashes.add(hash);
    count++;
  }
}
for(const name of ["world-globe.webp","city-landscape.webp"]) {
  const path=resolve(client,"public/ui",name);
  if(!existsSync(path) || statSync(path).size<1000) {
    console.error(`Missing painted HUD/world artwork: ${name}`);
    failures++;
  }
}
console.log(`Verified ${count} generated research sprites across ${pack.techs.length} technologies; violations: ${failures}.`);
if(failures) process.exitCode=1;

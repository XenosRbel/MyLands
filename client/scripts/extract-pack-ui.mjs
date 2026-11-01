// Extract original illustrated UI resources from the source-art archive.
// Node standard library only. Files are produced before Vite builds the client.
import { readFileSync, writeFileSync, mkdirSync, existsSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { inflateRawSync } from "node:zlib";

const root = resolve(import.meta.dirname, "../..");
const source = resolve(root, "docs/reference/MyLands_References_and_Ashen_Crowns_Resource_Pack.zip");
const target = resolve(root, "client/public/ui/painted");
const suffix = "/packs/ashen-crowns/assets/ui/";
const allowed = new Set([
  "avatar_ring.png", "banner.png", "button_primary.png", "button_secondary.png",
  "dialog.png", "divider.png", "panel_main.png", "panel_wide.png",
  "portrait_frame.png", "status_bar.png", "tabs.png",
]);

if (!existsSync(source)) throw Error(`Missing original art archive: ${source}`);
const zip = readFileSync(source);
const sig = (pos) => zip.readUInt32LE(pos);
const eocd = (() => {
  for (let i = zip.length - 22; i >= Math.max(0, zip.length - 65557); i--) {
    if (sig(i) === 0x06054b50) return i;
  }
  throw Error("Art archive is missing ZIP end-of-directory.");
})();
let cursor = zip.readUInt32LE(eocd + 16);
const total = zip.readUInt16LE(eocd + 10);
const entries = new Map();
for (let n = 0; n < total; n++) {
  if (sig(cursor) !== 0x02014b50) throw Error("Invalid ZIP central directory.");
  const method = zip.readUInt16LE(cursor + 10);
  const packedSize = zip.readUInt32LE(cursor + 20);
  const fileSize = zip.readUInt32LE(cursor + 24);
  const nameLength = zip.readUInt16LE(cursor + 28);
  const extraLength = zip.readUInt16LE(cursor + 30);
  const commentLength = zip.readUInt16LE(cursor + 32);
  const localOffset = zip.readUInt32LE(cursor + 42);
  const name = zip.toString("utf8", cursor + 46, cursor + 46 + nameLength);
  if (name.includes(suffix)) {
    const filename = name.slice(name.lastIndexOf(suffix) + suffix.length);
    if (allowed.has(filename)) entries.set(filename, { method, packedSize, fileSize, localOffset });
  }
  cursor += 46 + nameLength + extraLength + commentLength;
}
if (entries.size !== allowed.size) throw Error(`Expected ${allowed.size} original UI files; found ${entries.size}`);
mkdirSync(target, { recursive: true });
for (const name of allowed) {
  const entry = entries.get(name);
  const offset = entry.localOffset;
  if (sig(offset) !== 0x04034b50) throw Error(`Invalid ZIP entry: ${name}`);
  const dataStart = offset + 30 + zip.readUInt16LE(offset + 26) + zip.readUInt16LE(offset + 28);
  const packed = zip.subarray(dataStart, dataStart + entry.packedSize);
  const data = entry.method === 8 ? inflateRawSync(packed) : entry.method === 0 ? packed : null;
  if (!data || data.length !== entry.fileSize || data.subarray(0, 8).toString("hex") !== "89504e470d0a1a0a") {
    throw Error(`Invalid original art PNG: ${name}`);
  }
  const output = resolve(target, name);
  mkdirSync(dirname(output), { recursive: true });
  if (!existsSync(output) || !readFileSync(output).equals(data)) writeFileSync(output, data);
}
console.log(`Original UI resources installed: ${entries.size} illustrated panels, buttons and frames`);

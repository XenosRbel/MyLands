// Packages dist/ for Yandex Games: index.html at the archive root, size check.
// Usage: VITE_API_BASE=https://api.example.com npm run build:yandex && node scripts/yandex-zip.mjs
import { execFileSync } from "node:child_process";
import { readdirSync, statSync, existsSync } from "node:fs";
import { join } from "node:path";

const dist = "dist";
if (!existsSync(join(dist, "index.html"))) throw new Error("dist/index.html missing — run npm run build:yandex first");
let total = 0;
const walk = (d) => readdirSync(d).forEach((f) => { const p = join(d, f); const s = statSync(p); s.isDirectory() ? walk(p) : (total += s.size); });
walk(dist);
const LIMIT = 100 * 1024 * 1024; // see PLATFORM_REQUIREMENTS YG-* (verify current limit before submission)
if (total > LIMIT) throw new Error(`build is ${total} bytes, over the limit`);
execFileSync("zip", ["-qr", "../yandex-build.zip", "."], { cwd: dist, stdio: "inherit" });
console.log(`yandex-build.zip created (${(total / 1024).toFixed(0)} KiB uncompressed)`);

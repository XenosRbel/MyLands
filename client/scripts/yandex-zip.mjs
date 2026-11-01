// Packages dist/ for Yandex Games: index.html at the archive root, size check.
// Usage: VITE_API_BASE=https://api.example.com npm run build:yandex && node scripts/yandex-zip.mjs
import { execFileSync } from "node:child_process";
import { readdirSync, statSync, existsSync } from "node:fs";
import { join } from "node:path";
import { checkPackage } from "./check-package.mjs";

const dist = "dist";
if (!existsSync(join(dist, "index.html"))) throw new Error("dist/index.html missing — run npm run build:yandex first");
// Package rules: root index.html, ASCII names, size limit, no external hosts except our API.
const allowHosts = [];
if (process.env.VITE_API_BASE) allowHosts.push(new URL(process.env.VITE_API_BASE).hostname);
const problems = checkPackage(dist, { allowHosts });
if (problems.length) throw new Error(`package check failed:\n${problems.map((p) => `  - ${p}`).join("\n")}`);
let total = 0;
const walk = (d) => readdirSync(d).forEach((f) => { const p = join(d, f); const s = statSync(p); s.isDirectory() ? walk(p) : (total += s.size); });
walk(dist);
execFileSync("zip", ["-qr", "../yandex-build.zip", "."], { cwd: dist, stdio: "inherit" });
console.log(`yandex-build.zip created (${(total / 1024).toFixed(0)} KiB uncompressed)`);

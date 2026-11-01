// Checks a built client directory against the Yandex Games package rules in
// docs/spec/PLATFORM_REQUIREMENTS.md (YG-13 and the pre-submission checklist 5.1):
//   - index.html at the archive root;
//   - at most 100 MB uncompressed (verify the current limit before submission);
//   - file names in ASCII, without spaces;
//   - no external URLs, except hosts declared as ours.
// Usage: node scripts/check-package.mjs [dir] [--allow-host=api.example.com]...
import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import { join, relative } from "node:path";
import { pathToFileURL } from "node:url";

const LIMIT = 100 * 1024 * 1024;
const TEXT = /\.(html|js|mjs|css|json|svg|txt|webmanifest)$/i;
const URL_RE = /https?:\/\/[A-Za-z0-9._~:/?#@!$&'()*+,;=%-]+/g;

// These strings are never fetched: XML namespace identifiers and React's
// error-message links. Any other host is reported.
const NEVER_FETCHED = new Set(["www.w3.org", "react.dev"]);

function walk(dir, out = []) {
  for (const name of readdirSync(dir)) {
    const p = join(dir, name);
    if (statSync(p).isDirectory()) walk(p, out);
    else out.push(p);
  }
  return out;
}

export function checkPackage(dir, { allowHosts = [], limitBytes = LIMIT } = {}) {
  const problems = [];
  if (!existsSync(join(dir, "index.html"))) problems.push("index.html is missing at the package root");
  const allowed = new Set([...NEVER_FETCHED, ...allowHosts]);
  let total = 0;
  for (const file of walk(dir)) {
    const rel = relative(dir, file);
    total += statSync(file).size;
    // eslint-disable-next-line no-control-regex
    if (/[^\x21-\x7e/]/.test(rel)) problems.push(`file name is not plain ASCII without spaces: ${rel}`);
    if (!TEXT.test(file)) continue;
    const hosts = new Set();
    for (const m of readFileSync(file, "utf8").matchAll(URL_RE)) {
      try {
        const host = new URL(m[0]).hostname;
        if (!allowed.has(host)) hosts.add(host);
      } catch {
        /* a partial template literal, not a URL */
      }
    }
    for (const h of hosts) problems.push(`external host ${h} in ${rel}`);
  }
  if (total > limitBytes) problems.push(`package is ${total} bytes, over the limit of ${limitBytes}`);
  return problems;
}

if (import.meta.url === pathToFileURL(process.argv[1] ?? "").href) {
  const args = process.argv.slice(2);
  const allowHosts = args.filter((a) => a.startsWith("--allow-host=")).map((a) => a.slice("--allow-host=".length));
  const dir = args.find((a) => !a.startsWith("--")) ?? "dist";
  const problems = checkPackage(dir, { allowHosts });
  if (problems.length) {
    console.error(problems.map((p) => `FAIL: ${p}`).join("\n"));
    process.exit(1);
  }
  console.log(`package check passed for ${dir}`);
}

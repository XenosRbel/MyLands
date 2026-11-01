import { mkdtempSync, mkdirSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { describe, expect, it } from "vitest";
import { checkPackage } from "./check-package.mjs";

function build(files) {
  const dir = mkdtempSync(join(tmpdir(), "pkg-"));
  for (const [name, text] of Object.entries(files)) {
    const p = join(dir, name);
    mkdirSync(join(p, ".."), { recursive: true });
    writeFileSync(p, text);
  }
  return dir;
}

describe("checkPackage", () => {
  it("accepts a clean build", () => {
    const dir = build({
      "index.html": '<script src="/assets/a.js"></script>',
      "assets/a.js": 'const ns="http://www.w3.org/2000/svg";',
    });
    expect(checkPackage(dir)).toEqual([]);
  });

  it("reports a missing root index.html", () => {
    expect(checkPackage(build({ "assets/a.js": "1" })).join("\n")).toMatch(/index\.html/);
  });

  it("reports an external URL", () => {
    const dir = build({ "index.html": "", "assets/a.js": 'fetch("https://fonts.googleapis.com/css2")' });
    expect(checkPackage(dir).join("\n")).toMatch(/fonts\.googleapis\.com/);
  });

  it("allows the declared API host", () => {
    const dir = build({ "index.html": "", "assets/a.js": 'const api="https://api.example.com/v1";' });
    expect(checkPackage(dir, { allowHosts: ["api.example.com"] })).toEqual([]);
  });

  it("reports names with spaces or non-ASCII characters", () => {
    const dir = build({ "index.html": "", "assets/my file.png": "x", "assets/карта.png": "x" });
    const out = checkPackage(dir).join("\n");
    expect(out).toMatch(/my file\.png/);
    expect(out).toMatch(/карта\.png/);
  });

  it("reports a build over the size limit", () => {
    const dir = build({ "index.html": "x".repeat(2048) });
    expect(checkPackage(dir, { limitBytes: 1024 }).join("\n")).toMatch(/limit/);
  });
});

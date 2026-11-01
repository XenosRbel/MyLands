#!/usr/bin/env python3
"""Validate the My Lands specification: JSON/CSV, links, card graph, coverage."""
import csv, json, re, sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
REF = ROOT / "docs/reference/mylands"
MOD = ROOT / "docs/modernization"
CARDS = ROOT / "tasks/mylands"
errors = []


def err(msg):
    errors.append(msg)


for p in REF.glob("*.json"):
    try:
        json.loads(p.read_text())
    except ValueError as e:
        err(f"{p.name}: invalid JSON: {e}")

for p in [REF / "coverage-pages.csv", REF / "coverage-mechanics.csv", MOD / "art-inventory.csv"]:
    rows = list(csv.DictReader(p.open()))
    if not rows:
        err(f"{p.name}: empty")

# Cards: unique IDs, resolvable dependencies, required fields, no cycles.
cards, deps = {}, {}
for p in sorted(CARDS.glob("ML-*.md")):
    t = p.read_text()
    cid = p.stem
    if not t.startswith(f"# {cid} "):
        err(f"{p.name}: title does not start with ID")
    for field in ("Цель", "Зависимости", "Модули", "Критерии приёмки", "Обязательные тесты", "Результат"):
        if f"**{field}**" not in t:
            err(f"{p.name}: missing {field}")
    m = re.search(r"\*\*Зависимости\*\*: (.*)", t)
    deps[cid] = [] if not m or m.group(1).strip() == "нет" else [d.strip() for d in m.group(1).split(",")]
    cards[cid] = t
for cid, ds in deps.items():
    for d in ds:
        if d not in cards:
            err(f"{cid}: unknown dependency {d}")


def cyclic(node, stack):
    if node in stack:
        return True
    return any(cyclic(d, stack | {node}) for d in deps.get(node, []) if d in deps)


for cid in deps:
    if cyclic(cid, frozenset()):
        err(f"{cid}: dependency cycle")

# Page coverage: every navigation page has a non-empty status and none is unread.
nav = {p["page"] for p in json.loads((REF / "navigation-inventory.json").read_text())["pages"]}
pages = {r["page"]: r for r in csv.DictReader((REF / "coverage-pages.csv").open())}
for n in sorted(nav - set(pages)):
    err(f"coverage-pages.csv: missing page {n}")
for n, r in pages.items():
    if r["status"] == "NOT_READ":
        err(f"page {n} is NOT_READ")

extra = list(csv.DictReader((REF / "coverage-pages-extra.csv").open()))
if len(extra) < 90 or any(not r["status"] for r in extra):
    err("coverage-pages-extra.csv: expected at least 90 rows with status")

# Mechanics coverage: every research ID has an existing card.
catalog = set(re.findall(r"^\| ([A-Z]+-\d+) \|", (ROOT / "docs/research/MECHANICS_CATALOG.md").read_text(), re.M))
mech = {r["mechanic_id"]: r["card"] for r in csv.DictReader((REF / "coverage-mechanics.csv").open())}
for i in sorted(catalog - set(mech)):
    err(f"coverage-mechanics.csv: missing {i}")
for i, c in mech.items():
    if c not in cards:
        err(f"mechanic {i}: card {c} does not exist")

# Exact tables: roster size and consistent row widths.
roster = list(csv.DictReader((REF / "units-roster.csv").open()))
if len(roster) != 280:
    err(f"units-roster.csv: expected 280 rows, got {len(roster)}")
for page, tabs in json.loads((REF / "tables-dom.json").read_text())["pages"].items():
    for name, rows in tabs.items():
        width = len(rows[0])
        for r in rows:
            if len(r) != width:
                err(f"tables-dom.json: {page}/{name}: row width {len(r)} != {width}")

# Markdown links to local files must resolve.
for md in list(REF.glob("*.md")) + list(MOD.glob("*.md")) + list(CARDS.glob("*.md")):
    for target in re.findall(r"\]\(([^)#\s]+)", md.read_text()):
        if target.startswith(("http", "mailto")):
            continue
        if not (md.parent / target).resolve().exists():
            err(f"{md.relative_to(ROOT)}: broken link {target}")

# Sources must be unique by ID per registry.
for name in ("sources-world.json", "sources-economy.json", "sources-combat-social.json", "sources-supplement.json"):
    ids = [s["id"] for s in json.loads((REF / name).read_text())]
    if len(ids) != len(set(ids)):
        err(f"{name}: duplicate source IDs")

if errors:
    print("\n".join(errors))
    sys.exit(1)
print(f"OK: {len(cards)} cards, {len(pages)} pages, {len(mech)} mechanics")

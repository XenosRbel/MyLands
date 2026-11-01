#!/usr/bin/env python3
"""Download temporary debug sprites from the My Lands wiki into an untracked folder.

Basis: the customer states that the project is a contractor demo for the rights holder.
Files go to art-reference/wiki-temp/ (git-ignored). They are for local debugging only.
Run from a machine with internet access: python3 scripts/fetch_wiki_sprites.py [--dry-run]
"""
import argparse, csv, hashlib, json, re, sys, time, urllib.parse, urllib.request
from html.parser import HTMLParser
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
REF = ROOT / "docs/reference/mylands"
OUT = ROOT / "art-reference/wiki-temp"
BASE = "https://ru.mlgame.org"
PREFIX_RACE = [("wiz", "wizard"), ("chi", "chimera"), ("dr", "drow"), ("de", "demon"), ("el", "elemental"),
               ("un", "undead"), ("m", "monster"), ("e", "elf")]


class Imgs(HTMLParser):
    def __init__(self):
        super().__init__()
        self.srcs = []

    def handle_starttag(self, tag, attrs):
        if tag == "img":
            src = dict(attrs).get("src") or ""
            if "/wiki-ru/attach/" in src:
                self.srcs.append(src)


def page_names():
    nav = json.loads((REF / "navigation-inventory.json").read_text())["pages"]
    names = {p["page"] for p in nav}
    extra = REF / "coverage-pages-extra.csv"
    if extra.exists():
        names |= {r["page"] for r in csv.DictReader(extra.open())}
    return sorted(names - {"UserPreferences"})


def image_urls(html):
    p = Imgs()
    p.feed(html)
    return sorted({urllib.parse.urljoin(BASE, s) for s in p.srcs})


KNOWN = {"scavenger", "melee", "cavalry", "flyer", "range", "healer", "miner", "mage", "houses", "market",
         "warehouse", "shop", "windmill", "sawmill", "mine_gold", "mine_iron", "mine_stone", "alchemist",
         "recovery", "prison", "tavern", "artefacts", "temple", "terraformer", "resurrector", "colony", "exped",
         "fortification", "watch_tower", "antiwatch_tower", "spy_guild", "antispy_guild", "merch_guild", "tower",
         "magic_tower", "diplomat", "clan_warehouse", "clan_reservation", "cityhall", "gates", "castlehall"}


def guess_race(filename):
    stem = re.sub(r"(_\d.*|\d+.*)$", "", Path(filename).stem.lower())
    if stem in KNOWN:
        return "knight"
    for prefix, race in PREFIX_RACE:
        if stem.startswith(prefix) and stem[len(prefix):] in KNOWN:
            return race
    return "unknown"


def get(url):
    req = urllib.request.Request(url, headers={"User-Agent": "realmforge-demo-debug-fetch/1.0"})
    with urllib.request.urlopen(req, timeout=30) as r:
        return r.read()


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--dry-run", action="store_true", help="list images only")
    ap.add_argument("--delay", type=float, default=0.5, help="seconds between requests")
    args = ap.parse_args()
    manifest = []
    for name in page_names():
        url = f"{BASE}/wiki-ru/Wiki.jsp?page={urllib.parse.quote(name)}"
        try:
            html = get(url).decode("utf-8", "replace")
        except Exception as e:
            print(f"SKIP page {name}: {e}", file=sys.stderr)
            continue
        for img in image_urls(html):
            rel = urllib.parse.unquote(img.split("/wiki-ru/attach/", 1)[1])
            fname = rel.split("/")[-1]
            entry = {"page": name, "url": img, "file": f"{name}/{fname}", "raceGuess": guess_race(fname)}
            if not args.dry_run:
                dest = OUT / name / fname
                if not dest.exists():
                    try:
                        data = get(img)
                    except Exception as e:
                        print(f"SKIP image {img}: {e}", file=sys.stderr)
                        continue
                    dest.parent.mkdir(parents=True, exist_ok=True)
                    dest.write_bytes(data)
                    time.sleep(args.delay)
                data = dest.read_bytes()
                entry.update(bytes=len(data), sha256=hashlib.sha256(data).hexdigest())
            manifest.append(entry)
        time.sleep(args.delay)
    OUT.mkdir(parents=True, exist_ok=True)
    (OUT / "manifest.json").write_text(json.dumps(manifest, ensure_ascii=False, indent=1))
    total = sum(e.get("bytes", 0) for e in manifest)
    print(f"{len(manifest)} images, {total} bytes, manifest: {OUT / 'manifest.json'}")


if __name__ == "__main__":
    main()

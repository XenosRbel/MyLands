#!/usr/bin/env python3
"""Normalises generated art into a Game Pack's assets directory.

Usage: python3 scripts/import_art.py <source assets dir> <pack dir>
  e.g. python3 scripts/import_art.py ~/Ashen_Crowns_Resource_Pack_v0.1/packs/ashen-crowns/assets packs/ashen-crowns

For each category it trims transparent margins to the visible content,
fits the result into the category's target box and writes lossy WebP with
alpha. Prints a JSON summary of written files (relative to <pack dir>/assets).
Requires Pillow.
"""
import json
import sys
from pathlib import Path

from PIL import Image

# category -> (target box, mode). "fit": keep aspect inside box; "square": centre-crop to square.
SPEC = {
    "resources": ((128, 128), "fit"),
    "items": ((128, 128), "fit"),
    "buildings": ((384, 320), "fit"),
    "units": ((256, 256), "fit"),
    "heroes": ((384, 384), "fit"),
    "terrain": ((256, 256), "square"),
    "ui": ((600, 440), "fit"),
}
SKIP_SUFFIX = ("_aux.png", "ui_atlas.png")


def trim(im: Image.Image) -> Image.Image:
    box = im.getchannel("A").point(lambda a: 255 if a > 8 else 0).getbbox()
    return im.crop(box) if box else im


def process(src: Path, box, mode) -> Image.Image:
    im = trim(Image.open(src).convert("RGBA"))
    if mode == "square":
        s = min(im.size)
        l, t = (im.width - s) // 2, (im.height - s) // 2
        return im.crop((l, t, l + s, t + s)).resize(box, Image.LANCZOS)
    im.thumbnail(box, Image.LANCZOS)
    return im


def main(src_dir: str, pack_dir: str) -> None:
    src, out = Path(src_dir), Path(pack_dir) / "assets"
    written = {}
    for cat, (box, mode) in SPEC.items():
        for f in sorted((src / cat).glob("*.png")):
            if f.name.endswith(SKIP_SUFFIX):
                continue
            dst = out / cat / (f.stem + ".webp")
            dst.parent.mkdir(parents=True, exist_ok=True)
            process(f, box, mode).save(dst, "WEBP", quality=86, method=6)
            written[f"{cat}/{f.stem}"] = {"path": f"{cat}/{f.stem}.webp", "bytes": dst.stat().st_size}
    json.dump(written, sys.stdout, indent=1)
    print(f"\n{len(written)} files, {sum(v['bytes'] for v in written.values()) // 1024} KiB", file=sys.stderr)


if __name__ == "__main__":
    if len(sys.argv) != 3:
        sys.exit(__doc__)
    main(sys.argv[1], sys.argv[2])

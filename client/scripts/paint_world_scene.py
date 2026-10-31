#!/usr/bin/env python3
"""Build an original raster castle-landscape from the Game Pack's painted terrain.

No SVG is used. The building sprites remain independently clickable in React.
Requires Pillow. The generated world scene is a visual prototype pending art review.
"""
import math
import random
from pathlib import Path
from PIL import Image, ImageDraw, ImageEnhance, ImageFilter, ImageOps

ROOT = Path(__file__).resolve().parents[2]
PACK = ROOT / "packs" / "ashen-crowns" / "assets"
OUTPUT = ROOT / "client" / "public" / "ui" / "city-landscape.webp"
W,H = 1600,900
rng = random.Random("ashen-crowns-landscape-v2")

def terrain(name, size):
    src = Image.open(PACK/"terrain"/f"{name}.webp").convert("RGB")
    return ImageOps.fit(src,size,method=Image.Resampling.LANCZOS,centering=(.5,.5))

def patch(scene, kind, bounds, blur=14, alpha=230):
    x0,y0,x1,y1=bounds
    width=max(1,x1-x0);height=max(1,y1-y0)
    layer=terrain(kind,(width,height)).convert("RGBA")
    mask=Image.new("L",(width,height),0)
    d=ImageDraw.Draw(mask)
    d.ellipse((-8,-6,width+8,height+5),fill=alpha)
    mask=mask.filter(ImageFilter.GaussianBlur(blur))
    layer.putalpha(mask)
    left=max(0,x0);top=max(0,y0);right=min(W,x1);bottom=min(H,y1)
    if right<=left or bottom<=top:return
    layer=layer.crop((left-x0,top-y0,right-x0,bottom-y0))
    scene.alpha_composite(layer,(left,top))

def main():
    OUTPUT.parent.mkdir(parents=True,exist_ok=True)
    base=terrain("plains",(W,H))
    base=ImageEnhance.Color(base).enhance(.85)
    scene=base.convert("RGBA")
    # No visible tile seams: large, feathered fragments of authored terrain art.
    for i in range(94):
        x=rng.randrange(-140,W+50);y=rng.randrange(-100,H+50)
        ww=rng.randrange(130,380);hh=rng.randrange(100,280)
        d=((x+ww/2-W/2)/(W*.49))**2+((y+hh/2-H/2)/(H*.42))**2
        kind=("forest" if d>.75 and i%3!=0 else
              "hills" if i%6==0 else "plains" if i%3 else "forest")
        patch(scene,kind,(max(-ww+2,x),max(-hh+2,y),min(W+ww-2,x+ww),min(H+hh-2,y+hh)),rng.randrange(15,34),rng.randrange(80,190))
    for x,y in [(-220,-160),(W-360,-210),(-260,H-360),(W-360,H-300)]:
        patch(scene,"mountains",(x,y,x+470,y+420),29,250)
    for x,y in [(-180,130),(W-230,110),(-160,H-270),(W-280,H-200)]:
        patch(scene,"lake",(x,y,x+345,y+360),22,240)
    # Softly painted footpaths connect the actual Game Pack building positions.
    roads=Image.new("RGBA",(W,H),(0,0,0,0))
    dr=ImageDraw.Draw(roads)
    ways=[
      [(180,480),(345,435),(590,470),(780,430),(1020,450),(1390,510)],
      [(345,195),(425,330),(680,440),(805,610),(1120,730)],
      [(250,710),(490,610),(680,560),(950,590),(1340,350)],
    ]
    for points in ways:
        dr.line(points,fill=(74,62,45,105),width=55,joint="curve")
        dr.line(points,fill=(162,142,98,98),width=37,joint="curve")
        dr.line(points,fill=(196,171,117,53),width=15,joint="curve")
    roads=roads.filter(ImageFilter.GaussianBlur(7))
    scene=Image.alpha_composite(scene,roads)
    # A fortress perimeter with visible towers, built from textured paint.
    perimeter=[(177,357),(298,211),(565,130),(855,103),(1128,165),
               (1410,335),(1467,625),(1281,778),(1015,824),
               (700,814),(390,746),(181,634)]
    outlines=Image.new("RGBA",(W,H),(0,0,0,0))
    d=ImageDraw.Draw(outlines)
    loop=perimeter+[perimeter[0]]
    d.line([(x+6,y+18) for x,y in loop],fill=(19,31,28,135),width=45,joint="curve")
    d.line(loop,fill=(84,88,75,255),width=32,joint="curve")
    d.line([(x,y-9) for x,y in loop],fill=(186,185,153,240),width=12,joint="curve")
    for i in range(len(perimeter)):
        x,y=perimeter[i]
        d.rounded_rectangle((x-20,y-29,x+20,y+20),6,fill=(105,111,97,255),outline=(57,64,57,255),width=4)
        d.polygon([(x-25,y-29),(x,y-71),(x+25,y-29)],fill=(77,70,84,255))
        d.line([(x-22,y-29),(x,y-71),(x+22,y-29)],fill=(189,175,141,210),width=4)
        d.rectangle((x-5,y-9,x+5,y+13),fill=(52,52,47,255))
    scene=Image.alpha_composite(scene,outlines)
    # Waterfall spray and atmospheric sunlight on the edges.
    mist=Image.new("RGBA",(W,H),(0,0,0,0))
    d=ImageDraw.Draw(mist)
    for cx,cy in [(53,280),(1548,300),(130,765),(1530,795)]:
        for i in range(160):
            x=cx+rng.gauss(0,59);y=cy+rng.gauss(0,75)
            r=rng.randrange(2,17)
            d.ellipse((x-r,y-r,x+r,y+r),fill=(210,233,221,rng.randrange(9,35)))
    mist=mist.filter(ImageFilter.GaussianBlur(13))
    scene=Image.alpha_composite(scene,mist)
    # Preserve texture. Never draw an opaque repeating pattern above scenery.
    vignette=Image.new("L",(W,H),0)
    pix=vignette.load()
    for y in range(H):
        for x in range(W):
            dist=((x-W/2)/(W*.68))**2+((y-H/2)/(H*.68))**2
            pix[x,y]=min(125,max(0,int((dist-.48)*107)))
    shadow=Image.new("RGBA",(W,H),(15,29,22,0))
    shadow.putalpha(vignette)
    scene=Image.alpha_composite(scene,shadow)
    scene=scene.convert("RGB")
    scene.save(OUTPUT,format="WEBP",quality=89,method=5)
    print(f"Created original raster city environment: {OUTPUT}.")

if __name__=="__main__":
    main()

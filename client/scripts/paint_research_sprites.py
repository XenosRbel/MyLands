#!/usr/bin/env python3
"""Generate original raster technology art. Never publish SVG fallbacks.

The illustrations are layered, textured and antialiased directly in RGB pixels.
They are not conversions of the previous ResearchEmblem SVG artwork.
Run with: python3 client/scripts/paint_research_sprites.py
Requires: Pillow >= 10.
"""
from __future__ import annotations

import math
import random
from pathlib import Path
from PIL import Image, ImageDraw, ImageFilter, ImageFont

ROOT = Path(__file__).resolve().parents[1]
OUTPUT = ROOT / "public" / "ui" / "research"
SIZE = 256
SCALE = 2
TECHS = {
    "forestry": ("tree", (110, 150, 93)),
    "agriculture": ("wheat", (180, 156, 75)),
    "masonry": ("castle", (154, 157, 142)),
    "metallurgy": ("hammer", (190, 136, 86)),
    "exchequer": ("coin", (207, 161, 75)),
    "husbandry": ("horn", (163, 139, 98)),
    "tactics": ("swords", (175, 168, 144)),
    "armour": ("shield", (158, 178, 170)),
    "logistics": ("wagon", (159, 145, 104)),
    "espionage": ("eye", (134, 166, 173)),
    "beastlore": ("claw", (171, 137, 104)),
    "engineering": ("gear", (178, 155, 125)),
    "arcana": ("crystal", (164, 129, 204)),
    "plunder": ("dagger", (182, 117, 91)),
    "heroics": ("star", (217, 173, 106)),
    "path_architect": ("tower", (195, 160, 113)),
    "path_steward": ("crown", (150, 174, 111)),
    "path_warlord": ("helmet", (197, 133, 115)),
}

def polygon(draw, pts, color, width=0, border=None):
    draw.polygon(pts, fill=color)
    if border:
        draw.line(pts + [pts[0]], fill=border, width=width, joint="curve")

def motif(draw, kind, light, shadow):
    x, y = 128, 132
    gold = tuple(min(255, c+32) for c in light)
    dark = (46, 37, 32)
    def L(points, color=light, w=13):
        draw.line(points, fill=dark, width=w+7, joint="curve")
        draw.line(points, fill=color, width=w, joint="curve")
    def P(points, fill=light):
        polygon(draw, points, fill, 7, dark)
    if kind == "tree":
        P([(x,48),(x-65,148),(x-29,139),(x-53,176),(x+52,176),(x+30,139),(x+65,148)])
        L([(x,156),(x,202)], gold, 15)
        for i in range(10): L([(x+(-30+i*7),125+i%3*10),(x+(-12+i*4),94+i%2*11)], gold, 3)
    elif kind == "book":
        P([(48,71),(112,72),(128,91),(145,72),(211,71),(211,185),(145,186),(128,203),(112,187),(48,185)])
        L([(128,92),(128,195)],gold,9)
        L([(63,104),(105,105)],dark,5)
        L([(153,105),(195,104)],dark,5)
    elif kind == "mail":
        P([(49,87),(210,87),(210,174),(49,174)],gold)
        L([(50,88),(128,140),(209,88)],shadow,9)
        L([(51,175),(104,133)],shadow,6)
        L([(207,175),(151,133)],shadow,6)
    elif kind == "clock":
        draw.ellipse((54,58,202,204),fill=gold,outline=dark,width=9)
        draw.ellipse((69,73,187,190),fill=shadow,outline=(245,220,163),width=5)
        L([(128,95),(128,135),(163,159)],gold,11)
    elif kind == "compass":
        draw.ellipse((54,58,202,204),fill=gold,outline=dark,width=9)
        P([(128,67),(161,128),(128,193),(95,128)],shadow)
        P([(128,70),(128,128),(161,128)],gold)
    elif kind == "cross":
        L([(77,80),(180,181)],gold,18)
        L([(180,80),(77,181)],gold,18)
    elif kind == "wheat":
        L([(125,210),(125,65)], gold, 9)
        for i in range(7):
            yy=83+i*17
            P([(122,yy),(98,yy-24),(88,yy-42),(90,yy-10),(120,yy+10)], gold if i%2 else light)
            P([(129,yy),(153,yy-25),(163,yy-42),(161,yy-10),(132,yy+10)])
    elif kind in ("castle","tower"):
        P([(60,196),(60,104),(74,104),(74,85),(88,85),(88,104),(110,104),
           (110,68),(126,68),(126,84),(142,84),(142,68),(158,68),(158,104),
           (180,104),(180,85),(194,85),(194,104),(206,104),(206,196)])
        polygon(draw,[(111,196),(111,153),(127,140),(145,153),(145,196)], dark)
        for xx in (82,172): draw.rectangle((xx,132,xx+11,154),fill=dark)
        L([(54,197),(213,197)],gold,8)
    elif kind=="hammer":
        L([(88,200),(166,81)],(138,102,65),23)
        P([(119,60),(177,60),(198,85),(181,113),(119,111),(106,86)], gold)
        L([(115,61),(181,62)],(238,205,153),5)
    elif kind=="coin":
        draw.ellipse((70,72,186,190),fill=light,outline=dark,width=10)
        draw.ellipse((83,84,175,179),outline=gold,width=6)
        L([(128,98),(128,165)],dark,8)
        L([(104,110),(151,110)],dark,7)
        L([(105,152),(154,152)],dark,7)
    elif kind=="horn":
        P([(63,84),(94,68),(115,98),(141,110),(166,98),(192,65),(201,100),
           (174,151),(157,190),(102,192),(80,146)])
        draw.ellipse((101,140,115,155),fill=dark)
        draw.ellipse((148,139,162,155),fill=dark)
        L([(112,177),(152,177)],dark,8)
    elif kind=="swords":
        L([(73,203),(186,66)],gold,17);L([(184,204),(74,65)],light,17)
        L([(68,175),(105,204)],gold,10);L([(156,204),(194,174)],gold,10)
    elif kind=="shield":
        P([(128,55),(194,77),(187,148),(163,184),(128,206),(93,184),(68,145),(62,76)],gold)
        P([(128,73),(174,91),(169,144),(151,174),(128,189),(106,172),(86,143),(82,91)],shadow)
        L([(128,87),(128,178)],gold,9)
    elif kind=="wagon":
        P([(63,113),(180,113),(200,176),(53,176)],gold)
        P([(82,93),(177,90),(183,115),(78,116)],shadow)
        for xx in (86,167):
            draw.ellipse((xx-25,155,xx+25,207),fill=dark,outline=gold,width=9)
            draw.ellipse((xx-8,172,xx+8,190),fill=(211,183,126))
    elif kind=="eye":
        P([(39,134),(76,98),(128,83),(179,99),(216,133),(184,169),(128,189),(72,167)],gold)
        draw.ellipse((95,99,162,169),fill=shadow,outline=dark,width=8)
        draw.ellipse((118,117,145,145),fill=(27,36,42))
        draw.ellipse((121,121,132,133),fill=(211,233,224))
    elif kind=="claw":
        for xx,angle in ((89,0),(128,12),(166,22)):
            P([(xx,74),(xx+23,113),(xx+17,170),(xx-7,209),(xx+7,150)],gold)
        L([(73,185),(173,200)],shadow,7)
    elif kind=="gear":
        pts=[]
        for i in range(24):
            a=i*math.pi/12
            r=88 if i%3!=0 else 102
            pts.append((int(x+math.cos(a)*r),int(y+math.sin(a)*r)))
        P(pts,gold)
        draw.ellipse((98,102,158,162),fill=dark,outline=light,width=9)
    elif kind=="crystal":
        P([(128,38),(193,105),(161,186),(128,217),(88,184),(62,106)],gold)
        P([(128,38),(155,106),(128,217),(105,105)],shadow)
        L([(62,106),(193,105)],(230,211,240),5)
    elif kind=="dagger":
        P([(166,49),(194,82),(131,167),(111,149)],gold)
        L([(103,148),(79,186)],shadow,15)
        L([(75,143),(126,176)],gold,14)
    elif kind=="star":
        pts=[]
        for i in range(10):
            a=-math.pi/2+i*math.pi/5
            rr=94 if i%2==0 else 45
            pts.append((int(x+math.cos(a)*rr),int(y+math.sin(a)*rr)))
        P(pts,gold)
        draw.ellipse((119,123,139,143),fill=(255,227,151))
    elif kind=="crown":
        P([(58,91),(85,128),(105,76),(128,127),(154,75),(175,124),(200,91),(185,186),(72,186)],gold)
        L([(76,174),(185,174)],shadow,8)
        for xx,yy in ((58,91),(105,76),(154,75),(200,91)):
            draw.ellipse((xx-7,yy-7,xx+7,yy+7),fill=(225,210,165),outline=dark,width=3)
    elif kind=="helmet":
        P([(64,151),(68,113),(91,77),(127,67),(164,80),(189,113),(191,154),(169,164),
           (161,197),(146,196),(142,158),(114,159),(110,197),(93,196),(86,165)],gold)
        L([(79,130),(176,130)],shadow,13)
        L([(128,67),(128,131)],(233,198,149),8)

def paint(tech, kind, tone, stage):
    rng = random.Random(tech+"_"+str(stage))
    n=SIZE*SCALE
    im=Image.new("RGBA",(n,n),(0,0,0,0))
    px=im.load()
    for y in range(n):
        for x in range(n):
            cx=(x-n/2)/(n/2);cy=(y-n/2)/(n/2)
            rr=(cx*cx+cy*cy)**.5
            rough=rng.randrange(-12,13)
            light=max(0,1-rr*.64)
            px[x,y]=(int(48+52*light+rough),int(38+35*light+rough),int(35+23*light+rough),255)
    d=ImageDraw.Draw(im)
    for i in range(18):
        x=rng.randrange(n);y=rng.randrange(n)
        size=rng.randrange(3,30)
        col=(180+rng.randrange(50),153+rng.randrange(45),99+rng.randrange(49),rng.randrange(10,38))
        glaze=Image.new("RGBA",(n,n));ImageDraw.Draw(glaze).ellipse((x,y,x+size,y+size),fill=col)
        im=Image.alpha_composite(im,glaze)
    d=ImageDraw.Draw(im)
    center=n//2
    ring=(min(246,196+stage*5),min(235,157+stage*7),min(192,92+stage*9))
    d.ellipse((12,12,n-12,n-12),fill=None,outline=(37,28,25),width=24)
    d.ellipse((24,24,n-24,n-24),outline=ring,width=12)
    d.ellipse((40,40,n-40,n-40),outline=(106,75,53),width=5)
    for i in range(4+stage*2):
        a=2*math.pi*i/(4+stage*2)-math.pi/2
        cx=center+int(math.cos(a)*(n//2-30))
        cy=center+int(math.sin(a)*(n//2-30))
        d.ellipse((cx-8,cy-8,cx+8,cy+8),fill=ring,outline=(72,51,32),width=3)
    # Draw large original icon silhouette with oversampling and shadow.
    symbol=Image.new("RGBA",(SIZE,SIZE))
    sd=ImageDraw.Draw(symbol)
    shadow=(62,44,34)
    motif(sd,kind,tone,shadow)
    symbol=symbol.resize((n,n),Image.Resampling.LANCZOS)
    shade=Image.new("RGBA",(n,n))
    shade.paste(symbol,(8,11))
    shade=shade.filter(ImageFilter.GaussianBlur(9))
    shade.putalpha(shade.getchannel("A").point(lambda alpha: int(alpha*.48)))
    im=Image.alpha_composite(im,shade)
    im=Image.alpha_composite(im,symbol)
    d=ImageDraw.Draw(im)
    d.arc((26,26,n-26,n-26),198,299,fill=(247,220,161),width=5)
    if stage>=4:
        d.arc((36,36,n-36,n-36),-8,10+stage*13,fill=(244,224,175),width=3+stage//3)
    if stage>=7:
        d.arc((56,56,n-56,n-56),160,160+stage*9,fill=(222,199,142),width=3+stage//4)
    if stage>=9:
        for a in (0,90,180,270):
            x=center+int(math.cos(math.radians(a))*185)
            y=center+int(math.sin(math.radians(a))*185)
            d.ellipse((x-10,y-10,x+10,y+10),fill=(242,222,169),outline=(70,45,27),width=3)
    im=im.convert("RGB").resize((SIZE,SIZE),Image.Resampling.LANCZOS)
    return im

UTILITY = {
    "log":"tree", "wheat":"wheat", "stone":"castle", "ingot":"hammer",
    "coin":"coin", "scroll":"book", "crystal":"crystal", "leaf":"tree",
    "bolt":"star", "chip":"gear", "sword":"swords", "shield":"shield",
    "heart":"star", "boot":"wagon", "bag":"crown", "people":"horn",
    "castle":"castle", "clock":"clock", "bell":"crown", "map":"compass",
    "flag":"crown", "star":"star", "book":"book", "mail":"mail",
    "chart":"gear", "user":"helmet", "gear":"gear", "scale":"swords",
    "menu":"book", "eye":"eye", "skull":"helmet", "close":"cross",
}

def paint_utility_icons():
    folder = ROOT / "public" / "ui" / "glyphs"
    folder.mkdir(parents=True,exist_ok=True)
    for name,kind in UTILITY.items():
        picture=Image.new("RGBA",(SIZE,SIZE),(0,0,0,0))
        draw=ImageDraw.Draw(picture)
        motif(draw,kind,(203,171,119),(98,71,43))
        shadow=Image.new("RGBA",(SIZE,SIZE))
        shadow.paste(picture,(3,4))
        shadow=shadow.filter(ImageFilter.GaussianBlur(5))
        shadow.putalpha(shadow.getchannel("A").point(lambda a:int(a*.45)))
        picture=Image.alpha_composite(shadow,picture)
        picture=picture.resize((96,96),Image.Resampling.LANCZOS)
        picture.save(folder/f"{name}.webp",format="WEBP",quality=89,method=4)
    print(f"Painted {len(UTILITY)} original raster utility icons.")

def paint_world_globe():
    n=480
    rng=random.Random("ashen-crowns-world-globe")
    image=Image.new("RGBA",(n,n),(0,0,0,0))
    draw=ImageDraw.Draw(image)
    draw.ellipse((4,4,n-4,n-4),fill=(48,31,40,255),outline=(188,144,79),width=16)
    draw.ellipse((27,27,n-27,n-27),fill=(36,46,48,255),outline=(238,197,117),width=9)
    draw.ellipse((47,47,n-47,n-47),fill=(34,92,124,255),outline=(160,130,79),width=8)
    # Soft speckled ocean pigment within a spherical mask.
    mask=Image.new("L",(n,n))
    ImageDraw.Draw(mask).ellipse((51,51,n-51,n-51),fill=255)
    sea=Image.new("RGBA",(n,n),(0,0,0,0))
    pixels=sea.load()
    for y in range(51,n-51):
        for x in range(51,n-51):
            if mask.getpixel((x,y))==0:continue
            d=((x-n*.38)**2+(y-n*.3)**2)**.5/n
            light=max(0.0,1.0-d)
            noise=rng.randrange(-13,14)
            pixels[x,y]=(int(29+49*light+noise),int(81+71*light+noise),int(104+87*light+noise),255)
    draw=ImageDraw.Draw(sea)
    continents=[
      [(72,167),(103,122),(145,105),(168,126),(179,164),(160,196),(175,221),(155,263),(121,275),(101,315),(78,282),(65,234)],
      [(226,85),(279,88),(314,120),(354,136),(388,173),(367,199),(329,213),(315,258),(286,306),(254,338),(232,297),(244,256),(217,213),(237,177)],
      [(294,355),(329,333),(364,362),(382,406),(347,425),(313,411)],
    ]
    for pts in continents:
        draw.polygon(pts,fill=(97,136,83,255))
        draw.line(pts+[pts[0]],fill=(180,183,118,240),width=5,joint="curve")
        for _ in range(22):
            x=rng.randrange(75,392);y=rng.randrange(105,412)
            if rng.random()<.48:
                draw.ellipse((x,y,x+rng.randrange(2,7),y+rng.randrange(2,7)),fill=(139,167,107,60))
    sea.putalpha(mask)
    image=Image.alpha_composite(image,sea)
    draw=ImageDraw.Draw(image)
    draw.ellipse((41,41,n-41,n-41),outline=(236,194,111),width=8)
    draw.arc((70,70,n-70,n-70),192,282,fill=(236,238,215),width=10)
    draw.arc((84,84,n-84,n-84),204,277,fill=(181,219,214),width=7)
    for i in range(12):
        a=2*math.pi*i/12
        cx=n/2+math.cos(a)*(n*.47)
        cy=n/2+math.sin(a)*(n*.47)
        draw.ellipse((cx-8,cy-8,cx+8,cy+8),fill=(211,169,94),outline=(74,49,33),width=3)
    image.save(ROOT/"public"/"ui"/"world-globe.webp",format="WEBP",quality=91,method=5)
    print("Painted original raster world globe.")

def main():
    OUTPUT.mkdir(parents=True,exist_ok=True)
    for tech,(kind,tone) in TECHS.items():
        max_level=3 if tech.startswith("path_") else 5 if tech in {"beastlore","engineering","arcana","plunder","heroics"} else 10
        for stage in range(1,max_level+1):
            out=OUTPUT / f"{tech}-{stage}.webp"
            image=paint(tech,kind,tone,stage)
            image.save(out,format="WEBP",quality=86,method=4)
    print(f"Painted all 134 level-specific technology WebP icons in {OUTPUT}.")
    paint_utility_icons()
    paint_world_globe()

if __name__ == "__main__":
    main()

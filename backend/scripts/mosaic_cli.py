#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Diamond mosaic generator — Square 5D drills only.
Adds STYLE PRESETS via --style: default | contrast | soft | glossy-dark
(Русские синонимы тоже поддерживаются: "контрастный", "мягкий", "глянец", "глянец-для-тёмных-цветов").

- Preview: realistic 5D square stones, **no legend panel**, **no outer frame** on stones.
- Scheme: color-filled cells + high-contrast symbols + grid.
- CSV legend saved separately with --legend.
- Performance: per-color tile cache + optional --threads.
- Separate DPI: --preview-dpi and --scheme-dpi (fallback to --dpi).
- NEW: --style preset tunes lighting/shape params.

Usage example:
  python mosaic_pro_final_v5.py image.jpg palette.xlsx \
      --stones-x 196 --stones-y 277 --stone-size-mm 2.52 \
      --dpi 300 --preview-dpi 240 --scheme-dpi 600 \
      --mode both --legend --threads 8 --style contrast
"""
import argparse
import sys
import os
import re
from pathlib import Path
import csv
from concurrent.futures import ThreadPoolExecutor, as_completed

import pandas as pd
import numpy as np
from PIL import Image, ImageDraw, ImageFont, ImageFilter

# ---------- Appearance controls (base defaults) ----------
LIGHT_DIR = (-0.7, -0.6, 0.9)   # light direction (x,y,z)
AMBIENT = 0.18
SPECULAR_STRENGTH = 0.55
SHININESS = 32
RIM_LIGHT = 0.12
NOISE_AMOUNT = 0.035

# Square drill character
ROUND_RADIUS = 0.02      # near-sharp corners
FRAME_FACTOR = 0.65      # darker outer frame (used only if show_frame=True)
GAP_INSET = 0.00         # visual gap from stone edge
SHADOW_BLUR = 0.5
SHADOW_OPACITY = 0.45
FACET_TILT = 0.8         # tilt of facet normals toward center (0.6..1.0 typical)

# Grid color for scheme
GRID_RGBA = (200, 200, 200, 255)

# ---------- Styles ----------
def normalize_style(name):
    if not name:
        return 'default'
    s = str(name).strip().lower()
    # Russian synonyms
    mapping = {
        'default': 'default',
        'стандарт': 'default',
        'по-умолчанию': 'default',
        'контрастный': 'contrast',
        'контраст': 'contrast',
        'contrast': 'contrast',
        'мягкий': 'soft',
        'soft': 'soft',
        'глянец': 'glossy-dark',
        'глянец-для-тёмных-цветов': 'glossy-dark',
        'глянец-для-темных-цветов': 'glossy-dark',
        'glossy': 'glossy-dark',
        'glossy-dark': 'glossy-dark',
    }
    return mapping.get(s, s)

def apply_style(style_name):
    """Override globals according to a preset."""
    s = normalize_style(style_name)
    g = globals()
    # Reset to base first
    g['LIGHT_DIR'] = (-0.7, -0.6, 0.9)
    g['AMBIENT'] = 0.18
    g['SPECULAR_STRENGTH'] = 0.55
    g['SHININESS'] = 32
    g['RIM_LIGHT'] = 0.12
    g['NOISE_AMOUNT'] = 0.035
    g['ROUND_RADIUS'] = 0.02
    g['FRAME_FACTOR'] = 0.65
    g['GAP_INSET'] = 0.06
    g['SHADOW_BLUR'] = 0.5
    g['SHADOW_OPACITY'] = 0.45
    g['FACET_TILT'] = 0.8

    if s == 'contrast':
        # Punchier highlights, deeper shadows
        g['AMBIENT'] = 0.14
        g['SPECULAR_STRENGTH'] = 0.72
        g['SHININESS'] = 56
        g['RIM_LIGHT'] = 0.14
        g['NOISE_AMOUNT'] = 0.03
        g['GAP_INSET'] = 0.065
        g['SHADOW_OPACITY'] = 0.50
        g['SHADOW_BLUR'] = 0.45
        g['FACET_TILT'] = 0.90

    elif s == 'soft':
        # Gentler, broader highlights
        g['AMBIENT'] = 0.24
        g['SPECULAR_STRENGTH'] = 0.45
        g['SHININESS'] = 20
        g['RIM_LIGHT'] = 0.10
        g['NOISE_AMOUNT'] = 0.03
        g['GAP_INSET'] = 0.055
        g['SHADOW_OPACITY'] = 0.40
        g['SHADOW_BLUR'] = 0.60
        g['ROUND_RADIUS'] = 0.03
        g['FACET_TILT'] = 0.75

    elif s == 'glossy-dark':
        # Extra glossy look tailored for darker colors
        g['AMBIENT'] = 0.16
        g['SPECULAR_STRENGTH'] = 0.80
        g['SHININESS'] = 72
        g['RIM_LIGHT'] = 0.16
        g['NOISE_AMOUNT'] = 0.025
        g['GAP_INSET'] = 0.06
        g['SHADOW_OPACITY'] = 0.52
        g['SHADOW_BLUR'] = 0.48
        g['FACET_TILT'] = 0.92
        g['LIGHT_DIR'] = (-0.6, -0.5, 1.0)  # slightly more frontal for crisp highlights

    elif s == 'default':
        pass
    else:
        print(f"[warn] Unknown style '{style_name}', using defaults.", file=sys.stderr)

# ---------- Small math helpers ----------
def _normalize(v):
    x, y, z = v
    n = (x*x + y*y + z*z) ** 0.5
    if n == 0:
        return (0.0, 0.0, 1.0)
    return (x/n, y/n, z/n)

def _dot(a, b):
    return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]

def _clamp01(x):
    return 0.0 if x < 0.0 else (1.0 if x > 1.0 else x)

def _rgb_scale(rgb, s):
    return tuple(max(0, min(255, int(c * s))) for c in rgb)

def _luma(rgb):
    r, g, b = rgb
    return 0.2126*r + 0.7152*g + 0.0722*b

# ---------- IO ----------
def load_palette(path):
    if path.endswith('.csv'):
        df = pd.read_csv(path, dtype=str)
    else:
        df = pd.read_excel(path, dtype=str)
    
    df.columns = df.columns.str.strip()
    lower = {c.lower(): c for c in df.columns}
    for req in ('code', 'rgb', 'symbol'):
        if req not in lower:
            sys.exit(f"Palette file missing column: {req}")
    codes = df[lower['code']].tolist()
    symbols = df[lower['symbol']].tolist()
    colors = []
    for val in df[lower['rgb']].tolist():
        nums = re.findall(r"\d+", str(val))
        if len(nums) != 3:
            sys.exit(f"Invalid RGB value: {val}")
        colors.append(tuple(map(int, nums)))
    return codes, colors, symbols

def map_to_palette(img_small, palette):
    arr = np.array(img_small).reshape(-1, 3).astype(int)
    pal = np.array(palette)
    d2 = ((arr[:, None] - pal[None])**2).sum(2)
    idx = np.argmin(d2, axis=1)
    return idx.reshape(img_small.height, img_small.width)

# ---------- 5D Square stone (returns RGBA tile) ----------
def render_stone_tile(size, base_color, show_frame=True):
    """
    Square 'drill' 5D look: 13 facets.
    - Square outline with tiny rounding
    - 4 corners + 8 split edge facets + 1 center
    - show_frame=True draws a darker outer frame; set False for clean preview look.
    """
    rr_px = max(1, int(ROUND_RADIUS * size))

    tile = Image.new("RGBA", (int(np.ceil(size)), int(np.ceil(size))), (0,0,0,0))
    tdraw = ImageDraw.Draw(tile)

    def rounded_rect_bbox(m=0):
        s = size - m*2
        return [m, m, m+s, m+s]

    # Mask (outer square)
    mask = Image.new("L", tile.size, 0)
    mdraw = ImageDraw.Draw(mask)
    try:
        mdraw.rounded_rectangle(rounded_rect_bbox(), radius=rr_px, fill=255)
    except Exception:
        mdraw.rectangle(rounded_rect_bbox(), fill=255)

    # Shadow (soft, down-right)
    shadow = Image.new("L", tile.size, 0)
    sdraw = ImageDraw.Draw(shadow)
    offset = int(size * 0.08)
    sdraw.ellipse([size-offset*2, 0, size, offset*2], fill=int(255*SHADOW_OPACITY*0.6))
    sdraw.rectangle([offset, offset, size, size], fill=int(255*SHADOW_OPACITY))
    blur = int(max(1, size * SHADOW_BLUR * 0.5))
    shadow = shadow.filter(ImageFilter.GaussianBlur(blur))
    with_shadow = Image.new("RGBA", tile.size, (0,0,0,0))
    with_shadow.paste((0,0,0,int(255*SHADOW_OPACITY)), (0,0), shadow)

    # Lighting model
    L = _normalize(LIGHT_DIR)
    def shade_color(rgb, nx, ny, nz):
        n = _normalize((nx, ny, nz))
        lambert = max(0.0, _dot(n, L))
        nl = _dot(n, L)
        rx, ry, rz = (2*nl*n[0]-L[0], 2*nl*n[1]-L[1], 2*nl*n[2]-L[2])
        spec = max(0.0, rz) ** SHININESS
        rim = (1.0 - max(0.0, n[2])) * RIM_LIGHT
        intensity = _clamp01(AMBIENT + lambert + SPECULAR_STRENGTH*spec + rim)
        col = _rgb_scale(rgb, 0.6 + 0.7*intensity)
        import random as _rnd
        col = _rgb_scale(col, 1.0 + (_rnd.random()*2-1)*NOISE_AMOUNT)
        return col

    # Inner area for facets
    inset = size * GAP_INSET
    inner0 = inset
    inner1 = size - inset

    # 3×3 grid cut -> base for facets
    gx = [inner0, inner0 + (inner1-inner0)/3, inner0 + 2*(inner1-inner0)/3, inner1]
    gy = gx[:]

    # Center & normalization scale
    cx = size/2.0; cy = size/2.0
    sx = (inner1 - inner0)/2.0; sy = sx
    tilt = FACET_TILT  # facet tilt towards the center

    def facet(poly):
        mx = sum(p[0] for p in poly)/len(poly)
        my = sum(p[1] for p in poly)/len(poly)
        nx = ((mx - cx)/sx) * tilt
        ny = ((my - cy)/sy) * tilt
        col = shade_color(base_color, nx, ny, 0.95)
        tdraw.polygon(poly, fill=col)

    # 4 corner triangles
    facet([(gx[0],gy[0]), (gx[1],gy[0]), (gx[0],gy[1])])  # TL
    facet([(gx[3],gy[0]), (gx[2],gy[0]), (gx[3],gy[1])])  # TR
    facet([(gx[0],gy[3]), (gx[1],gy[3]), (gx[0],gy[2])])  # BL
    facet([(gx[3],gy[3]), (gx[2],gy[3]), (gx[3],gy[2])])  # BR

    # 8 edge triangles
    facet([(gx[1],gy[0]), (gx[2],gy[0]), (gx[2],gy[1])])
    facet([(gx[1],gy[0]), (gx[2],gy[1]), (gx[1],gy[1])])
    facet([(gx[1],gy[3]), (gx[2],gy[3]), (gx[2],gy[2])])
    facet([(gx[1],gy[3]), (gx[2],gy[2]), (gx[1],gy[2])])
    facet([(gx[0],gy[1]), (gx[1],gy[1]), (gx[1],gy[2])])
    facet([(gx[0],gy[1]), (gx[1],gy[2]), (gx[0],gy[2])])
    facet([(gx[3],gy[1]), (gx[2],gy[1]), (gx[2],gy[2])])
    facet([(gx[3],gy[1]), (gx[2],gy[2]), (gx[3],gy[2])])

    # center quad
    facet([(gx[1],gy[1]), (gx[2],gy[1]), (gx[2],gy[2]), (gx[1],gy[2])])

    # central "table" highlight
    table = Image.new("RGBA", tile.size, (0,0,0,0))
    t2 = ImageDraw.Draw(table)
    k = 0.85
    t_half = (inner1-inner0)*k/2.0
    bb = [cx - t_half, cy - t_half, cx + t_half, cy + t_half]
    try:
        t2.rounded_rectangle(bb, radius=int(ROUND_RADIUS*size*0.5), fill=_rgb_scale(base_color, 1.06))
    except Exception:
        t2.rectangle(bb, fill=_rgb_scale(base_color, 1.06))
    hi = Image.new("L", tile.size, 0)
    ImageDraw.Draw(hi).rectangle(bb, fill=140)
    hi = hi.filter(ImageFilter.GaussianBlur(max(1, int(size*0.04))))
    table.putalpha(hi)
    tile.alpha_composite(table)

    # outer frame (gap accent) — only if requested
    if show_frame:
        frame = Image.new("RGBA", tile.size, (0,0,0,0))
        fdraw = ImageDraw.Draw(frame)
        try:
            fdraw.rounded_rectangle(rounded_rect_bbox(inset*0.25), outline=_rgb_scale(base_color, FRAME_FACTOR), width=max(1, int(size*0.03)), radius=rr_px)
        except Exception:
            fdraw.rectangle(rounded_rect_bbox(inset*0.25), outline=_rgb_scale(base_color, FRAME_FACTOR), width=max(1, int(size*0.03)))
        tile.alpha_composite(frame)

    # compose final: shadow + masked stone
    final_layer = Image.new("RGBA", tile.size, (0,0,0,0))
    final_layer.alpha_composite(with_shadow)
    stone_rgb = Image.new("RGBA", tile.size, (0,0,0,0))
    stone_rgb.paste(tile, (0,0), mask)
    final_layer.alpha_composite(stone_rgb)
    return final_layer

# ---------- Caches & helpers ----------
def _build_tile_cache(mapping, palette, stone_px, threads=0, show_frame=True):
    used = np.unique(mapping).tolist()
    cache = {}

    def work(idx):
        color = palette[idx]
        return idx, render_stone_tile(stone_px, color, show_frame=show_frame)

    if threads is None or threads <= 0:
        try:
            threads = max(1, min(8, (os.cpu_count() or 4) - 1))
        except Exception:
            threads = 1

    if threads > 1:
        with ThreadPoolExecutor(max_workers=threads) as ex:
            futs = [ex.submit(work, idx) for idx in used]
            for f in as_completed(futs):
                k, v = f.result()
                cache[k] = v
    else:
        for idx in used:
            k, v = work(idx)
            cache[k] = v
    return cache

def _best_font_size(draw, text, target_px, font_candidates=("arial.ttf",)):
    size = max(10, int(target_px * 0.9))  # a bit larger symbols by default
    for try_size in range(size, 6, -1):
        for name in font_candidates:
            try:
                font = ImageFont.truetype(name, try_size)
                bbox = draw.textbbox((0,0), text, font=font)
                if (bbox[3]-bbox[1]) <= target_px:
                    return font
            except OSError:
                continue
    return ImageFont.load_default()

def _build_glyph_cache_dual(mapping, symbols, stone_px, font):
    used = np.unique(mapping).tolist()
    cache_dark = {}
    cache_light = {}
    dark = (20, 20, 20, 255)
    light = (245, 245, 245, 255)
    for idx in used:
        sym = str(symbols[idx]) if idx < len(symbols) else "?"
        for color, bucket in ((dark, cache_dark), (light, cache_light)):
            tile = Image.new("RGBA", (stone_px, stone_px), (0,0,0,0))
            d = ImageDraw.Draw(tile)
            bbox = d.textbbox((0,0), sym, font=font)
            tw = bbox[2] - bbox[0]
            th = bbox[3] - bbox[1]
            tx = (stone_px - tw) // 2
            ty = (stone_px - th) // 2
            d.text((tx, ty), sym, fill=color, font=font)
            bucket[idx] = tile
    return cache_dark, cache_light

# ---------- Outputs ----------
def generate_preview(mapping, palette, stone_px, dpi, out_path, threads=0):
    h, w = mapping.shape
    canvas_w = w * stone_px
    canvas_h = h * stone_px

    img = Image.new("RGBA", (canvas_w, canvas_h), (245, 245, 245, 255))
    # show_frame=False for clean preview
    cache = _build_tile_cache(mapping, palette, stone_px, threads=threads, show_frame=False)

    for y in range(h):
        y_off = y * stone_px
        for x in range(w):
            idx = int(mapping[y, x])
            tile = cache[idx]
            img.paste(tile, (x*stone_px, y_off), tile)

    out = img.convert("RGB")
    out.save(out_path, dpi=(dpi, dpi))

def generate_scheme(mapping, palette, symbols, stone_px, dpi, out_path):
    """
    Color-filled scheme: each cell filled with the mapped palette color,
    high-contrast symbol over it, and a visible grid.
    """
    h, w = mapping.shape
    img = Image.new("RGBA", (w * stone_px, h * stone_px), (255, 255, 255, 255))
    draw = ImageDraw.Draw(img)

    # Font auto-sizing
    sample_symbol = "?"
    for s in symbols:
        if isinstance(s, str) and s.strip():
            sample_symbol = s.strip()[0]
            break
    font = _best_font_size(draw, sample_symbol, int(stone_px * 0.9))

    glyph_dark, glyph_light = _build_glyph_cache_dual(mapping, symbols, stone_px, font)

    # Fill cells by color + paste glyphs
    for y in range(h):
        y_off = y * stone_px
        for x in range(w):
            px = x * stone_px
            idx = int(mapping[y, x])
            rgb = palette[idx] if idx < len(palette) else (255,255,255)
            draw.rectangle([px, y_off, px + stone_px, y_off + stone_px], fill=rgb, width=0)
            glyph = glyph_dark[idx] if _luma(rgb) > 140 else glyph_light[idx]
            img.paste(glyph, (px, y_off), glyph)

    # Grid lines
    for x in range(w+1):
        x0 = x * stone_px
        draw.line([(x0, 0), (x0, h*stone_px)], fill=GRID_RGBA, width=1)
    for y in range(h+1):
        y0 = y * stone_px
        draw.line([(0, y0), (w*stone_px, y0)], fill=GRID_RGBA, width=1)

    img.convert("RGB").save(out_path, dpi=(dpi, dpi))

def _safe_open_csv(path_base):
    base = Path(path_base)
    stem = base.stem
    suffix = base.suffix or ".csv"
    folder = base.parent

    candidates = [base] + [folder / f"{stem}_{i}{suffix}" for i in range(1, 100)]
    last_err = None
    for p in candidates:
        try:
            f = open(p, "w", newline="", encoding="utf-8")
            return f, str(p)
        except PermissionError as e:
            last_err = e
            continue
    if last_err:
        raise last_err
    raise OSError("Unable to open CSV for writing.")

def write_legend(codes, mapping, symbols, palette, out_csv_path):
    unique, counts = np.unique(mapping, return_counts=True)
    total = mapping.size
    rows = []
    for idx, cnt in zip(unique.tolist(), counts.tolist()):
        code = codes[idx] if idx < len(codes) else f"#{idx:03d}"
        sym = symbols[idx] if idx < len(symbols) else ""
        rgb = palette[idx] if idx < len(palette) else (0,0,0)
        percent = 100.0 * cnt / total
        rows.append([code, sym, cnt, f"{percent:.2f}", rgb[0], rgb[1], rgb[2]])

    f, final_path = _safe_open_csv(out_csv_path)
    with f:
        writer = csv.writer(f, delimiter=";")
        writer.writerow(["Code","Symbol","Count","Percent","R","G","B"])
        writer.writerows(rows)
    return final_path

# ---------- CLI ----------
def main():
    parser = argparse.ArgumentParser(description="Square 5D diamond mosaic generator (styles, preview without legend/frame, color-filled scheme).")
    parser.add_argument('input')
    parser.add_argument('palette')
    parser.add_argument('--stones-x', type=int, required=True)
    parser.add_argument('--stones-y', type=int, required=True)
    parser.add_argument('--stone-size-mm', type=float, required=True)
    parser.add_argument('--dpi', type=int, default=300, help='Fallback DPI for both outputs if specific DPIs not set.')
    parser.add_argument('--preview-dpi', type=int, default=None, help='DPI for preview output only.')
    parser.add_argument('--scheme-dpi', type=int, default=None, help='DPI for scheme output only.')
    parser.add_argument('--mode', choices=['preview','scheme','both'], default='both')
    parser.add_argument('--legend', action='store_true', help='Write CSV legend (no preview legend panel).')
    parser.add_argument('--threads', type=int, default=0, help='Threads for pre-rendering tiles (0=auto)')
    parser.add_argument('--style', type=str, default='default',
                        help="Preset: default | contrast | soft | glossy-dark (русские: 'контрастный', 'мягкий', 'глянец')")
    args = parser.parse_args()

    if not os.path.isfile(args.input) or not os.path.isfile(args.palette):
        sys.exit('Missing files')

    # Apply style preset
    apply_style(args.style)

    # Resolve per-output DPI
    preview_dpi = args.preview_dpi if args.preview_dpi is not None else args.dpi
    scheme_dpi  = args.scheme_dpi  if args.scheme_dpi  is not None else args.dpi

    # Per-output stone pixel sizes
    preview_stone_px = round(args.stone_size_mm/25.4 * preview_dpi)
    scheme_stone_px  = round(args.stone_size_mm/25.4 * scheme_dpi)

    codes, palette, symbols = load_palette(args.palette)

    img = Image.open(args.input).convert('RGB')
    small = img.resize((args.stones_x, args.stones_y), Image.LANCZOS)
    mapping = map_to_palette(small, palette)
    base = Path(args.input).stem

    if args.mode in ('preview','both'):
        generate_preview(mapping, palette, preview_stone_px, preview_dpi, f"{base}_preview.png", threads=args.threads)

    if args.mode in ('scheme','both'):
        generate_scheme(mapping, palette, symbols, scheme_stone_px, scheme_dpi, f"{base}_scheme.png")

    if args.legend:
        legend_path = write_legend(codes, mapping, symbols, palette, f"{base}_legend.csv")
        print(f"Legend saved to: {legend_path}")

if __name__ == '__main__':
    main()

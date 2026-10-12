import csv, json, re, urllib.parse, urllib.request
from pathlib import Path

SOURCE = Path('/home/javi/.hermes/cache/scratch/venao-checkpoints.csv')
OUT = Path(__file__).with_name('index.html')
TEMPLATE = Path(__file__).with_name('template.html')


def coords(url):
    if not url: return None
    text = urllib.parse.unquote(url)
    patterns = [r'[?&]q=(-?\d+\.\d+),(-?\d+\.\d+)', r'@(-?\d+\.\d+),(-?\d+\.\d+)', r'!3d(-?\d+\.\d+)!4d(-?\d+\.\d+)']
    for pattern in patterns:
        m = re.search(pattern, text)
        if m:
            lat, lon = map(float, m.groups())
            if 6 < lat < 10 and -83 < lon < -78: return [lat, lon]
    if 'maps.app.goo.gl' in url:
        try:
            req = urllib.request.Request(url, headers={'User-Agent':'Mozilla/5.0'})
            with urllib.request.urlopen(req, timeout=7) as response:
                return coords(response.url)
        except Exception: pass
    return None

with SOURCE.open(newline='', encoding='utf-8-sig') as f:
    rows = list(csv.reader(f))[1:]
records = []
for n, row in enumerate(rows, 1):
    row += [''] * (26-len(row))
    if not (row[1].strip() or row[2].strip() or row[9].strip()): continue
    if not (row[2].strip() or row[11].strip()): continue
    code = row[2].strip() or f'S/C-{len([r for r in records if r["code"].startswith("S/C-")])+1}'
    records.append({'code':code,'area':row[3].strip() or 'Sin ubicación',
        'place':row[5].strip(),'directions':row[7].strip(),
        'needs':row[1].strip(),'notes':row[6].strip(), 'extra':row[25].strip(),
        'people':row[11].strip(),'children':row[13].strip(),'babies':row[14].strip(),
        'service':row[21].strip(),'serviceNotes':row[22].strip(),
        'gps':row[24].strip(),'coords':coords(row[24].strip())})
html = TEMPLATE.read_text().replace('__CHECKPOINT_DATA__', json.dumps(records, ensure_ascii=False).replace('</', '<\\/'))
OUT.write_text(html, encoding='utf-8')
print(f'{len(records)} points, {sum(bool(r["coords"]) for r in records)} mapped, {sum(not r["coords"] for r in records)} unmapped -> {OUT}')

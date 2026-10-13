"""Combine the current public template with the previously approved snapshot."""

import json
import re
import sys
from pathlib import Path


PRIVATE_FIELDS = {'contact', 'phone', 'secondContact', 'secondPhone', 'thirdContact', 'thirdPhone'}
SNAPSHOT_LINE = re.compile(r'^const snapshot=(\[.*\]);$', re.MULTILINE)


def render(template: str, approved_html: str) -> str:
    matches = SNAPSHOT_LINE.findall(approved_html)
    if len(matches) != 1 or template.count('__CHECKPOINT_DATA__') != 1:
        raise ValueError('Expected one approved snapshot and one template placeholder')
    records = json.loads(matches[0])
    if any(not isinstance(record, dict) or PRIVATE_FIELDS.intersection(record) for record in records):
        raise ValueError('Approved snapshot contains private contact fields')
    data = json.dumps(records, ensure_ascii=False).replace('</', '<\\/')
    return template.replace('__CHECKPOINT_DATA__', data)


if __name__ == '__main__':
    if len(sys.argv) != 4:
        raise SystemExit('Usage: render_index.py TEMPLATE APPROVED_HTML OUTPUT')
    template, approved, output = map(Path, sys.argv[1:])
    output.write_text(render(template.read_text(encoding='utf-8'),
                             approved.read_text(encoding='utf-8')), encoding='utf-8')

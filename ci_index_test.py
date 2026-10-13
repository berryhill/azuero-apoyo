from pathlib import Path
import unittest

ROOT = Path(__file__).parent

class CIIndexTests(unittest.TestCase):
    def test_checkout_build_uses_tracked_index_without_secret_or_generation(self):
        workflow = (ROOT / '.github/workflows/ci.yml').read_text()
        self.assertNotIn('INDEX_HTML_B64', workflow)
        self.assertNotIn('render_index.py', workflow)
        self.assertIn('test -s index.html', workflow)
        page = (ROOT / 'index.html').read_text()
        self.assertNotIn('__CHECKPOINT_DATA__', page)
        self.assertFalse((ROOT / 'template.html').exists())
        self.assertFalse((ROOT / 'build_data.py').exists())
        self.assertFalse((ROOT / 'render_index.py').exists())

if __name__ == '__main__':
    unittest.main()

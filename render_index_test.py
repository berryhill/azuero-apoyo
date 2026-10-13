import json
import unittest

from render_index import render


class RenderIndexTest(unittest.TestCase):
    def test_uses_current_template_and_approved_data(self):
        old = 'old UI\nconst snapshot=[{"code":"A","place":"</script>"}];\n'
        current = 'new UI\nconst snapshot=__CHECKPOINT_DATA__;\n'
        result = render(current, old)
        self.assertTrue(result.startswith('new UI'))
        self.assertNotIn('old UI', result)
        self.assertIn('<\\/script>', result)
        self.assertEqual(json.loads(result.split('const snapshot=')[1].split(';')[0]),
                         [{'code': 'A', 'place': '</script>'}])

    def test_rejects_missing_snapshot_or_private_contacts(self):
        template = 'const snapshot=__CHECKPOINT_DATA__;'
        with self.assertRaises(ValueError):
            render(template, 'const snapshot=[];\nconst snapshot=[];')
        with self.assertRaises(ValueError):
            render(template, 'const snapshot=[{"phone":"65551234"}];')
        with self.assertRaises(ValueError):
            render('no placeholder', 'const snapshot=[];')


if __name__ == '__main__':
    unittest.main()

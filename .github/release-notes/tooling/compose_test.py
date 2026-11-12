import tempfile
import unittest
from pathlib import Path

from compose import END_MARKER, START_MARKER, insert_container, prepare_release, render_container


class ComposeTest(unittest.TestCase):
    def test_prepare_release_with_curated_notes_and_container(self):
        with tempfile.TemporaryDirectory() as directory:
            notes = Path(directory) / "0.2.0.md"
            notes.write_text("## Upgrade notes\n")

            result = prepare_release(notes, True)

        self.assertEqual(
            result,
            f"## Upgrade notes\n\n{START_MARKER}\n{END_MARKER}\n",
        )

    def test_prepare_release_without_curated_notes(self):
        result = prepare_release(Path("missing.md"), False)

        self.assertEqual(result, "")

    def test_render_container_with_digest(self):
        template = "Image {{VERSION}}\n\n{{DIGEST}}\n"

        result = render_container(template, "0.2.0", "sha256:abc")

        self.assertEqual(result, "Image 0.2.0\n\nDigest: `sha256:abc`\n")

    def test_render_container_without_digest(self):
        template = "Image {{VERSION}}\n\n{{DIGEST}}\n"

        result = render_container(template, "0.2.0", "")

        self.assertEqual(result, "Image 0.2.0\n")

    def test_insert_container_is_idempotent(self):
        release = f"## Upgrade\n\n{START_MARKER}\n{END_MARKER}\n\n## What's Changed\n"
        container = "### Container image\n"

        first = insert_container(release, container)
        second = insert_container(first, container)

        self.assertEqual(first, second)
        self.assertEqual(first.count("### Container image"), 1)

    def test_insert_container_requires_one_marker_pair(self):
        with self.assertRaisesRegex(ValueError, "exactly one container marker pair"):
            insert_container("release notes", "container notes")


if __name__ == "__main__":
    unittest.main()

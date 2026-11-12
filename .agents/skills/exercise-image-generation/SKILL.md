---
name: exercise-image-generation
description: Create and review consistent male and female instructional exercise images for Simple Workout. Use when adding exercises, preparing image-generation prompts, reviewing generated exercise artwork, or processing approved image assets.
license: MIT
---

# Exercise image generation

Read `references/style-guide.md` and `references/review-checklist.md` before preparing or reviewing an image.

## Workflow

1. Read the exercise definition in the application catalog and confirm the technically correct start and finish positions.
2. Select the canonical model reference from `tools/exercise-images/models/`.
3. Prepare one generation prompt using the application guidance and requirements in the style guide.
4. Ask the user to generate the image externally and save it as `reference/exercises/<slug>-<variant>.png`.
5. Inspect the result against every item in the review checklist. Reject technically incorrect form, anatomy, equipment, or framing even when the image looks polished.
6. Run `tools/exercise-images/process.sh` to validate and convert approved source images.
7. Inspect the generated WebP assets under `assets/static/images/exercises/`.
8. Remove temporary source files from `reference/` after the approved WebP assets are preserved.

Do not silently accept uncertain exercise form. Call out the specific concern and request a correction or user decision.

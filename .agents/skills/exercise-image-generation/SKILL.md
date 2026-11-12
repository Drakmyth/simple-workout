---
name: exercise-image-generation
description: Create and review consistent male and female instructional exercise images for Simple Workout. Use when adding exercises, preparing image-generation prompts, reviewing generated exercise artwork, or processing approved image assets.
license: MIT
---

# Exercise image generation

Read `references/style-guide.md` and `references/review-checklist.md` before preparing or reviewing an image.

## Workflow

1. Confirm the exercise variant and technically correct start and finish positions.
2. Select the canonical model reference from `tools/exercise-images/assets/models/`.
3. Prepare one generation prompt using the requirements in the style guide.
4. Ask the user to generate the image externally and save it as `reference/exercises/<slug>-<variant>.png`.
5. Inspect the result against every item in the review checklist. Reject technically incorrect form, anatomy, equipment, or framing even when the image looks polished.
6. Record the approved prompt and generation metadata using `assets/manifest-template.json`.
7. Run `tools/exercise-images/process.sh` to produce application assets.
8. Remove temporary source files from `reference/` after approved assets and metadata are preserved.

Do not silently accept uncertain exercise form. Call out the specific concern and request a correction or user decision.

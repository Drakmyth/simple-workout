# Exercise image assets

`models/` contains the canonical synthetic model references used when generating exercise images. `requirements.json` records the reusable exercise-specific pose guidance.

Generate a reviewed 1536 x 1024 PNG for each `<exercise>-<variant>` pair in `reference/exercises/`, then run:

```sh
./scripts/process-exercise-images.sh
```

The pinned ImageMagick container validates dimensions, removes metadata, and writes optimized composite WebP application assets. Source PNGs under `reference/` are temporary and must not be committed.

See `.agents/skills/exercise-image-generation/` for the complete generation and review workflow.

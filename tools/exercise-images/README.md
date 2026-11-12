# Exercise image tooling

`models/` contains the canonical synthetic model references used when generating exercise images.

Generate and review a 1536 x 1024 PNG for each `<exercise>-<variant>` pair, save the approved sources under `reference/exercises/`, then run:

```sh
./tools/exercise-images/process.sh
```

The pinned ImageMagick container validates dimensions, removes metadata, and writes optimized WebP assets under `assets/static/images/exercises/`. Existing assets are replaced only when a matching source PNG is present. Source PNGs under `reference/` are temporary and must not be committed.

See `.agents/skills/exercise-image-generation/` for the complete generation and review workflow.

# Simple Workout

Simple Workout is a mobile-first progressive web app for defining and playing guided workouts. It uses a Go backend, server-rendered HTML, plain JavaScript, and SQLite.

## Features

- Exercise catalog with male and female movement guides
- Renameable, reorderable workouts with editable sets, reps, and weights
- Guided set and rest playback with optional JavaScript timers
- Per-exercise male, female, alternating, or random playback images
- Optional working-weight updates after workout completion
- Optional exercise performance notes and a transient completion summary
- Installable PWA that remains network-dependent

## Requirements

- Go 1.26 or newer
- Docker with Docker Compose (optional)

## Development

Run the server:

```sh
go run ./cmd/simple-workout
```

Then open <http://localhost:8080>. The health endpoint is available at <http://localhost:8080/health>.

Run validation:

```sh
go test ./...
go vet ./...
go build ./cmd/simple-workout
```

Test progressive playback enhancements with Node.js (no dependencies required):

```sh
node --test assets/tests/*.test.cjs
```

## Docker

### Source builds

Build and run the application from a source checkout:

```sh
docker compose up --build
```

By default, the app is available at <http://localhost:8080>. Copy `.env.example` to `.env` to change the bind address or port. Set `HTTP_BIND_ADDRESS=0.0.0.0` to allow direct access from other LAN devices; the app does not yet provide authentication.

Runtime data is stored in the `workout-data` volume.

### Release packages

Automated releases attach a minimal source deployment package named `simple-workout-<version>.zip`. Download and extract it, then run:

```sh
cp .env.example .env
docker compose up -d --build
```

The package contains only the application source and files required for Docker deployment. Automated releases also publish AMD64 and ARM64 images to `ghcr.io/drakmyth/simple-workout` for deployments that use a container registry.

### PWA installation

PWA installation requires a secure context. Localhost works during development; LAN devices should access the app through an HTTPS reverse proxy. The service worker intentionally does not cache application responses, so the installed app still requires network access.

Before a release, verify the deployed PWA on a target device:

1. Open the HTTPS URL and install the app.
2. Confirm the app launches in a standalone window with the correct icon.
3. Complete a workout and confirm the rest timer, Next actions, and optional weight saving.
4. Restart the container and confirm workout changes persist.
5. Disconnect the network and confirm the expected network-dependent failure.

### Reverse proxies

The application serves plain HTTP on container port 8080 and does not manage TLS. A reverse proxy can reach it through the published host port or by sharing its Docker network. Configure the proxy to forward requests without changing their path and to provide HTTPS externally.

When the proxy shares the Docker network, it should route to service `app` on port 8080. The published host port can remain bound to `127.0.0.1` for local diagnostics.

## Releases

[`VERSION`](VERSION) is the authoritative release version. Development builds report version `dev`; the release workflow injects `VERSION` into the application binary and container metadata.

After a version change reaches `master`, run the **Package Release Publish** workflow from GitHub Actions. Simulation mode validates and packages the release and builds both container architectures without creating a tag, pushing an image, or publishing a GitHub release.

## Known limitations

- The application is single-user and does not provide authentication; deploy it only on a trusted network or behind external access controls.
- It must be hosted at the origin root and does not support a configurable URL base path.
- The installed PWA requires a network connection.
- Active workouts cannot be resumed after closing the page. Performance notes and completion summaries are not stored as workout history.

## License

[MIT](LICENSE)

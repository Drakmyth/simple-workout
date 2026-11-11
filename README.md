# Simple Workout

Simple Workout is a mobile-first progressive web app for defining and playing guided workouts. It uses a Go backend, server-rendered HTML, plain JavaScript, and SQLite.

## Features

- Exercise catalog with male and female movement guides
- Renameable, reorderable workouts with editable sets, reps, and weights
- Guided set and rest playback with optional JavaScript timers
- Per-exercise male, female, alternating, or random playback images
- Optional working-weight updates after workout completion
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

## Docker

Build and run the application:

```sh
docker compose up --build
```

By default, the app is available at <http://localhost:8080>. Copy `.env.example` to `.env` to change the bind address or port. Set `HTTP_BIND_ADDRESS=0.0.0.0` to allow direct access from other LAN devices; the app does not yet provide authentication.

Runtime data is stored in the `workout-data` volume.

### PWA installation

PWA installation requires a secure context. Localhost works during development; LAN devices should access the app through an HTTPS reverse proxy. The service worker intentionally does not cache application responses, so the installed app still requires network access.

### Reverse proxies

The application serves plain HTTP on container port 8080 and does not manage TLS. A reverse proxy can reach it through the published host port or by sharing its Docker network. Configure the proxy to forward requests without changing their path and to provide HTTPS externally.

When the proxy shares the Docker network, it should route to service `app` on port 8080. The published host port can remain bound to `127.0.0.1` for local diagnostics.

## License

[MIT](LICENSE)

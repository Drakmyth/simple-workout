# Simple Workout

Simple Workout is a mobile-first progressive web app for defining and playing guided workouts. The planned application uses a Go backend, server-rendered HTML, plain JavaScript, HTMX, and SQLite.

## Status

Initial project scaffolding. Exercise catalog, workout editing, playback, and PWA support are forthcoming.

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

### Reverse proxies

The application serves plain HTTP on container port 8080 and does not manage TLS. A reverse proxy can reach it through the published host port or by sharing its Docker network. Configure the proxy to forward requests without changing their path and to provide HTTPS externally.

When the proxy shares the Docker network, it should route to service `app` on port 8080. The published host port can remain bound to `127.0.0.1` for local diagnostics.

## License

[MIT](LICENSE)

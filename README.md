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

### Traefik

Set `APP_HOST` and, if necessary, `TRAEFIK_NETWORK` and `TRAEFIK_ENTRYPOINT` in `.env`. The named Traefik network must already exist. Start the app with the Traefik overlay:

```sh
docker compose -f compose.yaml -f compose.traefik.yaml up --build -d
```

Traefik routes HTTPS traffic to the app's internal port 8080. The direct localhost port remains available for diagnostics. Certificate resolver configuration remains owned by the existing Traefik deployment.

## License

[MIT](LICENSE)

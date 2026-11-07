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

The container listens on port 8080 within its Docker network. Production access is expected to use an HTTPS reverse proxy such as Traefik; publish the port locally only when direct access is required.

Runtime data is stored in the `workout-data` volume.

## License

[MIT](LICENSE)

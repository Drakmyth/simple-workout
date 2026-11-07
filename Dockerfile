# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/simple-workout ./cmd/simple-workout

FROM alpine:3.23
RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app
COPY --from=build /out/simple-workout /usr/local/bin/simple-workout
USER app
EXPOSE 8080
ENTRYPOINT ["simple-workout"]

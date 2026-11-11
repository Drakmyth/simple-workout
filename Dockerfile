# syntax=docker/dockerfile:1
ARG VERSION=dev
ARG VCS_REF=unknown

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
ARG TARGETOS
ARG TARGETARCH
ARG VERSION
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath \
    -ldflags="-s -w -X main.version=$VERSION" \
    -o /out/simple-workout ./cmd/simple-workout

FROM alpine:3.23
ARG VERSION
ARG VCS_REF
LABEL org.opencontainers.image.title="Simple Workout" \
    org.opencontainers.image.description="Mobile-first guided workout PWA" \
    org.opencontainers.image.source="https://github.com/Drakmyth/simple-workout" \
    org.opencontainers.image.revision=$VCS_REF \
    org.opencontainers.image.version=$VERSION \
    org.opencontainers.image.licenses="MIT"
RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app \
    && install -d -o app -g app /data
COPY --from=build /out/simple-workout /usr/local/bin/simple-workout
VOLUME ["/data"]
USER app
EXPOSE 8080
ENTRYPOINT ["simple-workout"]

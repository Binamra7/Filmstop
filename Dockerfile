# syntax=docker/dockerfile:1

FROM golang:1.27.2-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/filmstop . && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/seed ./cmd/seed
# The runtime image runs as uid 65532 and has no shell, so make the data dir here.
RUN mkdir -p /out/data

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/filmstop /out/seed /
COPY --from=build --chown=65532:65532 /out/data /data
ENV DATABASE_PATH=/data/filmstop.db PORT=8080
VOLUME /data
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s CMD ["/filmstop", "-healthcheck"]
ENTRYPOINT ["/filmstop"]

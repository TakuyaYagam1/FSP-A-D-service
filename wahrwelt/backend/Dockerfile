FROM golang:1.27.3-alpine@sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb AS build

WORKDIR /src/backend
ENV GOMAXPROCS=4
ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/root/.cache/go-build
COPY go.mod go.sum ./
RUN --mount=type=cache,id=wahrwelt-modules-gomod,target=/go/pkg/mod go mod download
COPY cmd ./cmd
COPY config ./config
COPY db ./db
COPY internal ./internal
RUN --mount=type=cache,id=wahrwelt-modules-go-build,target=/root/.cache/go-build \
    --mount=type=cache,id=wahrwelt-modules-gomod,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -mod=readonly -buildvcs=false -trimpath -ldflags="-s -w" -o /out/wahrwelt ./cmd/wahrwelt

FROM alpine:3.24.3@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a

RUN apk add --no-cache ca-certificates && addgroup -S -g 10001 app && adduser -S -D -H -u 10001 -G app app && mkdir -p /var/lib/wahrwelt/archive /var/lib/wahrwelt/preview && chown -R 10001:10001 /var/lib/wahrwelt
WORKDIR /app
COPY --from=build /out/wahrwelt /usr/local/bin/wahrwelt
USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/wahrwelt"]

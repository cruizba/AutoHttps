FROM golang:1.25.3 AS builder

ARG TARGETOS
ARG TARGETPLATFORM
ARG TARGETARCH

WORKDIR /workspace
COPY . .
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build

FROM caddy/caddy:2.10.2-alpine

COPY --from=builder /workspace/caddyfile-generate /usr/bin/caddyfile-generate
COPY --from=builder /workspace/entrypoint.sh /entrypoint.sh
RUN chmod +x /usr/bin/caddyfile-generate /entrypoint.sh

# Run the binary.
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/usr/bin/caddy", "run", "--config", "/config/caddy/Caddyfile"]
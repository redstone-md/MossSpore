# MossSpore container image.
#
# Self-contained: the `moss` dependency is fetched from
# github.com/redstone-md/moss (pinned in go.mod / go.sum), so the build context
# is just THIS repo — no sibling checkout needed:
#
#   docker build -t mossspore .
#
# ---- builder ----
FROM golang:1.25 AS builder
WORKDIR /src
# Prime the module cache first so it is reused across source-only rebuilds.
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /mossspore .

# ---- runtime ----
FROM gcr.io/distroless/static-debian12
COPY --from=builder /mossspore /mossspore
COPY docker/config.json /etc/mossspore/config.json

# Peer port: moss binds ONE port number for both the TCP transport and the
# UDP transport (hole-punch / uTP). 4001 is this image's default, baked into
# docker/config.json; override listen_port by mounting a different config over
# /etc/mossspore/config.json if you need another port.
EXPOSE 4001/tcp
EXPOSE 4001/udp
# Monitor HTTP endpoint: /health, /info, /version.
EXPOSE 9800/tcp

# Persistent identity (see README.md "Identity Persistence"). Mount a volume
# here so the spore keeps the same peer id / SuperNode identity across
# container restarts.
VOLUME ["/var/lib/mossspore"]

ENTRYPOINT ["/mossspore", "--config", "/etc/mossspore/config.json"]

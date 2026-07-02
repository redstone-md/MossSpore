# MossSpore container image.
#
# BUILD CONTEXT MUST BE THE PARENT DIRECTORY, not MossSpore/ itself.
# MossSpore/go.mod has `replace moss => ../moss`: the moss module lives one
# level up, as a sibling of this repo. `docker build MossSpore/` only sees
# inside MossSpore/ and cannot resolve ../moss, so the build fails. Build
# from the directory that contains both MossSpore/ and moss/:
#
#   docker build -f MossSpore/Dockerfile -t mossspore ..
#
# MossSpore/Dockerfile.dockerignore (matched by BuildKit against this exact
# Dockerfile path) trims that parent-dir context down to just MossSpore/ and
# moss/ — everything else up there (other repos, editor/AI-tool state) is
# unrelated to this build.

# ---- builder ----
FROM golang:1.25 AS builder
WORKDIR /src
COPY MossSpore/ ./MossSpore/
COPY moss/ ./moss/
WORKDIR /src/MossSpore
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /mossspore ./cmd/mossspore

# ---- runtime ----
FROM gcr.io/distroless/static-debian12
COPY --from=builder /mossspore /mossspore
COPY MossSpore/docker/config.json /etc/mossspore/config.json

# Peer port: moss binds ONE port number for both the TCP transport and the
# UDP transport (hole-punch / uTP) — see moss/internal/transport/listener.go
# (ListenPair). 4001 is this image's default, baked into docker/config.json;
# override listen_port by mounting a different config over
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

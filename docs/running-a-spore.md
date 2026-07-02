# Running a spore, in 2 minutes

MossSpore joins the shared relay mesh (`moss-relay/1`) by default and tries
to become a **SuperNode** — a relay other peers can route through. That only
works if its **peer port is reachable from the internet, inbound, on both
UDP and TCP** (moss binds one port number for both transports). If it's
behind a NAT/firewall that doesn't forward that port, the spore still relays
fine for itself, but it can never be promoted to SuperNode — check this with
the `/health` and `/info` calls at the end of each path below.

Pick one:

- [(a) VPS — one-line install](#a-vps--one-line-install)
- [(b) Docker](#b-docker)
- [(c) Fly.io](#c-flyio)

---

## (a) VPS — one-line install

```bash
curl -sSL https://raw.githubusercontent.com/redstone-md/MossSpore/main/install.sh | sh
sudo systemctl start mossspore
```

This installs the binary, writes `/etc/mossspore/config.json` (relay-mesh
mode on, persistent identity at `/var/lib/mossspore/identity.key`), and
sets up a systemd service. The generated config uses an OS-assigned
(ephemeral) peer port by default.

**If your provider's firewall/security-group defaults to deny-all inbound**
(common on AWS/GCP; most plain VPS boxes like Hetzner/DigitalOcean/Vultr
have no inbound firewall by default), pin the peer port to something you can
open before starting the service. Edit `/etc/mossspore/config.json`:

```json
{ "listen_port": 4001, "...": "..." }
```

Then open UDP **and** TCP `4001` in your firewall/security group and:

```bash
sudo systemctl restart mossspore
```

Verify:

```bash
curl <host>:9800/health
# {"status":"ok","nat_type":"public"}   <- "public" (or a cone type) means reachable

curl <host>:9800/info
# supernode_ready flips to true within relay.min_uptime_sec (60s default)
# after the node confirms it's reachable.
```

---

## (b) Docker

Build context is the PARENT directory of this repo — see the comment at the
top of [`../Dockerfile`](../Dockerfile) for why (`go.mod` has
`replace moss => ../moss`, a sibling directory outside `MossSpore/`):

```bash
# from the directory that contains both MossSpore/ and moss/
docker build -f MossSpore/Dockerfile -t mossspore ..
```

Run it, mapping the peer port (UDP **and** TCP) and the monitor port, with a
volume for the persistent identity:

```bash
docker run -d --name mossspore \
  -p 4001:4001/tcp -p 4001:4001/udp \
  -p 9800:9800/tcp \
  -v mossspore_data:/var/lib/mossspore \
  --restart=always \
  mossspore
```

The image ships with a default `moss-relay/1` config baked in at
`/etc/mossspore/config.json` (see [`../docker/config.json`](../docker/config.json)).
To use your own (different mesh, PSK, static peers, etc.), bind-mount over
that path instead: `-v ./my-config.json:/etc/mossspore/config.json:ro`.

Verify:

```bash
curl <host>:9800/health
# {"status":"ok","nat_type":"public"}

curl <host>:9800/info
# supernode_ready: true once reachability is confirmed (~60s after start)
```

`<host>` is `localhost` if you're checking from the Docker host itself, or
the VPS's public address if the host has its own firewall to open 4001/tcp
and 4001/udp through.

---

## (c) Fly.io

Fly gives every app a public IP, so reachability is handled for you once the
dedicated IPv4 below is allocated. Deploy from the same PARENT directory as
the Docker build (`../fly.toml` and `../Dockerfile` both need `../moss` in
context — see the comment block at the top of
[`../fly.toml`](../fly.toml)):

```bash
# from the directory that contains both MossSpore/ and moss/
fly apps create <your-app-name>
fly volumes create mossspore_data --app <your-app-name> --region <region> --size 1
fly ips allocate-v4 --app <your-app-name>
fly deploy . --config MossSpore/fly.toml --dockerfile MossSpore/Dockerfile --app <your-app-name>
```

The `fly ips allocate-v4` step matters: Fly's default shared/anycast address
only proxies the `[http_service]` block (the monitor on `:9800`); the raw
peer TCP/UDP port needs its own dedicated IPv4, which `fly.toml`'s
`[[services]]` blocks bind to.

Verify:

```bash
curl https://<your-app-name>.fly.dev/health
# {"status":"ok","nat_type":"public"}

curl https://<your-app-name>.fly.dev/info
# supernode_ready: true once reachability is confirmed (~60s after start)
```

---

## Collecting the peer address

Once a spore's `/info` shows `supernode_ready: true`, its `host:port` peer
address (from `/info`'s `advertised_addr`) is what gets added to a mosh
client's bundled relay-spore bootstrap list. See the main
[README.md](../README.md#relay-mesh-mode) for the relay-mesh overview.

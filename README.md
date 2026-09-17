# Fizz-Buzz REST server

A small, production-ready HTTP service exposing a generalised fizz-buzz, plus a
statistics endpoint reporting the most frequently requested parameter set.

Written in Go on the chi router, with a cobra CLI, viper-based `.env`
configuration and a lifecycle (health gate, errgroup, graceful shutdown) laid
out the way a service of this shape usually is.

---

## Quick start

```bash
make run                      # go run . serve, listens on :8080
curl "http://localhost:8080/api/v1/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz"
```

`make` on its own lists every target: `run`, `build`, `dist`, `fmt`, `lint`,
`check`, `image`, `up`, `env`, `clean`.

```json
{"count":15,"result":["1","2","fizz","4","buzz","fizz","7","8","fizz","buzz","11","fizz","13","14","fizzbuzz"]}
```

With Docker — a container image is Linux only, so only the architecture varies:

```bash
make image                    # docker build -t fizzbuzz .
make up                       # runs it on :8080, or PORT=9000 make up

# Another Linux architecture, cross-compiled on the build host, no emulation:
docker buildx build --platform linux/arm64 -t fizzbuzz:arm64 .

# Both at once needs a container driver, since the default one is single-platform:
docker buildx create --name multi --driver docker-container --use
docker buildx build --platform linux/amd64,linux/arm64 -t fizzbuzz:multi .
```

### Configuring the container

The image only sets `ENVIRONMENT=production`, so the logs come out as JSON, and
`APP_PORT=8080` to match the exposed port. Everything else keeps the default it
has in the code, and `.env` files are kept out of the image by `.dockerignore`.

```bash
# One setting at a time
docker run --rm -p 8080:8080 -e FIZZBUZZ_MAX_LIMIT=50 -e APP_LOG_LEVEL=warn fizzbuzz

# A file of them
docker run --rm -p 8080:8080 --env-file .env fizzbuzz

# Or mount a .env; the working directory is /app, which is where it is read from
docker run --rm -p 8080:8080 -v "$PWD/.env:/app/.env:ro" fizzbuzz
```

`-e` and `--env-file` become real environment variables, so they win over a
mounted `.env`, which itself loses to `.env.local`. A malformed value stops the
container at start-up with the name of the offending variable, rather than
letting it run with a wrong setting.

Requires Go 1.22+ (the version pinned in `go.mod`; `log/slog` and `context.WithoutCancel` need 1.21 at least).

### Commands

```bash
make build                    # into bin/fizzbuzz

bin/fizzbuzz serve [--port 8080]
bin/fizzbuzz --help
```

`--verbose` on any command forces the log level to `debug`.

### Building for macOS, Linux and Windows

Docker covers Linux; the other platforms are plain cross-compilations. CGO is
off, so every binary is static and builds from any host with no extra toolchain.
`make dist` builds all seven at once into `dist/`; one at a time:

```bash
# macOS, Apple silicon
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 \
	go build -trimpath -ldflags='-s -w' -o dist/fizzbuzz_darwin_arm64 .

# Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
	go build -trimpath -ldflags='-s -w' -o dist/fizzbuzz_windows_amd64.exe .

# Linux on 32-bit ARM, a Raspberry Pi for instance
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 \
	go build -trimpath -ldflags='-s -w' -o dist/fizzbuzz_linux_arm .
```

| `GOOS`    | `GOARCH`                        |
|-----------|---------------------------------|
| `darwin`  | `amd64` (Intel), `arm64` (Apple silicon) |
| `linux`   | `amd64`, `arm64`, `arm` (with `GOARM=7`) |
| `windows` | `amd64`, `arm64`                |

`go tool dist list` prints every combination Go supports.

---

## API

Base path: `/api/v1`. Every response — success or failure — is JSON.

### `GET /api/v1/fizzbuzz`

Returns the numbers from 1 to `limit`, where multiples of `int1` are replaced by
`str1`, multiples of `int2` by `str2`, and multiples of both by `str1str2`.

| Parameter | Type    | Constraints                          |
|-----------|---------|--------------------------------------|
| `int1`    | integer | required, > 0                        |
| `int2`    | integer | required, > 0                        |
| `limit`   | integer | required, >= 0 and <= `FIZZBUZZ_MAX_LIMIT` |
| `str1`    | string  | required, 1 to 100 characters        |
| `str2`    | string  | required, 1 to 100 characters        |

All five parameters are mandatory; there is no implicit default, so a request
always says exactly what it wants.

```bash
curl "http://localhost:8080/api/v1/fizzbuzz?int1=3&int2=5&limit=5&str1=fizz&str2=buzz"
```
```json
{"count":5,"result":["1","2","fizz","4","buzz"]}
```

`200 OK`. `count` is the length of `result`, handy for clients that stream or
paginate. A `limit` of `0` returns `{"count":0,"result":[]}` — an empty array,
never `null`.

### `GET /api/v1/stats`

Takes no parameter. Returns the most frequently served fizz-buzz request and its
number of hits.

```bash
curl "http://localhost:8080/api/v1/stats"
```
```json
{"hits":42,"request":{"int1":3,"int2":5,"limit":100,"str1":"fizz","str2":"buzz"}}
```

Before any request has been served:

```json
{"hits":0,"request":null}
```

Semantics worth knowing:

- Only **successfully served** requests are counted, so the endpoint reports
  parameters that actually work rather than client mistakes.
- Two requests are "the same" when the five parameters are equal. `limit=10` and
  `limit=11` are two distinct requests.
- On a tie, the parameter set that reached the count **first** is reported, which
  keeps the answer stable instead of flapping between equals.

### `GET /health`

Liveness/readiness probe for orchestrators. Returns `{"status":"ok"}`.

### Errors

```json
{"error":{"code":"invalid_parameter","message":"must be a positive integer","field":"int1"}}
```

| Status | `code`               | When                                              |
|--------|----------------------|---------------------------------------------------|
| 400    | `missing_parameter`  | a required parameter is absent                    |
| 400    | `invalid_parameter`  | a parameter is present but unusable                |
| 404    | `not_found`          | unknown endpoint                                   |
| 405    | `method_not_allowed` | known endpoint, wrong method (`Allow` is set)      |
| 429    | `rate_limited`       | too many requests (`Retry-After` is set)           |
| 503    | `unavailable`        | the health check is failing                        |
| 500    | `internal_error`     | unexpected failure — details stay in the logs      |

`code` is stable and meant to be branched on; `message` is for humans; `field`
points at the offending parameter. Internal error details are never echoed to
the client.

Validation is evaluated field by field in a fixed order (`int1`, `int2`,
`limit`, `str1`, `str2`), so a given bad request always returns the same error.

---

## Configuration

Configuration comes from the environment, from `.env` (committed defaults) and
from `.env.local` (developer overrides), in increasing order of precedence, with
the real environment winning over both. Every setting has a default, so the
server runs with no `.env` at all; `make env` copies `.env.example` to `.env`.

A malformed value aborts start-up rather than being silently ignored.

| Variable                  | Default  | Description                                          |
|---------------------------|----------|------------------------------------------------------|
| `ENVIRONMENT`             | —        | `development`, `staging` or `production`              |
| `APP_PORT`                | `8080`   | Listen port                                           |
| `APP_READ_HEADER_TIMEOUT` | `5s`     | Header read timeout                                   |
| `APP_READ_TIMEOUT`        | `10s`    | Full request read timeout                             |
| `APP_WRITE_TIMEOUT`       | `30s`    | Response write timeout                                |
| `APP_IDLE_TIMEOUT`        | `60s`    | Keep-alive idle timeout                               |
| `APP_SHUTDOWN_TIMEOUT`    | `10s`    | Grace period for in-flight requests on SIGTERM        |
| `APP_REQUEST_TIMEOUT`     | `20s`    | Per-request deadline; must be under the write timeout |
| `APP_LOG_LEVEL`           | `info`   | `debug`, `info`, `warn` or `error`                    |
| `APP_CORS`                | —        | Allowed browser origins, comma separated; empty means no CORS at all |
| `APP_TRUST_PROXY`         | `false`  | Parse `X-Forwarded-For` (only behind a trusted proxy) |
| `APP_RATE_LIMIT`          | `100`    | Requests per IP and per window; `0` disables it       |
| `APP_RATE_WINDOW`         | `1m`     | Rate-limiting window                                  |
| `FIZZBUZZ_MAX_LIMIT`      | `100000` | Maximum accepted `limit`; must be positive            |
| `FIZZBUZZ_MAX_STATS_KEYS` | `10000`  | Distinct requests kept in memory; `0` means unbounded |

---

## Project layout

```
main.go              Starts the CLI
cli/                 cobra root command
cli/cmd/             serve, version
server/launch.go     Lifecycle: signals, errgroup, graceful shutdown
server/routes.go     chi router, middleware stack, route table
server/middleware.go Request id, security headers, access log, recovery
server/health.go     The health check behind /health and the start-up gate
server/rest/         Handlers: fizzbuzz, stats, health
internal/fizzbuzz/   The rules: Request, Validate, Generate — no HTTP, no config
internal/stats/      Generic concurrency-safe frequency counter
utils/config/        Environment and .env loading, one struct per namespace
utils/resp/          JSON responses and the API error type
```

The dependency arrow points one way: `rest` knows `fizzbuzz`, `fizzbuzz` knows
nothing. `internal/` holds the rules, `server/` the transport, `utils/` what
both sides need. The rules of the exercise can therefore be tested, reused or
re-exposed (CLI, gRPC, queue consumer) without touching a line of HTTP code.

---

## Design decisions

**Logs read differently while developing.** `ENVIRONMENT=development`, or an
unset one, switches the handler from JSON to `slog`'s text format; deployed
environments keep the JSON a log collector expects.

**chi for routing and middlewares.** It brings `NotFound`/`MethodNotAllowed`
hooks, sub-routers and a middleware set worth more than its size: `CleanPath`,
`RealIP`, `Compress`, `RequestSize`, `Timeout`, plus `httprate` and `cors` from
the same family. Access logging and panic recovery stay hand-written, because
chi answers those in plain text where this API answers JSON everywhere.

**One config struct per namespace.** `APP_*` and `FIZZBUZZ_*` each map to a
struct that loads and validates itself, so adding a setting is local to its
namespace. Loading is explicit and happens once, before the server starts.

**The server lifecycle is one function.** `server.Launch` runs the health gate,
the listener and the shutdown watcher in an errgroup: the process refuses to
serve if it is not healthy, and a failure in either goroutine cancels the other.

**Domain isolated from transport.** `fizzbuzz.Request` validates itself and
returns a `*FieldError` naming the offending field; the HTTP layer maps that to
a 400 with the field name. Adding a parameter means touching the domain and its
parsing, nothing else.

**`Request` is the statistics key.** All five fields are comparable, so the
struct is used directly as a map key — no string concatenation, no ambiguity
between `str1="a b"` and `str1="a"`+`str2="b"`.

**Storage is bounded on purpose.** The counter is capped at `FIZZBUZZ_MAX_STATS_KEYS`
distinct requests. Without that bound, a client could grow the map indefinitely
by sending unique parameter combinations — a slow memory leak reachable from the
outside. Beyond the cap, known requests keep counting and new ones are dropped;
since the goal is to surface the *most frequent* request, losing the tail of
one-off requests is the right trade-off. See "Going further" for the alternative.

**`FIZZBUZZ_MAX_LIMIT` caps the response.** `limit=1000000000` is an easy way to
make a server allocate gigabytes, and a failed allocation of that size is fatal:
no recover catches it. The cap is therefore mandatory, and `Generate` bounds its
initial allocation on top, so the domain stays safe on its own.

**Interface on the consumer side.** `rest.StatsRecorder` is declared where it
is used, so the in-memory counter can be swapped for a shared store (Redis,
Postgres) without changing a handler.

**Errors are structured.** A stable `code`, a human `message`, and the `field`
at fault. Internals are logged, never returned.

### Security

The middleware order is the design: each one wraps those registered after it.

- **Security headers on every response**: `X-Content-Type-Options: nosniff`,
  `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer` and a
  `Content-Security-Policy` of `default-src 'none'; frame-ancestors 'none'` —
  the API only ever returns JSON, so nothing may be loaded, framed or referred.
  HSTS is announced only over TLS, directly or through `X-Forwarded-Proto`,
  since a browser ignores it on plain HTTP.
- **CORS is opt-in.** With no `APP_CORS`, the middleware is not installed at
  all, so no cross-origin request is ever answered — the library's own default
  is to allow every origin, which is the wrong way round. Credentials stay off:
  the API is stateless, so there is no cookie a hostile page could replay.
- **Rate limiting per IP** (`httprate`), on `/api/v1` only. A throttled probe
  would take a healthy instance out of rotation, so `/health` is exempt.
- **`X-Forwarded-For` is ignored unless `APP_TRUST_PROXY` is set.** The header is
  client-controlled: honouring it unconditionally lets anyone mint a fresh
  rate-limit bucket per request.
- **Request ids are opaque, and an inbound one is untrusted.** chi's generator
  prefixes ids with the machine's host name, which would publish it in every
  response, so ids are 16 random bytes. An id supplied by the caller is reused
  only when it is short and printable, since it is echoed back and written into
  every log line.
- **Bounded inputs**: request bodies are capped at 64 KiB, `limit` at
  `FIZZBUZZ_MAX_LIMIT`, the statistics map at `FIZZBUZZ_MAX_STATS_KEYS`, and
  every request carries a deadline (`APP_REQUEST_TIMEOUT`).
- **Errors say nothing useful to an attacker**: a stable code, a generic
  message, and the details in the logs.

### Production readiness

- **Timeouts on every phase** (header, read, write, idle): without them a slow
  client can pin a connection and its goroutine indefinitely.
- **Graceful shutdown** on SIGINT/SIGTERM: stop accepting connections, let
  in-flight requests finish within `APP_SHUTDOWN_TIMEOUT`, then exit.
- **Panic recovery**: a panicking handler becomes a logged 500 instead of a
  dropped connection, and the process stays up.
- **Structured JSON logs** to stdout, one access log line per request, with
  status, size and duration — what a log collector expects from a container.
- **Correlation id**: `X-Request-Id` is honoured when a gateway sets it, and
  generated otherwise; it is echoed back and present in every log line.
- **`/health`** for liveness and readiness probes, backed by the same
  `server.Health` check that gates start-up, so a probe and a boot agree.
- **Minimal image**: a 9.5 MB `scratch` image holding the static binary and the
  root certificates, nothing else — no shell, no libc, no package manager, so
  the attack surface is the binary itself. It runs as the unprivileged uid
  65532, and with no zoneinfo on board its logs are UTC.

## Continuous integration

`.github/workflows/ci.yml` runs one job on every push and pull request, with
read-only permissions:

```yaml
- run: make check      # gofmt, go vet, go build
- run: make image      # the container image still builds
```

Both are the same commands used locally, so a green pipeline and a green
`make check` mean the same thing.

## Going further

Deliberately left out. Each item carries a `TODO(lifetime)` in the source, on
the line that would change — `grep -rn "TODO(lifetime)" .` lists all seven.

**As traffic grows**

- The statistics counter lives in the process and dies with it, so behind
  several replicas each one answers with its own view. A Redis `ZINCRBY` keyed
  on the parameters, or a table with an upsert, makes it global and durable;
  `rest.StatsRecorder` is the seam that keeps handlers untouched.
- Rate limiting counts in memory, so N replicas allow N times the configured
  rate. `httprate-redis`, or throttling at the ingress, restores one window.
- One access log line per request stops being free at a few thousand rps:
  sample the successes and keep every error.

**As the service stays up**

- `/health` answers for liveness and readiness at once. They should split, and
  readiness should fail while draining, so the load balancer stops sending
  traffic before the process goes.
- Nothing reports saturation. Rate, errors and duration as metrics is what an
  autoscaler and an SLO both read.
- A recovered panic only reaches the logs; an error tracker would page someone.

**As the API evolves**

- `/api/v1` is a promise. A payload change ships as `/api/v2` alongside it with
  a deprecation window, not as an edit in place.

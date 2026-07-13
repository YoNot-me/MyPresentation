# Presentator

A backend service for **privately hosting presentations**. Each brand (tenant) signs in and sees only its own materials (works), while an administrator centrally manages brands and their presentations through a protected admin panel.

Built in **Go** with a clean, feature-based architecture: JWT sessions stored in Redis, PostgreSQL for persistence, per-brand isolated file serving, and full containerization via Docker Compose.

---

## Features

- **Private brand accounts.** A brand signs in with a login/password and sees only its own presentations.
- **Two-level authorization.** `brand` and `admin` roles are separated at the middleware level (`Protected` vs `ProtectedAdmin`).
- **Redis-backed JWT sessions.** Each token carries a unique `jti`, stored in Redis under `sess:<jti>` with a 12h TTL. Logout invalidates the session; every protected request re-checks that the session still exists.
- **Brute-force protection.** Failed sign-in attempts are counted per IP in Redis (`brute:auth:<ip>` / `brute:admin:auth:<ip>`) and rejected after 5 failures for a 3-minute window.
- **Admin panel.** Create/rename/delete brands, change brand passwords, add/edit/delete works, upload presentation images and a preview.
- **Protected file serving.** Presentation files are served per-brand from `./works/<brand>/<work>/…` with access isolation and path-traversal protection (`filepath.Clean` + `..` rejection). Only image files (`.jpg`, `.jpeg`, `.png`, `.gif`) are accepted on upload.
- **Production practices.** Graceful shutdown, structured logging (zap), HTTP timeouts, and a request body size limit (200 MB).

---

## Tech stack

| Layer | Technology |
|------|-----------|
| Language | Go 1.26 |
| HTTP framework | [Gin](https://github.com/gin-gonic/gin) |
| Database | PostgreSQL 18 (driver [pgx/v5](https://github.com/jackc/pgx), connection pool) |
| Sessions / rate limiting | Redis 8 |
| Authentication | JWT ([golang-jwt/v5](https://github.com/golang-jwt/jwt), HS256) |
| Password hashing | bcrypt (`golang.org/x/crypto`) |
| Logging | [zap](https://github.com/uber-go/zap) |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Configuration | godotenv (`.env`) |
| Orchestration | Docker Compose + Makefile |

---

## Architecture

The project follows a **feature-based** structure: each feature is self-contained and split into `transport → service → repository` layers.

```
.
├── cmd/                        # entry point
│   ├── main.go                 # dependency wiring, server start, graceful shutdown
│   └── init.go                 # loads .env, connects to Postgres/Redis, builds the logger
├── internal/
│   ├── core/                   # shared application core
│   │   ├── server/             # HTTP server setup and routing
│   │   ├── middleware/         # route protection (brand/admin), request body size limit
│   │   ├── repository/         # database connection pool setup
│   │   ├── entity/             # entities, config, error → HTTP status mapping
│   │   └── logger/             # zap logger initialization
│   └── features/               # business features
│       ├── auth/               # brand sign-in, logout
│       │   └── token/          # issuing/validating JWTs, session storage in Redis
│       ├── admin/              # admin auth + brand and work management
│       └── fileserving/        # listing and serving presentation files
├── migrations/                 # SQL migrations (up/down)
├── public/                     # static frontend (auth, admin panel, presentation gallery)
├── docker-compose.yaml
├── Dockerfile
├── Makefile
├── go.mod
└── go.sum
```

Each feature is layered consistently:

- **transport** — HTTP handlers (Gin): request parsing, response formatting.
- **service** — business logic.
- **repository** — data access (SQL via pgx, Redis via go-redis).

Errors are declared as sentinel values in `entity/err.go`, and `entity.FindStatus` centrally maps them to HTTP status codes, returned through a single `entity.Response` shape.

---

## Authentication & sessions

- **Brands** authenticate at `POST /auth/check` with `{ "name", "password" }`. On success the server issues an HS256 JWT (role `guest`) and sets it as an **HttpOnly, Secure, SameSite=Lax** cookie named `Pres-Access` (12h lifetime), then redirects to `/works/serve`.
- **Admin** authenticates at `POST /admin/auth` with `{ "login", "password" }`. Credentials are checked against `PRES_ADMIN_LOGIN` / `PRES_ADMIN_PASSWORD` (a bcrypt hash) from the environment; on success a JWT with role `admin` is issued in the same `Pres-Access` cookie.
- **Session validity** is enforced on every protected request: the middleware validates the token signature, confirms the session key `sess:<jti>` still exists in Redis, and (for admin routes) checks the `admin` role. Logout (`POST /logout`, `POST /logout/admin`) deletes the session key, invalidating the token before its natural expiry.

> **Note:** the `Pres-Access` cookie is issued with the `Secure` flag, so the browser only stores it over **HTTPS**. Run the service behind a TLS-terminating reverse proxy (or add TLS) in production.

---

## Data model

```sql
CREATE SCHEMA presentation;

-- brands (tenants), identified by a unique name, with a bcrypt-hashed password
CREATE TABLE presentation.brands (
    name     VARCHAR(256) UNIQUE NOT NULL,
    password TEXT         NOT NULL
);

-- presentations linked to a brand
CREATE TABLE presentation.works (
    brand       VARCHAR(256) NOT NULL,
    workName    VARCHAR(256) NOT NULL,
    url         TEXT         DEFAULT '',
    description TEXT         NOT NULL DEFAULT '',

    CONSTRAINT presentation_work FOREIGN KEY (brand)
        REFERENCES presentation.brands(name)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    CONSTRAINT unique_brand_work UNIQUE (brand, workName)
);

CREATE INDEX work_name_idx ON presentation.works(workName);
```

Cascading rules keep the file store and the database consistent: deleting a brand removes all of its works, and renaming a brand cascades to its works via `ON UPDATE CASCADE`.

---

## API

### Public routes

| Method | Path | Description |
|-------|------|-------------|
| `GET`  | `/` | Redirects to `/works/serve` |
| `GET`  | `/auth` | Brand sign-in page |
| `POST` | `/auth/check` | Brand authentication, issues a JWT cookie |
| `POST` | `/logout` | Brand sign-out (session invalidation) |
| `GET`  | `/admin` | Admin sign-in page |
| `POST` | `/admin/auth` | Admin authentication |
| `POST` | `/logout/admin` | Admin sign-out |

Static assets: `/auth/*` (sign-in pages), `/static/presentation/*` (gallery frontend).

### Brand routes (require a valid brand JWT)

| Method | Path | Description |
|-------|------|-------------|
| `GET` | `/works` | JSON list of the current brand's presentations |
| `GET` | `/works/serve` | Serves the gallery HTML shell |
| `GET` | `/works/files/:name` | JSON list of image files for a presentation |
| `GET` | `/presentation/:name/*filepath` | Serves a file of a specific presentation |

### Admin routes (require an admin JWT, `/admin` prefix)

| Method   | Path | Description |
|----------|------|-------------|
| `GET`    | `/admin/brands` | List all brands |
| `POST`   | `/admin/brands/add` | Create a brand |
| `PUT`    | `/admin/:brandName/rename` | Rename a brand |
| `DELETE` | `/admin/:brandName` | Delete a brand |
| `PUT`    | `/admin/:brandName/password` | Change a brand's password |
| `GET`    | `/admin/:brandName/works` | List a brand's works |
| `POST`   | `/admin/:brandName/works/add` | Add a work (multipart: `data` JSON + `files`/`preview` images) |
| `DELETE` | `/admin/:brandName/remove/:workName` | Delete a work |
| `PUT`    | `/admin/:brandName/:workName/change` | Update a work's fields / rename / replace preview |
| `GET`    | `/admin/:brandName/files/:workName` | List image files for a work |
| `GET`    | `/admin/:brandName/serve/:workName/*filepath` | Serve a work's file as admin |

Admin panel static assets are served under `/admin/panel/*`.

---

## Configuration

The application reads its settings from environment variables (loaded from a `.env` file in the project root when present).

| Variable            | Purpose                                              | Example |
|---------------------|------------------------------------------------------|---------|
| `PRES_ADDR`         | HTTP server listen address                           | `:8080` |
| `PRES_DATABASE_URL` | PostgreSQL connection string                         | `postgres://user:pass@postgres-database:5432/presentation_db?sslmode=disable` |
| `PRES_JWT_SECRET`   | Secret key for signing JWTs (HS256)                  | `super-secret-random-key` |
| `PRES_JWT_ISSUER`   | JWT `iss` claim                                      | `my-presentation-db` |
| `PRES_ADMIN_LOGIN`  | Admin login                                          | `admin` |
| `PRES_ADMIN_PASSWORD` | Admin password, stored as a **bcrypt hash**        | `$2a$10$...` |
| `REDIS_ADDR`        | Redis address                                        | `redis:6379` |
| `REDIS_PASSWORD`    | Redis password                                       | `my_pres_redis123` |
| `POSTGRES_USER`     | Database user (for Docker & migrations)              | `postgres` |
| `POSTGRES_PASSWORD` | Database password (for Docker & migrations)          | `postgres` |
| `POSTGRES_DB`       | Database name (for Docker & migrations)              | `presentation_db` |

> Generate the admin password hash with any bcrypt tool, e.g.
> `htpasswd -bnBC 10 "" 'yourpassword' | tr -d ':\n'`, and put the resulting `$2a$...` string in `PRES_ADMIN_PASSWORD`.

---

## Running the project

### Prerequisites

- Docker and Docker Compose
- Go 1.26+ (only for running the app locally without Docker)

### Full stack via Docker Compose (recommended)

Create a `.env` file with the variables above, then start everything — Postgres, Redis, migrations, and the app:

```bash
docker compose up --build -d
```

Compose brings services up in order:
1. `postgres-database` and `redis` start and become healthy;
2. `postgres-migrate` applies all migrations and exits;
3. `app` builds and starts once its dependencies are healthy.

The server is then available on `http://localhost:8080`.

### Local development (infrastructure in Docker, app on host)

Bring up the infrastructure and apply migrations, then run the app from source:

```bash
make run          # start Postgres + Redis, apply migrations
make run-local    # go run ./cmd
```

When running the app on the host, make sure `PRES_DATABASE_URL` and `REDIS_ADDR` in `.env` point to host-reachable addresses for Postgres and Redis.

### Useful Makefile commands

| Command | Action |
|---------|--------|
| `make run` | Start Postgres + Redis and apply migrations |
| `make run-local` | Run the application from source (`go run ./cmd`) |
| `make compose-up` | Start Postgres and Redis and wait for the DB |
| `make compose-down` | Stop the containers |
| `make migrate-up` / `make migrate-down` | Apply / roll back migrations |
| `make migrate-create seq=<name>` | Create a new migration |
| `make clean-env` | Remove containers, local images, volumes, and DB data |
| `make logs-db` | Show PostgreSQL container logs |

---

## Implementation notes

- **Graceful shutdown.** The server listens for `SIGINT`/`SIGTERM` and cleanly shuts down the HTTP server (10s timeout), the database pool, and the Redis client.
- **Timeouts & limits.** A `TimeoutHandler` (180s) wraps the router; the HTTP server uses a 120s read and 180s write timeout, and request bodies are capped at 200 MB.
- **Centralized error handling.** Sentinel errors are mapped to HTTP statuses via `entity.FindStatus`, with a single `entity.Response` response format.
- **Single dependency-wiring point.** All services are constructed in `cmd/main.go` via explicit dependency injection (no global state).

---

## License

See the [LICENSE](./LICENSE) file.

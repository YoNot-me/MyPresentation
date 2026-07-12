# Presentator

A backend service for **privately hosting presentations**. Each brand gets personal access to its own materials (works), while an administrator centrally manages brands and their presentations.

Built in **Go** with a clean feature-based architecture, JWT authentication backed by Redis, PostgreSQL for storage, and full containerization via Docker Compose.

---

## Features

- **Private brand accounts.** A brand signs in with a login/password and sees only its own presentations.
- **Two-level authorization.** `brand` and `administrator` permissions are separated at the middleware level.
- **Admin panel (API).** Create/delete brands, change passwords, add/edit/delete works.
- **Protected file serving.** HTML presentations are served per-brand with access isolation and path-traversal protection (`filepath.Clean`).
- **JWT sessions in Redis.** Tokens are stored in Redis; logout invalidates the session.
- **Production practices.** Graceful shutdown, structured logging (zap), request timeouts, request body size limit.

---

## Tech stack

| Layer | Technology |
|------|-----------|
| Language | Go 1.26 |
| HTTP framework | [Gin](https://github.com/gin-gonic/gin) |
| Database | PostgreSQL 18 (driver [pgx/v5](https://github.com/jackc/pgx), connection pool) |
| Sessions / cache | Redis 8 |
| Authentication | JWT ([golang-jwt/v5](https://github.com/golang-jwt/jwt)) |
| Logging | [zap](https://github.com/uber-go/zap) |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Configuration | godotenv (`.env`) |
| Orchestration | Docker Compose + Makefile |

---

## Architecture

The project follows a **feature-based** structure: each feature is self-contained and split into `transport → service → repository` layers.

```
.
├── cmd/                        # entry point: dependency initialization and server startup
│   ├── main.go                 # service wiring, graceful shutdown
│   └── init.go                 # loads .env, connects to Postgres/Redis, logger
├── internal/
│   ├── core/                   # shared application core
│   │   ├── server/             # HTTP server initialization and routing
│   │   ├── middleware/         # route protection (brand/admin), body size limit
│   │   ├── repository/         # database connection pool setup
│   │   ├── entity/             # entities, config, error → HTTP status mapping
│   │   └── logger/             # zap logger initialization
│   └── features/               # business features
│       ├── auth/               # brand sign-in, JWT tokens, logout
│       │   └── token/          # issuing/validating JWTs, storage in Redis
│       ├── admin/               # brand and work management
│       └── fileserving/         # listing and serving presentation files
├── migrations/                 # SQL migrations (up/down)
├── public/                     # static assets (auth and presentation pages)
├── docker-compose.yaml
├── Makefile
├── go.mod
└── public                      # frontend
```

Each feature is layered consistently:

- **transport** — HTTP handlers (Gin), request parsing, response formatting.
- **service** — business logic.
- **repository** — data access (SQL via pgx).

Errors are declared as sentinel values in `entity/err.go`, and the `FindStatus` function centrally maps them to the correct HTTP status codes.

---

## Data model

```sql
CREATE SCHEMA presentation;

-- brands (tenants), identified by a unique name, with a hashed password
CREATE TABLE brands (
    name     VARCHAR(256) UNIQUE NOT NULL,
    password TEXT         NOT NULL
);

-- presentations linked to a brand
CREATE TABLE works (
    brand       VARCHAR(256) NOT NULL,
    workName    VARCHAR(256) NOT NULL,
    url         TEXT         DEFAULT '',
    description TEXT         NOT NULL DEFAULT '',

    CONSTRAINT presentation_work FOREIGN KEY (brand)
        REFERENCES brands(name)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    CONSTRAINT unique_brand_work UNIQUE (brand, workName)
);

CREATE INDEX work_name_idx ON works(workName);
```

Cascading delete guarantees that removing a brand also removes all of its works; renaming a brand cascades to its works via `ON UPDATE CASCADE`.

---

## API

### Public routes

| Method | Path | Description |
|-------|------|----------|
| `GET`  | `/auth` | Brand sign-in page/endpoint |
| `POST` | `/auth/check` | Brand authentication, issues a JWT |
| `POST` | `/logout` | Sign out, session invalidation |
| `GET`  | `/admin` | Admin sign-in page |
| `POST` | `/admin/auth` | Administrator authentication |
| `POST` | `/logout/admin` | Administrator sign-out |

### Brand routes (require a brand JWT)

| Method | Path | Description |
|-------|------|----------|
| `GET` | `/works` | List of the current brand's presentations |
| `GET` | `/works/files/:name` | List of files for a presentation |
| `GET` | `/works/serve` | Serves the gallery HTML shell |
| `GET` | `/presentation/:name/*filepath` | Serves files for a specific presentation |

### Administrator routes (require an admin JWT, `/admin` prefix)

| Method   | Path | Description |
|----------|------|-------------|
| `GET`    | `/admin/brands` | List of all brands |
| `POST`   | `/admin/brands/add` | Create a brand |
| `PUT`    | `/admin/:brandName/rename` | Rename a brand |
| `DELETE` | `/admin/:brandName` | Delete a brand |
| `PUT`    | `/admin/:brandName/password` | Change a brand's password |
| `GET`    | `/admin/:brandName/works` | List a brand's works |
| `POST`   | `/admin/:brandName/works/add` | Add a work |
| `DELETE` | `/admin/:brandName/remove/:workName` | Delete a work |
| `PUT`    | `/admin/:brandName/:workName/change` | Update a work's fields |
| `GET`    | `/admin/:brandName/files/:workName` | List files for a work |
| `GET`    | `/admin/:brandName/serve/:workName/*filepath` | Serve a work's files as an admin |

---

## Configuration

The application reads its settings from a `.env` file in the project root.

| Variable            | Purpose                               | Example |
|---------------------|---------------------------------------|--------|
| `PRES_ADDR`         | HTTP server listen address            | `:8080` |
| `PRES_DATABASE_URL` | PostgreSQL connection string          | `postgres://user:pass@localhost:5433/presentation?sslmode=disable` |
| `PRES_JWT_SECRET`   | Secret key for signing JWTs           | `super-secret-key` |
| `PRES_REQ_TIMEOUT`  | Request timeout                       | `180s` |
| `POSTGRES_USER`     | Database user (for Docker/migrations) | `postgres` |
| `POSTGRES_PASSWORD` | Database password                     | `postgres` |
| `POSTGRES_DB`       | Database name                         | `presentation` |
| `REDIS_PASSWORD`    | Redis password                        | `my_pres_redis123` |
| `REDIS_ADDR`        | Redis addr                            | `localhost:6379` |

> Redis is exposed on `localhost:6379` by default.

---

## Running the project

### Prerequisites

- Docker and Docker Compose
- Go 1.26+ (for running the application locally)

### Via Docker Compose (infrastructure + migrations)

```bash
# start Postgres, Redis, the port-forwarder, and apply migrations
make run
```

`make run` runs, in order:
1. starts the `postgres-database` and `redis` containers and waits for the database to become ready;
2. starts the `port-forwarder` (socat) on `127.0.0.1:5433 → postgres:5432`;
3. applies migrations (`migrate-up`).

Then start the application itself:

```bash
make run-local   # go run ./cmd
```

The server starts on `localhost:8080`.

### Useful Makefile commands

| Command | Action |
|---------|--------|
| `make compose-up` | Start Postgres and Redis |
| `make compose-down` | Stop the containers |
| `make migrate-up` / `make migrate-down` | Apply / roll back migrations |
| `make migrate-create seq=<name>` | Create a new migration |
| `make run-local` | Run the application locally |
| `make clean-env` | Remove containers, images, and database data |
| `make logs-db` | Show PostgreSQL container logs |

---

## Implementation notes

- **Graceful shutdown.** The server listens for `SIGINT`/`SIGTERM` and cleanly closes the HTTP server, the database pool, and Redis with a 10s timeout.
- **Centralized error handling.** Sentinel errors are mapped to HTTP statuses via `entity.FindStatus`, with a single `entity.Response` response format.
- **Infrastructure isolation.** Access to Postgres is proxied through a socat port-forwarder, simplifying local development and deployment.
- **Single dependency-wiring point.** All services are constructed in `cmd/main.go` via explicit dependency injection (no global state).

---

## License

See the [LICENSE](./LICENSE) file.
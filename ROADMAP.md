# Relay — Design & Development Roadmap

## What is Relay?

A one-click deployment platform. Users submit a GitHub repo URL → Relay clones it, detects the project type, builds it in an isolated Docker container, streams live logs, and serves the deployed site/app at a unique URL.

---

## Tech Stack & Tools

### Core Backend
- **Go** — main backend language (already in use)
- **Chi Router** — HTTP routing (already in use)
- **godotenv** — env config loading (already in use)
- **golang-jwt** — JWT auth (already in use)
- **bcrypt** — password hashing (already in use)

### Database
- **PostgreSQL** — primary data store (users, projects, deployments, logs)
- **sqlx** or **pgx** — Go Postgres driver (`github.com/jackc/pgx/v5` recommended)
- **golang-migrate** — database migrations (`github.com/golang-migrate/migrate/v4`)

### Docker & Build
- **Docker Engine** — isolated build containers (Go Docker SDK: `github.com/docker/docker/client`)
- **git** — repo cloning (via `os/exec` calling `git clone`)

### Real-Time
- **Server-Sent Events (SSE)** — live log streaming (no extra library needed, just `text/event-stream` over HTTP)

### Storage
- **Local filesystem** — artifact storage for dev/MVP
- **AWS S3** (or compatible like MinIO) — production artifact storage (`github.com/aws/aws-sdk-go-v2`)

### Reverse Proxy / Site Serving
- **Caddy** or **Chi subrouter** — subdomain-based routing to serve deployed sites
- Serves static files with correct Content-Type headers

### Containerization & Deployment
- **Docker** — containerize the Relay backend itself
- **Docker Compose** — local dev (Relay + Postgres + optional MinIO)
- Cloud deployment target: any container service (ECS, Cloud Run, Fly.io, Railway, etc.)

### Optional / Future
- **Redis** — job queue for build orchestration (if goroutines aren't enough)
- **WebSocket** — alternative to SSE if bidirectional comms needed later
- **GitHub Webhooks** — auto-deploy on push

---

## Supported Project Types

Relay auto-detects and builds these project types:

### Frontend Frameworks
| Type | Detection | Base Image | Install | Build Command | Output Dir |
|------|-----------|------------|---------|---------------|------------|
| React | `react` in package.json deps | `node:20-alpine` | `npm ci` | `npm run build` | `build/` or `dist/` |
| Next.js | `next` in package.json deps | `node:20-alpine` | `npm ci` | `npm run build` | `.next/` (or `out/` for static export) |
| Vue | `vue` in package.json deps | `node:20-alpine` | `npm ci` | `npm run build` | `dist/` |
| Angular | `@angular/core` in package.json deps | `node:20-alpine` | `npm ci` | `npx ng build --configuration=production` | `dist/{project}/` |
| Svelte | `svelte` or `@sveltejs/kit` in deps | `node:20-alpine` | `npm ci` | `npm run build` | `build/` or `public/` |
| Static HTML | No package.json or no known framework | `alpine:latest` | — | — | `.` (root) |

### Backend / Full-Stack
| Type | Detection | Base Image | Install | Build/Run Command | Exposed |
|------|-----------|------------|---------|-------------------|---------|
| Node.js (JS) | `package.json` exists, no framework, has `start` script | `node:20-alpine` | `npm ci` | `npm start` | Port from `process.env.PORT` |
| Node.js (TS) | `typescript` in devDeps, has `start`/`build` script | `node:20-alpine` | `npm ci` | `npm run build && npm start` | Port from `process.env.PORT` |
| Go | `go.mod` exists | `golang:1.22-alpine` | `go mod download` | `go build -o app . && ./app` | Port from env |
| Python | `requirements.txt` or `pyproject.toml` exists | `python:3.12-slim` | `pip install -r requirements.txt` | `python main.py` or `gunicorn` | Port from env |
| Java | `pom.xml` or `build.gradle` exists | `eclipse-temurin:21-jdk-alpine` | `mvn package` or `gradle build` | `java -jar target/*.jar` | Port from env |

### Detection Priority (order matters)
```
1. go.mod           → Go
2. pom.xml          → Java (Maven)
3. build.gradle     → Java (Gradle)
4. requirements.txt → Python
5. pyproject.toml   → Python
6. package.json     → inspect deps:
   a. next          → Next.js
   b. @angular/core → Angular
   c. svelte        → Svelte
   d. vue           → Vue
   e. react         → React
   f. has start script, no framework → Node.js
7. index.html       → Static HTML
8. fallback         → Static HTML
```

---

## Architecture Overview

```
User → [Chi Router + Auth Middleware] → Handlers → Services → Repositories → PostgreSQL
                                            ↓
                                    Builder Service
                                    ├── git clone
                                    ├── detect project type
                                    ├── Docker build (isolated container)
                                    └── stream logs → Log Streamer → SSE → User
                                            ↓
                                    Storage Service
                                    ├── local filesystem (dev)
                                    └── S3 (prod)
                                            ↓
                                    Reverse Proxy
                                    └── subdomain → serve artifacts/proxy to running container
```

### Pipeline Flow
```
User submits repo URL
    → POST /api/projects (create project)
    → POST /api/projects/{id}/deployments (trigger deploy)
        → Status: pending
        → Status: cloning    (git clone to temp dir)
        → Status: building   (Docker container: install + build)
        → Status: deploying  (store artifacts / start container)
        → Status: live       (assign URL, serve site)
        → Status: failed     (at any step, with error in logs)
```

### For Frontend Projects (static output)
- Build produces static files → stored in artifact directory → served by reverse proxy as static files

### For Backend Projects (running process)
- Build produces a binary/app → runs inside a long-lived container → reverse proxy forwards traffic to the container's port

---

## Data Models

```
USER (already exists in auth)
├── id, email, name, role, created_at

PROJECT
├── id, user_id (FK), name, repo_url, project_type, active_deployment_id, created_at, updated_at

DEPLOYMENT
├── id, project_id (FK), status, deployed_url, subdomain, failure_reason, created_at, updated_at

DEPLOYMENT_LOG
├── id, deployment_id (FK), message, created_at
```

---

## API Endpoints

### Public
Already done: login, signup, OAuth, logout

### Protected (behind AuthZ + AuthN middleware)

| Method | Endpoint | What it does |
|--------|----------|-------------|
| `POST` | `/api/projects` | Create project from GitHub URL |
| `GET` | `/api/projects` | List user's projects |
| `GET` | `/api/projects/{projectID}` | Get project details |
| `DELETE` | `/api/projects/{projectID}` | Delete project + all deployments |
| `POST` | `/api/projects/{projectID}/deployments` | Trigger new deployment |
| `GET` | `/api/projects/{projectID}/deployments` | List deployments for project |
| `GET` | `/api/projects/{projectID}/deployments/{deploymentID}` | Get deployment details |
| `DELETE` | `/api/projects/{projectID}/deployments/{deploymentID}` | Delete a deployment |
| `GET` | `/api/projects/{projectID}/deployments/{deploymentID}/logs` | SSE log stream (or static JSON if done) |

---

## New Files to Create

```
relay/
├── models/
│   ├── project.go
│   ├── deployment.go
│   └── deploymentLog.go
├── repositories/
│   ├── projectRepository.go
│   └── deploymentRepository.go
├── services/
│   ├── projectService.go
│   ├── deploymentService.go
│   ├── builderService.go
│   ├── storageService.go
│   └── logStreamer.go
├── handlers/
│   ├── projectHandler.go
│   ├── deploymentHandler.go
│   └── logStreamHandler.go
├── errors/
│   ├── notFoundError.go
│   └── conflictError.go
├── proxy/
│   └── reverseProxy.go
├── migrations/
│   ├── 001_create_users.up.sql
│   ├── 001_create_users.down.sql
│   ├── 002_create_projects.up.sql
│   ├── 002_create_projects.down.sql
│   ├── 003_create_deployments.up.sql
│   ├── 003_create_deployments.down.sql
│   ├── 004_create_deployment_logs.up.sql
│   └── 004_create_deployment_logs.down.sql
├── Dockerfile
└── docker-compose.yml
```

---

## New Environment Variables

```env
# Database
DATABASE_URL="postgres://user:pass@localhost:5432/relay?sslmode=disable"

# Deployment Pipeline
ARTIFACTS_DIR="./artifacts"
DOCKER_TIMEOUT="600"
RELAY_DOMAIN="relay.host"
PROXY_PORT=":8080"
```

---

## Development Phases — What to Build & When

### Phase 1: Foundation (Database + Models + Repo Layer)
**Goal:** Get the data layer working so everything else can plug into it.

1. Set up PostgreSQL (local via Docker Compose or installed)
2. Create migration files for users, projects, deployments, deployment_logs tables
3. Add `pgx` driver and connect to DB in `configs/`
4. Create model structs in `models/`
5. Implement repository layer in `repositories/` (CRUD for each model)
6. Wire up the auth repo layer (the TODOs in `authService.go`)
7. Test: verify you can create users, projects, deployments via repo functions

**Go packages to install:**
```bash
go get github.com/jackc/pgx/v5
go get github.com/golang-migrate/migrate/v4
```

---

### Phase 2: Project CRUD
**Goal:** Users can create, list, view, and delete projects.

1. Create `services/projectService.go` — GitHub URL validation (regex), duplicate check, ownership validation
2. Create `handlers/projectHandler.go` — parse requests, call service, return JSON
3. Create `errors/notFoundError.go` and `errors/conflictError.go`
4. Register routes in `main.go` under the protected group
5. Test: curl/Postman to create a project, list it, get it, delete it

---

### Phase 3: Builder Service — Clone & Detect
**Goal:** Relay can clone a repo and figure out what it is.

1. Implement `Clone()` in `services/builderService.go` — `exec.Command("git", "clone", ...)` to a temp dir
2. Implement `DetectProjectType()` — read `go.mod`, `pom.xml`, `requirements.txt`, `package.json`, etc. following the detection priority
3. Test: clone a few different repos and verify detection works

---

### Phase 4: Builder Service — Docker Build
**Goal:** Relay can build any supported project type in an isolated container.

1. Install Docker SDK: `go get github.com/docker/docker/client`
2. Implement `Build()` in `services/builderService.go`:
   - Create a Docker container with the right base image
   - Mount the cloned source code
   - Run install + build commands
   - Stream stdout/stderr to the log streamer
   - Enforce 10-minute timeout
   - Extract build output (for frontend) or keep container running (for backend)
   - Clean up container on completion/failure
3. Test: trigger a build for a React app, a Go app, a Python app — verify output

---

### Phase 5: Storage Service
**Goal:** Built artifacts are persisted and retrievable.

1. Implement `services/storageService.go`:
   - `Store()` — copy build output to `ARTIFACTS_DIR/{deploymentID}/`
   - `Delete()` — remove artifact directory
   - `ServePath()` — return the filesystem path for serving
2. For backend projects, this step is different — the container itself is the "artifact"
3. Test: verify files are stored and retrievable after a build

---

### Phase 6: Log Streamer (SSE)
**Goal:** Users can watch builds happen in real-time.

1. Implement `services/logStreamer.go`:
   - In-memory pub/sub with channels
   - `Subscribe()` — returns a channel of log lines, supports `Last-Event-ID` for reconnect
   - `Publish()` — sends a log line to all subscribers
   - `Complete()` — sends completion event, closes channels
2. Implement `handlers/logStreamHandler.go`:
   - If deployment is in progress → open SSE connection, stream logs
   - If deployment is done → return full log as JSON array
3. Register the route
4. Test: trigger a deployment, open the SSE endpoint in browser/curl, watch logs stream

---

### Phase 7: Deployment Service (The Orchestrator)
**Goal:** Wire the full pipeline together.

1. Implement `services/deploymentService.go`:
   - `TriggerDeployment()` — create record (pending), launch async goroutine:
     - Clone → Detect → Build → Store → Assign URL → Mark live
     - Update status at each transition
     - On any failure: mark failed, record reason
   - `ListDeployments()`, `GetDeployment()`, `DeleteDeployment()` with ownership checks
   - Redeployment: same as trigger, updates active deployment on success
2. Implement `handlers/deploymentHandler.go`
3. Register routes
4. Test: full end-to-end — submit repo → trigger deploy → watch logs → get deployed URL

---

### Phase 8: Reverse Proxy — Serve Deployed Sites
**Goal:** Deployed sites are accessible at unique URLs.

1. Implement `proxy/reverseProxy.go`:
   - Extract subdomain from Host header
   - Look up deployment by subdomain
   - For frontend: serve static files from artifact directory
   - For backend: proxy HTTP traffic to the running container's port
   - 404 with Relay-branded page for unknown subdomains
2. URL generation: `{project-slug}-{short-id}.relay.host`
3. Wire into `main.go` — separate listener on `PROXY_PORT` or subdomain-aware middleware
4. Test: deploy a site, hit the generated URL, see it load

---

### Phase 9: Dockerize Relay Itself
**Goal:** Package the entire Relay backend as a Docker image.

1. Create `Dockerfile` — multi-stage build (build Go binary → slim runtime image)
2. Create `docker-compose.yml` — Relay + PostgreSQL + (optional MinIO for S3)
3. Make sure Docker-in-Docker or Docker socket mounting works for builds
4. Test: `docker compose up` and run through the full flow

---

### Phase 10: Polish & Production Prep
**Goal:** Harden for real usage.

1. Add rate limiting on deployment triggers
2. Add build caching (node_modules, go mod cache)
3. Swap local storage for S3 (if going to prod)
4. Add GitHub webhooks for auto-deploy on push
5. Add custom domain support
6. Add HTTPS/TLS for deployed sites
7. Add cleanup job for old/orphaned artifacts and containers
8. Add monitoring and alerting

---

## Key Design Decisions

| Decision | Choice | Why |
|----------|--------|-----|
| Build isolation | Docker containers | Process-level isolation, no interference between builds |
| Log streaming | SSE (Server-Sent Events) | Unidirectional, lightweight, native browser support, reconnect built-in |
| URL scheme | `{slug}-{shortID}.relay.host` | Human-readable, collision-resistant |
| Artifact storage | Local filesystem → S3 | Start simple, swap later without code changes |
| Pipeline execution | Async goroutine per deployment | Simple concurrency, upgradeable to job queue later |
| Backend project serving | Long-lived Docker container + reverse proxy | Backend apps need a running process, not just static files |
| Project detection | File-based inspection (go.mod, package.json, etc.) | No user config needed, works automatically |

---

## Quick Reference — Go Packages to Install

```bash
# Database
go get github.com/jackc/pgx/v5
go get github.com/golang-migrate/migrate/v4

# Docker SDK
go get github.com/docker/docker/client
go get github.com/docker/docker/api/types

# S3 (when ready for prod)
go get github.com/aws/aws-sdk-go-v2
go get github.com/aws/aws-sdk-go-v2/service/s3

# UUID generation
go get github.com/google/uuid
```

---

Happy building.

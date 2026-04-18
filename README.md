<h1 align="center">Relay</h1>

<p align="center">
  One-click deployment platform — submit a GitHub repo, Relay clones it, detects the stack, builds it in Docker, streams live logs, and serves it at a unique URL.
</p>

---

## Overview

Relay is a full-stack deployment platform that automates the entire pipeline from source code to live site. Users paste a GitHub repository URL, and Relay handles everything — cloning, project type detection, containerized builds, artifact storage, and serving the deployed site at a unique subdomain.

Build logs stream in real-time via Server-Sent Events, deployment status updates live, and each project gets its own isolated Docker build environment.

| Component | Stack | Purpose |
|-----------|-------|---------|
| Backend | Go (Chi router), PostgreSQL, Docker SDK | REST API, auth, build orchestration, reverse proxy |
| Frontend | React 19, Vite, Framer Motion | SPA with project management, deployment triggers, live log viewer |
| Database | PostgreSQL (Supabase) | Users, projects, deployments, logs |
| Builder | Docker containers | Isolated builds per project type |

---

## System Architecture

```
app.relay.host (Frontend)  ──HTTPS──►  relay.host (Go Backend)
                                              │
                                              ▼
                                    ┌──────────────────-┐
                                    │  Relay Backend    │
                                    │  Go + Chi Router  │
                                    │  Port 3000 (API)  │
                                    │  Port 8080 (Proxy)│
                                    └────────┬─────────-┘
                                             │
                              ┌──────────────┼──────────────┐
                              ▼              ▼              ▼
                        ┌──────────┐  ┌──────────┐  ┌──────────┐
                        │ Docker   │  │ Supabase │  │ Artifacts│
                        │ Builds   │  │ Postgres │  │ Storage  │
                        │ (per     │  │ (remote) │  │ (local)  │
                        │ deploy)  │  │          │  │          │
                        └──────────┘  └──────────┘  └──────────┘
```

### Deployment Pipeline

```
User submits repo URL
    → POST /api/projects (create project, clone, detect type)
    → POST /api/projects/{id}/deployments (trigger deploy)
        → Status: pending
        → Status: cloning    (git clone --depth 1)
        → Status: detecting  (inspect go.mod, package.json, etc.)
        → Status: building   (Docker container: install + build)
        → Status: deploying  (store artifacts)
        → Status: live       (assign subdomain, serve site)
        → Status: failed     (at any step, with error details)
```

---

## Features

- **One-Click Deploy** — Paste a GitHub URL. Relay clones, detects, builds, and serves. No config files needed.
- **Auto-Detection** — Supports 11 project types. Detection follows a priority chain: Go → Java → Python → Node.js frameworks → Static HTML.
- **Isolated Docker Builds** — Every build runs in its own container with the correct base image, install commands, and build commands.
- **Live Build Logs** — Real-time log streaming via Server-Sent Events. Watch `npm ci` and `npm run build` happen line by line.
- **Unique Subdomain URLs** — Each deployment gets a URL like `my-app-a1b2c3d4.relay.host`.
- **Deployment History** — Track all deployments per project with status, timestamps, and failure reasons.
- **JWT Authentication** — Signup/login with bcrypt-hashed passwords. Google and GitHub OAuth supported.
- **Rate Limiting** — One deployment per project per 2 minutes to prevent spam.
- **Automatic Cleanup** — Hourly job removes orphaned Docker containers and stale temp directories.
- **Reverse Proxy** — Subdomain-based routing serves deployed frontend sites directly from stored artifacts.

---

## Supported Project Types

| Type | Detection | Base Image | Build Command |
|------|-----------|------------|---------------|
| Go | `go.mod` exists | `golang:1.22-alpine` | `go mod download && go build` |
| Java (Maven) | `pom.xml` exists | `eclipse-temurin:21-jdk-alpine` | `mvn package` |
| Java (Gradle) | `build.gradle` exists | `eclipse-temurin:21-jdk-alpine` | `gradle build` |
| Python | `requirements.txt` or `pyproject.toml` | `python:3.12-slim` | `pip install -r requirements.txt` |
| Next.js | `next` in package.json deps | `node:20-alpine` | `npm ci && npm run build` |
| Angular | `@angular/core` in deps | `node:20-alpine` | `npm ci && npx ng build` |
| Svelte | `svelte` or `@sveltejs/kit` in deps | `node:20-alpine` | `npm ci && npm run build` |
| Vue | `vue` in deps | `node:20-alpine` | `npm ci && npm run build` |
| React | `react` in deps | `node:20-alpine` | `npm ci && npm run build` |
| Node.js | package.json with `start` script | `node:20-alpine` | `npm ci && npm start` |
| Static HTML | `index.html` or fallback | `alpine:latest` | None |

Detection priority follows the order listed above — first match wins.

---

## Tech Stack

**Backend**
- Go 1.25, Chi v5 (router)
- pgx v5 (PostgreSQL driver + connection pool)
- Docker SDK (`github.com/docker/docker/client`)
- golang-jwt v5, bcrypt
- Server-Sent Events (native `text/event-stream`)
- Structured logging via `slog`
- Embedded SQL migrations (`embed.FS`)

**Frontend**
- React 19, React Router v7
- Vite 8
- Framer Motion (animations)
- Axios (HTTP client)

**Database**
- PostgreSQL with UUID primary keys
- Tables: `users`, `admins`, `projects`, `deployments`, `deployment_logs`

**Infrastructure & Deployment**
- Docker (multi-stage builds for Relay itself)
- Docker SDK (for building user projects in isolated containers)
- GCP Compute Engine (VM hosting)
- Supabase (managed PostgreSQL)
- Wildcard DNS (`*.relay.host` → VM IP)

---

## Project Structure

```
.
├── configs/              # Server config, OAuth provider setup
├── db/                   # Database connection + embedded migrations
│   └── migrations/       # SQL migration files
├── errors/               # Typed HTTP error constructors (400, 404, 409, 500)
├── handlers/             # HTTP handlers (auth, project, deployment, log stream)
├── middlewares/           # CORS, logging, AuthZ (JWT), AuthN (role), rate limiting
├── models/               # DB models, request/response structs
├── proxy/                # Reverse proxy for serving deployed sites
├── repositories/         # Database queries (auth, project, deployment)
├── services/             # Business logic
│   ├── builderService    # Clone, detect, Docker build
│   ├── storageService    # Artifact storage (copy, delete, serve path)
│   ├── logStreamer        # In-memory pub/sub for SSE log streaming
│   ├── deploymentService # Pipeline orchestrator
│   └── cleanupService    # Hourly cleanup of containers and temp dirs
├── utils/                # JWT generation/validation, password hashing
├── frontend/             # React SPA (Vite)
│   └── src/
│       ├── components/   # Navbar, Card, Button, Input
│       └── pages/        # Landing, Login, Signup, Dashboard, ProjectDetail, DeploymentDetail
├── artifacts/            # Build output storage (gitignored)
├── main.go               # Entrypoint, DI wiring, router setup, graceful shutdown
├── Dockerfile            # Multi-stage build (Go binary + alpine runtime)
└── docker-compose.yml    # Relay + Docker socket mount
```

---

## API Endpoints

### Public

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `POST` | `/auth/signup` | Register new user |
| `POST` | `/auth/login` | Login, returns JWT |
| `GET` | `/auth/google/login` | Google OAuth redirect |
| `GET` | `/auth/google/callback` | Google OAuth callback |
| `GET` | `/auth/github/login` | GitHub OAuth redirect |
| `GET` | `/auth/github/callback` | GitHub OAuth callback |
| `GET` | `/auth/logout` | Clear auth cookie |

### Protected (requires `Authorization: Bearer <token>`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/projects` | Create project from GitHub URL |
| `GET` | `/api/projects` | List user's projects |
| `GET` | `/api/projects/{projectID}` | Get project details |
| `DELETE` | `/api/projects/{projectID}` | Delete project + all deployments |
| `POST` | `/api/projects/{projectID}/deployments` | Trigger new deployment |
| `GET` | `/api/projects/{projectID}/deployments` | List deployments for project |
| `GET` | `/api/projects/{projectID}/deployments/{deploymentID}` | Get deployment details |
| `DELETE` | `/api/projects/{projectID}/deployments/{deploymentID}` | Delete a deployment |
| `GET` | `/api/projects/{projectID}/deployments/{deploymentID}/logs` | SSE log stream (or JSON if done) |

---

## Authentication Flow

1. User signs up or logs in via `/auth/signup`, `/auth/login`, or OAuth (Google/GitHub)
2. Backend hashes passwords with bcrypt, generates a JWT containing `userID`, `email`, and `role`
3. Frontend stores the JWT in `localStorage` and attaches it as a `Bearer` token on all subsequent requests
4. OAuth flow: backend redirects to provider → callback receives user info → upserts user → generates JWT → sets cookie
5. Protected routes pass through two middleware layers:
   - **AuthZ** — Validates the JWT and extracts user claims into request context
   - **AuthN** — Checks role-based access (e.g., `/admin/*` routes require `admin` role)

---

## How Deployment Works

1. User creates a project by submitting a GitHub URL — Relay clones the repo and auto-detects the project type
2. User triggers a deployment — backend creates a deployment record with `pending` status and returns immediately
3. A goroutine runs the pipeline asynchronously:
   - **Clone** — `git clone --depth 1` to a temp directory
   - **Detect** — Inspect files (`go.mod`, `package.json`, etc.) to determine project type
   - **Build** — Pull the correct Docker image, create a container with the source mounted, run install + build commands
   - **Stream** — Build logs are published line-by-line to an in-memory pub/sub; SSE subscribers receive them in real-time
   - **Store** — Copy build output to `artifacts/{deploymentID}/`
   - **Live** — Generate subdomain URL, update deployment record, update project's active deployment
4. On any failure, the deployment is marked `failed` with the last 5 lines of build output as the error reason
5. The reverse proxy serves deployed sites by extracting the subdomain from the `Host` header and serving static files from the corresponding artifacts directory

---

## Local Development

### Prerequisites

- Go 1.25+
- Node.js 20+
- Docker Desktop
- PostgreSQL (local or Supabase)

### Setup

```bash
git clone https://github.com/sash2721/Relay.git
cd Relay
cp template.env .env
# Edit .env with your values
go mod download
mkdir -p artifacts
```

### Run

```bash
# Terminal 1 — Backend
go run main.go

# Terminal 2 — Frontend
cd frontend && npm install && npm run dev
```

Open `http://localhost:5173`

### Run with Docker

```bash
docker compose up --build
```

---

## Deployment to GCP

### 1. Create a Compute Engine VM

```bash
gcloud compute instances create relay-server \
  --zone=us-central1-a \
  --machine-type=e2-medium \
  --image-family=ubuntu-2204-lts \
  --image-project=ubuntu-os-cloud \
  --boot-disk-size=50GB \
  --tags=http-server,https-server
```

### 2. Open Firewall Ports

```bash
gcloud compute firewall-rules create allow-relay \
  --allow=tcp:80,tcp:443,tcp:3000,tcp:8080 \
  --target-tags=http-server
```

### 3. Install Docker on the VM

```bash
sudo apt update && sudo apt install -y docker.io docker-compose-plugin git
sudo usermod -aG docker $USER && newgrp docker
```

### 4. Clone, Configure, and Run

```bash
git clone https://github.com/sash2721/Relay.git
cd Relay
nano .env  # Add production values
mkdir -p artifacts
docker compose up --build -d
```

### 5. DNS Setup

Configure your domain with a wildcard A record:

```
A     relay.host      → <VM_EXTERNAL_IP>
A     *.relay.host    → <VM_EXTERNAL_IP>
```

This ensures all deployed sites (`my-app-a1b2c3d4.relay.host`) resolve to your server automatically.

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `PORT` | API server port (e.g. `:3000`) |
| `HOST` | Server host |
| `ENV` | `development` or `production` |
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | Secret key for JWT signing |
| `ARTIFACTS_DIR` | Path to store build artifacts (e.g. `./artifacts`) |
| `PROXY_PORT` | Reverse proxy port (e.g. `:8080`) |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret |
| `GITHUB_CLIENT_ID` | GitHub OAuth client ID |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth client secret |

---

## License

MIT

---

<p align="center">Built with 🤍 and Go</p>

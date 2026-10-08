# AutoForge — Autonomous Software Engineering Fabric

Complete implementation of the autonomous build system for macOS.

## Quick Start

```bash
# Build locally
make build

# Run services
make run

# Submit a job
./bin/autoforge submit https://github.com/myorg/myrepo abc123

# Check status
./bin/autoforge status <saga-id>

# Or use Docker
make docker
make up
```

## Architecture

- **CKG Service** (:8080) — Code Knowledge Graph with SQLite backend
- **Saga Supervisor** (:8081) — State machine orchestrator
- **CLI** — Client tool for job submission
- **SandboxCell CRD** — Kubernetes custom resource for ephemeral builds

## GitHub Actions

Automatically builds:
- macOS binaries (native)
- Linux containers (pushed to GHCR)
- Runs integration tests
- Deploys K8s CRDs

## Project Structure

```
autoforge/
├── bin/                    # Compiled binaries
├── cmd/
│   ├── ckg-service/       # Code Knowledge Graph API
│   ├── saga-supervisor/   # Workflow orchestrator
│   └── cli/               # Command line tool
├── crds/                  # Kubernetes CRDs
├── deployments/           # Dockerfiles
├── .github/workflows/     # CI/CD
└── docker-compose.yml     # Local dev stack
```

## Environment Variables

- `CKG_URL` — CKG service endpoint
- `SUPERVISOR_URL` — Saga supervisor endpoint
- `PORT` — Service port

## License

MIT

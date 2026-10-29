# FlowPilot

**AI-powered workflow and integration platform** — a portfolio project exploring reliable asynchronous execution, visual workflow design, and AI-assisted document processing.

> Status: early development. Features described in the roadmap are planned until implemented and tested.

## Product vision

FlowPilot lets a user define a workflow made of triggers, processing steps, conditions, and actions. It executes workflows asynchronously and provides a traceable history so users can understand successes, failures, and retries.

## Planned stack

- **Frontend:** React, TypeScript, Vite, React Flow
- **Backend:** Go, REST API
- **Persistence:** PostgreSQL
- **Asynchronous jobs:** Redis-backed queue and Go worker pool
- **Testing:** Go test, Vitest, React Testing Library, integration tests
- **Local development:** Docker Compose
- **Delivery:** GitHub Actions and a deployable demo

The first version will use a modular monolith. Separate services will only be introduced if a demonstrated scaling or ownership need justifies them.

## Initial scope

1. Create and edit workflow definitions.
2. Run a small set of workflow step types asynchronously.
3. Track each execution and its step-level status.
4. Retry eligible failures safely and make duplicate effects less likely through idempotency controls.
5. Inspect execution history and diagnostic logs.
6. Add one AI-powered document-processing workflow with structured output and an approval step.

## Development principles

- Ship in small, reviewable increments.
- Keep the database as the source of truth for workflow and execution history.
- Make failure paths observable and test them deliberately.
- Never log credentials or unredacted secrets.
- Document measured performance only after running reproducible benchmarks.

## Repository layout (planned)

```text
flowpilot/
├── backend/       # Go API, domain logic, worker and tests
├── frontend/      # React + TypeScript application
├── deploy/        # Local and deployment configuration
├── docs/          # Architecture, decisions and development notes
└── .github/       # CI workflows
```

## Local development

The application scaffold is not implemented yet. Follow the commit history and roadmap as features are added. Setup instructions will be added with the first runnable backend and frontend commits.

## Roadmap

See [docs/roadmap.md](docs/roadmap.md) for the incremental implementation plan and intended commit sequence. See [docs/architecture.md](docs/architecture.md) for initial architecture and reliability decisions.

## License

The project license will be selected before public release.

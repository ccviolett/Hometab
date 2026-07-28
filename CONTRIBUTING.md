# Contributing to Hometab

Thanks for your interest in contributing! This document explains how to set up the
project, the development workflow, and the conventions we follow.

## Development setup

Prerequisites: the Go version declared in `apps/backend-go-fiber/go.mod`, **Node.js 22+**, and **npm**.

```bash
# 1. Backend (terminal 1, with CORS for the frontend dev server)
cd apps/backend-go-fiber
HOME_CORS_ENABLED=true go run ./cmd/server/

# 2. Frontend (terminal 2, hot reload on http://localhost:5173)
cd apps/frontend-react-vite
npm install
npm run dev
```

To build the production single binary (frontend embedded):

```bash
cd apps/backend-go-fiber
make build            # → bin/hometab
./bin/hometab        # http://127.0.0.1:52173
```

Before opening a pull request, run the complete local quality gate from the repository root:

```bash
make check             # lint, format/vet, frontend tests/build, Go race + E2E tests
```

## Proposing changes

For anything non-trivial, open an issue first to discuss the goal and approach.
Then submit a PR against `main` with a clear description of what changed and why.
Smaller, focused PRs are easier to review and merge.

## Conventions

- **i18n**: every user-visible string must go through `t()` (react-i18next). Do **not**
  hardcode UI text. See [`src/i18n/`](./apps/frontend-react-vite/src/i18n).
- **TypeScript**: lint, Vitest, and `tsc -b` must pass.
- **Backend**: formatting, `go vet ./...`, and `go test -race ./...` (including E2E) must pass.
- **Secrets**: never commit keys, cookies, `.env`, or internal/company assets.

## Pull requests

- Open PRs against `main`.
- Fill in the PR template; link the related issue.
- CI must be green (frontend build + backend build/test).

## Releases

Maintainers cut releases by pushing a `v*` tag. GitHub Actions builds and packages
binaries for macOS, Linux, and Windows with SHA-256 checksums. See
`.github/workflows/release.yml`.

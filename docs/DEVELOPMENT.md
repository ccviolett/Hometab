# Development

## Prerequisites

- Go version declared in `apps/backend-go-fiber/go.mod`
- Node.js 22 or newer
- npm

## Setup

```bash
make setup
```

## Run With Hot Reload

Terminal 1:

```bash
cd apps/backend-go-fiber
make run-dev
```

Terminal 2:

```bash
cd apps/frontend-react-vite
npm run dev
```

The frontend runs on `http://localhost:5173` and proxies API requests to `http://127.0.0.1:52173`.

## Build The Production Binary

```bash
make build
./apps/backend-go-fiber/bin/hometab
```

The build compiles the frontend, copies it into the Go embed directory, and creates the binary.
Running the binary opens Hometab in the default browser after the server is ready. Use `--no-open` when testing startup without browser interaction.

## Quality Gate

```bash
make check
```

This runs frontend linting and tests, Go formatting and vet checks, race-enabled backend tests, and production builds.

## Repository Layout

```text
apps/backend-go-fiber/    Go server, SQLite, embedded frontend
apps/frontend-react-vite/ React application
docs/                     Architecture and development documentation
.github/                   Community templates
.github/workflows/        CI and release workflows
```

## Release

Create a `vX.Y.Z` tag. GitHub Actions builds and packages binaries for macOS, Linux, and Windows and publishes them with checksums in a GitHub Release.

```bash
git tag v1.0.0
git push origin v1.0.0
```

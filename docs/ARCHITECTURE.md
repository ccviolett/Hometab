# Hometab Architecture

Hometab is a local-first, single-user start page distributed as one Go binary.

## Runtime

```text
Browser
  -> Fiber HTTP server
     -> REST handlers
        -> services
           -> GORM repositories
              -> SQLite
     -> embedded React application
     -> managed icon files
     -> automatic restore-point backups
```

The server listens on `127.0.0.1:52173` by default. The frontend is built with Vite and embedded with `go:embed`; production deployment does not require a separate web server.

## Data

SQLite is the source of truth for groups, links, link flows, settings, search engines, saved requests, and icon metadata. Link flows use a many-to-many membership table so group order and flow order remain independent.

Runtime files live in the operating system user configuration directory:

- macOS: `~/Library/Application Support/Hometab/`
- Linux: `~/.config/Hometab/`
- Windows: `%AppData%\Hometab\`

Existing installations under the previous `Home` directory are detected automatically and remain usable.

## Security Boundary

Hometab has no built-in authentication and assumes a trusted local user. Binding to a non-loopback address requires an authenticated HTTPS reverse proxy.

Saved HTTP requests use a constrained transport with URL, DNS, redirect, timeout, response-size, and concurrency validation. Execution is disabled by default on non-loopback listeners unless explicitly enabled.

## Backup And Restore

Exports are versioned ZIP archives containing structured JSON and managed icon files. Import supports:

- `merge`: add missing records without replacing existing data;
- `replace`: validate the archive, create a restore point, then replace data transactionally.

The five most recent automatic restore points are retained.

## Frontend State

TanStack Query owns server state. Local UI preferences use local storage or small Zustand stores. Demo-only adapters are isolated under `src/lib/demo` and are never used by the production binary.

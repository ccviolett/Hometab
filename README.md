# Hometab

> A local-first personal start page delivered as a single executable.

[![CI](https://github.com/ccviolett/Hometab/actions/workflows/ci.yml/badge.svg)](https://github.com/ccviolett/Hometab/actions/workflows/ci.yml)
[![Pages](https://github.com/ccviolett/Hometab/actions/workflows/pages.yml/badge.svg)](https://ccviolett.github.io/Hometab/)

[Project page](https://ccviolett.github.io/Hometab/?lang=en) · [中文说明](./README.zh-CN.md)

![Hometab dashboard with grouped links and search](./site/assets/hometab-demo.png)

Hometab combines grouped links, multi-engine search, saved HTTP actions, wallpapers, icon management, and versioned backups in one self-hosted application. The React frontend is embedded in the Go server, and all persistent data stays in a local SQLite database.

## Features

- **Links and flows**: groups, drag-and-drop ordering, and reusable link flows.
- **Search**: Google, Bing, DuckDuckGo, and custom engines.
- **Saved requests**: structured headers, query parameters, bodies, and JSON-path result extraction.
- **Request safety**: URL, DNS, redirect, timeout, response-size, and concurrency controls.
- **Wallpapers**: Bing daily or random images, custom URLs, local fallback, and caching.
- **Domain icons**: automatic discovery, upload, replacement, retry, conflict resolution, and local storage.
- **Backup and restore**: versioned ZIP export, merge import, replace restore, and automatic restore points.
- **Bilingual UI**: Simplified Chinese and English.
- **Focused mode**: double-click the wallpaper to switch between the dashboard and search-only view.

## Requirements

For a source build:

- Go version declared in `apps/backend-go-fiber/go.mod`
- Node.js 22 or newer
- npm

## Build And Run

```bash
git clone https://github.com/ccviolett/Hometab.git
cd Hometab
make setup
make build
./apps/backend-go-fiber/bin/hometab
```

Hometab opens the application in your default browser after the server is ready. The default URL is:

```text
http://127.0.0.1:52173
```

If the default port is occupied, Hometab automatically tries subsequent ports and opens the actual URL. Use `--no-open` when running in a background or automated environment.

## Data Location

Override the database path with `HOME_DATABASE_PATH`. Default locations are:

- macOS: `~/Library/Application Support/Hometab/data.db`
- Linux: `~/.config/Hometab/data.db`
- Windows: `%AppData%\Hometab\data.db`

Existing installations under the previous `Home` directory are detected automatically.

## macOS Background Service

On macOS, open **Settings > Login startup** to register or remove Hometab for
the next user login. The current process keeps running when login startup is
disabled. The same lifecycle operations remain available from the CLI:

```bash
./apps/backend-go-fiber/bin/hometab --install
./apps/backend-go-fiber/bin/hometab --status
./apps/backend-go-fiber/bin/hometab --stop
./apps/backend-go-fiber/bin/hometab --start
./apps/backend-go-fiber/bin/hometab --uninstall
```

Uninstalling the service preserves user data.

## Development

```bash
# Terminal 1
cd apps/backend-go-fiber
make run-dev

# Terminal 2
cd apps/frontend-react-vite
npm run dev
```

Run the complete quality gate before committing:

```bash
make check
```

## Security

Hometab is a trusted, single-user application without built-in authentication. Keep the default loopback binding. Remote access requires an authenticated HTTPS reverse proxy.

Saved-request execution is disabled by default when the server binds to a non-loopback address. See [SECURITY.md](./SECURITY.md) for the complete deployment boundary.

## Documentation

- [Architecture](./docs/ARCHITECTURE.md)
- [Development](./docs/DEVELOPMENT.md)
- [Login startup](./docs/AUTOSTART.md)
- [Contributing](./CONTRIBUTING.md)
- [Security](./SECURITY.md)
- [Changelog](./CHANGELOG.md)

## Release

Create and push a semantic release tag:

```bash
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions builds macOS, Linux, and Windows archives with SHA-256 checksums and publishes a GitHub Release.

## License

[MIT](./LICENSE)

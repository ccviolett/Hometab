# Hometab Server

Go Fiber server with GORM, SQLite, and an embedded React frontend.

## Commands

```bash
make run-dev
make test
make build
./bin/hometab
```

The server listens on `127.0.0.1:52173` by default, automatically tries subsequent ports when necessary, and opens the actual URL in the system browser after it is ready. Pass `--no-open` for background or automated use.

## Configuration

Configuration can be provided through `config.yaml` or `HOME_`-prefixed environment variables.

| Setting | Environment variable | Default |
|---|---|---|
| `server.host` | `HOME_SERVER_HOST` | `127.0.0.1` |
| `server.port` | `HOME_SERVER_PORT` | `52173` |
| `database.path` | `HOME_DATABASE_PATH` | OS user config directory |
| CORS | `HOME_CORS_ENABLED` | `false` |
| external requests | `HOME_EXTERNAL_REQUEST_ENABLED` | loopback listeners only |

Default database locations:

- macOS: `~/Library/Application Support/Hometab/data.db`
- Linux: `~/.config/Hometab/data.db`
- Windows: `%AppData%\Hometab\data.db`

Existing installations under `Home` are detected automatically.

## macOS Service

```bash
./bin/hometab --install
./bin/hometab --status
./bin/hometab --stop
./bin/hometab --start
./bin/hometab --uninstall
```

New installations use `~/Library/Application Support/Hometab`, `~/Library/Logs/Hometab`, and the `com.species.hometab` LaunchAgent. Existing installations keep their current paths.

See [the architecture document](../../docs/ARCHITECTURE.md) for component boundaries.

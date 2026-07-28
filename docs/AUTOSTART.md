# Login Startup Specification

## Scope

Hometab can register its single executable to start when the current macOS user
signs in. The embedded React settings UI controls this through the local Go API.
No native macOS GUI or privileged helper is required.

Windows and Linux report the feature as unsupported in this phase. Their
implementations must keep the same API contract when added later.

## User Experience

Settings exposes **Start Hometab at login**, which registers or removes the
current user's login item. The LaunchAgent always starts with `--no-open`.
Opening a browser is intentionally outside this phase because a `KeepAlive`
restart must not create additional browser tabs.

Changes apply to the next login. Enabling must not launch a second Hometab
process in the current session. Disabling must not stop the process serving the
current page. When a previously loaded LaunchAgent remains active after being
disabled, the UI reports that the current process continues to run.

## API

### `GET /api/system/startup`

Returns the operating system state, not a database preference:

```json
{
  "platform": "darwin",
  "supported": true,
  "enabled": true,
  "active": false
}
```

- `enabled` means a login registration exists.
- `active` means the registered service is currently loaded by the OS.

### `PUT /api/system/startup`

```json
{
  "enabled": true
}
```

The operation is idempotent and returns the resulting status. Unsupported
platforms return `501`. Requests from non-loopback clients return `403`.

## macOS Registration

The backend manages a per-user LaunchAgent:

- Label: `com.species.hometab`
- Plist: `~/Library/LaunchAgents/com.species.hometab.plist`
- Binary: `~/Library/Application Support/Hometab/bin/hometab`
- Logs: `~/Library/Logs/Hometab/`

Registration copies the currently running executable atomically and writes a
LaunchAgent with `RunAtLoad`, `KeepAlive`, and the `--no-open` argument.

Removing login startup deletes the plist but preserves the copied executable,
logs, database, and current process. The explicit CLI `--uninstall` command
continues to stop the loaded service and remove its copied executable.

Existing `com.powerbase.home` installations remain detectable and manageable.

## Data and Security Boundaries

Login startup is machine state and must not be stored in SQLite, exported in a
backup, or restored onto another machine. The API reads the LaunchAgent state
on every request.

System endpoints require a loopback client, loopback Host, and (when present) a
loopback Origin, even if Hometab is configured to listen on a non-loopback
interface. No endpoint accepts arbitrary paths, labels, executables, or command
arguments from the frontend.

## Acceptance Criteria

- macOS status reflects the plist and loaded LaunchAgent independently.
- Enabling copies the current binary and writes the expected plist.
- Repeating enable or disable succeeds without changing semantics.
- Enabling does not start another process in the current session.
- Disabling does not terminate the current process or remove user data.
- CLI service commands and HTTP configuration share one Go implementation.
- Non-macOS builds compile and return `supported: false`.
- Settings handles loading, mutation, unsupported, active, and error states in
  English and Simplified Chinese.

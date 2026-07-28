# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in Hometab, please **do not open a public
issue**. Instead, report it privately so we can address it before disclosure.

Contact the maintainer privately through the repository host. If a private channel is
not available, ask for security contact details without including vulnerability specifics
in a public issue. See `CODEOWNERS` for the current maintainer handle.

Please include:

- A description of the vulnerability and its impact
- Steps to reproduce (or a proof of concept)
- Affected version(s) / commit(s)

We will acknowledge your report and aim to provide a remediation plan within a reasonable
timeframe. Credit will be given (with your consent) in the release notes.

## Deployment Boundary

Hometab is designed as a trusted, single-user application and listens on `127.0.0.1` by
default. Its API has no built-in authentication. Do not bind it to `0.0.0.0`, expose it to
a LAN, or publish it to the internet unless an authenticated reverse proxy and HTTPS are
in front of it.

The saved-request and icon features intentionally make HTTP requests to user-provided
URLs, including loopback, private-network, and public services. Saved-request execution
is disabled by default when Hometab binds to a non-loopback address; enabling it requires
the explicit `HOME_EXTERNAL_REQUESTS_ALLOW_REMOTE_EXECUTION=true` setting. Target
validation, redirect checks, timeouts, response-size limits, and concurrency limits reduce
risk but do not replace authentication. Treat every user with API access as fully trusted.
Never import an untrusted backup because importing changes application data.

## Supported Versions

Only the latest released `v*` version receives security fixes. Older versions are
supported at the maintainer's discretion.

# Configuration

Every knob `ai-beacon` exposes. The [main README](../README.md) and the in-app setup guide (rocket icon, top bar) cover the variables you need to start. Reach here when you outgrow the defaults.

- [Agent (client) configuration](#agent-client-configuration)
- [Server configuration](#server-configuration)
- [Helm values](#helm-values)
- [Config file (`config.toml`)](#config-file-configtoml)
- [Data directory layout](#data-directory-layout)

## Agent (client) configuration

These apply to the machine running the wrapped coding agent (`ai-beacon session -- claude`, or the auto-installed hooks).

| Variable | Purpose | Default |
|----------|---------|---------|
| `AI_BEACON_URL` | Dashboard server URL. | `http://localhost:8080` |
| `AI_BEACON_AUTH_TOKEN` | Bearer token for agent → dashboard heartbeats. The dashboard generates this on first boot; copy it from the setup guide. | _(required)_ |
| `AI_BEACON_DEVICE_NAME` | Friendly device name shown on session cards and the device-grouped grid. | hostname |
| `AI_BEACON_PROJECTS_DIR` | Base directories of your repositories. Enables the dashboard's "spawn new session" flow and worktree workflows. Accepts a single path or a list joined by the OS path separator (`:` on Unix, `;` on Windows) for multiple roots — e.g. `~/work:~/oss`. Disabled when unset. | _(unset)_ |
| `AI_BEACON_WORKTREE_LOCATION` | Where new worktrees are created relative to the source repo. `sibling` (default) places them next to the repo; `subdirectory` places them inside it. | `sibling` |
| `AI_BEACON_LOG_FILE` | Override the log file location. Absolute path, or a bare filename rooted under `<data-dir>/logs/`. | _(unset, logs to data dir)_ |

Set these in your shell profile (`~/.zshrc`, `~/.bashrc`, etc.) so every session inherits them:

```bash
export AI_BEACON_URL=https://ai-beacon.example.com
export AI_BEACON_AUTH_TOKEN=…
export AI_BEACON_DEVICE_NAME=macbook
export AI_BEACON_PROJECTS_DIR=~/work:~/oss
```

`ai-beacon install` persists `AI_BEACON_URL` and `AI_BEACON_AUTH_TOKEN` into the config file, so hooks keep working even if the env vars aren't exported from the shell that launches the agent. See [Connecting an agent](connect-agent.md).

**Symlinks and case.** The agent resolves symlinks before validating each `cwd` it accepts from the dashboard, so a symlink that points outside a configured root is rejected even if it lives lexically inside one. Path comparison matches the host filesystem: case-insensitive on macOS and Windows, case-sensitive on Linux — a dashboard-echoed `cwd` that differs only in case from the configured root validates on macOS/Windows and is rejected on Linux.

### Session-command flags

`ai-beacon session -- <agent-cmd>` accepts these (most have an env-var fallback above):

| Flag | Purpose |
|------|---------|
| `--dashboard-url <url>` | Override `$AI_BEACON_URL` for this run. |
| `--device <name>` | Override `$AI_BEACON_DEVICE_NAME`. |
| `--project <dir>` | Project directory shown on the session card. Defaults to the current working directory. |
| `--projects-dir <dirs>` | Override `$AI_BEACON_PROJECTS_DIR`. Accepts a single path or a list joined by `:` (Unix) / `;` (Windows) for multiple roots. |
| `--worktree-location sibling\|subdirectory` | Override `$AI_BEACON_WORKTREE_LOCATION`. |
| `--session-id <uuid>` | Pre-assign a session ID. By default a UUID is generated. |
| `--log-file <path>` | Override `$AI_BEACON_LOG_FILE`. |
| `--capture-bytes <dir>` | Capture raw byte streams under `<dir>`: `agent-pty.bin` (pre-compositor PTY chunks), `local-stdout.bin` (post-compositor terminal output), `events.jsonl` (timing/resize), and `manifest.json`. Useful when reporting terminal rendering issues (especially Windows ConPTY); share the directory with maintainers so the session can be replayed offline. Each session overwrites `<dir>`, so this flag fits a single targeted capture; for recurring capture use `--capture-bytes-auto-routed`. |
| `--capture-bytes-auto-routed` | Capture each session into its own subdirectory at `<data-dir>/sessions/agent-session-<id>.capture/`, keeping the 45 most recent runs (older runs are pruned automatically). Use when you want to leave capture on across several sessions (add the flag to your launcher script, reproduce a few sessions, then remove it). Wins over `--capture-bytes` if both are set. |

Flags marked `(internal)` in `--help` are populated by the dashboard when it spawns a session and aren't intended for direct invocation.

## Server configuration

These apply to the machine running `ai-beacon server` (or the container / pod).

| Variable | Purpose | Default |
|----------|---------|---------|
| `AI_BEACON_AUTH_TOKEN` | Agent bearer token the server validates. When unset, the server generates a 64-hex-char token and writes it to `<data-dir>/token`. | _(auto-generated)_ |
| `AI_BEACON_AUTH_PASSWORD` | Browser login password (default auth mode). When unset, generated and written to `<data-dir>/password`. | _(auto-generated)_ |
| `AI_BEACON_DATA_DIR` | Where the server keeps the auth token, password, logs, and (future) session history. | `~/.config/ai-beacon` |
| `AI_BEACON_ALLOWED_USERS` | Comma-separated allowlist of usernames permitted to sign in. Required by OIDC and proxy-header modes. | _(unset)_ |
| `AI_BEACON_OIDC_ISSUER` | OIDC issuer URL (e.g. `https://accounts.google.com`). | _(unset)_ |
| `AI_BEACON_OIDC_CLIENT_ID` | OIDC client ID. | _(unset)_ |
| `AI_BEACON_OIDC_CLIENT_SECRET` | OIDC client secret. Use `--oidc-client-secret-file` instead when running on shared hosts to keep the secret out of `argv`. | _(unset)_ |
| `AI_BEACON_OIDC_SCOPES` | Comma- or space-separated OIDC scopes. `openid` is always included. | `openid,profile,email` |
| `AI_BEACON_OIDC_REDIRECT_URL` | OIDC callback URL registered with your IdP. | _(unset)_ |
| `AI_BEACON_OIDC_DISPLAY_NAME` | Label rendered on the "Sign in with X" button. | Derived from issuer host |

> OIDC env vars also accept the Hugging Face Spaces aliases (`OPENID_PROVIDER_URL`, `OAUTH_CLIENT_ID`, `OAUTH_CLIENT_SECRET`, `OAUTH_SCOPES`) as a fallback. See [Authentication](auth.md).

### `ai-beacon server` flags

| Flag | Purpose | Default |
|------|---------|---------|
| `-a, --address <addr>` | Listen address (e.g. `:8080`, `127.0.0.1:9000`). | `:8080` |
| `--auth <mode>` | `""` (password, default), `none`, `proxy-header`, or `oidc`. | `""` |
| `--auth-token <token>` | Agent bearer token (overrides `$AI_BEACON_AUTH_TOKEN`). | _(env / generated)_ |
| `--password-file <path>` | Read the browser password from a file instead of env / data dir. | _(unset)_ |
| `--data-dir <path>` | Override `$AI_BEACON_DATA_DIR`. | `~/.config/ai-beacon` |
| `--trusted-proxies <cidrs>` | CIDRs whose immediate-peer requests are trusted to relay `X-Forwarded-*`. Required by `--auth=proxy-header`. | _(empty, trust none)_ |
| `--allowed-users <names>` | Sign-in allowlist (overrides `$AI_BEACON_ALLOWED_USERS`). Required by `--auth=proxy-header` and `--auth=oidc`. | _(env / unset)_ |
| `--oidc-issuer`, `--oidc-client-id`, `--oidc-client-secret`, `--oidc-client-secret-file`, `--oidc-scopes`, `--oidc-redirect-url`, `--oidc-display-name` | OIDC configuration. Each falls back to the matching `AI_BEACON_OIDC_*` env var. See [Authentication](auth.md). | — |
| `--log-file <path>` | Override `$AI_BEACON_LOG_FILE`. | _(unset)_ |

## Helm values

The Helm chart's [`README.md`](../charts/ai-beacon/README.md) is the source of truth — auto-generated from `values.yaml`, it covers every value with its default and full description. Don't duplicate it; consult it directly.

A few footguns worth surfacing here:

- **`replicaCount` must stay at `1`.** Session state is in-process; multiple replicas shard sessions across pods and break dashboards/terminals. Horizontal scaling is not currently supported.
- **`persistence.enabled=true` is required when auth values are auto-generated.** Without a volume the auth token regenerates on every restart and silently invalidates installed agent hooks. The chart leaves the PVC behind on `helm uninstall` (`helm.sh/resource-policy: keep`) so credentials survive a reinstall.
- **`oauthProxy.enabled=true` requires `allowedUsers`.** Without an allowlist the proxy admits every cluster user with OAuth access. See [Authentication § OpenShift OAuth Proxy](auth.md#openshift-oauth-proxy).

## Config file (`config.toml`)

`ai-beacon install` writes to `<data-dir>/config.toml`. The agent and server also read it at startup. You rarely need to edit it by hand — use it to persist values you'd otherwise re-export on every session.

```toml
url         = "https://ai-beacon.example.com"
token_path  = "/Users/you/.config/ai-beacon/token"
binary_path = "/Users/you/.local/bin/ai-beacon"

# Single base directory (legacy):
# projects_dir = "/Users/you/projects"
# Multiple base directories (path-list semantics, but as a TOML array):
projects_dirs = ["/Users/you/work", "/Users/you/oss"]

[workflow.implement_issue]
prompt = """
Implement this issue: {issue_url} ({issue_title}).
Ask before coding.
"""

[workflow.review_pr]
prompt = "/mn-review {pr_url}"
```

`[workflow.*]` sections override the built-in workflow prompt templates — see [GitHub integration § Workflow prompts](github.md#workflow-prompts).

### Multiple dashboards

Hook events (SessionStart/Stop/Notify/StatusLine) can fan out to additional dashboards in parallel by adding `[[dashboard]]` entries. The top-level `url` / `token_path` remain the **singleton** (the dashboard `ai-beacon install`, the session wrapper, and `ai-beacon heartbeat` work with). Secondary dashboards receive **heartbeats only** — they observe the live session metadata (model, tokens, context %, state, task, branch, PR) but cannot attach a terminal, spawn, kill, or rename the session.

```toml
url        = "https://butler.example.com"
token_path = "/Users/you/.config/butler/token"

[[dashboard]]
url        = "http://localhost:8080"
token_path = "/Users/you/.config/ai-beacon/token"

[[dashboard]]
url        = "https://company.example.com"
token_path = "/Users/you/.config/company/token"
```

Each entry needs its own dashboard URL and a readable `token_path`. Entries with a missing or unreadable token file are skipped with a warning at hook time — secondaries are never posted unauthenticated. Duplicate URLs (case difference, trailing slash) fold to a single target.

## Data directory layout

`AI_BEACON_DATA_DIR` (default `~/.config/ai-beacon`) holds:

| Path | Contents |
|------|----------|
| `config.toml` | Persisted CLI configuration (see above). |
| `token` | Agent bearer token (file mode `0600`). Read by the server on boot; written by `ai-beacon install`. |
| `password` | Browser login password (file mode `0600`). Server-side only. |
| `logs/` | Log files. `ai-beacon session …` writes here unless `--log-file` overrides. |

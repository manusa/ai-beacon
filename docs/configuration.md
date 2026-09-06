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
| `AI_BEACON_BAR_COST` | Whether the local terminal status bar shows the session cost. Set to `off`, `0`, or `false` to hide it (context usage is unaffected) — handy when pairing, screen-sharing, or recording. `--hide-cost` overrides this. | show |
| `AI_BEACON_LOG_FILE` | Override the log file location. Absolute path, or a bare filename rooted under `<data-dir>/logs/`. | _(unset, logs to data dir)_ |
| `AI_BEACON_JIRA_URL` | Jira instance base URL (e.g. `https://acme.atlassian.net`, or `https://jira.example.com/jira` for a Data Center install under a subfolder). Enables pasting a Jira ticket into the *Implement Issue* workflow. Disabled when unset. Falls back to a generic `JIRA_URL` if unset — but it is **never** inferred from the ticket URL you paste. | _(unset)_ |
| `AI_BEACON_JIRA_EMAIL` | Your Atlassian account email. Set it for **Atlassian Cloud** (the API token is used as a Basic-auth password); leave it unset for **Data Center**, where the token is sent as a Bearer PAT. Falls back to a generic `JIRA_USER` if unset. | _(unset)_ |
| `AI_BEACON_JIRA_TOKEN` | Jira API token. Prefer `jira_token_path` in the config file so the secret stays in a file rather than your shell environment. Falls back to a generic `JIRA_TOKEN` if unset. | _(unset)_ |

Set these in your shell profile (`~/.zshrc`, `~/.bashrc`, etc.) so every session inherits them:

```bash
export AI_BEACON_URL=https://ai-beacon.example.com
export AI_BEACON_AUTH_TOKEN=…
export AI_BEACON_DEVICE_NAME=macbook
export AI_BEACON_PROJECTS_DIR=~/work:~/oss
```

`ai-beacon install` persists `AI_BEACON_URL` and `AI_BEACON_AUTH_TOKEN` into the config file, so hooks keep working even if the env vars aren't exported from the shell that launches the agent. See [Connecting an agent](connect-agent.md).

**Symlinks and case.** The agent resolves symlinks before validating each `cwd` it accepts from the dashboard, so a symlink that points outside a configured root is rejected even if it lives lexically inside one. Path comparison matches the host filesystem: case-insensitive on macOS and Windows, case-sensitive on Linux — a dashboard-echoed `cwd` that differs only in case from the configured root validates on macOS/Windows and is rejected on Linux.

### Jira tickets

With `AI_BEACON_JIRA_URL` and `AI_BEACON_JIRA_TOKEN` set (plus `AI_BEACON_JIRA_EMAIL` on Atlassian Cloud — leave it unset on Data Center), pasting a ticket URL (`https://acme.atlassian.net/browse/ABC-123`) into the *Implement Issue* workflow starts a session seeded with the whole ticket: description, comments, links, subtasks and metadata. Jira is an issue **tracker**, not a git host — the branch and the pull/merge request still go to the project's own GitHub or GitLab remote, so the project needs an authenticated `gh` or `glab` just as it does for a GitHub/GitLab issue.

A few things worth knowing:

- **Mint a fresh API token.** Atlassian has been force-expiring older unscoped tokens; an expired one fails with no useful message. Create one at *Account settings → Security → API tokens*.
- **The ticket is read once, at spawn.** Nothing in the agent session re-reads Jira, and there is no Jira CLI there, so the snapshot is all it has. If it needs something the ticket does not contain, it is told to ask you rather than go looking.
- **`jira_token_path` keeps the token out of the agent's reach; the env var does not.** The token never goes on the session's command line either way. But a session inherits its parent's environment and re-runs your login profile, so a `AI_BEACON_JIRA_TOKEN` you exported is readable from inside the session; a token in a file the wrapper reads is not.
- **Already export `JIRA_*` for another tool?** If the `AI_BEACON_JIRA_*` vars and the config file are all unset, ai-beacon falls back to the generic `JIRA_URL`, `JIRA_USER` (read as your account email), and `JIRA_TOKEN` many Jira tools use. The `AI_BEACON_`-prefixed vars and the config file always take precedence. The base URL is still required from one of these sources — it is **never** inferred from the ticket URL you paste, so the token only ever reaches the instance you configured.
- **Paste the URL; you can't search Jira here.** The typeahead searches GitHub and GitLab only, and the field says so if you type a bare ticket key.
- **If the read fails, the session still starts.** A network blip, an expired token or a deleted ticket doesn't block the spawn — you get a *"Couldn't read ABC-123 from Jira"* notice, and the agent opens by asking you for the ticket rather than guessing at it.

If the dialog says the agent has no Jira credentials, the agent process is not seeing your settings. Either export the variables in the shell profile that launches your agent and restart the session, **or** add `jira_url` and `jira_token_path` (plus `jira_email` on Atlassian Cloud) to the config file below — the config file is re-read each time the dialog asks, so that route needs no restart.

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
| `--hide-cost` | Hide the session cost from the local terminal status bar (context usage is unaffected). Useful when pairing, screen-sharing, or recording. Overrides `$AI_BEACON_BAR_COST`. |

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

# Jira as an issue source for the Implement Issue workflow. The token is a
# *path*, never the value: it is read at spawn time and never handed to the
# agent process. Omit jira_email on a Data Center instance (Bearer PAT).
# jira_url        = "https://acme.atlassian.net"
# jira_email      = "you@example.com"
# jira_token_path = "/Users/you/.config/ai-beacon/jira-token"

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
| `jira-token` | Jira API token, if you point `jira_token_path` here. Create it yourself with mode `0600`. |
| `password` | Browser login password (file mode `0600`). Server-side only. |
| `logs/` | Log files. `ai-beacon session …` writes here unless `--log-file` overrides. |

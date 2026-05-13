# Connecting an agent

The in-app **setup guide** (rocket icon, top bar) walks new users through:

1. Downloading the `ai-beacon` binary for your OS.
2. Setting `AI_BEACON_URL` and `AI_BEACON_AUTH_TOKEN`, then running `ai-beacon install`.
3. Verifying with `ai-beacon session -- claude` and watching the session appear on the dashboard.

This page picks up after that. It explains what `install` actually changes, how to rotate or update credentials without re-walking the guide, and what to check when a session doesn't show up.

- [What `ai-beacon install` does](#what-ai-beacon-install-does)
- [Updating the dashboard URL or token](#updating-the-dashboard-url-or-token)
- [Running the wrapper without `install`](#running-the-wrapper-without-install)
- [Uninstalling](#uninstalling)
- ["Session not appearing" troubleshooting](#session-not-appearing-troubleshooting)
- [Logs](#logs)

## What `ai-beacon install` does

`install` is idempotent — it can be re-run any time. It performs three steps:

1. **Resolves the dashboard URL and auth token.** Precedence: `--url` flag → `$AI_BEACON_URL` → existing `config.toml` → `http://localhost:8080`. The token follows the same chain, ending at `$AI_BEACON_AUTH_TOKEN`. If a token is provided, it's written to `<data-dir>/token` (mode `0600`).
2. **Writes `config.toml`.** Stored at `<data-dir>/config.toml` (default `~/.config/ai-beacon/config.toml`). Records the URL, the path to the token file, and the absolute path of the `ai-beacon` binary itself.
3. **Installs hooks for each registered agent plugin.** For Claude Code, this means editing `~/.claude/settings.json` so each `claude` invocation transparently wraps itself in `ai-beacon session`. Hooks are tagged with the `# ai-beacon-managed` marker so `uninstall` can find and remove them without disturbing your own edits.

After install, running the agent normally (`claude`) is equivalent to `ai-beacon session -- claude`. No env vars need to be exported in the agent's shell — the hook reads them from `config.toml`.

> The `--yes` flag skips the consent prompt. Useful for CI / scripted setup.

## Updating the dashboard URL or token

Both values live in `<data-dir>/config.toml`. To rotate either, re-run `install`:

```bash
export AI_BEACON_URL=https://ai-beacon.example.com
export AI_BEACON_AUTH_TOKEN=$(pbpaste)  # or however you got the new token
ai-beacon install --yes
```

Hooks pick up the new values on the next session. No restart of running sessions is needed; long-running sessions keep their cached token until they reconnect.

> The current token is stored at `<data-dir>/token`, owned by your user, mode `0600`. To inspect: `cat ~/.config/ai-beacon/token`. Never check this file in.

## Running the wrapper without `install`

If you'd rather not modify the agent's settings, skip `install` and call the wrapper explicitly:

```bash
export AI_BEACON_URL=https://ai-beacon.example.com
export AI_BEACON_AUTH_TOKEN=…
ai-beacon session -- claude
```

The trade-off is reach: only sessions you launch through this command appear on the dashboard. Direct `claude` invocations stay invisible.

Per-invocation overrides work too:

```bash
ai-beacon session --dashboard-url https://other.example.com --device server-1 -- claude
```

See [Configuration § Session-command flags](configuration.md#session-command-flags) for the full list.

## Uninstalling

```bash
ai-beacon uninstall
```

By default this removes hooks and leaves the binary and data directory in place — handy if you're just pausing.

| Flag | Effect |
|------|--------|
| `--remove-binary` | Also delete the `ai-beacon` binary at its installed location. |
| `--purge-data` | Also delete the entire data directory (`config.toml`, token, password, logs). Cannot be undone. |
| `--purge` | Shorthand for both of the above. |
| `--yes` | Skip the interactive confirmations. Without this, removal of the binary and data dir each prompt individually with a default of `No`. |

After `--purge`, the next `ai-beacon install` starts from a clean slate.

## "Session not appearing" troubleshooting

Walk these in order — they cover ~90% of cases.

1. **Confirm the dashboard URL is reachable from the agent host.** From the machine running `claude`:

   ```bash
   curl -fsS "$AI_BEACON_URL/api/v1/health" && echo OK
   ```

   A `401` or `OK` means the network path works. A `404` means the URL is wrong. A connection error means firewall, DNS, or VPN.

2. **Confirm the token is current.**

   ```bash
   cat ~/.config/ai-beacon/token
   ```

   Compare it to the **Install** step of the setup guide on the dashboard. The most common cause of "everything looks fine but nothing shows up" is a container deployment with no persistent volume — every restart regenerates a fresh token and silently invalidates the one baked into your installed hooks. Either re-run `ai-beacon install` with the new token, or — preferably — mount `/data` to a persistent volume so the token survives restarts (see the [main README](../README.md)'s container instructions and the chart's `persistence.enabled` value).

3. **Confirm the hook ran.** Look for the `# ai-beacon-managed` marker in `~/.claude/settings.json`. If it's missing, `install` didn't run or your editor stripped it — re-run `ai-beacon install`.

4. **Look at the wrapper logs.** See [Logs](#logs) below. A session that fails to authenticate writes an unambiguous error there before exiting.

5. **Run the wrapper directly to bypass hooks.**

   ```bash
   AI_BEACON_URL=… AI_BEACON_AUTH_TOKEN=… ai-beacon session -- claude
   ```

   If this works but hook-launched sessions don't, the hook isn't picking up `config.toml` — usually a stale binary path. Re-run `ai-beacon install` so `binary_path` updates.

## Logs

Logs are written to `<data-dir>/logs/` by default. The session command suppresses log output to stderr (it would garble the wrapped agent's TUI), so the log file is the only place errors land.

| Override | Behavior |
|----------|----------|
| `--log-file <path>` | Absolute path writes there directly. Bare filename roots it under `<data-dir>/logs/`. |
| `$AI_BEACON_LOG_FILE` | Same as the flag; flag wins when both are set. |

For an active session, `tail -F ~/.config/ai-beacon/logs/<session>.log` will surface heartbeat failures, auth issues, and reconnect attempts in real time.

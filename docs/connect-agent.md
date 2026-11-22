# Connecting an agent

The in-app **setup guide** (rocket icon, top bar) walks new users through:

1. Downloading the `ai-beacon` binary for your OS.
2. Setting `AI_BEACON_URL` and `AI_BEACON_AUTH_TOKEN`, then running `ai-beacon install`.
3. Verifying with `ai-beacon session -- claude` and watching the session appear on the dashboard.

This page picks up after that. It documents the download URLs for the six native binaries, explains what `install` actually changes, how to rotate or update credentials without re-walking the guide, and what to check when a session doesn't show up.

- [Native binaries](#native-binaries)
- [What `ai-beacon install` does](#what-ai-beacon-install-does)
- [Updating the dashboard URL or token](#updating-the-dashboard-url-or-token)
- [Running the wrapper without `install`](#running-the-wrapper-without-install)
- [Uninstalling](#uninstalling)
- [Windows consoles and raw escape codes](#windows-consoles-and-raw-escape-codes)
- [The terminal title shows the agent's state](#the-terminal-title-shows-the-agents-state)
- [Subagents on the session card](#subagents-on-the-session-card)
- [Initial prompts](#initial-prompts)
- [opencode sessions starting and ending](#opencode-sessions-starting-and-ending)
- [Your opencode configuration](#your-opencode-configuration)
- [opencode 2](#opencode-2)
- [A smaller browser is watching](#a-smaller-browser-is-watching)
- ["Session not appearing" troubleshooting](#session-not-appearing-troubleshooting)
- [The terminal keeps disconnecting](#the-terminal-keeps-disconnecting)
- [The status bar says "update required"](#the-status-bar-says-update-required)
- [The status bar says "rate limited" or "sign-in failing"](#the-status-bar-says-rate-limited-or-sign-in-failing)
- [Logs](#logs)

## Native binaries

The in-dashboard setup guide picks the right binary for the operating system and CPU architecture it detects in your browser. The same six artifacts are also available directly from the [`snapshot`](https://github.com/manusa/ai-beacon/releases/tag/snapshot) rolling pre-release — useful if you're installing on a headless machine, scripting fleet onboarding, or you want to verify the asset before downloading.

| OS | Architecture | Asset |
|------|------|-------|
| Linux | amd64 | [`ai-beacon-linux-amd64`](https://github.com/manusa/ai-beacon/releases/download/snapshot/ai-beacon-linux-amd64) |
| Linux | arm64 | [`ai-beacon-linux-arm64`](https://github.com/manusa/ai-beacon/releases/download/snapshot/ai-beacon-linux-arm64) |
| macOS | amd64 (Intel) | [`ai-beacon-darwin-amd64`](https://github.com/manusa/ai-beacon/releases/download/snapshot/ai-beacon-darwin-amd64) |
| macOS | arm64 (Apple silicon) | [`ai-beacon-darwin-arm64`](https://github.com/manusa/ai-beacon/releases/download/snapshot/ai-beacon-darwin-arm64) |
| Windows | amd64 | [`ai-beacon-windows-amd64.exe`](https://github.com/manusa/ai-beacon/releases/download/snapshot/ai-beacon-windows-amd64.exe) |
| Windows | arm64 | [`ai-beacon-windows-arm64.exe`](https://github.com/manusa/ai-beacon/releases/download/snapshot/ai-beacon-windows-arm64.exe) |

POSIX example:

```bash
curl -fL -o ai-beacon \
  https://github.com/manusa/ai-beacon/releases/download/snapshot/ai-beacon-linux-amd64
chmod +x ai-beacon
./ai-beacon version
```

PowerShell example:

```powershell
Invoke-WebRequest `
  -Uri https://github.com/manusa/ai-beacon/releases/download/snapshot/ai-beacon-windows-amd64.exe `
  -OutFile ai-beacon.exe
.\ai-beacon.exe version
```

`snapshot` is the rolling pre-release alias that always points at the latest build. A stable release tag will replace it once published.

Snapshot builds report `0.0.0-snapshot.<YYYYMMDDHHMMSS>`, the UTC time of the commit they were built from, and the container image built from the same commit reports the same version. Keep your wrappers on the dashboard's build by updating both together, and compare `ai-beacon version` with the dashboard's: `docker exec <container> /app/ai-beacon version`, or `kubectl exec -n ai-beacon deploy/ai-beacon -- /app/ai-beacon version` for the Helm chart.

## What `ai-beacon install` does

`install` is idempotent — it can be re-run any time. It performs three steps:

1. **Resolves the dashboard URL and auth token.** Precedence: `--url` flag → `$AI_BEACON_URL` → existing `config.toml` → `http://localhost:8080`. The token follows the same chain, ending at `$AI_BEACON_AUTH_TOKEN`. If a token is provided, it's written to `<data-dir>/token` (mode `0600`).
2. **Writes `config.toml`.** Stored at `<data-dir>/config.toml` (see [Data directory layout](configuration.md#data-directory-layout) for the per-OS location). Records the URL, the path to the token file, and the absolute path of the `ai-beacon` binary itself.
3. **Installs hooks for each registered agent plugin.** For Claude Code, this means editing `~/.claude/settings.json` so each `claude` invocation reports its session to the dashboard. Hooks are tagged with the `# ai-beacon-managed` marker so `uninstall` can find and remove them without disturbing your own edits.

After install, running the agent normally (`claude`) puts the session on the dashboard marked **Monitor only**: you see its state and details, but you can't attach its terminal from the browser. For that, start it through the wrapper: `ai-beacon session -- claude`. Neither needs env vars exported in the agent's shell; both read them from `config.toml`.

> The `--yes` flag skips the consent prompt. Useful for CI / scripted setup.

## Updating the dashboard URL or token

Both values live in `<data-dir>/config.toml`. To rotate either, re-run `install`:

```bash
export AI_BEACON_URL=https://ai-beacon.example.com
export AI_BEACON_AUTH_TOKEN=$(pbpaste)  # or however you got the new token
ai-beacon install --yes
```

Hooks pick up the new values on the next session. No restart of running sessions is needed; long-running sessions keep their cached token until they reconnect.

> The current token is stored at `<data-dir>/token`, owned by your user, mode `0600`. Inspect it with `cat`, and never check this file in.

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

## Windows consoles and raw escape codes

`ai-beacon session` works in every Windows console: Windows Terminal, the VS Code and JetBrains terminals, and the classic console window (`conhost.exe`). You get the classic window by default on Windows 10, and on Windows 11 whenever the **Default terminal application** setting is **Windows Console Host** (Settings → System → For developers → Terminal).

A classic console window shows terminal colour and cursor codes as raw text, such as `←[H←[1;29r`, unless the program running in it asks the window to interpret them. `ai-beacon session` asks for that while the session runs and puts the window back the way it found it when the session ends. If you still see raw codes like these, you are running an older `ai-beacon`: download the current binary from [Native binaries](#native-binaries), or start the session from Windows Terminal.

Windows Terminal is still recommended. [Sessions spawned from the dashboard](multi-machine.md#platform-requirements) open in a new Windows Terminal window when it is installed, and in a classic console window otherwise.

## The terminal title shows the agent's state

When the agent reports its own state, `ai-beacon session` puts a mark in front of your terminal's window title:

- `◐` while the agent works;
- `✋` while it waits for you, at a permission prompt or a question;
- `✳` when it is idle.

The mark goes away when the agent stops reporting, as Claude Code does on `/exit`.

The dashboard shows the same state. Claude Code 2.1.295 and later report it on macOS and Linux. On Windows, with `CLAUDE_CODE_DISABLE_TERMINAL_TITLE` set, or with agents that don't report their state, the title keeps the agent's own text and the dashboard state comes from the agent's hooks. While the dashboard can't be reached, the title reads `⚠ ai-beacon offline` instead.

## Subagents on the session card

When the agent hands work to its own subagents (Claude Code subagents, background agents and forks, or opencode subagent sessions), the session card counts them:

- `Idle · 3 subagents running` when the agent waits for input while its subagents keep working, or `Working · 2 subagents running` while it works too;
- `Idle · 1 result waiting` when a subagent finished while the agent was idle. It clears on the agent's next turn, or when you open the session's terminal.

The card keeps its idle or working color, the session stays under the **Idle** filter, and the **Done** notification waits until no subagent is running. Claude Code needs the `SubagentStart` and `SubagentStop` hooks: run `ai-beacon install` again after upgrading to add them. opencode picks up the updated plugin the next time it starts.

## Initial prompts

A session that starts with a prompt, such as the New session dialog's *Implement issue* and *Review PR* modes, hands it to the agent on its command line. Claude Code gets it as its last argument. opencode 1 and 2 get it as `--prompt` and send it as the first message, except opencode 2.0.0, which only fills it into the input box (see [opencode 2](#opencode-2)). `opencode run` gets it as its message. If your opencode command already passes its own `--prompt`, the session's prompt replaces it.

## opencode sessions starting and ending

A wrapped opencode session shows as idle as soon as opencode has started, so a message sent to it with `tell` is delivered at once, as for Claude Code, instead of waiting for its first prompt. This holds for opencode 1 and for an opencode 2 session with its own server (`--standalone`, which the wrapper adds itself); a wrapped opencode 2 that shares the background service follows what [opencode 2](#opencode-2) describes.

An opencode started without the wrapper gets its card at its first prompt, since opencode has no session before then. When you quit opencode, each of its sessions leaves the dashboard, earlier ones you left with `/new` included: ai-beacon's plugin starts a small background process, `ai-beacon hook watch`, that ends them when opencode exits and then exits itself. Continuing a session later (`opencode -c`, `opencode -s <id>` or the session list) brings its card back at its first prompt, or, if you continue it within two minutes of quitting, at its first prompt after those two minutes: the dashboard ignores a session for two minutes after it ended. A session open in two opencode windows stays on the dashboard until the window that reported it last quits. With opencode 2 run without `--standalone`, the plugin runs in the background service every opencode window shares, so those sessions end when the service stops (`opencode service stop`), not when a window closes. Between prompts nothing reports such a session, so a card left idle for two minutes leaves the dashboard until its next prompt, as for a Claude Code session started without the wrapper.

Both need the current plugin. A wrapped session refreshes it before opencode starts. Otherwise the plugin is refreshed the next time opencode reports to ai-beacon, or when you run `ai-beacon install` again, and opencode loads it at its next start (opencode 2: after `opencode service restart`).

## Your opencode configuration

A wrapped opencode session, opencode 1 or 2, loads the same configuration as a plain `opencode`, including a file you point `OPENCODE_CONFIG` at. The wrapper adds its messaging tools through `OPENCODE_CONFIG_CONTENT`, which opencode loads after your global, `OPENCODE_CONFIG` and project files, so its `ai-beacon` entry wins over one of yours there.

- **Turning the tools off.** Set `OPENCODE_CONFIG_CONTENT='{"mcp":{"ai-beacon":{"enabled":false}}}'` before starting the session. The wrapper then adds nothing. On opencode 2, `"disabled": true` works too, but opencode 1 rejects that key and refuses to start, so use `"enabled": false` there. `"enabled": false` in another opencode config file no longer turns them off.
- **Your own `OPENCODE_CONFIG_CONTENT`.** The wrapper adds its tools to it, so it must be plain JSON: no comments, and `{env:…}` or `{file:…}` only inside quoted strings. Otherwise the session starts without the messaging tools, and the wrapper's log says why (see [Logs](#logs)).
- **After the session ends.** Programs started from the session, such as a terminal or editor launched from its shell, inherit `OPENCODE_CONFIG_CONTENT`. Its tools entry finds the session's token through `AI_BEACON_AUTH_TOKEN_FILE`, a file the wrapper deletes when the session ends, or, if the wrapper crashed or was killed, at the next session start. Once it is gone, an opencode started from those programs loads normally but without the messaging tools, which belonged to the ended session.

## opencode 2

ai-beacon supports opencode 1 and opencode 2. opencode 2 is the `@opencode/cli` npm package, installed with `curl -fsSL https://opencode.ai/v2/install | bash` or `opencode upgrade --major`; its installer also adds an `opencode2` command. Each of these starts a session on the dashboard:

```bash
ai-beacon session -- opencode
ai-beacon session -- opencode2
ai-beacon session -- npx -y @opencode/cli
```

`ai-beacon install` writes one plugin for both versions, `~/.config/opencode/plugins/ai-beacon.ts`, and a wrapped session refreshes it when the binary carries a newer one.

**What works:** sessions started without the wrapper report their state (working, idle, waiting at a permission prompt or a question), their subagents and the notifications, and the browser terminal works for wrapped sessions. Where a wrapped session's state shows depends on how opencode runs.

**Model, tokens, cost and context.** The card shows the session's model, tokens, cost, context percentage and task, and a wrapped session's status bar its model, percentage and cost, as with opencode 1. ai-beacon's plugin reads them from opencode 2 itself as the session runs, so nothing extra starts per update. The cost and output tokens are opencode 2's own totals for the session, which count its title generation too, so they can read a little higher than opencode 1 shows for the same turns. The tokens and percentage appear once the session's first model step has finished in the running opencode, so a session you continue in a new opencode shows them after its next step.

**One background service for every opencode window.** By default opencode 2 runs its plugins and tools in a background service (`opencode serve --service`) that all opencode windows share and that keeps running after they close. ai-beacon's plugin reports what happens in that service under opencode's own session ids, so it never moves a wrapped session's card on another session's behalf.

**Wrapped sessions get their own server.** `opencode --standalone` runs a private server for that window instead, started with its own environment, which stops with it. So `ai-beacon session -- opencode` starts opencode 2's interface, `opencode run` and `opencode mini` with `--standalone`. Everything then appears on the wrapped session's card, as with opencode 1. The messaging tools act for that session, and killing it from the dashboard stops its turn. Other subcommands, such as `opencode serve`, and opencode 1 start exactly as before.

To tell the versions apart, the wrapper looks at what you launch:

- `npx` or `bunx` with `@opencode/cli` is opencode 2, and with `opencode-ai` opencode 1; nothing is run to find out.
- `opencode2`, and on Linux and macOS an `opencode` installed with `npm install -g` or `bun add -g`, are recognized from their install, also without running anything. On Windows those installs write launcher scripts, which are asked for their version as below.
- Any other `opencode` is asked for its version (`opencode --version`) before the launch. opencode 1 takes about a second to answer, so those launches start about a second later; opencode 2 answers in about a quarter of a second.
- A command run through another program, such as `ssh`, `docker` or `npx opencode`, is left as you typed it.

If the wrapper can't tell, the session starts without the flag; with `--log-file` set, the wrapper's log says so (see [Logs](#logs)). Pass `--standalone` yourself to fix that. When you pass `--standalone` or `--server <url>` yourself, the wrapper adds nothing.

A wrapped session that runs without `--standalone` shares the background service. That covers one you point at `--server`, a subcommand the wrapper leaves alone, and an opencode 2 it couldn't recognise. Such a session:

- shows its prompts, questions, subagents and notifications on a second card, under its opencode session id;
- shows on its own card only the state opencode reports to the terminal, with opencode 2.1 and later on Linux and macOS. With opencode 2.0, or on Windows, that card has no state, and with several opencode sessions open the report can be another session's activity;
- has messaging tools that act for the session that started the service, and none once that session has ended;
- when killed from the dashboard, ends its window while its turn keeps running in the service.

**Messaging tools on opencode 2.** opencode 2.2 and later offer `tell` and `list_sessions` as ordinary tools, as opencode 1 does. opencode 2.0 and 2.1 offer them only inside Code Mode: the model calls them from the JavaScript it runs in the `execute` tool, which those versions explain to it, and messages still go through.

**Known gaps:**

- opencode 2.0.0 shows a session's initial prompt in its input box without sending it: press Enter to send it. 2.1.0 and later send it.
- The card's activity line (the last tool and the last message) stays empty with opencode 2.
- A running opencode 2 service keeps the plugin it loaded, so the model, tokens, cost and context appear only once it runs the plugin of an ai-beacon that reads them. After updating ai-beacon, restart the service with `opencode service restart`, or close every opencode window and run `opencode service stop`. A "Plugin failed" message when opencode 2 starts means it loaded a plugin from before opencode 2 support: update ai-beacon, run `ai-beacon install` again (or start a wrapped session), then restart the service.
- With `OPENCODE_CONFIG_DIR` set, opencode 2 reads that directory instead of `~/.config/opencode`, so it does not load the plugin `ai-beacon install` writes there, and sessions do not report their state through it. opencode 1 reads both directories.

## A smaller browser is watching

A session is drawn at the size of its smallest viewer. When you open it in a browser window smaller than your terminal, the agent shrinks to the browser's size, and your terminal keeps leftover space on the right, at the bottom, or both:

Anything to the right of where the status bar ends, or below the status bar, is outside the session. Dim lines mark that edge:

- **Browser narrower than your terminal:** a short vertical line closes the status bar on the right.
  - On Windows with right margins on (see below), the vertical line runs down the whole right edge of the session.
  - On other systems it stays on the status bar row: a full-height line would scroll into your terminal's scrollback.
- **Browser shorter than your terminal:** a horizontal line sits below the status bar. It is as wide as the session, or as wide as your terminal when the full vertical line is drawn.
- The lines go away once the browser closes or grows to your terminal's size.

**On Windows, `ai-beacon` sets your terminal's right margin to the session's width**, so the agent's long lines wrap where the agent expects. Without it, lines wrap at your terminal's full width, and the screen fills with garbled rows until the browser closes. Windows Terminal and recent Windows 11 consoles support right margins.

- **Scrollback:** while the margin is set, lines that scroll off the top of the session do not reach your terminal's scrollback, and they are lost from it for good. With Claude Code's default inline mode, that is the conversation transcript. The browser still shows the session.
- **Rows still garble:** your terminal ignores right margins. Use Windows Terminal, or make the browser window at least as large as your terminal.
- **VS Code and JetBrains terminals:** these may ignore right margins too, and may lose the cursor position they saved. Turn margins off there if you see either.
- **Turning margins off:** do this if they cause glitches, for example the cursor jumping when a browser connects. With margins off, the garbled rows come back while a smaller browser is attached (see [Configuration](configuration.md#agent-client-configuration)).
  - For the current PowerShell window, run `$env:AI_BEACON_LOCAL_MARGINS = "0"` before `ai-beacon session`.
  - To make it permanent, run `setx AI_BEACON_LOCAL_MARGINS 0`; it applies to terminals opened afterwards.

## "Session not appearing" troubleshooting

Walk these in order — they cover ~90% of cases.

1. **Confirm the dashboard URL is reachable from the agent host.** From the machine running `claude`:

   ```bash
   curl -fsS "$AI_BEACON_URL/api/v1/health" && echo OK
   ```

   A `401` or `OK` means the network path works. A `404` means the URL is wrong. A connection error means firewall, DNS, or VPN.

2. **Confirm the token is current.**

   ```bash
   cat ~/.config/ai-beacon/token                         # Linux
   cat ~/Library/Application\ Support/ai-beacon/token    # macOS
   ```

   Compare it to the **Install** step of the setup guide on the dashboard. The most common cause of "everything looks fine but nothing shows up" is a container deployment with no persistent volume — every restart regenerates a fresh token and silently invalidates the one baked into your installed hooks. Either re-run `ai-beacon install` with the new token, or — preferably — mount `/data` to a persistent volume so the token survives restarts (see the [main README](../README.md)'s container instructions and the chart's `persistence.enabled` value).

3. **Confirm the hook ran.** Look for the `# ai-beacon-managed` marker in `~/.claude/settings.json`. If it's missing, `install` didn't run or your editor stripped it — re-run `ai-beacon install`.

4. **Look at the wrapper logs.** See [Logs](#logs) below. A session that fails to authenticate writes an unambiguous error there before exiting.

5. **Run the wrapper directly to bypass hooks.**

   ```bash
   AI_BEACON_URL=… AI_BEACON_AUTH_TOKEN=… ai-beacon session -- claude
   ```

   If this works but hook-launched sessions don't, the hook isn't picking up `config.toml` — usually a stale binary path. Re-run `ai-beacon install` so `binary_path` updates.

## The terminal keeps disconnecting

If a session's terminal disconnects every time you open it in the browser and reconnects a moment later, the wrapper is probably older than the dashboard. The dashboard refuses any single terminal message over 1 MiB, and older wrappers send a session's whole screen and scrollback as one message. The wrapper log (see [Logs](#logs)) shows `the dashboard closed the connection: a frame was over its read limit`. Download the current binary (see [Native binaries](#native-binaries)) and restart the session.

Behind IIS Application Request Routing, a session whose connection drops a few seconds after every connect, with reconnects failing with HTTP 502, is ARR failing on compressed WebSocket messages. The wrapper log shows `connection closed before it was stable` or `HTTP 502`. See [IIS Application Request Routing (ARR)](configuration.md#iis-application-request-routing-arr) for the fix. Older wrappers could stop reconnecting altogether after such an outage, leaving the session disconnected until restarted; download the current binary and restart those sessions.

## The status bar says "update required"

A yellow ` ⚠ update required: update ai-beacon, then restart this session ` row (shortened to ` ⚠ update ai-beacon, then restart ` in a narrow terminal), with `⚠` added to the terminal title, means the dashboard was upgraded and no longer accepts this version of ai-beacon. Every session on the machine still running that version shows it, including one started after the dashboard upgrade from a binary that was not updated. The agent keeps working locally, but the session no longer reports to the dashboard and its messaging tools stop working.

To fix it:

1. Download the current binary (see [Native binaries](#native-binaries)) and put it where the old one was, or run `ai-beacon install` again from the new binary so the hooks point at it (`binary_path` in `config.toml`). A binary saved anywhere else leaves hook-launched sessions on the old one.
2. On Windows a running `ai-beacon.exe` cannot be overwritten: exit the sessions, and quit every opencode started without the wrapper (with opencode 2, also run `opencode service stop`), since ai-beacon's opencode plugin keeps an `ai-beacon hook watch` running for each; or rename the old file and then download.
3. Restart each session that shows the row: exit the agent and start it again (`claude --continue` resumes the conversation). A running session keeps the binary it started with, so it cannot pick up the new one by itself.

Until then the wrapper stops calling the dashboard, so an old session does not use up the sign-in allowance of your network. The wrapper log shows `dashboard rejected this client as outdated` with the client's and the dashboard's protocol numbers (these are not the `ai-beacon version` numbers). Messaging tools tell the agent to ask you to update and restart.

## The status bar says "rate limited" or "sign-in failing"

A red row ending in `rate limited, retrying in <N>s` means the dashboard is refusing new sign-ins from your network for a while: too many came from the same address in the last minute (several sessions starting at once, or a dashboard restart that signed every session out). The wrapper waits as long as the dashboard asks, then signs in again on its own; no action is needed. Sessions that are already signed in keep reporting meanwhile. Messaging tools answer `dashboard rate limited; messaging unavailable for <N>s` during the wait.

A red row ending in `sign-in failing, retrying in <N>s` means the dashboard rejected the session's credentials. The wrapper retries with growing pauses, up to five minutes apart. If it does not clear after a dashboard restart, check the credentials the session uses (see [Updating the dashboard URL or token](#updating-the-dashboard-url-or-token)) and the wrapper log (see [Logs](#logs)). Messaging tools tell the agent to point you at the log.

## Logs

No log file is written unless you set one, with `--log-file` or `AI_BEACON_LOG_FILE`. The session command suppresses log output to stderr (it would garble the wrapped agent's TUI), so the log file is the only place its errors land; without one they are not recorded.

| Override | Behavior |
|----------|----------|
| `--log-file <path>` | Absolute path writes there directly. Bare filename roots it under `<data-dir>/logs/`. |
| `$AI_BEACON_LOG_FILE` | Same as the flag; flag wins when both are set. |

For a session started with `--log-file <session>.log`, `tail -F <data-dir>/logs/<session>.log` will surface heartbeat failures, auth issues, and reconnect attempts in real time.

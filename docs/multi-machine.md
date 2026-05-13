# Multi-machine setup

AI Beacon was built around a single dashboard watching many machines. One server, any number of agent hosts — your laptop, a workstation, a cloud VM, a Raspberry Pi running a long-form refactor — all reporting into the same view.

This page covers the three parts of that setup: naming machines, pointing them at the same dashboard, and enabling the dashboard to spawn sessions on them.

- [The one-dashboard, many-agents model](#the-one-dashboard-many-agents-model)
- [Naming devices](#naming-devices)
- [Sharing one token across hosts](#sharing-one-token-across-hosts)
- [Spawning sessions from the dashboard](#spawning-sessions-from-the-dashboard)
- [Worktrees](#worktrees)
  - [Cleanup on session exit](#cleanup-on-session-exit)
- [Common patterns](#common-patterns)

## The one-dashboard, many-agents model

There is exactly one `ai-beacon server` (a pod, container, or local process). Every agent host runs the `ai-beacon` CLI with `AI_BEACON_URL` pointing back at that server. The session card UI groups sessions by their reported **device name**, so each machine gets its own row even when several sessions are running on it.

```
   ┌──────────────┐
   │  dashboard   │   ai-beacon server (one)
   │  ai-beacon.… │
   └───────▲──────┘
           │ heartbeats over HTTPS
   ┌───────┴───────────────────┐
   │              │            │
 macbook    workstation     vm-1     ← devices
 (Claude)   (Claude×2)    (Claude)   ← sessions
```

Adding a machine = running `ai-beacon install` there with the dashboard URL and token. There is no server-side registration step; the dashboard discovers each device the first time a session from it reports in.

## Naming devices

Every session reports a device name in its heartbeats, and the dashboard groups by that name. The name is resolved once at session startup, in this order — first non-empty wins:

1. `--device <name>` on `ai-beacon session`.
2. `$AI_BEACON_DEVICE_NAME`.
3. `device_name` in `~/.config/ai-beacon/config.toml`.
4. The system hostname.
5. Literal `"unknown"`.

You can also **rename a device live** from the dashboard: click the device-group header and type a new name. The change propagates within one heartbeat tick (≤ 5 s) and is persisted to `config.toml` on the affected host, so it survives restarts.

> **Pinned names are immune to dashboard renames.** If a session reported a name that came from `--device` or `$AI_BEACON_DEVICE_NAME`, the dashboard can't rename it — those sources are treated as operator-pinned. To rename a pinned session, restart it with the new flag/env value. Hostname- and `config.toml`-sourced names are renameable.

Duplicate names across hosts are allowed and the dashboard merges them into one group. That's fine if you actually want a `dev` row spanning two boxes; rename one if you don't.

## Sharing one token across hosts

The agent token (`AI_BEACON_AUTH_TOKEN`) is the same on every host — the server generates one PSK at first boot and every agent presents that same value. Copy it to each machine via your usual secret-sharing channel (1Password, password manager, `ssh agent-host 'echo … > ~/.config/ai-beacon/token'`, etc.) and run `ai-beacon install` there.

There is no per-user token rotation today — PSK is the only agent auth mode in the standalone server. Rotating means generating a new token on the server and re-running `install` on every agent host.

> If you're deploying with the [Helm chart](../README.md#deploy-to-any-kubernetes-cluster) on persistent storage and leave `auth.token` unset, the token is auto-generated. Retrieve it with:
>
> ```bash
> kubectl exec -n ai-beacon deploy/ai-beacon -- cat /data/token
> ```

## Spawning sessions from the dashboard

By default the dashboard is a viewer: it surfaces what each agent host reports, and you can attach a terminal to anything running. To go further — **starting** a fresh session on a remote host from the dashboard — that host needs to tell the dashboard which projects live there.

Set `AI_BEACON_PROJECTS_DIR` on the agent host before running `ai-beacon install` (or any `ai-beacon session …`):

```bash
export AI_BEACON_PROJECTS_DIR=~/projects
```

With that set, the host periodically reports its discovered repositories. The dashboard shows them under that device's "Spawn new session" picker; selecting one launches `ai-beacon session -- claude` (or your configured agent) inside the repo. The spawn request goes through the existing agent → server WebSocket — no new ports to open, no SSH involved.

If `AI_BEACON_PROJECTS_DIR` is unset on a host, that host's row stays read-only on the dashboard. You can still watch and attach to terminals, just not spawn.

## Worktrees

When a dashboard-spawned session asks to work on an issue or PR, ai-beacon can put it in a fresh git worktree so it doesn't disturb your main checkout. `AI_BEACON_WORKTREE_LOCATION` controls where:

| Value | Layout |
|-------|--------|
| `sibling` (default) | Worktree is created next to the source repo, named `<repo>-<branch>`. |
| `subdirectory` | Worktree is created inside the source repo at `.worktrees/<branch>`. |

`sibling` is the right default for most setups (cleaner top-level project view, easier `cd` between trees). Choose `subdirectory` when your environment doesn't allow writing alongside the project (e.g. a workspace mount that's read-only above the repo root).

Set it per host:

```bash
export AI_BEACON_WORKTREE_LOCATION=sibling
```

…or per session with `--worktree-location`.

### Cleanup on session exit

When a dashboard-spawned session exits, ai-beacon cleans up after itself:

- **Clean worktree → removed.** No modified, staged, or untracked files? The worktree directory is deleted.
- **Dirty worktree → preserved.** Anything uncommitted stays on disk so you can recover the work. The local branch is preserved too.
- **PR merged → local branch deleted.** When the session's GitHub PR is in the `MERGED` state and the worktree was removed cleanly, the corresponding local branch in the main repo is also deleted. The PR's head commit is matched against the local tip first, so any commits you added locally after the merge (e.g. a follow-up tweak that didn't make it into the PR) cause the deletion to be skipped and a warning to be logged. The repo's default branch (`main`, `master`, or whatever `origin`/`upstream` reports) is always protected.
- **PR not merged → local branch preserved.** `OPEN`, `DRAFT`, `CLOSED`, or no-PR-at-all leaves the local branch alone.

PR state is re-fetched at exit, so a PR merged moments before the session ends still triggers the cleanup.

## Common patterns

**A laptop + a beefier workstation, both spawning.**

- Workstation: `AI_BEACON_DEVICE_NAME=workstation`, `AI_BEACON_PROJECTS_DIR=~/projects` — heavy refactors, long test runs, headless agents.
- Laptop: `AI_BEACON_DEVICE_NAME=macbook`, `AI_BEACON_PROJECTS_DIR=~/projects` — interactive work.
- Dashboard: pick whichever device you want to spawn into. Watch both from anywhere.

**A homelab box that only runs scheduled / long-form sessions.**

- Just set `AI_BEACON_DEVICE_NAME=homelab` and run `ai-beacon install`. Skip `AI_BEACON_PROJECTS_DIR` — the box doesn't need to expose project pickers, you'll `ssh` in to start sessions yourself.
- Dashboard shows the row; you attach a terminal when something needs you.

**Mixed OS (macOS + Linux + Windows).**

- Same token, same `AI_BEACON_URL` on each. The binaries are per-platform — see the download step in the in-app setup guide on the dashboard.
- Windows paths in `AI_BEACON_PROJECTS_DIR` work too (`C:\Users\you\projects`); set via PowerShell `$env:AI_BEACON_PROJECTS_DIR`.

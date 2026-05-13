# GitHub integration

When an agent host has the [`gh` CLI](https://cli.github.com) installed and authenticated, ai-beacon automatically enriches each session's heartbeat with GitHub state: the branch's pull request, its checks, mergeability, and draft/state flags. The dashboard surfaces all of that on the session card.

Plus, the dashboard ships two opinionated **workflow buttons** — *Implement Issue* and *Review PR* — that spawn a session pre-loaded with a battle-tested prompt template. You can override those templates per-host.

- [Setup](#setup)
- [What the session card shows](#what-the-session-card-shows)
- [Forks and upstreams](#forks-and-upstreams)
- [Workflow prompts](#workflow-prompts)
- [Privacy](#privacy)

## Setup

Install and authenticate `gh` on every agent host that should report PR state:

```bash
# macOS
brew install gh

# Linux (Debian/Ubuntu)
sudo apt install gh

# All platforms
gh auth login
```

That's it — there are no AI Beacon env vars to set. The agent calls `gh auth token` at session start; if a token is found, GitHub features light up. If not, the session still reports normally; the PR chip just doesn't render.

> **Token scope.** The default `gh auth login` flow grants the `repo` scope, which is enough for ai-beacon to read PRs and check status. Read-only scopes work; ai-beacon never pushes, merges, or comments on your behalf.

## What the session card shows

| Element | Source |
|---------|--------|
| **Repo link** (clickable project name) | The repo your session is working in. Built from the heartbeat's `github_owner` / `github_repo`. |
| **Fork glyph** | Rendered next to the repo link when an `upstream` remote is detected. See [Forks and upstreams](#forks-and-upstreams). |
| **PR chip** (`#42 · Ready`, etc.) | Live state from `gh pr view`. Priority order: Merged → Closed → Draft → Failing → Ready → Pending. |

PR state is polled per session, with ETags on the no-PR-yet path so creation detection is near-free. Once a PR exists, the agent refreshes every ~30 s — fast enough that the chip flips from Pending to Ready shortly after CI finishes, without hammering the API.

## Forks and upstreams

ai-beacon detects fork repositories by looking for an `upstream` git remote. When found:

- The session card's **primary** repo link points to **upstream** (the canonical project). Most users think *about* the upstream — its issues, docs, CI — even when they work *in* the fork.
- A small fork glyph next to the link links to **your fork**.
- PR detection uses the right repo and head reference automatically (`gh pr view fork-owner:branch -R upstream-repo`).

Non-forks render unchanged: the link points to `origin` and no glyph is rendered.

This works out of the box. The convention assumed is `origin` = your fork, `upstream` = the canonical repo — the same convention `gh repo fork` and most contributor guides use.

## Workflow prompts

The dashboard's session-spawn UI exposes two workflows out of the box:

| Workflow | Trigger | What the agent receives as its initial prompt |
|----------|---------|------------------------------------------------|
| **Implement Issue** | Pick an issue from the GitHub picker, click *Implement* | A multi-step plan-then-implement template with TDD enforcement, ask-before-assume gate, scope guard, and multi-persona peer review |
| **Review PR** | Pick a PR, click *Review* | A code-review template that walks the diff, raises blockers, and produces a summary |

The full built-in templates are non-trivial — they encode several engineering-discipline guardrails that took real iteration to land. You usually don't need to touch them.

### Customizing a workflow

If you do want to tweak one, override the prompt in `~/.config/ai-beacon/config.toml` on the **agent host**:

```toml
[workflow.implement_issue]
prompt = """
Implement this issue: {issue_url} ({issue_title}).
Ask before coding.
"""

[workflow.review_pr]
prompt = "/mn-review {pr_url}"
```

The following placeholders are substituted at spawn time:

| Placeholder | Filled with |
|-------------|-------------|
| `{issue_url}` | Full issue URL |
| `{issue_number}` | Issue number (e.g. `42`) |
| `{issue_title}` | Issue title |
| `{pr_url}` | Full PR URL |
| `{pr_number}` | PR number |
| `{repo}` | `owner/repo` derived from the issue or PR URL |
| `{branch}` | Branch name being worked on |

Missing placeholders are left as the literal `{placeholder}` so the signal is visible to the agent rather than silently dropped. The override applies to every session spawned on the host that owns the `config.toml`. Removing the section restores the built-in default on the next session.

### Bypassing the workflow

If you want to spawn from the dashboard with a one-off prompt that ignores both your override and the built-in default, use the **Custom prompt** input on the spawn form. Whatever you type wins, no template machinery involved.

## Privacy

GitHub data goes from your agent host's `gh` CLI straight to the dashboard via the agent → server WebSocket. The dashboard caches just enough to render — PR number, state, checks rollup, mergeability, draft flag, branch. Issue and PR titles for the workflow picker are fetched live; nothing is stored on the server beyond the running session's heartbeat.

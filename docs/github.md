# GitHub integration

When an agent host has the [`gh` CLI](https://cli.github.com) installed and authenticated, ai-beacon automatically enriches each session's heartbeat with GitHub state: the branch's pull request, its checks, mergeability, and draft/state flags. The dashboard surfaces all of that on the session card.

Plus, the dashboard's **New session** dialog ships two opinionated workflow modes, *Implement issue* and *Review PR*, that spawn a session pre-loaded with a battle-tested prompt template. You can override those templates per-host.

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
| **Repo link** (clickable project name) | The repo your session is working in. Built from the heartbeat's `forge_owner` / `forge_repo`. |
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

Click the **+** button in a device's header (its tooltip reads *Start session on {device}*) to open the **New session** dialog. Pick a project, then a mode: *Plain session*, *Implement issue* or *Review PR*. The workflow modes stay disabled until the project reports that `gh` (or `glab`) can serve them. The dialog remembers your last project and mode per device, so a repeat run is a few keystrokes: Enter picks the remembered project, then pick the issue or PR and press Enter again.

| Mode | Trigger | What the agent receives as its initial prompt |
|------|---------|------------------------------------------------|
| **Implement issue** | Pick an issue from the list (*Assigned to you* by default; search or paste its URL), click *Implement issue* | A multi-step plan-then-implement template with TDD enforcement, ask-before-assume gate, scope guard, and multi-persona peer review |
| **Review PR** | Pick a PR (*Open PRs* by default; search or paste its URL), click *Review PR* | A code-review template that walks the diff, raises blockers, and produces a report with the linked issues and a verdict (`APPROVE`, `REQUEST_CHANGES` or `COMMENT`). It **presents the full review for your approval and posts nothing to GitHub until you explicitly sign off** |

Four things the *Review PR* template does on your machine are worth knowing about:

- **It reads every issue and ticket the PR references.** The PR's closing issues, every `#123`, `owner/repo#123` and issue URL in its title and description, and every Jira key (`ABC-123`) in its title, description and branch name. Up to ten issues and ten tickets are preloaded at spawn: GitHub/GitLab issues with `gh`/`glab`, Jira tickets through the [Jira settings](configuration.md#jira-tickets) when they are configured. A `#123` that turns out to be a pull request is only named. The prompt lists the rest, with the command to read each, and the agent must read them all, say in its report when there are none, and name any it could not read, saying the review is incomplete. If the PR's closing issues could not be read at spawn, the prompt says so rather than reporting none. Without Jira settings the agent reads the tickets with whatever Jira access it has (an MCP server or a CLI), and never prints, asks for or passes credentials. When a fetch fails, you see the same *"Couldn't preload PR #N details"* notice as for the PR itself: it doesn't say which reference failed or why, but the agent's prompt lists every reference it still has to read, with the command for each. Identifiers that only look like Jira keys, such as `UTF-8`, `SHA-256` or a CVE id, are ignored, and so are the upstream issues bots quote in collapsed `<details>` blocks (Dependabot release notes). Because these reads use your credentials, on a public PR the posted review only *names* a Jira ticket or an issue from a private repository; it never quotes or summarizes it, and anything the agent concluded from one stays in the report shown to you. A PR counts as public unless its repository is confirmed private.
- **It may build and test before reviewing.** If your working tree holds the change under review, the agent builds it and runs the tests covering the changed area **once**, up front, and shares that result with its reviewers instead of each of them re-running it. If the tree is on the base branch (the default — the worktree toggle is off), it skips the build rather than baselining unrelated code. It will **not** build or test a change it doesn't trust — an unfamiliar author, or a diff touching build/test scripts, CI config, or `CLAUDE.md`/`AGENTS.md` — and asks you first instead.
- **A red baseline stops the review.** If the build or those tests fail on their own, the agent stops and asks whether to continue, rather than issuing verdicts over a broken tree. That is a second point (besides the post-approval gate) where it waits for you.
- **It posts one review, against the commit it reviewed.** The prompt names the PR's head commit at spawn. After you approve, the agent checks the PR is still at the head it reviewed and posts nothing if the author pushed since. On GitHub it posts the overview and every inline comment as a single review through `gh api repos/<owner>/<repo>/pulls/<n>/reviews`, pinned to that commit, with the verdict you approved; on a PR you opened yourself GitHub only accepts `COMMENT`. Posting needs [`jq`](https://jqlang.org) on the agent machine next to `gh`, to build that review from the approved texts; without it the agent stops and tells you instead of posting.

The full built-in templates are non-trivial — they encode several engineering-discipline guardrails that took real iteration to land. You usually don't need to touch them.

To save the agent its first turn, the issue/PR body and discussion (inline review threads included) are **preloaded into the prompt at spawn time** (fetched on the agent host with your `gh` credentials) and fenced as untrusted data. The built-in templates still keep the `gh … view` command as a refresh/fallback, so if the preload can't be fetched the agent loads it itself — you'll see a brief *"Couldn't preload … details"* note and a small marker on the session card, and nothing else changes.

If the forge CLI isn't installed or authenticated for the project's remote, no workflow can run there — the dialog says so and gives you the command to run on the agent machine: `gh auth login` for github.com, and for any other host, `glab auth login --hostname <host>` *if that host runs GitLab*. The hostname alone doesn't say which forge is behind it, so anywhere off github.com the command is offered conditionally: GitHub Enterprise, Gitea and Bitbucket aren't supported, and logging `glab` into one of them would mislabel the project as GitLab.

> **GitLab.** The same two workflows work on a GitLab checkout when the [`glab` CLI](https://gitlab.com/gitlab-org/cli) is installed and authenticated (`glab auth login`) for the project's host, including self-managed instances. The prompt is forge-accurate: it uses `glab` commands (`glab issue view`, `glab mr view`/`diff`, and `glab mr note` for posting a review), says "merge request" / `!123`, and preloads the issue/MR body plus its full discussion via `glab api`: every thread, inline diff comments included, each with its file:line anchor and resolved state. The *Review MR* post is an overview note plus one diff note per finding (`glab mr note create --file <path> --line <n>`, an experimental glab feature). Before posting inline notes the agent checks that your glab has those flags (1.119 or later), and otherwise folds the file:line references into the overview note instead. For an `APPROVE` or `COMMENT` verdict the overview is a non-resolvable note, so it never blocks a project that requires every thread resolved; for `REQUEST_CHANGES` it is a thread the author has to resolve. Only `APPROVE` approves, with `glab mr approve --sha` pinned to the reviewed head. If a step fails, the agent stops and tells you what was and was not posted instead of posting again. In the New session dialog you can search GitLab issues/MRs or paste an issue/MR URL (self-managed hosts included); on a GitLab project the picker brands its copy to "merge request" / `!123` (and the mode to *Review MR*), and it ignores a pasted GitHub PR URL. Everything else (the guardrails, the consent gate, the placeholders) is identical.

### Customizing a workflow

If you do want to tweak one, override the prompt in `config.toml` (in the [data directory](configuration.md#data-directory-layout)) on the **agent host**:

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
| `{repo}` | `owner/repo` (GitHub) or `group/…/project` (GitLab) derived from the issue or PR URL |
| `{branch}` | Branch name being worked on |
| `{issue_context}` | The issue body + comments, preloaded at spawn (empty if it couldn't be fetched) |
| `{pr_context}` | The PR body + comments + reviews + inline review threads, preloaded at spawn (empty if it couldn't be fetched) |
| `{referenced_context}` | The issues and tickets the PR references: the preloaded ones, then a list of the rest with the command to read each (empty if the PR itself couldn't be fetched). Only read when your template uses it |

Missing placeholders are left as the literal `{placeholder}` so the signal is visible to the agent rather than silently dropped, except `{issue_context}` / `{pr_context}` / `{referenced_context}`, which collapse to nothing when their content couldn't be preloaded (so your template never shows a stray token). The override applies to every session spawned on the host that owns the `config.toml`. Removing the section restores the built-in default on the next session.

### Bypassing the workflow

If you want to spawn from the dashboard with a one-off prompt that ignores both your override and the built-in default, use the **Custom prompt** input on the spawn form. Whatever you type wins, no template machinery involved.

## Privacy

GitHub data goes from your agent host's `gh` CLI straight to the dashboard via the agent → server WebSocket. The dashboard caches just enough to render — PR number, state, checks rollup, mergeability, draft flag, branch. Issue and PR titles for the New session dialog are fetched live; nothing is stored on the server beyond the running session's heartbeat.

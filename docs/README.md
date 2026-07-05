# AI Beacon Documentation

The [main README](../README.md) covers deploy and first-session onboarding. Past that, these pages take over.

| Page | When you need it |
|------|------------------|
| [Configuration](configuration.md) | Every environment variable, CLI flag, and Helm value. |
| [Connecting an agent](connect-agent.md) | What `ai-beacon install` does, rotating credentials, running the wrapper without install, uninstalling, "session not appearing" diagnostics. |
| [Multi-machine setup](multi-machine.md) | Device names, multiple hosts pointing at one dashboard, worktree layouts, `AI_BEACON_PROJECTS_DIR`. |
| [GitHub integration](github.md) | PR and check status on session cards, workflow prompts (`implement_issue`, `review_pr`), `gh` setup, customization. |
| [Inter-session messaging](inter-session-messaging.md) | Letting sessions message each other ("tell Session A about …") so you stop copy-pasting between terminals. |
| [Authentication](auth.md) | Choosing between password, PSK-only, OIDC, proxy-header, and no auth. Allowlists, OpenShift OAuth proxy. |

Found a gap? [Open an issue.](https://github.com/manusa/ai-beacon/issues)

---
title: ai-beacon
emoji: 🛰️
colorFrom: blue
colorTo: purple
sdk: docker
app_port: 8080
hf_oauth: true
hf_oauth_scopes:
  - email
pinned: false
---

# ai-beacon on Hugging Face Spaces

This directory is a one-shot recipe for deploying [ai-beacon](https://github.com/manusa/ai-beacon) to a free [Hugging Face Spaces](https://huggingface.co/spaces) instance with **OIDC sign-in via your Hugging Face account** — no IdP to configure, no password to manage.

The Space wraps the published `ghcr.io/manusa/ai-beacon:latest` image. The `entrypoint.sh` translates HF's auto-injected OAuth environment variables onto ai-beacon's generic OIDC flags; nothing HF-specific lives in the ai-beacon binary itself.

## What you get

- A public dashboard at `https://<your-user>-<space-name>.hf.space` gated by Hugging Face login.
- Browser users sign in with their HF account; ai-beacon admits only the usernames you allowlist.
- Agents (your wrapped Claude Code sessions on your local machines) connect with the pre-shared token you set as a Space secret.

## Deploy

1. Create a new Hugging Face Space (any owner, any name, **SDK: Docker**).
2. Clone this directory into the Space's repo:

   ```bash
   git clone https://huggingface.co/spaces/<your-user>/<space-name>
   cp huggingface-space/Dockerfile huggingface-space/entrypoint.sh huggingface-space/README.md <space-name>/
   cd <space-name>
   git add . && git commit -m "Deploy ai-beacon" && git push
   ```

   (The recipe directory in this repo also ships a Go black-box test (`recipe_test.go`) that pins the recipe's contract. It is deliberately excluded from the copy above — HF Spaces doesn't run Go tests, and it would just add confusing-looking code to the public Space repo.)

3. In the Space's **Settings → Variables and secrets**, add two **secrets**:

   | Secret | Value |
   |--------|-------|
   | `AI_BEACON_AUTH_TOKEN` | A long random string, e.g. `openssl rand -hex 32`. Your agents will present this on every heartbeat. |
   | `AI_BEACON_ALLOWED_USERS` | Comma-separated Hugging Face usernames permitted to sign in (yours, plus anyone you share the dashboard with). Required — see [Allowlists](https://github.com/manusa/ai-beacon/blob/main/docs/auth.md#allowlists). |

4. Restart the Space (Settings → Restart). The dashboard URL is the Space URL itself.

5. Open the Space URL in a browser → you'll be bounced through Hugging Face login → after you approve, the dashboard loads.

6. Connect an agent: install the ai-beacon CLI on your machine and point it at the Space:

   ```bash
   ai-beacon install \
     --url https://<your-user>-<space-name>.hf.space \
     --token <the AI_BEACON_AUTH_TOKEN you set>
   ```

   See [Connecting an agent](https://github.com/manusa/ai-beacon/blob/main/docs/connect-agent.md) for the rest.

## Caveats (free CPU tier)

- **Auto-sleep.** Free CPU Spaces sleep after **48 h of dashboard inactivity**. The first request after sleep waits ~30–60 s for cold start. ai-beacon's agent heartbeat retry loop tolerates this transparently — sessions reappear on the dashboard as soon as the Space wakes — but the human user opening the dashboard will see the cold-start latency once.
- **Public README.** Free Spaces are public by default; the README you push lives at `https://huggingface.co/spaces/<your-user>/<space-name>`. The OIDC gate protects the running app's responses — the Space's *existence* and this README are not hidden. Use a Private Space (paid) if you need that.
- **No custom domain.** Free tier serves only `*.hf.space`.
- **Single instance.** One container, no horizontal scale-out on the free tier.

## What this recipe does NOT do

- Run ai-beacon with persistent data (paid HF addon).
- Pre-register a redirect URL with HF — HF accepts any URL targeting your Space, so the recipe derives `https://${SPACE_HOST}/login/oidc/callback` from the env var HF auto-injects.
- Customize the Space metadata further (title, emoji, colors). Edit the YAML frontmatter at the top of this file if you want to.

## How the recipe wires up

Refer to [`docs/auth.md`](https://github.com/manusa/ai-beacon/blob/main/docs/auth.md#hugging-face-spaces) for the full env-var → flag mapping and a discussion of why no Go code in ai-beacon is HF-aware.

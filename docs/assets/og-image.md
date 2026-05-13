# OG image regeneration

`og-image.png` is the social-card image rendered by GitHub, Slack, X, Discord, etc. when someone pastes a link to <https://github.com/manusa/ai-beacon>. It is configured manually in the repo's **Settings → Social preview**.

The current image is a lighthouse-and-archipelago composition generated via Option A below. To regenerate (new tagline, refreshed visual, etc.), use one of the two recipes below.

## Format requirements

GitHub renders OG images at **1200×630** (the open-graph standard). Anything smaller looks blurry; anything taller gets cropped to that ratio. The social-preview upload also caps at **1 MB** — if the generator produces a larger file, resize with:

```bash
sips -z 630 1200 ai-beacon/docs/assets/og-image.png
```

## Option A — ChatGPT / image-model prompt

Paste this into ChatGPT (GPT-Image / DALL·E) or any image-generation model that accepts art direction. It is intentionally opinionated to push the model away from generic "AI tool" tropes.

> Design a social-share card (open-graph image) for an open-source developer tool called **AI Beacon**.
>
> **Format requirements (hard):**
> - Exactly **1200 × 630 pixels**, PNG, no transparency, no rounded corners, no outer border.
> - Will be displayed at small sizes (Slack/X/LinkedIn previews) — all text must remain legible at ~600 × 315 px.
>
> **About the product:**
> AI Beacon is a self-hosted dashboard for AI coding agents (Claude Code, OpenCode). A developer running agents across multiple machines uses it to see what every agent is doing in real time and step in the moment one needs them — answer a permission prompt, attach to a terminal, unblock an idle session, all from a browser on any device, including a phone.
>
> **Wordmark:** "AI Beacon" in a clean modern sans-serif (Inter, SF Pro, or similar). Medium-bold, generous letter spacing, light slate color.
>
> **Tagline (must appear, exact wording):**
> "Watch every coding agent on every machine — and step in the moment one needs you."
>
> **Supporting text (smaller):**
> "Claude Code · OpenCode · Linux · macOS · Windows"
>
> **Repo URL (smallest, monospace, muted, bottom-right corner):**
> `github.com/manusa/ai-beacon`
>
> **Pick ONE visual concept — do not combine:**
> 1. **Lighthouse sweeping a beam over a sparse archipelago of machines at night.** Lighthouse on the left third, beam crosses toward the right where the wordmark sits. "Machines" are stylized as a sparse cluster of glowing terminal windows or rectangles on darker silhouettes — not a busy grid. Editorial, painterly, calm. NOT cartoonish.
> 2. **A single bright dot pulsing among a constellation of dimmer ones.** The bright dot is the agent that needs you. Thin radial pulse around it. Minimalist, generous negative space, like an elegant radar — but never cyberpunk neon.
> 3. **A dark dashboard interface visible at 25–30% opacity in the background**, wordmark + tagline floating in clean negative space on top. Suggests the product without literally being a screenshot.
>
> **Color palette (strict — do not deviate):**
> - Background: deep slate-navy, **#0F172A → #1E293B** (subtle vertical gradient, not radial, not flashy).
> - Beacon / accent: warm yellow **#FBBF24** — used sparingly, the only warm color on the card.
> - Primary text: **#E2E8F0**. Secondary text: **#94A3B8**. Repo URL: **#64748B**.
> - Optional thin red accent **#DC2626** — single small highlight, not a stripe across the image.
>
> **Mood:** calm, intentional, slightly maritime. The product is about *paying attention*, not about *flashing alerts*. Should feel like a developer-tool landing page, not a startup launch banner. Think Linear's marketing site, not a SaaS hero.
>
> **Do NOT include:**
> - Robots, humanoid AI characters, brains, neural-network node graphs, cyberpunk neon, Tron grids, lens flares, generic "tech blue" radial gradients.
> - Laptop-and-coffee stock-photo composition.
> - Multiple visual elements competing for attention. One focal point only.
> - Floating UI mockups with fake metrics or fake chart screenshots.
> - Emojis, sparkles, hand-drawn arrows, motion blur.
> - Any text other than what's explicitly listed above.
>
> **Composition rule:** wordmark + tagline sit in a clean column — left or right of the focal visual, NOT floating over it. Whitespace is the design.
>
> Output: a single PNG, exactly 1200 × 630, ready to upload as a GitHub social preview.

## Option B — Screenshot + post-process

Capture a real dashboard screenshot showing 3–4 sessions across two devices (`make screenshots` in `server-frontend` produces `ai-beacon-dashboard-dark.png` — see `src/agents/__screenshots__/ai-beacon-dashboard.screenshots.jsx` for the fixture). Crop to a 1200×630 frame, then either:

- Paste it into the prompt above and ask the model to compose the wordmark and tagline over the screenshot, or
- Composite in your image editor: dashboard screenshot at ~55% width on the right with a vignette, wordmark + tagline on the left.

## Verifying after replacement

1. Commit and push to `main` on `manusa/ai-beacon`.
2. In the public repo, go to **Settings → Social preview** and upload the PNG (GitHub does not auto-pick the in-repo file).
3. Paste `https://github.com/manusa/ai-beacon` into a fresh chat in Slack, X, or Discord and confirm the card renders the new image. GitHub may take a few minutes to invalidate its cached card.

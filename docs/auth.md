# Authentication

Two things authenticate to the server: **agents** (your wrapped Claude Code sessions) and **browsers** (you opening the dashboard). The agent side is fixed — a single pre-shared bearer token (`AI_BEACON_AUTH_TOKEN`). The browser side has four modes you choose between.

- [Picking a browser-auth mode](#picking-a-browser-auth-mode)
- [Password (default)](#password-default)
- [OIDC (bring your own IdP)](#oidc-bring-your-own-idp)
- [Proxy-header (identity-aware reverse proxy)](#proxy-header-identity-aware-reverse-proxy)
- [None (no browser auth)](#none-no-browser-auth)
- [Agent-side auth](#agent-side-auth)
- [Allowlists](#allowlists)
- [OpenShift OAuth Proxy](#openshift-oauth-proxy)

## Picking a browser-auth mode

| Mode | `--auth` value | Pick when… |
|------|----------------|------------|
| **Password** | `""` (default) | Solo user, small team, or anyone who already trusts a shared password. Lowest friction. |
| **OIDC** | `oidc` | You have an IdP (Google Workspace, GitHub, Keycloak, Authentik, Okta…) and want per-user identity. |
| **Proxy-header** | `proxy-header` | An identity-aware proxy already fronts the dashboard (OpenShift OAuth Proxy, oauth2-proxy, Cloudflare Access, AWS ALB OIDC). |
| **None** | `none` | The dashboard is behind a separate access boundary you fully trust (a single-user laptop, a private mesh network). Only choose this knowingly. |

The mode is set with `--auth=<value>` on `ai-beacon server`. The defaults below pair flag-style and env-var-style configuration; both work.

## Password (default)

What the [main README](../README.md) describes. One shared password admits all browser users. No allowlist, no per-user identity.

```bash
ai-beacon server \
  --address :8080 \
  --auth ""                # or just omit, it's the default
```

Or:

```bash
AI_BEACON_AUTH_PASSWORD=changeme ai-beacon server
```

If `AI_BEACON_AUTH_PASSWORD` is unset and no `--password-file` is given, a random 64-hex-char password is generated and written to `<data-dir>/password` on first boot. Retrieve it with:

```bash
cat ~/.config/ai-beacon/password                                # local
kubectl exec -n ai-beacon deploy/ai-beacon -- cat /data/password # Helm
```

Sign-in URL `https://your.host/?token=<password>` works for bookmarks / scripts; otherwise use the standard login form.

## OIDC (bring your own IdP)

Per-user identity via any OIDC-compliant IdP. Required flags:

```bash
ai-beacon server \
  --auth=oidc \
  --oidc-issuer https://accounts.google.com \
  --oidc-client-id <id> \
  --oidc-client-secret-file /run/secrets/oidc-secret \
  --oidc-redirect-url https://ai-beacon.example.com/login/oidc/callback \
  --allowed-users alice,bob,charlie
```

Notes:

- `--oidc-client-secret-file` is preferred over `--oidc-client-secret` for shared hosts — it keeps the secret out of `argv`. Public clients (PKCE-only) need neither.
- `--oidc-scopes` defaults to `openid,profile,email`. `openid` is always added.
- `--oidc-redirect-url` is the callback you register with the IdP. It must match exactly, including the `/login/oidc/callback` path.
- The display label for the sign-in button is auto-derived from the issuer host. Override with `--oidc-display-name "Sign in with Acme"`.
- An [`--allowed-users` allowlist](#allowlists) is required — without it any identity the IdP verifies could sign in.

### Hugging Face Spaces

Hugging Face Spaces is a first-class consumer of this generic OIDC flow. Setting `hf_oauth: true` in a Space's `README.md` metadata makes HF provision an OAuth/OIDC client for the Space and inject four env vars at container start:

| HF env var | ai-beacon flag | Passed via |
|------------|---------------|-----------|
| `OPENID_PROVIDER_URL` | `--oidc-issuer` | argv (`entrypoint.sh`) |
| `OAUTH_CLIENT_ID` | `--oidc-client-id` | argv (`entrypoint.sh`) |
| `OAUTH_CLIENT_SECRET` | `--oidc-client-secret` | env-var fallback only — the recipe deliberately does **not** put the secret on argv so it stays out of `/proc/<pid>/cmdline` |
| `OAUTH_SCOPES` | `--oidc-scopes` | argv (`entrypoint.sh`) |

The redirect URL is **not** an HF-injected var. HF instead injects `SPACE_HOST` (the public hostname, no scheme) and accepts any callback URL targeting the Space — so the recipe builds `--oidc-redirect-url https://${SPACE_HOST}/login/oidc/callback` at boot.

The Space's `README.md` frontmatter sets `hf_oauth_scopes: [email]` — **not** the full `openid, profile, email` list. HF auto-includes `openid` and `profile` whenever `hf_oauth: true` is set; the `hf_oauth_scopes` YAML key is a strict allowlist of *extra* permission scopes (email, repo scopes, billing, …) and HF's API rejects the git push with `must be one of [email, read-repos, …]` if `openid` or `profile` appear there. At runtime the injected `OAUTH_SCOPES` env var still resolves to `openid profile email`.

No HF-specific code lives in `pkg/auth/` — the recipe is a Dockerfile + entrypoint that maps env vars onto the same flags every OIDC deployment uses. The ai-beacon binary never learns it is running on HF; the only HF-shaped surface in the codebase is the [`huggingface-space/`](../huggingface-space/README.md) recipe directory and an entry in the issuer-host → display-name table (parallel to `Okta`, `GitLab`, etc.) so the sign-in button reads "Sign in with Hugging Face" without operator configuration.

For the deployment steps see [`huggingface-space/README.md`](../huggingface-space/README.md) and the [main README's Deploy section](../README.md#deploy-to-hugging-face-spaces).

If you're integrating with any other environment that ships OIDC config under the legacy `OPENID_PROVIDER_URL` / `OAUTH_*` names — Cloud Run with an identity-aware proxy, a vendored OAuth-proxy template, etc. — those env vars are honored as fallbacks by the binary itself; you do not need to re-prefix them to `AI_BEACON_OIDC_*`.

## Proxy-header (identity-aware reverse proxy)

If a fronting proxy already authenticates the user and forwards their identity via `X-Forwarded-User`, ai-beacon can trust that header — provided the proxy is in a CIDR you nominate.

```bash
ai-beacon server \
  --auth=proxy-header \
  --trusted-proxies 10.0.0.0/8,127.0.0.0/8 \
  --allowed-users alice,bob
```

Both `--trusted-proxies` and `--allowed-users` are **required** in this mode — without trusted-proxies any client could spoof `X-Forwarded-User`, and without an allowlist any identity the proxy verified would pass.

Typical fronting proxies:

- OpenShift OAuth Proxy (see [§ OpenShift OAuth Proxy](#openshift-oauth-proxy) for chart support).
- `oauth2-proxy` in front of any Kubernetes Ingress.
- Cloudflare Access — set `Cf-Access-Authenticated-User-Email` via a `request_headers` policy that rewrites it to `X-Forwarded-User`.
- AWS ALB OIDC — has `x-amzn-oidc-data` but no `X-Forwarded-User` natively; rewrite at the listener.

## None (no browser auth)

```bash
ai-beacon server --auth=none
```

The dashboard is accessible to anyone who can reach the port. Use only when there is a separate access boundary you trust completely — a single-user laptop, a private Tailscale-only mesh, an SSH-tunneled port. **Do not pair `--auth=none` with a public ingress.**

Agent auth still applies in this mode (PSK token required).

## Agent-side auth

There is one mode: PSK bearer token (`AI_BEACON_AUTH_TOKEN`).

- The server generates a 64-hex-char token on first boot (or accepts one via `--auth-token` / `$AI_BEACON_AUTH_TOKEN`), writes it to `<data-dir>/token` (mode `0600`), and exposes it to the dashboard's setup guide.
- Every agent presents that same token. There is no per-user agent rotation today; rotating means generating a new token and re-running `ai-beacon install` on every agent host (see [Connecting an agent § Updating the dashboard URL or token](connect-agent.md#updating-the-dashboard-url-or-token)).
- Per-user agent tokens via SSH challenge-response are a known follow-up. They are not in the standalone server today.

## Allowlists

`--allowed-users` (or `$AI_BEACON_ALLOWED_USERS`, comma-separated) is the canonical sign-in allowlist. Required by both `--auth=oidc` and `--auth=proxy-header`.

- Names are matched case-insensitively against the username returned by the upstream authenticator (OIDC `preferred_username` claim, or the `X-Forwarded-User` value).
- Blank entries (e.g. `--allowed-users ""` or `--allowed-users ,,`) are dropped; the CLI rejects launches that would result in an empty effective allowlist.
- For OpenShift OAuth Proxy users, the chart value `allowedUsers` populates this — `--set allowedUsers="{$(oc whoami)}"` is the one-liner the [main README](../README.md) uses.

## OpenShift OAuth Proxy

On OpenShift (Developer Sandbox or any cluster) the chart's `oauthProxy.enabled: true` adds an [OAuth Proxy](https://github.com/openshift/oauth-proxy) sidecar. The sidecar gates browser traffic through OpenShift's OAuth provider — no IdP to configure, no password to manage — while agent traffic under `/api/v1/agent/` skips the proxy and remains protected by the PSK token.

The chart wires up the proxy automatically:

- Flips ai-beacon's `--address` to `127.0.0.1:8080` so the only network path is through the proxy.
- Switches the Service to target the proxy port (`:8443`).
- Annotates the ServiceAccount for OpenShift OAuth redirects.
- When `openshift: true`, flips the Route to TLS reencrypt.

You still need `allowedUsers` set — without it the proxy will admit every cluster user with OAuth access.

```bash
helm install ai-beacon oci://ghcr.io/manusa/charts/ai-beacon \
  --version 0.0.0-snapshot \
  --set openshift=true \
  --set oauthProxy.enabled=true \
  --set persistence.enabled=false \
  --set auth.token="$(openssl rand -hex 32)" \
  --set allowedUsers="{$(oc whoami)}"
```

This is what the [main README's Sandbox recipe](../README.md#deploy-to-openshift-developer-sandbox) renders behind the scenes.

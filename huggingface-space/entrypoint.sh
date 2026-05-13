#!/bin/sh
# Hugging Face Spaces entrypoint for ai-beacon.
#
# HF Spaces injects four OAuth env vars when hf_oauth: true is set in
# the Space's README.md metadata (see ../docs/auth.md § Hugging Face
# Spaces). It also injects SPACE_HOST with the Space's public hostname
# (no scheme). This script maps them onto the generic --oidc-* flags
# the ai-beacon binary already understands, so the Go codebase never
# learns about HF.
#
# Agent token (AI_BEACON_AUTH_TOKEN) and browser allowlist
# (AI_BEACON_ALLOWED_USERS) come in via Space secrets and are resolved
# by the ai-beacon binary itself from the environment — they do not
# need explicit flag mapping here.
set -eu

# OAUTH_CLIENT_SECRET is intentionally NOT passed on argv. The ai-beacon
# binary already resolves it from the environment (see docs/auth.md
# § OIDC). Keeping the secret out of /proc/<pid>/cmdline matches the
# spec's preference for --oidc-client-secret-file over the argv flag.
exec /app/ai-beacon server \
  --address ":8080" \
  --auth=oidc \
  --oidc-issuer "${OPENID_PROVIDER_URL}" \
  --oidc-client-id "${OAUTH_CLIENT_ID}" \
  --oidc-scopes "${OAUTH_SCOPES}" \
  --oidc-redirect-url "https://${SPACE_HOST}/login/oidc/callback" \
  "$@"

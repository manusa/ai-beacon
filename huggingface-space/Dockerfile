# Hugging Face Spaces deployment recipe for ai-beacon.
#
# This is a thin wrapper around the published ai-beacon image. The
# entrypoint script translates HF's auto-injected OAuth env vars
# (OPENID_PROVIDER_URL, OAUTH_CLIENT_*, OAUTH_SCOPES) and the Space
# hostname (SPACE_HOST) onto ai-beacon's generic OIDC flags. No
# HF-specific Go code exists upstream — the recipe is pure config.
#
# Push the entire huggingface-space/ directory to a fresh HF Space
# repo with sdk: docker; see ./README.md for the user-facing steps.
FROM ghcr.io/manusa/ai-beacon:latest

# HF Spaces routes the public URL to whatever app_port the Space's
# README.md declares; ai-beacon listens on 8080 in this recipe.
EXPOSE 8080

# Switch to root briefly so we can COPY in the entrypoint with the
# right ownership, then drop back to the unprivileged user the base
# image already configures (65532:0).
USER 0:0
COPY --chmod=0755 entrypoint.sh /entrypoint.sh
USER 65532:0

ENTRYPOINT ["/entrypoint.sh"]

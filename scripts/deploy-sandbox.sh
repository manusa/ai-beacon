#!/bin/bash

set -euo pipefail

echo "==================================="
echo "AI Beacon - Developer Sandbox Deploy"
echo "==================================="
echo ""

# Check if oc is available and user is logged in
if ! command -v oc &> /dev/null; then
    echo "❌ Error: 'oc' command not found"
    echo "Please install the OpenShift CLI: https://docs.openshift.com/container-platform/latest/cli_reference/openshift_cli/getting-started-cli.html"
    exit 1
fi

if ! oc whoami &> /dev/null; then
    echo "❌ Error: Not logged in to OpenShift"
    echo ""
    echo "To log in to Developer Sandbox:"
    echo "1. Go to https://developers.redhat.com/developer-sandbox"
    echo "2. Click 'Start your sandbox for free'"
    echo "3. Click the 'DevSandbox' button and select 'Copy login command'"
    echo "4. Paste and run the login command in your terminal"
    exit 1
fi

# Check if helm is available
if ! command -v helm &> /dev/null; then
    echo "❌ Error: 'helm' command not found"
    echo "Please install Helm: https://helm.sh/docs/intro/install/"
    exit 1
fi

ALLOWED_USER=$(oc whoami)

echo "✓ Logged in as: $ALLOWED_USER"
echo "✓ Current namespace: $(oc project -q)"
echo ""

if helm status ai-beacon >/dev/null 2>&1; then
    echo "❌ Error: ai-beacon is already installed in namespace $(oc project -q)."
    exit 1
fi

# Generate random token for agent authentication
TOKEN=$(openssl rand -hex 32)

echo ""
echo "Deploying AI Beacon..."
echo "  Allowed user: $ALLOWED_USER"
echo "  Agent token: $TOKEN"
echo ""

# Deploy using Helm (browser auth via OpenShift OAuth Proxy sidecar)
helm install ai-beacon \
  oci://ghcr.io/manusa/charts/ai-beacon \
  --version 0.0.0-snapshot \
  --set openshift=true \
  --set oauthProxy.enabled=true \
  --set persistence.enabled=false \
  --set auth.token="$TOKEN" \
  --set allowedUsers="{$ALLOWED_USER}"

# Store agent token in a Secret for easy retrieval
echo ""
echo "Storing credentials..."
oc delete configmap ai-beacon-credentials >/dev/null 2>&1 || true
oc create secret generic ai-beacon-credentials \
  --from-literal=token="$TOKEN" \
  --dry-run=client -o yaml | oc apply -f -

echo ""
echo "✅ Deployment complete!"
echo ""
echo "==================================="
echo "Next Steps:"
echo "==================================="
echo ""
echo "1. Get your dashboard URL:"
echo "   oc get route ai-beacon -o jsonpath='https://{.spec.host}'"
echo ""
echo "2. Open the URL in your browser and sign in with your Red Hat account"
echo "   (OpenShift OAuth). Only $ALLOWED_USER is allowed."
echo ""
echo "3. Click the rocket icon in the dashboard to get agent setup instructions"
echo ""
echo "4. To retrieve the agent token later, run:"
echo "   oc extract secret/ai-beacon-credentials --keys=token --to=-"
echo ""
echo "==================================="

# Wait for the route host (route object can exist before spec.host is set)
echo ""
echo "Waiting for route to be ready..."
for _ in {1..30}; do
    host=$(oc get route ai-beacon -o jsonpath='{.spec.host}' 2>/dev/null) || true
    if [ -n "$host" ]; then
        break
    fi
    sleep 1
done

host=$(oc get route ai-beacon -o jsonpath='{.spec.host}' 2>/dev/null) || true
if [ -n "$host" ]; then
    echo ""
    echo "🚀 Dashboard URL: https://${host}"
else
    echo "⚠️  Route not ready yet. Run 'oc get route ai-beacon' to check status."
fi

echo ""

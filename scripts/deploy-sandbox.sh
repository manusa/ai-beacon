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

echo "✓ Logged in as: $(oc whoami)"
echo "✓ Current namespace: $(oc project -q)"
echo ""

# Prompt for password with default
read -p "Enter dashboard password [changeme]: " PASSWORD
PASSWORD=${PASSWORD:-changeme}

# Validate password is not empty
if [ -z "$PASSWORD" ]; then
    echo "❌ Error: Password cannot be empty"
    exit 1
fi

# Generate random token for agent authentication
TOKEN=$(openssl rand -hex 32)

echo ""
echo "Deploying AI Beacon..."
echo "  Dashboard password: $PASSWORD"
echo "  Agent token: $TOKEN"
echo ""

# Deploy using Helm
helm install ai-beacon \
  oci://ghcr.io/manusa/charts/ai-beacon \
  --version 0.0.0-snapshot \
  --set openshift=true \
  --set persistence.enabled=false \
  --set auth.token="$TOKEN" \
  --set auth.password="$PASSWORD"

# Store credentials in a ConfigMap for easy retrieval
echo ""
echo "Storing credentials..."
oc create configmap ai-beacon-credentials \
  --from-literal=token="$TOKEN" \
  --from-literal=password="$PASSWORD" \
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
echo "2. Open the URL in your browser and log in with:"
echo "   Password: $PASSWORD"
echo ""
echo "3. Click the rocket icon in the dashboard to get agent setup instructions"
echo ""
echo "4. To retrieve credentials later, run:"
echo "   oc get configmap ai-beacon-credentials -o yaml"
echo ""
echo "==================================="

# Wait for the route to be ready
echo ""
echo "Waiting for route to be ready..."
sleep 5

DASHBOARD_URL=$(oc get route ai-beacon -o jsonpath='https://{.spec.host}' 2>/dev/null)
if [ -n "$DASHBOARD_URL" ]; then
    echo ""
    echo "🚀 Dashboard URL: $DASHBOARD_URL"
else
    echo "⚠️  Route not ready yet. Run 'oc get route ai-beacon' to check status."
fi

echo ""

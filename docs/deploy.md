# Deploying the dashboard

The [main README](../README.md#quick-start) has the two quick starts: Docker on your own machine and the OpenShift Developer Sandbox. This page covers the other ways to run the dashboard. Whichever you pick, open the dashboard afterwards and click the **rocket icon** in the top bar: the setup guide walks you through [connecting your first agent](connect-agent.md).

- [Any Kubernetes cluster](#any-kubernetes-cluster)
- [Hugging Face Spaces](#hugging-face-spaces)
- [Native binary](#native-binary)

## Any Kubernetes cluster

```bash
export TOKEN=$(openssl rand -hex 32)
export PASSWORD=$(openssl rand -hex 16)
echo "Dashboard password: $PASSWORD"

helm install ai-beacon \
  oci://ghcr.io/manusa/charts/ai-beacon \
  --version 0.0.0-snapshot \
  --set ingress.host=ai-beacon.example.com \
  --set auth.token="$TOKEN" \
  --set auth.password="$PASSWORD" \
  -n ai-beacon --create-namespace
```

On clusters with persistent storage, you can omit `auth.token` and `auth.password`: credentials are auto-generated and persisted to the volume. Retrieve them with:

```bash
kubectl exec -n ai-beacon deploy/ai-beacon -- cat /data/password
kubectl exec -n ai-beacon deploy/ai-beacon -- cat /data/token
```

`--version 0.0.0-snapshot` is a rolling pre-release alias that tracks the latest build; it is required until a stable release is published.

Every chart value is documented in the [chart's README](../charts/ai-beacon/README.md), and [Configuration § Helm values](configuration.md#helm-values) lists the ones that are easy to get wrong. For OpenShift, the [Developer Sandbox recipe](../README.md#deploy-to-openshift-developer-sandbox) works on any cluster; [Authentication § OpenShift OAuth Proxy](auth.md#openshift-oauth-proxy) explains what it sets up.

## Hugging Face Spaces

Hugging Face's free CPU tier accepts arbitrary Docker images and projects the signed-in Hugging Face user's identity into the container via OAuth/OIDC. The dashboard's "who is allowed in" question is answered by your existing Hugging Face account, with no IdP to configure and no password to manage.

The full recipe lives in [`huggingface-space/`](../huggingface-space/README.md): push that directory to a fresh Space, set two secrets, restart the Space. The auth side is covered in [Authentication § Hugging Face Spaces](auth.md#hugging-face-spaces).

## Native binary

To run the dashboard without a container runtime, download the `ai-beacon` binary for your OS from [Native binaries](connect-agent.md#native-binaries) and start the server:

```bash
./ai-beacon server
```

It listens on port `8080` (change it with `--address`). Unless you set `AI_BEACON_AUTH_PASSWORD` or `--password-file`, the browser password is generated on first start and written to `password` in the [data directory](configuration.md#data-directory-layout), next to the agent `token`. Every server flag and variable is in [Configuration § Server configuration](configuration.md#server-configuration).

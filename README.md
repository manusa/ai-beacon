[![Contribute](https://www.eclipse.org/che/contribute.svg)](https://workspaces.openshift.com#https://github.com/manusa/ai-beacon)

# AI Beacon

A web dashboard for monitoring and managing AI coding agent sessions across your devices.

> [!NOTE]
> **Early access preview** — AI Beacon is under active development. Pre-built binaries and container images are available now; source code will be published once we have initial validation from early users.
> Currently supports **Claude Code**. Support for additional coding agents is planned.

![AI Beacon Dashboard](screenshots/dashboard.png)

## Deploy

Pick a deployment method and follow the steps — the built-in setup guide will walk you through connecting your first agent.

| Method | Best for |
|--------|----------|
| [OpenShift Developer Sandbox](#deploy-to-openshift-developer-sandbox) | Free cloud dashboard, no credit card |
| [Any Kubernetes cluster](#deploy-to-any-kubernetes-cluster) | Your own cluster with Helm |
| [Container image](#container-image) | Quick local look without a cluster |

## Recommended tools

AI Beacon works out of the box, but most features require these tools on the machines running your agents:

| Tool | What it unlocks |
|------|----------------|
| [`tmux`](https://github.com/tmux/tmux) | Terminal attach from the dashboard, remote session spawning |
| [`git`](https://git-scm.com/) | Current branch display, worktree management |
| [`gh`](https://cli.github.com/) (authenticated) | PR status and checks on session cards, review and merge PRs from the dashboard |

Without them the dashboard still tracks every session's model, context usage, cost, and duration.

## Deploy to OpenShift Developer Sandbox

The [Developer Sandbox](https://developers.redhat.com/developer-sandbox) is free and available to anyone with a Red Hat account.

```bash
# 1. Set credentials
#    The token authenticates agents to the server.
#    The password is for browser login — pick something you'll remember.
#    WARNING: don't reuse a real password here; the value is passed on the command line.
export TOKEN=$(openssl rand -hex 32)
export PASSWORD=changeme

# 2. Install (into your current namespace — the sandbox assigns one for you)
helm install ai-beacon \
  oci://ghcr.io/manusa/charts/ai-beacon \
  --version 0.0.0-snapshot \
  --set openshift=true \
  --set persistence.enabled=false \
  --set auth.token="$TOKEN" \
  --set auth.password="$PASSWORD"

# 3. Get the dashboard URL
oc get route ai-beacon -o jsonpath='https://{.spec.host}'
```

Open the dashboard URL in your browser and log in with the password you set above.
Once inside, click the **rocket icon** in the top bar — the built-in setup guide walks you through downloading the CLI and connecting your first agent.
Then head to [Agent configuration](#agent-configuration) for optional tuning.

> [!NOTE]
> `--version 0.0.0-snapshot` is a rolling pre-release alias that tracks the latest build.
> It is required until a stable release is published.

## Deploy to any Kubernetes cluster

```bash
export TOKEN=$(openssl rand -hex 32)
export PASSWORD=changeme

helm install ai-beacon \
  oci://ghcr.io/manusa/charts/ai-beacon \
  --version 0.0.0-snapshot \
  --set ingress.host=ai-beacon.example.com \
  --set auth.token="$TOKEN" \
  --set auth.password="$PASSWORD" \
  -n ai-beacon --create-namespace
```

On clusters with persistent storage, you can omit `auth.token` and `auth.password` — credentials are auto-generated and persisted to the volume. Retrieve them with:

```bash
kubectl exec -n ai-beacon deploy/ai-beacon -- cat /data/password
kubectl exec -n ai-beacon deploy/ai-beacon -- cat /data/token
```

See [Agent configuration](#agent-configuration) for optional tuning.

## Container image

To try the dashboard locally without a cluster:

```bash
podman volume create ai-beacon
podman run --pull=always \
  -e AI_BEACON_AUTH_PASSWORD=demo \
  -p 8080:8080 \
  -v ai-beacon:/data \
  ghcr.io/manusa/ai-beacon:latest
```

Open <http://localhost:8080> and log in with password **demo**.
The dashboard will be empty until you connect an agent — click the **rocket icon** in the top bar for setup instructions.
Then head to [Agent configuration](#agent-configuration) for optional tuning.

> [!IMPORTANT]
> Mount `/data` to a persistent volume (named volume above, or a bind mount). The agent auth token lives there; without a volume, every container restart regenerates it and silently invalidates the token baked into your installed agent hooks — sessions stop appearing on the dashboard until you re-run `ai-beacon install` with the new token.

## Agent configuration

The setup guide covers installing the CLI and connecting to the server.
These additional environment variables are optional but useful:

| Variable | Purpose | Default |
|----------|---------|---------|
| `AI_BEACON_PROJECTS_DIR` | Base directory for your repositories — enables spawning new sessions and worktree workflows from the dashboard | _(disabled)_ |
| `AI_BEACON_DEVICE_NAME` | Friendly name shown in the dashboard for this machine | hostname |

Set them in your shell profile (e.g. `~/.zshrc`) so they apply to every session:

```bash
export AI_BEACON_PROJECTS_DIR=~/projects
export AI_BEACON_DEVICE_NAME=macbook
```

## License

[Apache License 2.0](LICENSE)

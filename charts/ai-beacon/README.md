# ai-beacon

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: latest](https://img.shields.io/badge/AppVersion-latest-informational?style=flat-square)

Helm chart for AI Beacon — dashboard for monitoring and managing AI coding agent sessions.

**Homepage:** <https://github.com/manusa/ai-beacon>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Marc Nuri |  | <https://www.marcnuri.com> |

## Source Code

* <https://github.com/manusa/ai-beacon>

## Installing the Chart

The chart is published as an OCI artifact to `ghcr.io/manusa/charts/ai-beacon` and `quay.io/manusa/charts/ai-beacon` on every main push and on `v*` tags. The examples below use `ghcr.io`; substitute `quay.io` if you prefer.

Until a stable release is published every chart is a pre-release, which Helm only installs when you name a version:

- `--version 0.0.0-snapshot` follows the latest build from `main` (a rolling alias, used in the examples below).
- `--version 0.1.0-snapshot.<YYYYMMDDHHMMSS>` pins one chart build, named after the UTC time of its commit. It pins the chart only: the image is still `image.version` (`latest` by default).

### Plain Kubernetes

```shell
helm upgrade -i -n ai-beacon --create-namespace ai-beacon \
  oci://ghcr.io/manusa/charts/ai-beacon \
  --version 0.0.0-snapshot \
  --set ingress.host=ai-beacon.example.com
```

### OpenShift (including Red Hat Developer Sandbox)

```shell
helm upgrade -i -n ai-beacon --create-namespace ai-beacon \
  oci://ghcr.io/manusa/charts/ai-beacon \
  --version 0.0.0-snapshot \
  --set openshift=true \
  --set persistence.enabled=false \
  --set auth.token="$(openssl rand -hex 32)" \
  --set auth.password="$(openssl rand -hex 32)"
```

With `openshift: true`, the chart renders a `Route` instead of an `Ingress`. When `route.host` is empty the cluster auto-assigns a host from the default ingress domain.

On the Red Hat Developer Sandbox the storage quota is tight, so `persistence.enabled=false` and inline credentials are the reliable install path. On production clusters without quota limits you can drop both overrides and rely on the auto-generate flow.

### OpenShift with OAuth Proxy

When the cluster's OAuth provider should gate the dashboard (so cluster users sign in with their cluster identity instead of a chart-generated password), enable the OAuth Proxy sidecar:

```shell
helm upgrade -i -n ai-beacon --create-namespace ai-beacon \
  oci://ghcr.io/manusa/charts/ai-beacon \
  --version 0.0.0-snapshot \
  --set openshift=true \
  --set oauthProxy.enabled=true \
  --set persistence.enabled=false \
  --set auth.token="$(openssl rand -hex 32)" \
  --set allowedUsers="{$(oc whoami)}" \
  --set route.host=ai-beacon.apps.$CLUSTER_DOMAIN
```

What this enables:

- **OAuth Proxy sidecar** terminating TLS on `:8443` and gating browser traffic with the OpenShift OAuth flow. Cluster users see the dashboard with their cluster identity (`X-Forwarded-User` → ai-beacon's `proxy-header` authenticator).
- **Username allowlist** — only cluster users listed in `allowedUsers` can sign in. Without it, every cluster user with OAuth access would gain dashboard access; the chart `validate` helper rejects the combination at install time.
- **M2M agents bypass the proxy** for routes under `/api/v1/agent/`, so `ai-beacon install` / heartbeats / terminal attach work over the same Route with a Bearer token. No `--auth=none` required.
- **ai-beacon binds to `127.0.0.1:8080`** as defense in depth — a port-forward, NetworkPolicy mistake, or misconfigured Route cannot reach the dashboard without going through the proxy first.
- **TLS reencrypt Route** — external traffic stays on the Route's edge cert; pod-internal traffic stays on the OAuth Proxy's serving cert.

Set `oauthProxy.existingCookieSecret` to a Secret name with key `cookie-secret` if you want OAuth Proxy sessions to survive `helm upgrade`; otherwise the cookie regenerates on each upgrade and users re-sign-in via OAuth.

## Single-replica limitation

ai-beacon keeps session state in-process, so the chart pins `replicaCount: 1` as a default and warns against scaling up. Running multiple replicas shards sessions across pods and breaks dashboard and terminal flows.

## Authentication modes

ai-beacon needs two credentials: `AI_BEACON_AUTH_TOKEN` (agent bearer token) and `AI_BEACON_AUTH_PASSWORD` (browser login password). The chart supports three modes:

1. **Auto-generate (default)** — leave all `auth.*` values empty. The server generates both on first boot and persists them to `/data`. Requires `persistence.enabled: true`. Retrieve the password with:
   ```shell
   kubectl exec -n ai-beacon deploy/ai-beacon -- cat /data/password
   ```
2. **Chart-rendered Secret** — set `auth.token` and/or `auth.password`. The chart renders a Secret named after the release. `persistence.enabled` can then be set to `false`.
3. **Existing Secret** — set `auth.existingSecret` to the name of a pre-existing Secret with keys `token` and `password`. Useful with SealedSecrets / External Secrets Operator.

## WebSockets

ai-beacon uses WebSockets for dashboard push and terminal attach. OpenShift Routes and most Ingress controllers (nginx, HAProxy, Traefik) support this transparently over edge-terminated TLS. No additional annotations are required for default OpenShift deployments.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity rules. |
| allowedUsers | list | `[]` | Usernames permitted to sign in to the dashboard (case-insensitive). Required when `oauthProxy.enabled=true`: rendered into ai-beacon's `--allowed-users` flag and matched against the `X-Forwarded-User` value forwarded by the OpenShift OAuth Proxy (the cluster username from `oc whoami`). Without it the proxy-header authenticator admits every cluster user with OAuth access; the chart `validate` helper rejects the combination at install time. Example: `--set allowedUsers="{alice,bob}"`. See `docs/specs/ai-beacon/auth/implemented-features.md` § User Allowlist. |
| auth | object | `{"existingSecret":"","password":"","token":""}` | / External Secrets Operator. |
| auth.existingSecret | string | `""` | no Secret is rendered and the Deployment references this Secret directly. |
| auth.password | string | `""` | Browser login password (`AI_BEACON_AUTH_PASSWORD`). |
| auth.token | string | `""` | Agent bearer token (`AI_BEACON_AUTH_TOKEN`). |
| defaultPodSecurityContext | object | `{"seccompProfile":{"type":"RuntimeDefault"}}` | `restricted-v2` SCC can inject namespace-range values. |
| defaultSecurityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"runAsNonRoot":true}` | Default container SecurityContext applied when `securityContext` is empty. |
| env | list | `[]` | `env` entries, supports value / valueFrom / template expressions via `tpl`. |
| envFrom | list | `[]` | Extra `envFrom` sources for the container (ConfigMap / Secret refs). |
| extraContainers | list | `[]` | Extra sidecar containers. |
| extraVolumeMounts | list | `[]` | Extra volumeMounts on the main container. |
| extraVolumes | list | `[]` | Extra volumes on the Deployment. |
| fullnameOverride | string | `""` | Override the fully-qualified release name. |
| image | object | `{"pullPolicy":"Always","registry":"ghcr.io","repository":"manusa/ai-beacon","version":"latest"}` | Container image. |
| image.pullPolicy | string | `"Always"` | `IfNotPresent` once you pin `version` to an immutable tag or digest. |
| image.registry | string | `"ghcr.io"` | Image registry. |
| image.repository | string | `"manusa/ai-beacon"` | Image repository. |
| image.version | string | `"latest"` | kubernetes-mcp-server chart convention. |
| imagePullSecrets | list | `[]` | Image pull secrets, more information can be found here: https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/ |
| ingress | object | `{"annotations":{},"className":"","enabled":true,"host":"","path":"/","pathType":"ImplementationSpecific","tls":[]}` | Ingress configuration. Ignored when `openshift: true` (use `route` instead). |
| ingress.annotations | object | `{}` | Annotations for the Ingress. |
| ingress.className | string | `""` | IngressClassName. |
| ingress.enabled | bool | `true` | Whether to render an Ingress resource. |
| ingress.host | string | `""` | Host for the Ingress rule. Supports template expressions. |
| ingress.path | string | `"/"` | Path for the Ingress rule. |
| ingress.pathType | string | `"ImplementationSpecific"` | PathType for the Ingress rule. |
| ingress.tls | list | `[]` | secretName: ai-beacon-tls |
| initContainers | list | `[]` | Init containers. |
| livenessProbe | object | `{"httpGet":{"path":"/login","port":"http"},"initialDelaySeconds":5,"periodSeconds":15}` | the server is ready. |
| nameOverride | string | `""` | Override the chart name. |
| nodeSelector | object | `{}` | Node selector. |
| oauthProxy.enabled | bool | `false` | Whether to render the OAuth Proxy sidecar. When true, the chart also flips ai-beacon's `--address` to `127.0.0.1:8080`, switches the Service to target the proxy port (`:8443`), annotates the ServiceAccount for OpenShift OAuth redirects, and (when `openshift: true`) flips the Route to TLS reencrypt. |
| oauthProxy.existingCookieSecret | string | `""` | Existing Secret name with key `cookie-secret` for OAuth Proxy. When empty, the chart renders its own Secret with a random value; `helm upgrade` regenerates it and users re-authenticate via OAuth. |
| oauthProxy.image | string | `"quay.io/openshift/origin-oauth-proxy:4.22"` | OAuth Proxy container image. Pinned by tag (no entitlement required); override with `registry.redhat.io/openshift4/ose-oauth-proxy-rhel9` on entitled clusters. |
| oauthProxy.imagePullPolicy | string | `"IfNotPresent"` | Pull policy for the OAuth Proxy image. Independent of `image.pullPolicy` so a locally-built ai-beacon image (`Never`) does not also disable pulls of the upstream proxy image. |
| oauthProxy.skipAuthRegex | string | `"^/api/v1/agent/"` | OAuth Proxy `--skip-auth-regex` value (generated; do not edit). `make helm-gen-oauth-proxy-config` writes this from the Go-side `server.AgentRoutePrefix` constant; CI fails on drift. |
| openshift | bool | `false` | rendered (see `route` values below) and the `Ingress` is not rendered. |
| persistence | object | `{"accessModes":["ReadWriteOnce"],"annotations":{},"enabled":true,"existingClaim":"","resourcePolicy":"keep","size":"1Gi","storageClass":""}` | `auth.password` / `auth.existingSecret`. |
| persistence.accessModes | list | `["ReadWriteOnce"]` | PVC access modes. |
| persistence.annotations | object | `{}` | Annotations for the rendered PVC (merged after `resourcePolicy`). |
| persistence.enabled | bool | `true` | When false, `/data` is backed by an `emptyDir` volume (ephemeral). |
| persistence.existingClaim | string | `""` | at `/data` by the pod's UID/GID. |
| persistence.resourcePolicy | string | `"keep"` | always delete the PVC manually with `kubectl delete pvc` afterwards. |
| persistence.size | string | `"1Gi"` | PVC storage request. |
| persistence.storageClass | string | `""` | StorageClass to use. Empty string means use the cluster default. |
| podAnnotations | object | `{}` | Pod annotations. |
| podLabels | object | `{}` | Pod labels. |
| podSecurityContext | object | `{}` | `defaultPodSecurityContext`. |
| readinessProbe | object | `{"httpGet":{"path":"/login","port":"http"},"initialDelaySeconds":2,"periodSeconds":10}` | Readiness probe. |
| replicaCount | int | `1` | terminal flows. Leave at `1`. |
| resources | object | `{"limits":{"cpu":"500m","memory":"256Mi"},"requests":{"cpu":"50m","memory":"64Mi"}}` | Resource requests and limits. |
| route | object | `{"annotations":{},"enabled":true,"host":"","insecureEdgeTerminationPolicy":"Redirect","path":"","termination":"edge"}` | termination. |
| route.annotations | object | `{}` | Annotations for the Route. |
| route.enabled | bool | `true` | Whether to render a Route when `openshift: true`. |
| route.host | string | `""` | ingress domain (e.g. `*.apps.<cluster>.openshiftapps.com`). |
| route.insecureEdgeTerminationPolicy | string | `"Redirect"` | Behavior for insecure traffic: `Allow`, `Redirect`, or `None`. |
| route.path | string | `""` | Route path. Empty string means all paths. |
| route.termination | string | `"edge"` | TLS termination: `edge`, `passthrough`, or `reencrypt`. |
| securityContext | object | `{}` | `defaultSecurityContext`. |
| service | object | `{"annotations":{},"port":8080,"targetPort":"http","type":"ClusterIP"}` | Service configuration. |
| service.annotations | object | `{}` | Annotations for the Service. |
| service.port | int | `8080` | Service port. |
| service.targetPort | string | `"http"` | Target port on the pod. Defaults to the named `http` port. |
| service.type | string | `"ClusterIP"` | Service type. |
| serviceAccount | object | `{"annotations":{},"create":true,"name":""}` | so no RBAC is created. |
| serviceAccount.annotations | object | `{}` | Annotations to add to the ServiceAccount. |
| serviceAccount.create | bool | `true` | Whether to create a ServiceAccount for the release. |
| serviceAccount.name | string | `""` | when empty and `create` is true. |
| tolerations | list | `[]` | Tolerations. |

## Updating this README

This file is generated from `values.yaml` comments by [helm-docs](https://github.com/norwoodj/helm-docs). Regenerate after changing values:

```shell
make -C ai-beacon helm-docs
```

CI fails if the committed README is out of sync with `values.yaml`.

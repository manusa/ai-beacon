# AI Beacon

A web dashboard for monitoring and orchestrating AI coding agents across your devices.

AI Beacon gives you a single pane of glass over every `claude` (and other coding-agent) session you're running. See which agents are working, what they're working on, attach a terminal, spawn new sessions — all from a browser, on any device that can reach the dashboard.

## Status

This project is in **early access**. Source will be published to this repository once the distribution pipeline and onboarding story are validated with early users. In the meantime, pre-built artifacts are available below.

## Try It

### Container image

```bash
podman run -p 8080:8080 ghcr.io/manusa/ai-beacon:latest
# then: open http://localhost:8080
```

Images are also mirrored to `quay.io/manusa/ai-beacon`.

### Snapshot binary

A rolling `snapshot` pre-release tracks the current development tip. Cross-compiled for macOS, Linux, and Windows (amd64 + arm64):

```bash
# Linux x86_64
curl -L -o ai-beacon \
  https://github.com/manusa/ai-beacon/releases/download/snapshot/ai-beacon-linux-amd64
chmod +x ai-beacon

./ai-beacon server               # dashboard on :8080
./ai-beacon session -- claude    # wrap a coding agent in another terminal
```

See the [snapshot release](https://github.com/manusa/ai-beacon/releases/tag/snapshot) for all platform downloads.

### Helm chart (Kubernetes / OpenShift)

```bash
helm install ai-beacon \
  oci://ghcr.io/manusa/charts/ai-beacon \
  --version snapshot \
  -n ai-beacon --create-namespace
```

> [!NOTE]
> `--version snapshot` is a rolling alias that always points to the
> latest main-branch build. It is required until a stable `v0.1.0` is
> released, because Helm's OCI resolver skips pre-release versions
> (`0.1.0-dev.<sha>`) when no explicit version is given. Pin to a
> specific `0.1.0-dev.<sha>` tag for reproducible deployments.

## License

Licensed under the [Apache License 2.0](./LICENSE).

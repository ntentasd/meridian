# Meridian

Meridian is a Kubernetes controller that watches `Ingress` and `HTTPRoute`
resources and exposes them through a unified HTTP API. It maintains an in-memory
store of route entries and streams real-time updates to connected clients over SSE.

## API

| Endpoint          | Description                              |
| ----------------- | ---------------------------------------- |
| `GET /api/routes` | JSON snapshot of all current routes      |
| `GET /api/events` | SSE stream of route upsert/delete events |

Route entries are enriched from annotations
(`owner`, `description`, `docs-url`, `environment`, `tags`).

## Prerequisites

- Go v1.26+
- kubectl v1.11.3+
- Access to a Kubernetes cluster with [Gateway API CRDs](https://gateway-api.sigs.k8s.io/guides/#installing-gateway-api)
  installed

## Getting Started

### Run locally against a cluster

```sh
make run
```

### Deploy to cluster

```sh
make docker-build docker-push IMG=<registry>/meridian:tag
make deploy IMG=<registry>/meridian:tag
```

### Uninstall

```sh
make undeploy
```

## Development

```sh
make test       # run tests
make lint       # run golangci-lint (uses ./bin/golangci-lint with logcheck plugin)
make lint-fix   # run golangci-lint with auto-fix
```

## License

Copyright 2026. Licensed under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0).

# Spotlite Kubernetes

This folder now uses a simple per-service layout.

## Deploy command

```bash
kubectl apply -k k8s
```

## Files

- `k8s/services/`: one file per app service deployment
- `k8s/deps/`: databases, Redis, NATS, HDFS, Neo4j
- `k8s/platform/`: ingress and routing
- `k8s/secrets/spotlite-app-secrets.example.yaml`: app secret template

## Required secrets before deploy

Create namespace first:

```bash
kubectl apply -f k8s/namespace.yaml
```

1. App/env secret:

```bash
kubectl apply -n spotlite -f k8s/secrets/spotlite-app-secrets.example.yaml
```

2. Crypto secret (`spotlite-crypto`) with certs/keys:

Required keys:

- `rootCA.crt`
- `nats.crt`, `nats.key`
- `user-service.crt`, `user-service.key`
- `content-service.crt`, `content-service.key`
- `notification-service.crt`, `notification-service.key`
- `subscription-service.crt`, `subscription-service.key`
- `rating-service.crt`, `rating-service.key`
- `recommendation-service.crt`, `recommendation-service.key`
- `private.pem`, `public.pem`

Example:

```bash
kubectl -n spotlite create secret generic spotlite-crypto \
  --from-file=rootCA.crt=/path/to/rootCA.crt \
  --from-file=nats.crt=/path/to/nats.crt \
  --from-file=nats.key=/path/to/nats.key \
  --from-file=user-service.crt=/path/to/user-service.crt \
  --from-file=user-service.key=/path/to/user-service.key \
  --from-file=content-service.crt=/path/to/content-service.crt \
  --from-file=content-service.key=/path/to/content-service.key \
  --from-file=notification-service.crt=/path/to/notification-service.crt \
  --from-file=notification-service.key=/path/to/notification-service.key \
  --from-file=subscription-service.crt=/path/to/subscription-service.crt \
  --from-file=subscription-service.key=/path/to/subscription-service.key \
  --from-file=rating-service.crt=/path/to/rating-service.crt \
  --from-file=rating-service.key=/path/to/rating-service.key \
  --from-file=recommendation-service.crt=/path/to/recommendation-service.crt \
  --from-file=recommendation-service.key=/path/to/recommendation-service.key \
  --from-file=private.pem=/path/to/private.pem \
  --from-file=public.pem=/path/to/public.pem
```

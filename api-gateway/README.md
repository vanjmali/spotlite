# API Gateway

API Gateway is a Traefik-based edge for routing, TLS termination, and service discovery (started via [`docker-compose.yml`](./../docker-compose.yml)).

## Usage

- Exposes entrypoints `web` (:80) and `dashboard` (:8080).
- Loads static config from [`traefik.yml`](./traefik.yml).
- Uses Docker labels on services to create routers/services automatically.
- Watches [`dynamic/`](./dynamic/) for optional file-based middlewares/routers (hot-reloaded).

## Layout

- Route frontend at `/` and backend APIs under `/api/...`.
- Define routers via Docker labels (current Compose approach). In Kubernetes these move to Ingress/IngressRoute resources. The file provider (`dynamic/`) is optional for shared middlewares or non-Docker backends.

## Deployment

- Replace the dashboard’s `insecure: true` setting before any public exposure.

### Kubernetes

Kubernetes deployment outline:

- Deploy Traefik via Helm as an IngressController.
- Define `IngressRoute` (or standard `Ingress`) for each service path; reuse `Middleware` CRDs for shared concerns (auth, headers, rate limits).
- Expose entrypoints through a Service (NodePort/LoadBalancer); TLS handled via ACME or cert-manager.
- Keep route structure consistent with Compose: `/` -> frontend Service, `/api/users` -> user-service, etc.
- Drop Docker provider; rely on K8s resources for discovery instead of labels.

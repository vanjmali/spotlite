# Spotlite

A simple clone of Spotify, a music streaming platform built using Go microservices with an Angular frontend.
For education purposes only.

## Structure

- `frontend` – Angular web client for browsing, playback, and user features.
- `api-gateway` – Traefik-based edge for routing, TLS termination, and service discovery.
- `user-service` – Authentication, authorization, and account management.
- `content-service` – Catalog of artists, albums, songs, and genres.
- `rating-service` – Song rating endpoints and aggregation.
- `subscription-service` – Manages user subscriptions to artists/genres.
- `notification-service` – Queues and delivers user notifications.
- `recommendation-service` – Personalized recommendations and feeds.
- `analytics-service` – Activity tracking and analytics endpoints.

## Tech stack

- Backend: Go >= 1.22, REST/JSON APIs, containerized per service.
- Frontend: Angular and Node.js.
- Gateway: Traefik for routing and edge concerns.
- Tooling: Docker & Docker Compose for local development.

## How to run

Prerequisites: Go 1.22+, Node.js v24.x, Docker, and Docker Compose.

```bash
# start everything
docker compose up --build

# rebuild and restart one component (example: user-service)
docker compose up -d --build user-service
```

## Contributing

Create a feature branch, make changes, and submit a pull request.

`develop` is the default development branch. `stable` is the protected production branch.

## License

[MIT License](LICENSE)

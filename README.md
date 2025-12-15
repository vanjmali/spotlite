# Spotlite

A simple clone of Spotify, a music streaming platform built using Go microservices with an Angular frontend.
For education purposes only.

## Structure

- `frontend` – Angular web client for browsing, playback, and user features.
- `api-gateway` – Traefik-based edge for routing, TLS termination, and service discovery.
- `user-service` – Authentication, authorization, and account management.
- `content-service` – Catalog of artists, albums, songs, and genres.
- `ratings-service` – Song rating endpoints and aggregation.
- `subscriptions-service` – Manages user subscriptions to artists/genres.
- `notifications-service` – Queues and delivers user notifications.
- `recommendation-service` – Personalized recommendations and feeds.
- `analytics-service` – Activity tracking and analytics endpoints.

> `api-gateway` is the main entrypoint and serves both frontend and backend services:
>
> - `frontend` is served at the root (`/`).
> - other microservices are served at `/api/<service-name>/...` (e.g. `/api/users/...`); Traefik strips the `/api/<service-name>` prefix before forwarding the request.

> [!NOTE] Read `README.md` located in each service for more details.
> Such as the port acccssible and the environment variables used for development.

## Tech stack

- Backend: Go >= 1.22, REST/JSON APIs, containerized per service.
- Frontend: Angular and Node.js.
- Gateway: Traefik for routing and edge concerns.
- Tooling: Docker & Docker Compose for local development.

## Setup

Prerequisites: Go 1.22+, Node.js v24.x, Docker, and Docker Compose.

Setup:

1. Copy `.env.example` to `.env`
2. Populate .env with your own values if needed.
3. Start everything: `docker compose up --build`

To rebuild/restart one service (example: user-service): `docker compose up -d --build user-service`

## Development

To run in development mode, follow the [Setup](#setup) steps.

There are additional services available for local development:

- `localhost:8080` - Traefik dashboard
- `localhost:3101` - User Service's MongoDB direct connection
- `localhost:3000/dev/user-service` - [Mongo Express](https://github.com/mongo-express/mongo-express) to User Service

## Contributing

Create a feature branch, make changes, and submit a pull request.

The `develop` branch is the main development branch; `main` is the production branch.

### Creating new services

#### Docker

Each service has a `docker-compose.yml` file that gives instructions for Docker on how to run it as well as its dependencies.

You can use the `Dockerfile.microservice` for building the service image. See other services for examples.

The dependencies used should have a prefix, usually the name of the service, i.e. `user-service` has a prefix `user-` (e.g. `user-mongodb`).
This avoids name conflicts with other services during development.

The `COMPOSE_FILE` variable in `.env.example` and `.env` should be updated to include the new service, for both shells displayed.

> [!NOTE]
> In case you are getting a `WARN[0000] The "XYZ" variable is not set. Defaulting to a blank string.` message when running `docker compose up`; make sure to update your `.env` file.

## License

[MIT License](LICENSE)

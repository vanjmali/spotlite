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
- `common-lib` – Shared Go library for common utilities and types.

> `api-gateway` is the main entrypoint and serves both frontend and backend services:
>
> - `frontend` is served at the root (`/`).
> - other microservices are served at `/api/<service-name>/...` (e.g. `/api/users/...`); Traefik strips the `/api/<service-name>` prefix before forwarding the request.

> [!NOTE]
> Read `README.md` located in each service for more details.
> Such as the port acccssible and the environment variables used for development.

## Tech stack

- Backend: Go >= 1.22, REST/JSON APIs, containerized per service.
- Frontend: Angular and Node.js.
- Gateway: Traefik for routing and edge concerns.
- Tooling: Docker & Docker Compose for local development.

## Setup

Prerequisites: Go 1.22+, Node.js v24.x, Docker, and Docker Compose.

Steps:

1. Copy `.env.example` to `.env`
2. Populate .env with your own values if needed.
3. Start everything: `docker compose up --build -d`

To rebuild/restart one service (example: user-service): `docker compose up -d --build user-service`

## Development

### Golang Setup

Install [`golangci-lint`](https://golangci-lint.run/docs/welcome/install/local/) to lint Go code.

### Frontend Setup

See [`frontend/README.md`](frontend/README.md).

### Development Services

There are additional services available for local development:

Dashboards and tools:

- http://localhost:8025 - MailHog web interface for viewing sent emails
- http://localhost:8080 - Traefik dashboard
- http://localhost:3000/dev/jaeger - Jaeger UI for viewing traces (via Traefik)
- http://localhost:3000/dev/user-service - [Mongo Express](https://github.com/mongo-express/mongo-express) to User Service

Database connections for services:

- http://localhost:3101 - User Service's MongoDB direct connection
- http://localhost:3102 - Content Service's MongoDB direct connection

### Useful commands

#### Docker

- `docker compose up -d --build <service-name>` - Rebuild and restart a specific service.
  - `docker compose up -d --build` - Rebuild and (re)start all services.
- `docker compose logs <service-name>` - Show logs for a specific service.
  - `docker compose logs -f <service-name>` - Follow logs for a specific service.
- `docker compose down -v` - Stop and remove all containers, **volumes** (`-v`), and networks. (complete reset)

#### MongoDB

You can install [`mongosh`](https://www.mongodb.com/docs/mongodb-shell/) command line tool to communicate with the database via a shell.

- `mongosh mongodb://localhost:3101/user-service --username mongo --password 123456` - connect to User Service's database

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

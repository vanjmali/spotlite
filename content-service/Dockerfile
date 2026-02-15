FROM golang:1.25.5-alpine AS build

# Build common-lib package
WORKDIR /common-lib

COPY --from=common-lib go.mod ./
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	go mod download

COPY --from=common-lib ./ ./

# Build app-service binary
WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	go mod download

COPY ./ ./

RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	CGO_ENABLED=0 go build \
	-ldflags="-s -w" \
	-v \
	-trimpath \
	-o /app/build.bin \
	main.go

FROM alpine:3.21

RUN apk add --no-cache ffmpeg ca-certificates

WORKDIR /app

COPY --from=build /app/build.bin /usr/bin/service

RUN addgroup -S app && adduser -S app -G app
USER app

ENTRYPOINT ["/usr/bin/service"]

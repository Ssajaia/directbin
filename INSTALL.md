# Install and run

## Requirements

- Go 1.22 or newer to build and run from source, as declared in `go.mod`.
- Docker Engine to build and run the container.
- Docker Compose v2 (`docker compose`) to use `compose.yaml`.

## Run from source

```sh
git clone https://github.com/Ssajaia/directbin.git
cd directbin
go run .
```

The server listens on `:8080` by default and stores pastes under `data/` relative to its working directory. To use another address, set `ADDR`:

```sh
ADDR=:9090 go run .
```

Build a binary with:

```sh
go build -o directbin .
./directbin
```

Set `ADDR` in the environment in the same way when running the binary.

## Run with Docker

Build the image and start a container with port 8080 published and a named volume mounted at `/data`:

```sh
docker build -t directbin:local .
docker run -d --name directbin -p 8080:8080 -v directbin-data:/data directbin:local
```

The container listens on `:8080` by default. If changing `ADDR`, set it to the container's listening address and publish that container port to the desired host port. The image's health check currently probes port 8080, so changing the listening port also requires changing the health-check configuration if Docker health status is to remain accurate.

The named volume keeps paste files separate from the container and persists when the container is stopped or recreated. Docker runs the application as the non-root `directbin` user. The image health check requests `GET /health`.

Stop the container with:

```sh
docker stop directbin
```

## Run with Docker Compose

Start the service from the repository root:

```sh
docker compose up --build
```

The Compose service publishes host port 8080 to container port 8080 and mounts the named volume `directbin-data` at `/data`. The Dockerfile health check is used. Compose allows 15 seconds for graceful shutdown.

Stop the service with:

```sh
docker compose down
```

This removes the service container but keeps the named volume by default, so paste files persist across container recreation.

## Verify the service

Check the health endpoint:

```sh
curl http://localhost:8080/health
```

Create a paste and copy the returned ID:

```sh
curl -X POST http://localhost:8080/paste --data 'hello from computer A'
```

Retrieve its content using that ID:

```sh
curl http://localhost:8080/paste/<id>/raw
```

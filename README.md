# DirectBin

DirectBin is a small HTTP service for moving text and small file contents between machines with `curl`. A client sends content, receives a short ID, and uses that ID to retrieve the content or save it as a file.

## How it works

```text
Client
  ↓
HTTP API
  ↓
ID generation
  ↓
data/<id>
```

Each paste is stored as a file in `data/`. Its generated, URL-safe ID is the filename, and its contents are the request body without parsing or transformation.

## API

| Method | Endpoint | Behavior |
| --- | --- | --- |
| `POST` | `/paste` | Store the request body and return `{"id":"..."}` with status `201 Created`. |
| `GET` | `/paste/{id}` | Return the stored content. |
| `GET` | `/paste/{id}/raw` | Return the stored content directly for CLI use. |
| `DELETE` | `/paste/{id}` | Delete the paste and return `204 No Content`. |
| `GET` | `/health` | Return `{"status":"ok"}`. |

Missing pastes return `404 Not Found`; unsupported methods return `405 Method Not Allowed`.

## Running DirectBin

Run the service directly with `go run .`; it listens on `:8080` by default. Set `ADDR` to configure its listening address.

Send content from one machine:

```sh
curl -X POST http://localhost:8080/paste --data 'hello from another machine'
```

Use the returned ID to retrieve the content on another machine:

```sh
curl http://localhost:8080/paste/<id>/raw
curl http://localhost:8080/paste/<id>/raw > message.txt
```

Run the tests with `go test ./...`.

## Docker

The container runs the compiled Go application as its main process. Its health check requests the existing `/health` endpoint.

```text
Host
 │
 ├── DirectBin container
 │       └── Go application
 │
 └── Docker volume
         └── /data
```

The application is disposable; the named volume keeps `data/` separate so paste files can outlive a stopped or recreated container.

Build and run the image with persistent storage:

```sh
docker build -t directbin:local .
docker run -d --name directbin --stop-timeout 15 -p 8080:8080 -v directbin-data:/data directbin:local
docker stop directbin
```

The container listens on `:8080` by default. Configure `ADDR` and the published port together when using a different port.

Or start and stop the Compose service:

```sh
docker compose up --build
docker compose down
```

Compose also stores pastes in a named volume. Removing the service container does not remove this volume by default.

## Architecture and scaling

DirectBin v0.1.0 uses a single Go HTTP server and local filesystem storage instead of a database, keeping its deployment simple. A future deployment could run multiple DirectBin instances behind a service or load balancer, but **multiple DirectBin containers cannot safely rely on separate local container filesystems as shared application storage.** Shared storage must be addressed before meaningful horizontal scaling or Kubernetes deployment.

The [`k8s/`](./k8s/) directory is reserved for future Kubernetes configuration; Kubernetes resources are not included in this version.

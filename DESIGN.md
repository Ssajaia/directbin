# Design decisions

DirectBin is intentionally a small HTTP service with a narrow purpose: accept content, assign an ID, store it as a local file, and return it later.

## Go standard library

The HTTP server, routing, JSON response, cryptographic randomness, filesystem operations, and tests use Go's standard library. The API has only a few routes, so a web framework or additional dependency would add concepts and maintenance without solving a current need.

## Filesystem storage

Each paste is a file named after its generated ID under `data/`. This keeps local development simple, provides persistence on the host filesystem, and avoids operating a database or external storage service. It suits the current single-instance design.

The trade-off is that separate instances have separate filesystems by default. A Compose volume keeps one deployment's data across container recreation, but does not provide distributed shared storage. Meaningful horizontal scaling would require choosing and implementing a shared storage design.

## Short generated IDs

The server reads six bytes from `crypto/rand` and encodes them using URL-safe Base64 without padding, producing an eight-character ID suitable for URLs and filenames. Creation uses exclusive file creation and retries collisions up to ten times, preventing an existing paste from being overwritten.

IDs identify pastes; they do not authenticate users or grant controlled access. The API does not check ownership or permissions.

## Raw endpoint

`GET /paste/{id}/raw` returns the stored bytes directly in the HTTP response body. It supports CLI workflows such as:

```sh
curl http://localhost:8080/paste/a8K29xQ_/raw
curl http://localhost:8080/paste/a8K29xQ_/raw > config.yaml
```

The raw route currently uses the same retrieval handler as `GET /paste/{id}`.

## Docker

Docker packages a compiled binary in a runtime image rather than running a Go development environment. This makes the application easier to run consistently and as a non-root user. The `/data` volume separates paste files from the disposable container so a single instance's data can persist across container recreation.

The image includes a health check for `/health` and runs the binary directly as its main process. The Go server handles interrupt and termination signals with a bounded graceful-shutdown period.

## Docker Compose

Compose provides a one-service local container deployment with port 8080 published and a named volume mounted at `/data`. It avoids requiring a database or other service and makes single-instance persistence straightforward.

## Deliberate scope

PostgreSQL, Redis, Kafka, RabbitMQ, object storage, authentication systems, Kubernetes, and web frameworks are outside the current requirements. Their absence is a scope choice, not a claim that they are unsuitable in other architectures.

If DirectBin needs true horizontal scaling, it will need a shared storage strategy and deployment behavior that ensures all instances can consistently read and write the same paste data. Kubernetes resources alone would not provide that application-level storage behavior.

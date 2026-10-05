# API reference

DirectBin accepts raw content for paste creation and stores it without parsing or transformation. The default server address is `:8080`.

## Endpoints

| Method | Endpoint | Purpose |
| --- | --- | --- |
| `POST` | `/paste` | Create a paste |
| `GET` | `/paste/{id}` | Retrieve a paste |
| `GET` | `/paste/{id}/raw` | Retrieve raw paste content |
| `DELETE` | `/paste/{id}` | Delete a paste |
| `GET` | `/health` | Check that the HTTP handler responds |

## `POST /paste`

Stores the request body as the paste content. The server treats it as bytes and does not parse it as JSON or another format.

```sh
curl -X POST http://localhost:8080/paste --data 'hello from computer A'
```

On success, the response is `201 Created`, with `Content-Type: application/json` and a JSON object containing the generated ID:

```json
{"id":"a8K29xQ_"}
```

The ID in this example is illustrative. The generated ID consists of eight URL-safe characters.
The JSON encoder appends a newline to the response.

| Status | Response body | Headers and behavior |
| --- | --- | --- |
| `201 Created` | JSON object containing `id`, followed by a newline. | `Content-Type: application/json`; paste stored. |
| `400 Bad Request` | `could not read request body` followed by a newline. | Error response. |
| `405 Method Not Allowed` | `method not allowed` followed by a newline. | `Allow: POST`; error response. |
| `500 Internal Server Error` | `could not store paste` followed by a newline. | ID generation or storage failed. |

## `GET /paste/{id}`

Returns the stored paste bytes with `Content-Type: application/octet-stream`.

`id` is the eight-character URL-safe paste ID.

| Status | Response body | Headers and behavior |
| --- | --- | --- |
| `200 OK` | Stored bytes, unchanged. | `Content-Type: application/octet-stream`. |
| `404 Not Found` | `paste not found` followed by a newline for a valid ID without a file. | Invalid ID returns the standard not-found response. |
| `405 Method Not Allowed` | `method not allowed` followed by a newline. | `Allow: GET, DELETE`. |
| `500 Internal Server Error` | `could not read paste` followed by a newline. | Filesystem read failed. |

## `GET /paste/{id}/raw`

Returns the same stored bytes through the same retrieval handler as `GET /paste/{id}`. The response is intended for command-line use:

```sh
curl http://localhost:8080/paste/a8K29xQ_/raw
curl http://localhost:8080/paste/a8K29xQ_/raw > config.yaml
```

| Status | Response body | Headers and behavior |
| --- | --- | --- |
| `200 OK` | Stored bytes, unchanged. | `Content-Type: application/octet-stream`. |
| `404 Not Found` | `paste not found` followed by a newline for a valid ID without a file. | Invalid ID returns the standard not-found response. |
| `405 Method Not Allowed` | `method not allowed` followed by a newline. | `Allow: GET`. |
| `500 Internal Server Error` | `could not read paste` followed by a newline. | Filesystem read failed. |

## `DELETE /paste/{id}`

Deletes the paste named by the eight-character URL-safe `id`.

| Status | Response body | Headers and behavior |
| --- | --- | --- |
| `204 No Content` | Empty. | Paste deleted. |
| `404 Not Found` | `paste not found` followed by a newline for a valid ID without a file. | Invalid ID returns the standard not-found response. |
| `405 Method Not Allowed` | `method not allowed` followed by a newline. | `Allow: GET, DELETE`. |
| `500 Internal Server Error` | `could not delete paste` followed by a newline. | Filesystem removal failed. |

## `GET /health`

Returns `200 OK` with `Content-Type: application/json` and a JSON-encoded body (including a trailing newline):

```json
{"status":"ok"}
```

The Docker image uses this endpoint for its health check. The handler confirms that the HTTP endpoint responds; it does not check the storage directory.

| Status | Response body | Headers and behavior |
| --- | --- | --- |
| `200 OK` | `{"status":"ok"}` followed by a newline. | `Content-Type: application/json`. |
| `405 Method Not Allowed` | `method not allowed` followed by a newline. | `Allow: GET`. |

## Other paths and errors

Unrecognized paths, invalid IDs, and unsupported path shapes return `404 Not Found` with `404 page not found` followed by a newline. Error responses use Go's `http.Error` behavior, including `Content-Type: text/plain; charset=utf-8` and `X-Content-Type-Options: nosniff`. The server does not define a separate JSON error format.

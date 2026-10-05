# Security

DirectBin has a deliberately small security model. It is suitable for trusted, controlled environments, but the current implementation does not provide the protections expected of a public paste service.

## Current protections

- Paste IDs are generated from six bytes read with Go's `crypto/rand`, then encoded with URL-safe Base64 without padding. This produces eight-character IDs.
- Paste creation uses exclusive file creation, so an ID collision does not overwrite an existing paste. The server tries up to ten generated IDs.
- IDs are checked against an eight-character URL-safe character set before being used in a path.
- Paste files are created with mode `0600` on the filesystem.
- The Docker image runs the application as the non-root `directbin` user.
- The API accepts only the methods implemented for each route. Invalid IDs and unrecognized paths are not served as paste files.

These controls do not authenticate clients or make the service safe to expose publicly.

## IDs are not access control

An ID is a locator for a paste, not an authentication credential or authorization check. Anyone who knows or obtains an ID can retrieve or delete that paste. IDs must not be relied on to protect sensitive content.

## Storage and container data

Paste files are stored as `data/<id>` on the server's local filesystem. File permissions are not a substitute for API authorization: the service itself can read and delete every paste, and any process or operator with access to the storage can inspect them.

The Docker setup mounts a named volume at `/data`, which persists paste files beyond the container's lifetime. Anyone with access to the Docker host or that volume can access the stored data. Removing the volume destroys that persisted data.

There is no encryption at rest in the application. Filesystem or volume encryption, backups, access control, and retention are outside the current implementation.

## Network and API exposure

The API has no authentication or authorization. It has no TLS handling, rate limiting, request-size limit, or abuse protection. The HTTP server reads the entire POST body into memory before storing it, so large or repeated requests can consume memory and disk space.

Retrieval and deletion are available to any client that can reach the server and knows a paste ID. Exposing the service publicly can therefore lead to unauthorized reading or deletion, storage exhaustion, and disclosure of submitted data. Direct HTTP traffic is unencrypted unless protected by infrastructure outside this application.

Run DirectBin only on a trusted network or place it behind appropriately configured access controls and encrypted transport. Do not store passwords, tokens, personal data, or other sensitive content in the current service.

## Reporting a vulnerability

No dedicated security contact is documented in this repository. Do not publish exploitable details in a public issue. Use the repository's private vulnerability reporting feature if it is available; otherwise contact the maintainers privately through an existing repository channel.

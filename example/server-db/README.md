# Server-Db Example

This is an example project using Bob to build and run a rest server writing and reading from a database.

---
## Prerequisites

It requires oapi-codegen to generate server boilerplate from a openapi.yaml
As a gopher you could just do a:
```bash
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.9.0
```

---
## Build tasks

To build the server binary, run:

```bash
$ bob build server
```

The server depends on the `generate-api` task, which, in turn, generates the HTTP server code from the `openapi.yaml`
file.

Bob uses intelligent caching so that it doesn't have to do work that it already has done. If you run `bob build server`
again, it will be much faster, since it will use the cached build targets.

You can observe that if you modify the `openapi.yaml` or `main.go` and run `bob build server` again, the necessary tasks
will be re-run, and not use the cache, since their inputs have changed.

---
## Run tasks

To run the server in its environment, run:

```bash
$ bob run server
```

This will make sure the server is properly built (with all its dependencies), ramp up a docker-compose environment with
the redis database running as a service, and will then start the server binary. You can then check the outputs of both
the server and the database, restart them, check their logs and stop them through the built-in TUI Bob offers.

Example:

![bob-tui](example/server-db/assets/bob-tui.gif?raw=true "Bob TUI")


You can test the server's HTTP REST API by using an HTTP client.
- `GET /api/ping`: Q(^o^)
- `POST /api/items` *(with JSON string as request body)*: creates a new item.
- `GET /api/item/:itemId`: returns the item with the given id.

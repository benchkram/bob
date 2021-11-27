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

### Building the API

To generate the HTTP API server code from the openapi.yaml file:

```bash
$ bob build generate-api
```

![bob-build-generate-api](../../example/server-db/assets/bob-build-generate-api.png?raw=true "bob build generate-api")

Bob uses intelligent caching so that it doesn't have to do work that it already has done. If you run
`bob build generate-api` again, it will be much faster, since it will use the cached build targets.

![bob-build-generate-api-cached](../../example/server-db/assets/bob-build-generate-api-cached.png?raw=true "bob build generate-api cached")

### Building the server

```bash
$ bob build    # or `bob build build` (`build` is the default task)
```

The server depends on the `generate-api` task, which, in turn, generates the HTTP server code from the `openapi.yaml`
file. If the `generate-api` task has been run, the server will be built without having to run it again.

![bob-build](../../example/server-db/assets/bob-build.png?raw=true "bob build")

### Rebuilding

You can observe that if you modify the `openapi.yaml` or `main.go` and run `bob build server` again, the necessary tasks
will be re-run, and not use the cache, since their inputs have changed.

---
## Run tasks

To run the server in its environment in the Bob Terminal User Interface (TUI), run:

```bash
$ bob run server
```

![bob-run-server](../../example/server-db/assets/bob-run-server.gif?raw=true "bob run server")

This will make sure the server is properly built (with all its dependencies), ramp up a docker-compose environment with
the redis database running as a service, and will then start the server binary. You can then check the outputs of both
the server and the database, restart them, check their logs and stop them through the built-in TUI Bob offers.

You can test the server's HTTP REST API by using an HTTP client.
- `GET /api/ping`: Q(^o^)
- `POST /api/items` *(with JSON string as request body)*: creates a new item.
- `GET /api/item/:itemId`: returns the item with the given id.

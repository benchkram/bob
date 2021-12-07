# Server-Database Example

Real world example using Bob for a web server with access to a database.

In this example we guide you through the folowing steps to build and launch a http api-server.

1. Code generation using an [IDL](https://en.wikipedia.org/wiki/Interface_description_language), like openapi or
   protobuf
2. Compile a server binary
3. Run a database in docker
4. Run the server

These steps are already defined in [bob.yaml](./bob.yaml), check Bob's README for explanation.



## Prerequisites
This example requires deepmap/oapi-codegen (as well as Go) to generate the server boilerplate from an openapi.yaml file.

```bash
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.9.0
```



## Walk-through

### 1. Code Generation

To generate the HTTP API server code from an openapi.yaml spec we use [oapi-codegen](https://github.com/deepmap/oapi-codegen). 

Here is the task to do that:
```yaml
tasks:
  generate-api:
    input: openapi.yaml
    cmd: |-
      oapi-codegen -package generated -generate types -o server/rest-api/generated/types.gen.go openapi.yaml
      oapi-codegen -package generated -generate server -o server/rest-api/generated/server.gen.go openapi.yaml
    target: |-
      server/rest-api/generated/types.gen.go
      server/rest-api/generated/server.gen.go
```

To build this task use:

```bash
bob build generate-api
```

This will generate the boilerplate code in [server/rest-api/generated](./server/rest-api/generated).

![bob-build-generate-api](../../example/server-db/assets/bob-build-generate-api.png?raw=true "bob build generate-api")

Bob tracks inputs and performs intelligent caching so that it doesn't have to do unnecessary work. 

If you run `bob build generate-api` again, it will be much faster as it doesn't have to be generated again.

![bob-build-generate-api-cached](../../example/server-db/assets/bob-build-generate-api-cached.png?raw=true "bob build generate-api cached")

### 2. Building the Binary

As code generation is involved we have to make sure the actual build task is run after code-generation is completed.
Notice the *dependson:* entry for the build task:

```yaml
  build:
    input: |-
      ./server
      go.mod
      go.sum
      main.go
    cmd: go build -o ./build/server main.go
    dependson:
      - generate-api
    target: /build/server
```

Bob considers the Go source code as the input to the build command. The checksum is used to verify whether a task needs
to be run again or not. For small repositories an `input: *` would work out of the box, which means "track all files".
Bob also creates a checksum for the target and rebuilds only if it was modified. To ignore files just prepend a `!`, as
in the `.gitignore` syntax.

Now, let's build the binary:

```bash
bob build    # or bob build build
```

This compiles the server binary and saves it as *./build/server*

![bob-build](../../example/server-db/assets/bob-build.png?raw=true "bob build")

#### Rebuilding

If you modify the `openapi.yaml` or `main.go` and run `bob build server` again, the necessary tasks
will be re-run, and not use the cached targets, since their respective inputs have changed.

### 3. Run a Database in Docker

For a regular CRUD app we need a database, to store and modify data. Let's use Redis as defined in [docker-compose.yaml](./docker-compose.yaml).

```yaml
version: '3'
services:
  redis:
    image: redis:latest
    ports:
      - "6379:6379"
    command: ["redis-server", "--appendonly", "yes"]
```

We could use `docker compose up` in a separate shell instance to start the database but that means we have to take care
of shutting it down when it's no longer needed, or if we are switching to other contexts.

As bob can also understand compose files let's define a run-task to start the database:

```yaml
runs:
  database:
    type: compose
    path: docker-compose.yml
```

Now we can use:

```bash
bob run database
```

This will run the database in a container inside docker, and will allow us to control it through Bob's Terminal User
Interface (TUI).

Press `Tab` to switch between the status page and the logs from the database. Restart the database using `Ctrl-R`. Stop
and remove the database container using `Ctrl-C`. 

### 4. Run Server

To start the database before our server binary we can depend on the 'database' run task.

```yaml
runs:
  server:
    type: binary
    path: ./build/server
    dependson:
      - build
      - database
```

We don't just depend on the database but also on the build-task. This assures everything is up-to-date before starting
our environment.

To run the server in the Bob TUI, run:

```bash
bob run server
```

![bob-run-server](../../example/server-db/assets/bob-run-server.gif?raw=true "bob run server")

This assures the server is built (with all its dependencies), ramps up a docker-compose environment with
the Redis database server and will then start the server binary. You can check the outputs of both the server and the
database, restart them (along with rebuilding them if necessary), check their logs and stop them, all through the
built-in TUI.

---

Calling the server api:
* `curl http://localhost:8080/api/ping`
* `curl -X POST http://localhost:8080/api/items`
* `curl http://localhost:8080/api/item/:itemId`

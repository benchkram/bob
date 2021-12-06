Server-Database Example
===
Real world example using Bob for a api-server writing and reading from a database.

In this example we guide you through the folowing steps to build and launch a http api-server.
1. Code generation using a [IDL](https://en.wikipedia.org/wiki/Interface_description_language) like openapi or protobuf
2. Build server binary
3. Run a database in docker
4. Run server binary

These steps are alreeady defined in [bob.yaml](./bob.yaml), follow the Readme for explanations.

## Prerequisites
It requires oapi-codegen to generate the server boilerplate from a openapi.yaml.
As a gopher you could just do.
```bash
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.9.0
```

## Walktrough

### 1. Code Generation

To generate the HTTP API server code from a openapi.yaml spec we use [oapi-codegen](https://github.com/deepmap/oapi-codegen). 

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
Just type 
```bash
$ bob build generate-api
```
to generate the boilerplate code in  [server/rest-api/generated](./server/rest-api/generated).

![bob-build-generate-api](../../example/server-db/assets/bob-build-generate-api.png?raw=true "bob build generate-api")

Bob tracks tasks inputs and does intelligent caching so that it doesn't have to do unnecessary work. 

If you run 

&nbsp;&nbsp;&nbsp;&nbsp;`bob build generate-api`

 again, it will be much faster as it doesn't have to do any work.

![bob-build-generate-api-cached](../../example/server-db/assets/bob-build-generate-api-cached.png?raw=true "bob build generate-api cached")

### 2. Building the Binary
As code generation is involved we have to make sure the actual build task is run after code-generation completed. See the *dependson:* entry for the build task.
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
Take a look at the `input:` directive. Bob takes this list to generates a checksum covering all files. The checksum is used to verify wheather a task needs to be run again. For small repositories a `input: *` would also do the trick, which means "track all files". Bob also creates a checksum for the target and rebuilds only if it was modified. To ignore files just prepend a `!`.

Let's build the binary.
```bash
$ bob build    # or bob build build
```
This creates the server binary in *./build/server*

![bob-build](../../example/server-db/assets/bob-build.png?raw=true "bob build")

#### Rebuild

If you modify the `openapi.yaml` or `main.go` and run `bob build server` again, the necessary tasks
will be re-run, and not use the cache, since their inputs have changed.

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
We could  use `docker compose up` to start the database but that means we have to take care of shutting it down. 

As bob can also understand compose files let's define a run-task to start the database:
```yaml
runs:
  database:
    type: compose
    path: docker-compose.yml
```
That's It.

```bash
bob run database
```
spawns a TUI to control the database.

Hit `tab` to jump between the status page and the logs from the database. Shut it down using `ctrl-c`.
### 4. Run Server
To start the database before our server binary we can depend on the database run-task.
```yaml
runs:
  server:
    type: binary
    path: ./build/server
    dependson:
      - build
      - database
```
You see? We don't just depend on the database but also on the build-task. This assures everything is up to date before starting the binary.

To run the server in the Bob Terminal User Interface (TUI), run:

```bash
$ bob run server
```

![bob-run-server](../../example/server-db/assets/bob-run-server.gif?raw=true "bob run server")

This assures the server is built (with all its dependencies), ramps up a docker-compose environment with
the Redis database running as a service and will then start the server binary. You can check the outputs of both
the server and the database, restart them `ctr+r`, check their logs and stop them through the built-in TUI.


And finally call the server api:
* `curl http://localhost:8080/api/ping`
* `curl -X POST http://localhost:8080/api/items`
* `curl http://localhost:8080/api/item/:itemId`

Use `tab` to switch between the logs of the server and database.
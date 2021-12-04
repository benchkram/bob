Bob
======

*Inspired by Make and Bazel | Made for humans | Written by Gophers.*

<p>
    <a href="https://github.com/Benchkram/bob/releases">
        <img src="https://img.shields.io/github/release/Benchkram/bob.svg" alt="Latest Release">
    </a>
    <a href="https://pkg.go.dev/github.com/Benchkram/bob?tab=doc">
        <img src="https://godoc.org/github.com/golang/gddo?status.svg" alt="GoDoc">
    </a>
    <a href="https://github.com/Benchkram/bob/actions">
        <img src="https://github.com/charmbracelet/bubbletea/workflows/build/badge.svg" alt="Build Status">
    </a>
</p>

Bob is a build system, a task runner as well as a Git extension for multi-repos.

With bob, you can:
- **Build** your programs stupidly-fast, as Bob tracks build inputs and caches compiled outputs, providing fast 
  incremental builds.
- **Run** local development environments, whether they are simple binaries, docker-compose files, or a mix of both.
- **Multi-Repo Git** Easily manage multi-repo setups, with bulk Git operations built-in.

⚠️ Project in early phase.

📖 Documentation in the making.


<!-- ![bob-tui](example/server-db/assets/bob-tui.gif?raw=true "Bob TUI") -->



## Install
```bash
go install github.com/Benchkram/bob
```
For autocompletion add `source <(bob completion)` to your .bashrc.


## Build System
Bobs generates its internal build graph from tasks described in a `bob.yaml` file.    
The basic components of a task are:

* **input**: Whenever an input changes, the task is executed [default: *]
* **cmd**: Commands to be executed
* **target**: File(s) or directories that are created and can be reused in other tasks.

```yaml
tasks:
  build:
    input: ./main.go
    cmd: go build -o app
    target: app
```

`sh` and `bash` commands are entirely possible as it's powered by the incredible shell interpreter [sh](https://github.com/mvdan/sh) from Daniel Martí

Take a look into [server-db](./example/server-db) for an example.


## Local Development
*No cloud | No latency | Just you and your machine* 

Our goal is to create a world-class local development experience by integrating seemlessly with the build-system and enable you to start right after cloning a repository. Let the build-system create binaries and docker images, then execute & control them from one terminal using `bob run`.

Idealy those are the only two commands necessary to get started when joining a new project.
```bash
git clone
bob run
```

### Api-Server Example
Individual steps for web server development are likely similar to this:  

1. Code generation using a [IDL](https://en.wikipedia.org/wiki/Interface_description_language) like openapi or protobuf
2. Build server binary
3. Run a database in docker
4. Run server binary

Those builds and runs can be described in a *bob.yaml* file. Which allows `bob run`  to launch and control a local dev environment. 

See how it works in the [server-db](./example/server-db) example.
<!-- should we exclude the example here? it's already mentioned earlier -->


## Multi-repo Git Operations
<!-- *Monorepo, Multi-repo? Bob got you covered!* -->
Bob enables a natural feeling git workflow for multi-repo setups without relying on submodules.

<!-- &nbsp;&nbsp;&nbsp;&nbsp;❕This works best in combination with trunk-based-development. -->

To do this, Bob uses the concept of an "workspace" to track other git repositories. Inside a workspace you can use the usual every day git-operations through bob with visually enhanced output for the multi-repo case.

<!-- And the fun part is that all four operations are performed recursively on the entirety of the repository tree. -->

Here is an example of `bob git status` in a workspace:

<!-- ![bob-tui](doc/bob-git-status.png?raw=true "bob git status") -->
<img src="doc/bob-git-status.png" width="50%"  />

Highlighted directories indicate other git repository and therfore `bob git add` and `bob git commit` will only operate on the repository a file belongs to allowing to create repository overlapping commits all together.

⚠️ So far only `bob clone` and `bob git status` are implemented.⚠️

Take a look into the [git readme](./README-GIT.md) for the current status of `bob git`
## Setting Up a Workspace 

To set up a new bob workspace you first have to initialise it.

```bash
bob workspace init
```
This creates a *.bob.workspace* file and initialises a new git repository in the current directory.

Add repositorys to the workspace

```bash
bob workspace add git@github.com:Benchkram/bob.git
bob clone # calls git clone for missing repos
```


Clone an existing workspace from a remote git repository

```bash
bob clone git@github.com:Benchkram/bob.git
```

### Credits
We use [bubbletea](https://github.com/charmbracelet/bubbletea) for the TUI


We use a [rewrite of docker-compose in golang](https://github.com/docker/compose). That allows us to execute compose files the same way  `docker compose` does.

We use a integrated shell environment, [sh](https://github.com/mvdan/sh) from Daniel Martí.
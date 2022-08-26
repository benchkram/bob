
<img src="https://bob.build/assets/logo.654a7917.svg" />


<p>
    <a href="https://github.com/benchkram/bob/releases">
        <img src="https://img.shields.io/github/release/benchkram/bob.svg" alt="Latest Release">
    </a>
    <a href="https://pkg.go.dev/github.com/benchkram/bob?tab=doc">
        <img src="https://godoc.org/github.com/golang/gddo?status.svg" alt="GoDoc">
    </a>
    <a href="https://github.com/benchkram/bob/actions">
        <img src="https://github.com/benchkram/bob/actions/workflows/main.yml/badge.svg" alt="Build Status">
    </a>
</p>


Bob is a high-level build system that isolates pipelines from the host system by executing them in a sandboxed shell to run them anywhere - [Nix Inside™](https://nixos.org/).

Why is this useful?

* Get rid of "Works on My Machine".
* No more building in docker.
* Easily jump between different versions of a programming language.


## Vision
Write Once, Build Once, Anywhere.

**Designed to relieve the pain from microservice development.**
## Getting Started

Documentation is available at [bob.build](https://bob.build/docs)

## Install

[docs-installation](https://bob.build/docs/getting-started/installation)

If you wanna go wild and have Go 1.17 or later installed, the short version is:

```bash
git clone https://github.com/benchkram/bob
cd bob
go install
```

For shell autocompletion (bash and zsh supported) add `source <(bob completion)` to your `.bashrc`/`.zshrc`.

### How it works
Bobs generates its internal build graph from tasks described in a `bob.yaml` file (usually refered to as "Bobfile").
Each build step is executed in a sandbox shell only using the given dependecies required from the nix package manager.

The basic components of a build task are:

- **input**: Whenever an input changes, the task's commands need to be re-executed.
- **cmd**: Commands to be executed
- **target**: File(s) or directories that are created, cached locally.
- **dependencies** Dependencies managed by the Nix package manager

Example of a `bob.yaml` file:

```yaml
build:
  build:
    input: "*"
    cmd: |
      go build -o ./app
    target: ./app
    dependencies: [ git, go_1_18 ]
```

Multiline `sh` and `bash` commands are entirely possible, powered by [mvdan/sh](https://github.com/mvdan/sh).



### Comparison
* [Dagger vs. Bob](https://medium.com/benchkram/dagger-vs-bob-2e917cd185d3)
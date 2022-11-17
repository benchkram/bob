
<p align="center">
  <img  width="300" src="https://bob.build/assets/logo.070b920e.svg" />
</p>
<p align="center">
Write Once, Build Once, Anywhere
</p>

---

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


Bob is a high-level build tool for multi-language projects.

Use it to build codebases organized in multiple repositories or in a monorepo.

When to consider using Bob?

* You want a pipeline which runs locally and on CI.
* You want remote caching and never having to do the same build twice.
* You want to get rid of "Works on My Machine".
* You like Bazel and its features but think it's too complex.
* You want a build system which keeps frontend tooling functional.

# Getting Started

[Docs](https://bob.build/docs/) | [Install](https://bob.build/docs/getting-started/installation/)

## Installing From Source 

If you want to go wild, and have Go 1.17 or later installed, the short version is:

```bash
git clone https://github.com/benchkram/bob
cd bob
go install
```

For shell autocompletion (bash and zsh supported) add `source <(bob completion)` to your `.bashrc`/`.zshrc`.



# How it works
Bob generates its internal build graph from tasks described in a `bob.yaml` file (usually referred to as "Bobfile").
Each build step is executed in a sandboxed shell only using the given dependencies required from the nix package manager.

The basic components of a build task are:

- **input**: Whenever an input changes, the task's commands need to be re-executed.
- **cmd**: Commands to be executed
- **target**: Files, directories or docker images created during execution of *cmd*
- **dependencies** Dependencies managed by the Nix package manager

Example of a `bob.yaml` file:

```yaml
build:
  build:
    input: "*"
    cmd: go build -o ./app
    target: ./app
    dependencies: [ git, go_1_18 ]
```

Multiline `sh` and `bash` commands are entirely possible, powered by [mvdan/sh](https://github.com/mvdan/sh).



# Comparisons
* [Dagger vs. bob](https://medium.com/benchkram/dagger-vs-bob-2e917cd185d3)
* [Mage vs. bob](https://medium.com/benchkram/build-system-comparison-mage-vs-bob-aaf4665e3d5c)

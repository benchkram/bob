# build-tool
Polyglot build tool with the power to view distributed repos as a monorepo

### What we want to achieve
* Be fast (caching strategy)
* No internal build environment to avoid definition of external build tools
* Monorepo `view` for distributed repos without git submodules (like https://github.com/mateodelnorte/meta)
* Encourage and Enable to .gitignore generated code


## Desired Functionality
Description of the desired functionality and possible workflows.

### CLI
Git: When there is a need to depend on the cli output of `git` => use go-git    
Git: When the output of `git` can be passed along => use os/Exec(git)
```
bob add            // Add a new child repo (updates .gitignore)
bob rm             // Remove child repo (updates .gitignore)
bob build          // build the `default` target or `all`
bob build `target` // build the specified target
bob build list     // list build targets
bob git ...        // Run git cmd on all child repos
bob git status     // Run git status all child repos (Based on go-git/go-git)
bob git add        // Run git add all child repos (Based on os/Exec(git)?)
bob git commit     // Run git commit all child repos (Based on os/Exec(git)?)
bob git push       // Run git push all child repos (Based on os/Exec(git)?)
bob clone          // Clone top + child repositories
```

### Build
For now only check first level builds.. builds for nested bob repos might be too complex.
Create a directet-acyclic-graph to determine the build order.. and plan for async task execution.

Minimal configuration rquirements for a build task
* Inputs: Inputs to track for changes. Defaults to `*`.
* InputIgnore: Ignore inputs. Defaults to `[.git, .gitignore, etc.]`
* Visibility: There might be internal builds which have special requirements. `public` | `private`. Defaults to `public`
* Depends: Tasks this task depends on
* Exec: Command to execute

Not Sure (Solves some problems but hard to implement/maintain):
* RequiredExecutables: `go:1.16-1.8`, `npm:12.12-14.14` # Needs to know how to get the version for each tool...
* Output: Could be left out for the first version.. Bit it means that top level jobs need to access child repos build structure. I think it's needed to have a good user experience and fail fast. It can probably be implemented to only check at the end of the job that the build succeeded and the output was created without providing more infos to a parent job. That makes it easier to find errors early on.
* OutputType: `binary` | `docker` | `dir` .. Not sure.. would mean we have to provide default workflows for each type.





### Example File Tree
```
/r
/r/.bob
/r/.bob/config
/r/BT_BUILD
/r/.git
/r/.gitignore
/r/...
/r/docker-compose.yaml
/r/tests
/r/...
/r/repo_one
/r/repo_one/BT_BUILD
/r/repo_two
/r/repo_two/BT_BUILD
/r/repo_three
/r/repo_three/BT_BUILD
```

### ThinkTank - Multi Repo Workflows and Problems
Should we add gitsubmodule a like features to track child repo dependencies?
That means keeping old artifacts in a repo after each merge to the main branch.
Should we store them on a "remote-handcrafted-registry" or track it through a repo?
```
1: bob commit # Adds a deptracker to each commit when there are childrepo not on the main branch
1: bob push   # Adds a deptracker to a remote location
```
#### Merge order of child repos
Is this a thing we must consider? Think about your time at BMW using gerrit. I seems that there are cases that you have to fix the commit id. 

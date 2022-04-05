# Projectname

Mechanics of adding a project-name and Bobfile implications.

# Bobfile
A optional project-name can be added to a Bobfile.
```
project: "bob.build/equanox/bob"
build:
  ...
  ...
```
Requirements for a project-name:
* (1) Can identify a remote location for a project e.g. bob.build/equanox/bob
* (2) Can be a plain name without a valid remote location e.g. manhatten-project
* (3) Must consist out of those character [TODO: use the same as go module names?]
* (4) No project name set.

Desired functionality in case of:
* (1) `bob build` downloads artifacts from the remote store identified by the projectname. `bob cache push` uploads artifacts to the remote store. `bob cache sync` download artifacts (based on input hash). Those commands require a valid api-key with the user beeing authorize to read/write to the remote store. In case of a unreachable remote it should continue with a warning (remote cache disabled). On CI it is wise to do a `bob cache sync` (wich should return a failing exit code in case of a not reachable remote) before running a build. `bob cache sync -b taskname` allows to only get artifacts for a specific build.
* (2) Remote capabilities disabled. The projectname is stored in the artifact.
* (3) - 
* (4) Remote capabilities disabled. The local path of the top most Bobfile is used as project-name and is stored in the artifact metadata.

## Multi Level
By default the project-name of the top most Bobfile is used. In case there is no projectname those will apply:
* The remote cache is disabled.
* The local directory of the top most Bobfile is used as project name even if a lower level Bobfile has a valid project-name.

TODO: No matter on which level  `bob ....` is executed it should always behave as it was called from the top most level. (Not yet implemented) 

### Running sub-level Bobfiles
To run a sub-level Bobfile we could use `!` to force build execution in the context of the first found Bobfile  e.g. `bob build !taskname`.

## Artifacts & Artifact Inspection
A project-name allows to use the same artifacts for multiple checkouts on the same system (different local paths). It also simplifies artifact inspection as it becomes possible to query the (local-)store from all locations e.g. `bob inspect artifact manhatten-project`. Though, those capabilities don't add any fundamental value.
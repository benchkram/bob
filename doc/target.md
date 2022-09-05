

## Lifecycle of a target

```mermaid
flowchart TD;
    subgraph Target Initialisation
    Parse --> Create[Create Target]
    end
    subgraph Parallel
    subgraph Extended Initialisation
    Create --> InputHash
    InputHash --> Load[Load Expected Buildinfo]
    Load --> Resolve[Resolve Directories to Files]
    end
    Resolve --> Verify 
    Verify --> Compute[Compute Buildinfos]
    Compute --> Store[Store Buildinfos]
    end
```


## Lifecycle of a task
Parse
Resolve Inputs (Ignore Child Targets)
InputHash
BuildRequired(InputChanged)?
BuildRequired(TargetInvalid)
  Yes
    Clean Targets
    UnpackArtifact
  No
         
    Clean Targets
    Build
        PackArtifacts

               |  Build  | Load | Store |
-------------------------------------------------
InputChanged   |    x    |      |   x   |
TargetChanged  |         |   x  |       |


               |  InputChanged  | TargetChanged |
---------------------------------------------------
InputChanged   |                |     Build     |
TargetChanged  |                |               |



InputHash
BuildinfoExists
LoadTargetFromCache?
    No
        Rebuild

        
    
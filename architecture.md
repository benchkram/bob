# bob architecture
Fine grained infos on bob's features


## Buildinfo & Artifact Store
Bob uses buildinfos and artifacts store to determine if a rebuild is required or if a target could be loaded from the artifact store.

* check if input hash exists in buildinfo store
* check validity of tragets
* check if tragets can be loaded from artifact store
* rebuild
* write buildinfo
* write artifacts

```
@startuml
ditaa(scale=0.9)

                     /--------------------------------\     
                     |cFFF                            |     
                     | `bob build mytask`             |       
                     \----+----------------------+----/     
                          |                      |          
                          |                      |          
                          |                      |          
   /----+-----------------+---------+       /----+---------------------------+
   |cFFF                            |       |cFFF                            |
   | buildinfo store                |       | artifact store                 | 
   \--------------------------------/       \--------------------------------/     
                                          
@enduml
```

#### Buildinfo
A buildinfo item holds general information about the task triggered that build. It also contains a map of input:target combinations for each target in the build chain (child tragets included).
This allows to validate the integraty and existent of all targets.

It can be accessed by the input hash of a task.

#### Artifact
A artifact stores the targets and the path information of a task. 

Artifacts can be accesed by the input hash of a task.
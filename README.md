```
 ██████╗ ██████╗  ██████╗ █████╗ 
██╔═══██╗██╔══██╗██╔════╝██╔══██╗
██║   ██║██████╔╝██║     ███████║
██║   ██║██╔══██╗██║     ██╔══██║
╚██████╔╝██║  ██║╚██████╗██║  ██║
 ╚═════╝ ╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝
```

`orca` is a bash command orchestrator that can be used to run terminal commands in a directed acyclic graph

- [:computer: Installation](#computer-installation)
- [:pencil: DSL for .orca](#pencil-metadata)
- [:rocket: Example](#rocket-example)

**v0.4.4**

### :computer: Installation

Locally: `git clone git@github.com:5amCurfew/orca.git`; `make build`

via Homebrew: `brew tap 5amCurfew/5amCurfew; brew install 5amCurfew/5amCurfew/orca`

```bash
orca is a bash command orchestrator that can be used to run terminal commands in a directed acyclic graph

Usage:
  orca [PATH_TO_DAG_FILE] [flags]

Flags:
  -h, --help      help for orca
  -v, --version   version for orca
```

### :pencil: DSL for .orca

DAGs are defined in `yml` files in the relative path directory.

To define a tasks, use the `tasks` sequence, defining a name, description (`desc`) and `bash` command (`cmd`) per task. Note that `name` must be unique for desired behaviour. The `parentRule` (optional) is one of either `success` (default) that only executes the task if **all parents complete successfully** or `complete` that will execute the task when **all parents have completed (regardless of success or failure)**.

Dependencies are defined using the `dependencies` mapping with `<CHILD>: [<PARENT_1>, <PARENT_2>, ...]` syntax.

### :rocket: Example
```yml
tasks:
  - name: step-1
    desc: start the DAG
    cmd: sleep 1 && echo "DAG started!"
  
  - name: step-2-1
    desc: do something for this task
    cmd: sleep 3 && echo "Step 2.1"
  
  - name: step-2-2
    desc: do something that will fail!
    cmd: sleep 3 && cd into_a_directory_that_does_not_exist

  - name: step-3
    desc: do something for this task that is not skipped if a parent fails
    cmd: sleep 3 && echo "Step 3"
    parentRule: complete
  
  - name: step-4
    desc: do something for this task
    cmd: sleep 2 && echo "Step 4"
  
  - name: step-5
    desc: do something that will fail!
    cmd: sleep 5 && cd into_a_directory_that_does_not_exist

  - name: step-6
    desc: do something for this task
    cmd: sleep 2 && echo "Step 6"
  
  - name: step-7
    desc: do something for this task that will skip if a parent fails
    cmd: sleep 4 && echo "Step 7"
  
  - name: step-8
    desc: do something for this task
    cmd: sleep 1 && echo "Step 8"

dependencies:
  step-2-1: [step-1]
  step-2-2: [step-1]
  step-3: [step-2-2]
  step-4: [step-3]
  step-5: [step-2-1]
  step-6: [step-4, step-5]
  step-7: [step-4]
  step-8: [step-7]
```

Output:

```bash
[🚀 DAG START] executing tasks...

Task                 Status       Pid        Started         Ended          
---------------------------------------------------------------------------
step-1               [✓] Success  15804      18:49:39.8366   18:49:40.8449  
step-2-1             [✓] Success  15807      18:49:40.8452   18:49:43.8539  
step-2-2             [X] Failed   15806      18:49:40.8451   18:49:43.8538  
step-3               [✓] Success  15814      18:49:43.8541   18:49:46.8636  
step-4               [✓] Success  15825      18:49:46.8639   18:49:48.8733  
step-5               [X] Failed   15815      18:49:43.8540   18:49:48.8650  
step-6               [-] Skipped  -          -               18:49:48.8735  
step-7               [✓] Success  15828      18:49:48.8736   18:49:52.8814  
step-8               [✓] Success  15838      18:49:52.8816   18:49:53.8891  

[⚠️  DAG COMPLETE] execution completed with failures
```
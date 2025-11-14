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
- [:pencil: DSL for .orca](#pencil-dsl-for-orca)
- [:rocket: Example](#rocket-example)

**v0.5.0**

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

To define a tasks, use the `nodes` sequence, defining a name, description (`desc`) and `bash` command (`cmd`), `retries` (optional) and `retryDelay` (optional) per task. Note that `name` must be unique for desired behaviour. The `parentRule` (optional) is one of either `success` (default) that only executes the task if **all parents complete successfully** or `complete` that will execute the task when **all parents have completed (regardless of success or failure)**.

Dependencies are defined using the `dependencies` mapping with `<CHILD>: [<PARENT_1>, <PARENT_2>, ...]` syntax.

### :rocket: Example
```yml
nodes:
  - name: step-1
    desc: start the DAG
    cmd: sleep 1 && echo "DAG started!"
  
  - name: step-2-1
    desc: do something for this task
    cmd: sleep 3 && echo "Step 2.1"
  
  - name: step-2-2
    desc: do something that will fail!
    cmd: sleep 3 && cd into_a_directory_that_does_not_exist
    retries: 2
    retryDelay: 5

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

Node                 Status       Pid        Attempt    Started         Ended          
-------------------------------------------------------------------------------------
step-1               [✓] Success  14924      1/1        22:21:33.1958   22:21:34.2003  
step-2-1             [✓] Success  14927      1/1        22:21:34.2025   22:21:37.2097  
step-2-2             [X] Failed   14958      3/3        22:21:50.2217   22:21:53.2292  
step-3               [✓] Success  14966      1/1        22:21:53.2315   22:21:56.2404  
step-4               [✓] Success  14971      1/1        22:21:56.2428   22:21:58.2521  
step-5               [X] Failed   14936      1/1        22:21:37.2117   22:21:42.2194  
step-6               [-] Skipped  -                     -               22:21:58.2524  
step-7               [✓] Success  14976      1/1        22:21:58.2545   22:22:02.2621  
step-8               [✓] Success  14988      1/1        22:22:02.2645   22:22:03.2716  

[⚠️  DAG COMPLETE] execution completed with failures

```
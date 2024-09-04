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

**v0.2.1**

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
  step-5: [step-2-1, step-3]
  step-6: [step-4, step-5]
  step-7: [step-5]
  step-8: [step-7]
```

Output:

```bash
INFO[2024-09-04T23:16:17+01:00] [✔ DAG START] orca execution started         
INFO[2024-09-04T23:16:17+01:00] [START] step-1 task execution started        
INFO[2024-09-04T23:16:18+01:00] [✔ SUCCESS] step-1 task execution successful 
INFO[2024-09-04T23:16:18+01:00] [START] step-2-2 task execution started      
INFO[2024-09-04T23:16:18+01:00] [START] step-2-1 task execution started      
INFO[2024-09-04T23:16:21+01:00] [✔ SUCCESS] step-2-1 task execution successful 
ERRO[2024-09-04T23:16:21+01:00] [X FAILED] task step-2-2 execution failed    
INFO[2024-09-04T23:16:21+01:00] [START] step-3 task execution started        
INFO[2024-09-04T23:16:24+01:00] [✔ SUCCESS] step-3 task execution successful 
INFO[2024-09-04T23:16:24+01:00] [START] step-4 task execution started        
INFO[2024-09-04T23:16:24+01:00] [START] step-5 task execution started        
INFO[2024-09-04T23:16:26+01:00] [✔ SUCCESS] step-4 task execution successful 
ERRO[2024-09-04T23:16:29+01:00] [X FAILED] task step-5 execution failed      
WARN[2024-09-04T23:16:29+01:00] [~ SKIPPED] parent task step-5 failed, skipping step-7 
WARN[2024-09-04T23:16:29+01:00] [~ SKIPPED] parent task step-7 was skipped, skipping step-8 
WARN[2024-09-04T23:16:29+01:00] [~ SKIPPED] parent task step-5 failed, skipping step-6 
WARN[2024-09-04T23:16:29+01:00] [~ DAG COMPLETE] orca.orca execution completed with failures 
```
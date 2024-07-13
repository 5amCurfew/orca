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

**v0.1.7**

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

DAGs are defined in `.orca` files in the relative path directory.

To define a task, use the `task` keyword and provide a name, description and `bash` command.

Use the bit-shift operator `>>` to define dependencies using task names (using a list to define multiple parent tasks) using the syntax `[<PARENT_1>, <PARENT_2>, ...] >> <CHILD>` for each child task. See the example below for more information.

### :rocket: Example
```
task {
    name = step-1
    desc = start the DAG
    cmd  = sleep 1 && echo "DAG started!"
}

task {
    name = step-2-1
    desc = do something for this task
    cmd  = sleep 3 && echo "Step 2.1"
}

task {
    name = step-2-2
    desc = do something that will fail!
    cmd  = sleep 3 && cd into_a_directory_that_does_not_exist
}

task {
    name = step-3
    desc = do something for this task
    cmd  = sleep 3 && echo "Step 3"
}

task {
    name = step-4
    desc = do something for this task
    cmd  = sleep 2 && echo "Step 4"
}

task {
    name = step-5
    desc = do something for this task
    cmd  = sleep 5 && echo "Step 5"
}

task {
    name = step-6
    desc = do something for this task
    cmd  = sleep 2 && echo "Step 6"
}

task {
    name = step-7
    desc = do something for this task
    cmd  = sleep 2 && echo "Step 7"
}

step-1 >> step-2-1
step-1 >> step-2-2
step-2-2 >> step-3
step-3 >> step-4
step-2-1 >> step-5
[step-4, step-5] >> step-6
```

Output:

```bash
INFO[2024-07-13T15:14:13+01:00] [✔ DAG START] example execution started      
INFO[2024-07-13T15:14:13+01:00] [START] step-7 task execution started        
INFO[2024-07-13T15:14:13+01:00] [START] step-1 task execution started        
INFO[2024-07-13T15:14:14+01:00] [✔ SUCCESS] step-1 task execution successful 
INFO[2024-07-13T15:14:14+01:00] [START] step-2-2 task execution started      
INFO[2024-07-13T15:14:14+01:00] [START] step-2-1 task execution started      
INFO[2024-07-13T15:14:15+01:00] [✔ SUCCESS] step-7 task execution successful 
INFO[2024-07-13T15:14:17+01:00] [✔ SUCCESS] step-2-1 task execution successful 
ERRO[2024-07-13T15:14:17+01:00] [X FAILED] task step-2-2 execution failed    
INFO[2024-07-13T15:14:17+01:00] [START] step-5 task execution started        
WARN[2024-07-13T15:14:17+01:00] [~ SKIPPED] parent task step-2-2 failed, aborting step-3 
WARN[2024-07-13T15:14:17+01:00] [~ SKIPPED] parent task step-3 failed, aborting step-4 
WARN[2024-07-13T15:14:17+01:00] [~ SKIPPED] parent task step-4 failed, aborting step-6 
INFO[2024-07-13T15:14:22+01:00] [✔ SUCCESS] step-5 task execution successful 
WARN[2024-07-13T15:14:22+01:00] [~ DAG COMPLETE] example.orca execution completed with failures 
```
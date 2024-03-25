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
- [:wrench: Settings](#nut_and_bolt-using-with-singerio-targets)
- [:pencil: DSL for .orca](#pencil-metadata)
- [:rocket: Example](#rocket-example)
- [:bar_chart: UI](#bar_chart-ui)

**v0.1.0**

### :computer: Installation

Locally: `git clone git@github.com:5amCurfew/orca.git`; `make build`

via Homebrew: `brew tap 5amCurfew/5amCurfew; brew install 5amCurfew/5amCurfew/orca`

### :wrench: Settings

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
    desc = do something for this task
    cmd  = sleep 1 && echo "Step 2.2"
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
    cmd  = sleep 1 && echo "Step 5"
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

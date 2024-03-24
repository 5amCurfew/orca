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

**v0.0.1**

### :computer: Installation

Locally: `git clone git@github.com:5amCurfew/orca.git`; `make build`

via Homebrew: `brew tap 5amCurfew/5amCurfew; brew install 5amCurfew/5amCurfew/orca`

### :wrench: Settings

### :pencil: DSL for .orca

DAGs are defined in `.orca` files in the relative path directory.

To define a task, use the `task` keyword and provide a name, description and `bash` command.

Use the bit-shift operator `>>` to define dependencies using task names (using a list to define multiple parent tasks).

To define a schedule, use the `schedule` keyword and provide a `cron` expression.

### :rocket: Example
```
task {
    name = start
    desc = start the DAG
    cmd  = sleep 1.5 && echo "DAG started!"
}

task {
    name = check-xtkt-version
    desc = checking version
    cmd  = sleep 3 && xtkt --version && xtkt --help
}

task {
    name = extract
    desc = extract commit data from the Github API
    cmd  = cd test && xtkt config_github.json | jq .
}

task {
    name = transform
    desc = transform data
    cmd  = sleep 1.5 && echo "data transformed!"
}

task {
    name = send-another-message
    desc = send
    cmd  = sleep 2 && echo "message sent!"
}

task {
    name = finish
    desc = checkpoint
    cmd  = sleep 2 && echo "DAG finished!"
}

start >> check-xtkt-version
check-xtkt-version >> extract
start >> send-another-message
extract >> transform
[transform, send-another-message] >> finish

```

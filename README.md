```
 ██████╗ ██████╗  ██████╗ █████╗ 
██╔═══██╗██╔══██╗██╔════╝██╔══██╗
██║   ██║██████╔╝██║     ███████║
██║   ██║██╔══██╗██║     ██╔══██║
╚██████╔╝██║  ██║╚██████╗██║  ██║
 ╚═════╝ ╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝
```

`orca` is a bash command orchestration tool.

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

DAGs are defined in `.orca` files in a `dags/` directory (relative to where `orca` is run). `orca` follows a similar syntax to [Airflow](https://airflow.apache.org/docs/apache-airflow/stable/concepts.html), defining tasks, dependencies and a schedule.

To define a task, use the `task` keyword and provide a name, description and `bash` command.

Use the bit-shift operator `>>` to define dependencies using task names (using a list to define multiple parent tasks).

To define a schedule, use the `schedule` keyword and provide a `cron` expression.

### :rocket: Example
```
task {
    name = task1
    desc = Start
    cmd  = sleep 1.5 && echo "Task 1"
}

task {
    name = task2
    desc = Load data
    cmd  = sleep 3 && echo "Task 2"
}

task {
    name = task3
    desc = Transform data
    cmd  = sleep 4 && echo "Task 3"
}

task {
    name = task4
    desc = Send message
    cmd  = sleep 2 && echo "Task 4"
}

task1 >> task2
task1 >> task3
[task1, task2] >> task4

schedule = 0 */10 * * * *
```

### :bar_chart: UI


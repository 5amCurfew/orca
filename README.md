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
    name = some_task
    desc = start
    cmd  = sleep 1.5 && echo "Task 1"
}

task {
    name = create_placeholder_file
    desc = create a location to pipe data
    cmd  = sleep 3 && touch target.json
}

task {
    name = extract
    desc = extract data from Rick & Morty API into placeholder file
    cmd  = xtkt ~/git/xtkt/_config_json/config_rickandmorty.json > target.json
}

task {
    name = some_other_task
    desc = send
    cmd  = sleep 2 && echo "rick & morty data sent!"
}

some_task >> create_placeholder_file
some_task >> extract
extract >> some_other_task

schedule = 0 */10 * * * *
```

### :bar_chart: UI


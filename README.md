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
- [:gear: How it works](#gear-how-it-works)
  - [1. Parse](#1-parse)
  - [2. Relay channels](#2-relay-channels)
  - [3. Concurrent execution](#3-concurrent-execution)
- [:rocket: Example](#rocket-example)

**v0.6.1**

#### TODO

* Max parallel node excution config
* Introduce `queued` state for above
* Conditional edges (dependencies depending on results)
* Visualisation command

### :computer: Installation

Locally: `git clone git@github.com:5amCurfew/orca.git`; `make build`

via Homebrew: `brew tap 5amCurfew/5amCurfew; brew install 5amCurfew/5amCurfew/orca`

```bash
$ orca -h
orca is a bash command orchestrator that can be used to run terminal commands in a directed acyclic graph

Arguments:
  PATH_TO_DAG_FILE   path to the DAG YAML file to execute (default: dag.yml)

Usage:
  orca [PATH_TO_DAG_FILE] [flags]

Flags:
  -h, --help               help for orca
  -p, --max-parallel int   maximum number of tasks to run in parallel (default: unlimited)
  -v, --version            version for orca
```

### :pencil: DSL for .orca

DAGs are defined in `yml` files in the relative path directory.

To define a tasks, use the `nodes` sequence, defining a name, description (`desc`) and `bash` command (`cmd`), `retries` (optional) and `retryDelay` (optional) per task. Note that `name` must be unique for desired behaviour. The `parentRule` (optional) is one of either `success` (default) that only executes the task if **all parents complete successfully** or `complete` that will execute the task when **all parents have completed (regardless of success or failure)**.

Dependencies are defined using the `dependencies` mapping with `<CHILD>: [<PARENT_1>, <PARENT_2>, ...]` syntax.

### :gear: How it works

When `orca` runs a dag file it goes through three phases:

#### 1. Parse

The `.yml` file is read once and used to construct an in-memory `Graph`:

- **Nodes** — each entry under `nodes:` becomes a `Node` struct holding its name, bash command, retry configuration, and `parentRule`. Defaults are applied (`parentRule: success`, `retryDelay: 10s` when retries are set).
- **Edges** — the `dependencies:` map is walked to build two adjacency sets: `Parents` (child → set of parents) and `Children` (parent → set of children). Self-references and cycles are rejected at this stage.
- **`.orca/`** — the log output directory is created (if absent).

#### 2. Relay channels

Before execution begins, one **buffered channel** (capacity 1) is created for every directed edge in the graph and stored in `nodeRelay`, keyed by `"parent->child"`:

```
step-1 ──► step-2-1          nodeRelay["step-1->step-2-1"] = make(chan NodeStatus, 1)
       └──► step-2-2          nodeRelay["step-1->step-2-2"] = make(chan NodeStatus, 1)
                │
                ▼
            step-3             nodeRelay["step-2-2->step-3"] = make(chan NodeStatus, 1)
```

These channels are the only mechanism by which one goroutine signals another. A node goroutine **blocks** on receiving from every one of its parent relay channels before it starts its bash command. Once a node finishes (success, failure, or skip) it sends its final `NodeStatus` into each of its outbound relay channels and closes them.

#### 3. Concurrent execution

Every node is launched as its own goroutine immediately. The concurrency model is:

```
                        ┌─────────────────────────────────────────┐
                        │              Execute()                  │
                        │                                         │
   ┌────────────┐       │  for each node:                         │
   │  dag.yml   │──────►│    go func(nodeKey) {                   │
   └────────────┘       │      waitForParents()   ← blocks here   │
         │              │      node.execute()     ← runs bash cmd │
         │ NewGraph()   │      notifyChildren()   ← sends signals │
         ▼              │    }                                    │
   ┌────────────┐       │                                         │
   │   Graph    │       │  waitGroup.Wait()  ← all nodes done     │
   │  ┌──────┐  │       │  notifyWG.Wait()   ← all signals sent   │
   │  │Nodes │  │       └─────────────────────────────────────────┘
   │  ├──────┤  │
   │  │Edges │  │       Node goroutine lifecycle:
   │  ├──────┤  │
   │  │Relay │  │         Pending ──► Running ──► Success
   │  │Chans │  │                        │
   │  └──────┘  │                        └──────► Failed  (retried up to n times)
   └────────────┘
                                Skipped  (parentRule: success and a parent failed/skipped)
```

Each node's stdout and stderr are written to `.orca/<run-timestamp>/<node-name>_<attempt>.log`. A `StatusChannel` pipes `NodeStatusMsg` events to the Bubble Tea TUI, which renders the live table you see in the terminal.

---

### :rocket: Example
```yml
nodes:

  # ── Stage 1: setup ─────────────────────────────────────────────────────────

  - name: check-env
    desc: verify required tools are available
    cmd: which bash && which echo && echo "environment OK"

  # ── Stage 2: parallel data fetch (one will fail and retry) ─────────────────

  - name: fetch-users
    desc: fetch user data from source A
    cmd: sleep 2 && echo "users fetched"

  - name: fetch-orders
    desc: fetch order data from source B — will fail twice before succeeding
    cmd: |
      if [ ! -f /tmp/orca-orders-attempt ]; then
        echo "1" > /tmp/orca-orders-attempt
        sleep 2
        echo "attempt 1 failed" && exit 1
      elif [ "$(cat /tmp/orca-orders-attempt)" = "1" ]; then
        echo "2" > /tmp/orca-orders-attempt
        echo "attempt 2 failed" && exit 1
      else
        rm /tmp/orca-orders-attempt
        sleep 1 && echo "orders fetched"
      fi
    retries: 2
    retryDelay: 3

  - name: fetch-products
    desc: fetch product catalogue — will fail permanently
    cmd: sleep 1 && echo "catalogue service unavailable" && exit 1

  # ── Stage 3: transform ─────────────────────────────────────────────────────

  - name: transform-users
    desc: clean and normalise user records
    cmd: sleep 2 && echo "users transformed"

  - name: transform-orders
    desc: enrich orders with metadata
    cmd: sleep 3 && echo "orders transformed"

  - name: transform-products
    desc: transform product catalogue — skipped because fetch-products failed
    cmd: sleep 2 && echo "products transformed"

  - name: audit-log
    desc: write audit entry regardless of fetch outcomes (parentRule complete)
    cmd: sleep 1 && echo "audit entry written"
    parentRule: complete

  # ── Stage 4: load ──────────────────────────────────────────────────────────

  - name: load-user-and-order-warehouse
    desc: load user and order data into warehouse
    cmd: sleep 2 && echo "warehouse loaded"

  - name: load-products-warehouse
    desc: load product data — skipped because transform-products was skipped
    cmd: sleep 2 && echo "products loaded"

  # ── Stage 5: notify ────────────────────────────────────────────────────────

  - name: notify-success
    desc: send success notification — skipped because load-products was skipped
    cmd: sleep 1 && echo "pipeline complete — notification sent"

dependencies:
  fetch-users:                    [check-env]
  fetch-orders:                   [check-env]
  fetch-products:                 [check-env]
  transform-users:                [fetch-users]
  transform-orders:               [fetch-orders]
  transform-products:             [fetch-products]
  audit-log:                      [fetch-users, fetch-orders, fetch-products]
  load-user-and-order-warehouse:  [transform-users, transform-orders]
  load-products-warehouse:        [transform-products]
  notify-success:                 [load-user-and-order-warehouse, load-products-warehouse]
```

Output:

```bash
[🚀 DAG START] executing tasks...

Node                          Status       Pid        Attempt    Started         Ended
------------------------------------------------------------------------------------------------
check-env                     [✓] Success  30306      1/1        21:53:18.9032   21:53:18.9086
fetch-users                   [✓] Success  30310      1/1        21:53:18.9104   21:53:20.9196
fetch-orders                  [✓] Success  30360      3/3        21:53:26.9325   21:53:27.9485
fetch-products                [X] Failed   30309      1/1        21:53:18.9099   21:53:19.9156
transform-users               [✓] Success  30325      1/1        21:53:20.9212   21:53:22.9264
transform-orders              [✓] Success  30376      1/1        21:53:27.9515   21:53:30.9595
transform-products            [-] Skipped  -                     -               21:53:19.9159
audit-log                     [✓] Success  30375      1/1        21:53:27.9507   21:53:28.9568
load-user-and-order-warehouse [✓] Success  30389      1/1        21:53:30.9615   21:53:32.9706
load-products-warehouse       [-] Skipped  -                     -               21:53:19.9161
notify-success                [-] Skipped  -                     -               21:53:32.9709

[⚠️  DAG COMPLETE] execution completed with failures
```
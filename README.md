# drk
[wrk](https://github.com/wg/wrk) but for databases and pronounced [/dɜːk/](https://dictionary.cambridge.org/pronunciation/english/dirk).

### Contents

* [Installation](#installation)
* [Usage](#usage)
* [Supported Databases](#supported-databases)
* [Configuration](#configuration)
	* [VUs](#vus)
	* [Workflows](#workflows)
	* [Activities](#activities)
	* [Queries](#queries)
	* [Args](#args)
* [Running the binary](#running-the-binary)
* [Running with Docker](#running-with-docker)
* [Deploying workloads via Docker](#deploying-workloads-via-docker)
* [Metrics](#metrics)
* [Todos](#todos)

### Installation

Find the release that matches your architecture on the [releases](https://github.com/codingconcepts/drk/releases) page.

Download the tar, extract the executable, and move it into your PATH. For example:

```sh
tar -xvf drk_0.0.1_macos_arm64.tar.gz
```

### Usage

drk's main 4 settings (config, driver, duration, and url) can be configured with arguments or from the environment.

| Setting             | Argument              | Envrironment        | Description                         |
| --------            | ----------            | ------------        | -----------                         |
| Config              | --config              | CONFIG              | Path to config file                 |
| URL                 | --url                 | URL                 | Database URL                        |
| Driver              | --driver              | DRIVER              | Type of database driver to use      |
| Duration            | --duration            | DURATION            | Duration of test                    |
| Retries             | --retries             | RETRIES             | Retries per request                 |
| Query Timeout       | --query-timeout       | QUERY_TIMEOUT       | Request timeout                     |
| Debug               | --debug               | DEBUG               | Toggle debug-level logging          |
| Sensitive           | --sensitive           | SENSITIVE           | Toggle sensitive env var logging    |
| Average Window Size | --average-window-size | AVERAGE_WINDOW_SIZE | Change latency average window size  |
| NoColor             | --no-color            | NO_COLOR            | Remove console color formatting     |
| Connection Lifetime | --connection-lifetime | CONNECTION_LIFETIME | Duration a connection can be reused |
```
drk --help

Usage of drk:
  -average-window-size int
        number of request to derive an average latency for (default 1000)
  -clear
        clear the terminal before printing metrics
  -config string
        absolute or relative path to config file (default "drk.yaml")
  -connection-lifetime duration
        amount of time a connection can be reused (default 1m0s)
  -debug
        show debugging logs
  -driver string
        database driver to use [mysql, spanner, pgx] (default "pgx")
  -dry-run
        if specified, prints config and exits
  -duration duration
        total duration of simulation (default 10m0s)
  -no-color
        print logs without color
  -output string
        type of metrics output to print [log, table] (default "log")
  -query-timeout duration
        timeout for database queries (default 5s)
  -retries int
        number of request retries (default 1)
  -sensitive
        show sensitive logs
  -url string
        database connection string
  -version
        display the application version
```

### Supported databases

* AWS DSQL ([example](examples/dsql))
* CockroachDB ([example](examples/ecommerce))
* Google Cloud Spanner
  * Google Standard SQL ([example](examples/spanner/google_standard_sql))
  * PostgreSQL ([example](examples/spanner/postgresql))
* Postgres ([example](examples/postgres))
* MySQL ([example](examples/mysql))
* Oracle ([example](examples/oracle))

### Configuration

drk workloads are configured by way of a YAML file that's provided with the `--config` argument or `CONFIG` environment variable.

The following concepts are important to understand if you're to get the best out of drk.

##### VUs

A VU (or "Virtual User" is simply a thread that executes a given workflow).

##### Workflows

A workflow defines a series of behaviours representing an archetype/persona (and executed under a single VU). If you wish to simulate load against an eCommerce database, you might choose to simulate 100 casual customers and 50 return customers; each can be expressed as a workflow as follows:

```yaml
workflows:
  casual_customer:
    vus: 100
    setup_queries:
      - create_shopper
      - fetch_product_names
    queries:
      - name: browse_product
        rate: 2/1s
      - name: add_to_basket
        rate: 1/5s
      - name: checkout
        rate: 1/30s
      - name: check_order
        rate: 1/10s

  return_customer:
    vus: 50
    setup_queries:
      - create_shopper
      - fetch_product_names
    queries:
      - name: browse_product
        rate: 10/1s
      - name: add_to_basket
        rate: 1/3s
      - name: checkout
        rate: 1/10s
      - name: check_order
        rate: 1/5s
```

The setup queries (defined under `setup_queries`) define the initialization behaviour of the workflow and may involve activities such as the creation of a shopper and the fetching of reference data. These are executed once and in the order specified when the VU starts.

Regular queries (defined under `queries`) define the runtime behaviour of the workflow and are executed at a given rate, meaning their execution order is non-deterministic.

##### Activities

An activity is simply a query that is executed at a given rate. The rate is expressed as a number and Go `time.Duration` pair (e.g. `10/1s` means "run this query 10 times every second" while `1/10s` means "run this query once every 10 seconds").

Activities are referenced in the workflow by name but are created in the `activities` section of the drk config file. There are 2 main types of query:

* `exec` - Executes a query and does not return any data. These queries are suited to write operations, where the outcome of the query does not need to be persisted in the VU state.

* `query` - Executes a query and remembers the data returned. These queries are suited to read operations and write operations where the outcome of the write needs to be remembered for other queries in the workflow (e.g. the creation of a new row that yields an identifier to reference later).

##### Queries

A query is simply a SQL statement that can optionally accept arguments (see [Args](#args)) and is expressed in an activity as a string. For example, the following query inserts a new shopper into the shopper table and returns their id. This id can later be referenced by a combination of the activity name (in this case "create_shopper") and the field returned (in this case "id"):

```yaml
activities:
  create_shopper:
    type: query
    args:
      - type: gen
        value: email
    query: |-
      INSERT INTO shopper (email)
      VALUES ($1)
      RETURNING id
```

##### Global Args

At the top-level of a drk config file, you can optionally express global arguments that are parsed once during initialization and can be reused throughout the test run as "global" types:

```yaml
args:
  region:
    type: env
    name: REGION

# workflows omitted for brevity

activities:

  show_args:
    args:
      - type: global
        name: region
    ...
```

##### Args

If provided, arguments to a query are passed in the order they are expressed in the config file.

The following argument types are supported:

* `gen` - These arguments are generated once per query execution and provide random fake data to the query. See [gen.go](pkg/random/gen.go) for a complete list of fake data available.

For example, the following argument will generate a credit card number:

```yaml
- type: gen
  value: credit_card_number
```

* `ref` - These arguments make use of previously generated data (for instance, the id of an inserted row, or the name of a fetched product etc.).

For example, the following argument will provide the id of a previously created shopper:

```yaml
- type: ref
  query: create_shopper
  column: id
```

* `set` - These arguments provide a random value from a set of available values.

For example, the following argument will select between the values "admin", "regular", or "read_only" for the purposes of inserting a user type; with each option equally likely:

```yaml
- type: set
  values: [admin, regular, read_only]
```

To give the options different likelihoods of being selected, optional weights can be provided. The following example selects from the same set of options but selects regular users more frequently than either admin or read_only:

```yaml
- type: set
  values: [admin, regular, read_only]
  weights: [10, 70, 20]
```

Note that the weights don't have to sum to 100.

* `const` - If you need to parse a specific value for every execution of a query, use this generator.

The following example will provide the value `42` for every query execution:

```yaml
- type: const
  value: 42
```

* `env` - These arguments source a value from the runtime environment based on its name, prevening you from having to hardcode values in your configuration file.

The following example will provide the value for the "REGION" environment variable for every query execution:

```yaml
- type: env
  name: REGION
```

* `expr` - These arguments make use of the popular [expr](https://github.com/expr-lang/expr) Go package and allow for more complex logic to be used when generating argument values.

The following example will provide a different value, depending on the "REGION" environment variable:

```yaml
- type: expr
  value: |-
    env("REGION") == "iad" ? "us-east-1"
    : env("REGION") == "fra" ? "eu-central-1"
    : env("REGION") == "sgp" ? "ap-southeast-1"
    : "invalid"
```

* The last family of argument generators are the range generators, which generate a value of a given type between a minimum and a maximum value.

The following examples demonstrate the generators available and how to use them:

``` yaml
- type: int
  min: 1
  max: 10

- type: float
  min: 1.0
  max: 10.0

- type: timestamp
  min: 2024-01-01
  max: 2024-12-31
  fmt: 2006-01-02

- type: timestamp
  min: 2024-01-01T00:00:00
  max: 2024-01-01T23:59:59

- type: interval
  min: 1m
  max: 10m

- type: duration
  min: 1m
  max: 10m

- type: location
  lat: 51.538970
  lon: -0.141689
  distance_km: 100.0

- type: point
  lat: 51.538970
  lon: -0.141689
  distance_km: 100.0
```

### Running the binary

For more examples see [examples](examples/) but here's the gist:

```sh
# CockroachDB / Postgres / Spanner (Postgres)
drk \
--driver pgx \
--url "postgres://root@localhost:26257?sslmode=disable"
--config examples/postgres/drk.yaml \

# AWS DSQL
export PGPASSWORD="The value of your DSQL cluster's PGPASSWORD"

drk \
--driver pgx \
--url "postgres://YOUR_ENDPOINT.dsql.us-east-1.on.aws:5432/postgres?user=admin&sslmode=verify-full&sslrootcert=AmazonRootCA1.pem"
--config examples/dsql/drk.yaml \

# MySQL
drk \
--driver mysql
--url "root:password@tcp(localhost:3306)/mysql"
--config examples/mysql/drk.yaml \

# Oracle
drk \
--driver oracle
--url "oracle://system:password@localhost:1521/defaultdb" \
--config examples/oracle/drk.yaml \

# Spanner (Google Standard SQL)
drk \
--driver spanner \
--url projects/PROJECT/instances/INSTANCE/databases/DATABASE \
--config examples/spanner/google_standard_sql/drk.yaml
```

### Running with Docker

Run Docker, mounting a local volume containing your workload file.

```sh
docker run --rm -it \
-v ${PWD}/examples/docker_run:/docker_run \
codingconcepts/drk:v0.1.0 \
--driver pgx \
--url "postgres://root@host.docker.internal:26257?sslmode=disable" \
--config docker_run/workload.yaml \
--pretty
```

### Deploying workloads via Docker

Build a Docker image containing your workload files (suitable for deployments of drk to remote runtime locations where you won't have access to workload files).

```sh
(
  cd examples/docker_deploy && \
	docker build \
		--build-arg workload_dir=. \
		-t codingconcepts/drkd \
		.
)

# Via arguments.
docker run --rm -it \
codingconcepts/drkd \
--driver pgx \
--url "postgres://root@host.docker.internal:26257?sslmode=disable" \
--config workloads/workload.yaml

# Via environment.
docker run --rm -it \
-e DRIVER=pgx \
-e URL="postgres://root@host.docker.internal:26257?sslmode=disable" \
-e CONFIG=workloads/workload.yaml \
codingconcepts/drkd
```

### Metrics

drk exports Prometheus metrics on :2112/metrics and publishes a single histogram metric, grouped by workflow and query:

* drk_request_duration_bucket
* drk_request_duration_count
* drk_request_duration_sum

To show the requests per second by workflow and query, try the following PromQL expression:

```
rate(drk_request_duration_count[1m])
```

To show the request latencies by workflow and query, try the following PromQL expression:

```
histogram_quantile(0.99, sum by (le, workflow, query) (rate(drk_request_duration_bucket[1m])))
```

### Todos

* Calculate average latency ring size based on VU count and requests/s
* Post-run summary
* Global arguments
* Array support
* Cohorts (run these, then these)
* Commit and Rollback counts
* Support bulk activities (e.g. insert 1,000 instead of just 1)
* Add the ability to ensure uniqueness across two arg values (re-running until unique, or crashing after X attempts)
* Update ref to allow more than one item to be seleted (e.g. add multiple products to a basket)
* Optionally pass args in workflow queries
* Ability to ramp VU's down

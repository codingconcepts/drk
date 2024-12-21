# drk
[wrk](https://github.com/wg/wrk) but for databases and pronounced [/dɜːk/](https://dictionary.cambridge.org/pronunciation/english/dirk).

### Installation

Find the release that matches your architecture on the [releases](https://github.com/codingconcepts/drk/releases) page.

Download the tar, extract the executable, and move it into your PATH. For example:

```sh
tar -xvf drk_0.0.1_macos_arm64.tar.gz
```

### Usage

drk's main 4 settings (config, driver, duration, and url) can be configured with arguments or from the environment.

| Setting  | Argument   | Envrironment |
| -------- | ---------- | ------------ |
| Config   | --config   | CONFIG       |
| Driver   | --driver   | DRIVER       |
| Duration | --duration | DURATION     |
| URL      | --url      | URL          |

```
drk --help

Usage of drk:
  -config string
        absolute or relative path to config file (default "drk.yaml")
  -debug
        enable verbose logging
  -driver string
        database driver to use [pgx, mysql, dsql] (default "pgx")
  -dry-run
        if specified, prints config and exits
  -duration duration
        total duration of simulation (default 10m0s)
  -url string
        database connection string
  -version
        display the application version
```

### Running the binary

For more examples see [examples](examples/) but here's the gist:

```sh
# CockroachDB / Postgres
drk \
--driver pgx \
--url "postgres://root@localhost:26257?sslmode=disable"
--config examples/db_comparison/postgres.drk.yaml \

# DSQL
AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
AWS_REGION=${AWS_REGION} \
drk \
--driver dsql \
--url "postgres://YOUR_ENDPOINT.dsql.us-east-1.on.aws:5432/postgres?user=admin&sslmode=verify-full&sslrootcert=AmazonRootCA1.pem"
--config examples/db_comparison/postgres.drk.yaml \

# MySQL
drk \
--driver mysql
--url "root:password@tcp(localhost:3306)/mysql"
--config examples/db_comparison/mysql.drk.yaml \
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

Build a Docker image containing your workload files (suitable for deployments of drk to remote runtime locations
where you won't have access to workload files).

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

### Todos

* Fix: No logging or printing in the event that first queries are all errors
* Cohorts (run these, then these)
* Commit and Rollover counts
* Support bulk activities (e.g. insert 1,000 instead of just 1)
* Add the ability to ensure uniqueness across two arg values (re-running until unique, or crashing after X attempts)
* Update ref to allow more than one item to be seleted (e.g. add multiple products to a basket)
* Optionally pass args in workflow queries
* Ramp VU's up and down

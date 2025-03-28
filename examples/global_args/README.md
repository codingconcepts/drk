### Setup

Cluster

```sh
cockroach demo \
--no-example-database \
--insecure
```

### Run

Workload

```sh
FLY_REGION="fra" \
go run drk.go \
--config examples/global_args/drk.yaml \
--url "postgres://root@localhost:26257/store?sslmode=disable" \
--driver pgx \
--debug
```
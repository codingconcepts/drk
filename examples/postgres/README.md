### Setup

Create databases

```sh
docker run -d \
  --name=cockroach \
  -p 26257:26257 \
  cockroachdb/cockroach:v24.2.0 start-single-node \
    --insecure
```

Create and populate database objects

```sh
cockroach sql --insecure -f examples/db_comparison/postgres.create.sql
```

Run drk

```sh
# CockroachDB
drk \
--config examples/db_comparison/postgres.drk.yaml \
--url "postgres://root@localhost:26257?sslmode=disable"
```
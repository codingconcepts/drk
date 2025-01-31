### Setup

Create databases

```sh
docker run -d \
--name=cockroach \
-p 26257:26257 \
cockroachdb/cockroach:v24.2.0 start-single-node \
--insecure
```

Create objects

```sh
cockroach sql --insecure -f examples/batch/create.sql
```

Run drk

```sh
go run drk.go \
--config examples/batch/drk.yaml \
--url "postgres://root@localhost:26257?sslmode=disable"
```
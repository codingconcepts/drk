### Setup

Create databases

```sh
docker run -d \
--name=cockroach \
-p 26257:26257 \
-p 8080:8080 \
cockroachdb/cockroach:v24.2.0 start-single-node --insecure
```

Create and populate database objects

```sh
cockroach sql --insecure -f examples/error_handling/create.sql
```

Run drk

```sh
go run drk.go \
--config examples/error_handling/drk.yaml \
--url "postgres://root@localhost:26257?sslmode=disable" \
--debug
```
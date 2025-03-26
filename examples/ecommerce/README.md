### Setup

Create databases

```sh
docker run -d \
--name cockroach \
-p 26257:26257 \
cockroachdb/cockroach:v24.3.3 start-single-node --insecure
```

Create and populate database objects

```sh
cockroach sql --insecure -f examples/ecommerce/create.sql
```

Run drk

```sh
go run drk.go \
--config examples/ecommerce/drk.yaml \
--url "postgres://root@localhost:26257?sslmode=disable" \
--duration 1m \
--output table \
--clear
```
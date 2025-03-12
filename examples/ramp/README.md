### Setup

Create databases

```sh
docker run -d \
--name cockroach \
-p 26257:26257 \
cockroachdb/cockroach:v24.3.3 start-single-node --insecure
```

Run drk

```sh
go run drk.go \
--config examples/ramp/drk.yaml \
--url "postgres://root@localhost:26257?sslmode=disable" \
--pretty
```
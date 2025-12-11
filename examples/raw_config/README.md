### Setup

Create databases

```sh
docker run -d \
--name cockroach \
-p 26257:26257 \
cockroachdb/cockroach:v25.3.4 start-single-node --insecure
```

Create and populate database objects

```sh
cockroach sql --insecure -f examples/raw_config/create.sql
```

Capture base64 config

```sh
cat examples/raw_config/drk.yaml | base64 | pbcopy
```

Run drk

```sh
go run drk.go \
--raw-config "d29ya2Zsb3dzOgoKICBkZXZpY2U6CiAgICB2dXM6IDEwCiAgICBxdWVyaWVzOgogICAgICAtIG5hbWU6IHJlY29yZF9tZWFzdXJlbWVudAogICAgICAgIHJhdGU6IDEwLzFzCgphY3Rpdml0aWVzOgoKICByZWNvcmRfbWVhc3VyZW1lbnQ6CiAgICB0eXBlOiBleGVjCiAgICBhcmdzOgogICAgICAtIHR5cGU6IGZsb2F0CiAgICAgICAgbWluOiAxLjAKICAgICAgICBtYXg6IDEwLjAKICAgIHF1ZXJ5OiB8LQogICAgICBJTlNFUlQgSU5UTyBtZWFzdXJlbWVudCAodmFsdWUpCiAgICAgIFZBTFVFUyAoJDEp" \
--url "postgres://root@localhost:26257?sslmode=disable" \
--duration 1m \
--output table \
--clear \
--average-window-size 1000
```
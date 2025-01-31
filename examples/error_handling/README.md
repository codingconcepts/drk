### Setup

Create databases

```sh
docker run -d \
--name=cockroach \
-p 26257:26257 \
-p 8080:8080 \
cockroachdb/cockroach:v24.2.0 start-single-node \
  --insecure
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

### Debugging

View counts

```sh
cockroach sql --insecure -e "SELECT 'shopper', COUNT(*) FROM shopper UNION ALL SELECT 'product', COUNT(*) FROM product UNION ALL SELECT 'basket', COUNT(*) FROM basket;"
```

### Teardown

Data

```sh
cockroach sql --insecure -e "TRUNCATE basket; TRUNCATE shopper CASCADE; TRUNCATE product CASCADE;"
```
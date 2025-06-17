### Setup

Create databases

```sh
docker run -d \
--name cockroach \
-p 26257:26257 \
cockroachdb/cockroach:v25.1.5 start-single-node --insecure
```

Create and populate database objects

```sh
cockroach sql --insecure -f examples/points/create.sql
```

Run drk

```sh
go run drk.go \
--config examples/points/drk.yaml \
--url "postgres://root@localhost:26257?sslmode=disable" \
--duration 1m \
--average-window-size 1000
```

See locations coming in

```sql
SELECT
  s.email,
  s.gender,
  s.date_of_birth,
  ST_ASTEXT(s.location)
FROM shopper s
LIMIT 10;
```

See max distances from central point

```sql
SELECT
  s.email,
  s.gender,
  s.date_of_birth,
  ST_DISTANCE('POINT(-0.141689 51.538970)', s.location) AS distance
FROM shopper s
ORDER BY ST_DISTANCE('POINT(-0.141689 51.538970)', s.location) DESC
LIMIT 10;
```
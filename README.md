# drk
[wrk](https://github.com/wg/wrk) but for databases

Cluster

```sh
cockroach demo --insecure --no-example-database
```

Objects

```sh
cockroach sql \
--host localhost \
--insecure \
-f examples/ecommerce/create.sql
```

Seed data

```sh
dgs gen data \
--config examples/ecommerce/dgs.yaml \
--url "postgres://root@localhost:26257?sslmode=disable" \
--workers 1

```

Dry run

```sh
go run main.go \
--config "examples/ecommerce/drk.yaml" \
--url "postgres://root@localhost:26257?sslmode=disable" \
--dry-run
```

Run

```sh
go run main.go \
--config "examples/ecommerce/drk.yaml" \
--url "postgres://root@localhost:26257?sslmode=disable" \
--debug
```

Reset data

```sql
TRUNCATE basket CASCADE;
TRUNCATE shopper CASCADE;
TRUNCATE product CASCADE;
```

### Todos

* Give activites dependencies (e.g. add_to_basket can't run until browse_product has run)
* Fix exec; I don't think it's working
* Update ref to allow more than one item to be seleted (e.g. add multiple products to a basket)
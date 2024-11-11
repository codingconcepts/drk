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
go run drk.go \
--config "examples/ecommerce/drk.yaml" \
--url "postgres://root@localhost:26257?sslmode=disable" \
--dry-run
```

Run

```sh
go run drk.go \
--config "examples/ecommerce/drk.yaml" \
--url "postgres://root@localhost:26257?sslmode=disable"

# Send output to file
go run drk.go \
--config "examples/ecommerce/drk.yaml" \
--url "postgres://root@localhost:26257?sslmode=disable" > &> log.txt
```

Reset data

```sql
TRUNCATE basket CASCADE;
TRUNCATE shopper CASCADE;
TRUNCATE product CASCADE;

-- DON'T FORGET TO RESEED
```

Show purchase

```sql
SELECT
  p.total,
  pi.product_id,
  pi.quantity
FROM purchase p
JOIN purchase_item pi ON p.id = pi.purchase_id;
```

### Todos

* Give activites dependencies (e.g. add_to_basket can't run until browse_product has run)
* Fix exec; I don't think it's working
* Update ref to allow more than one item to be seleted (e.g. add multiple products to a basket)
* TEST!
* Optionally pass args in workflow queries
* Ramp VU's up and down
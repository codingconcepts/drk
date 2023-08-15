# drk
[wrk](https://github.com/wg/wrk) but for databases

### Running an example

Start a local CockroachDB cluster and hop onto it

``` sh
cockroach demo --insecure --no-example-database --max-sql-memory 1GiB
```

Create the database tables

``` sql
CREATE TABLE "product" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "name" STRING NOT NULL,
  "market" STRING NOT NULL,
  "amount" DECIMAL NOT NULL,
  "currency" STRING NOT NULL,

  INDEX ("name", "market") STORING ("amount", "currency")
);
```

Run the test

``` sh
go run drk.go \
  -c examples/drk.yaml \
  -u "postgres://root@localhost:26257/defaultdb?sslmode=disable"
```

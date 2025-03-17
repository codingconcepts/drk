### Setup

Cluster

```sh
cockroach demo \
--no-example-database \
--nodes 9 \
--demo-locality=region=us-east-1,az=a:region=us-east-1,az=b:region=us-east-1,az=c:region=eu-central-1,az=a:region=eu-central-1,az=b:region=eu-central-1,az=c:region=ap-southeast-1,az=a:region=ap-southeast-1,az=b:region=ap-southeast-1,az=c \
--insecure
```

Objects

```sql
CREATE DATABASE store
  PRIMARY REGION 'us-east-1'
  REGIONS 'eu-central-1', 'ap-southeast-1';

USE store;

CREATE TABLE product (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "name" STRING NOT NULL,
  "price" DECIMAL NOT NULL
) LOCALITY REGIONAL BY ROW;
```

Data

```sql
INSERT INTO product ("crdb_region", "name", "price") VALUES
  ('us-east-1', 'a', 1.99),
  ('us-east-1', 'b', 100.99),
  ('eu-central-1', 'c', 2.99),
  ('eu-central-1', 'd', 200.99),
  ('ap-southeast-1', 'e', 3.99),
  ('ap-southeast-1', 'f', 300.99);
```

### Run

Workload

```sh
FLY_REGION="sgp" \
go run drk.go \
--config examples/expressions/drk.yaml \
--url "postgres://root@localhost:26257/store?sslmode=disable" \
--driver pgx
```
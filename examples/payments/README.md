### Setup

Create database

```sh
docker compose -f examples/payments/compose.yaml up -d
docker exec -it node1 cockroach init --insecure
```

Create and populate database objects

```sh
cockroach sql --insecure -f examples/payments/create.sql
```

Test database objects

```sql
INSERT INTO customer (id, email) VALUES
  ('a55fa2ab-26e3-4399-83b7-728d43a1092a', 'a@example.com'),
  ('b53cdd85-872b-4d69-a1cf-1978e5f4f318', 'b@example.com');

INSERT INTO account (id, customer_id, balance) VALUES
  ('15873f45-b1bb-4cdc-86b7-06066d78ac8c', 'a55fa2ab-26e3-4399-83b7-728d43a1092a', 10000),
  ('2218f55c-efe8-41b7-8fd9-6c5f82905315', 'b53cdd85-872b-4d69-a1cf-1978e5f4f318', 10000);

CALL make_transfer('15873f45-b1bb-4cdc-86b7-06066d78ac8c', '2218f55c-efe8-41b7-8fd9-6c5f82905315', 10);
SELECT * FROM account;
```

Run drk

```sh
go run drk.go \
--config examples/payments/drk.yaml \
--url "postgres://root@localhost:26257?sslmode=disable" \
--duration 1m \
--output table \
--clear \
--average-window-size 1000
```

### Teardown

Teardown database

```sh
docker compose -f examples/payments/compose.yaml down
```
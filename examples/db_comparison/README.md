### Setup

Create databases

```sh
docker run -d \
  --name mysql \
  -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=password \
    mysql:8.2.0 \
      --server-id=1 \
      --log-bin=mysql-bin \
      --binlog-format=ROW \
      --gtid-mode=ON \
      --enforce-gtid-consistency \
      --log-slave-updates

docker run -d \
  --name=cockroach \
  -p 26257:26257 \
  cockroachdb/cockroach:v24.2.0 start-single-node \
    --insecure
```

Create and populate database objects

```sh
mysql -h localhost -u root -p mysql --protocol=tcp < examples/db_comparison/mysql.create.sql

cockroach sql --insecure -f examples/db_comparison/postgres.create.sql
```

Run drk

```sh
go run drk.go \
--config examples/db_comparison/mysql.drk.yaml \
--url "root:password@tcp(localhost:3306)/mysql" \
--driver mysql

go run drk.go \
--config examples/db_comparison/postgres.drk.yaml \
--url "postgres://root@localhost:26257?sslmode=disable"
```
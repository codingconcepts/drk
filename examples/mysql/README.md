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
```

Create and populate database objects

```sh
mysql -h localhost -u root -p mysql --protocol=tcp < examples/mysql/create.sql
```

Run drk

```sh
drk \
--config examples/mysql/drk.yaml \
--url "root:password@tcp(localhost:3306)/mysql" \
--driver mysql
```
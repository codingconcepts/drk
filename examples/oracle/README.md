Setup Oracle

```sh
docker run \
-d \
--name oracle \
-p 1521:1521 \
-e ORACLE_PDB=defaultdb \
-e ORACLE_PWD=password \
container-registry.oracle.com/database/free:23.6.0.0-lite
```

Connect

```sh
docker exec -it oracle sqlplus system/password@//localhost:1521/defaultdb
```

Run commands in cat examples/oracle/create.sql

Run drk

```sh
drk \
--config examples/oracle/drk.yaml \
--url "oracle://system:password@localhost:1521/defaultdb" \
--driver oracle \
--pretty
```
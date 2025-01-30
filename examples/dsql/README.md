### Setup

Create DSQL cluster

```sh
response=$(aws dsql create-cluster \
--region us-east-1 \
--tags Name=rob-sandbox-drk \
--no-deletion-protection-enabled)

export ENDPOINT=$(echo "$response" | jq -r .identifier)

aws dsql get-cluster \
--region us-east-1 \
--identifier ${ENDPOINT}

export PGPASSWORD=$(aws dsql generate-db-connect-admin-auth-token \
--region us-east-1 \
--expires-in 3600 \
--hostname ${ENDPOINT}.dsql.us-east-1.on.aws)

export DSQL_URL="postgres://${ENDPOINT}.dsql.us-east-1.on.aws:5432/postgres?user=admin&sslmode=verify-full&sslrootcert=AmazonRootCA1.pem"
```

Create objects

```sh
psql ${DSQL_URL} -f examples/dsql/create.sql
```

Run drk

```sh
drk \
--config examples/dsql/drk.yaml \
--driver pgx \
--url ${DSQL_URL}
```

Teardown

```sh
aws dsql delete-cluster \
--region us-east-1 \
--identifier ${ENDPOINT}
```
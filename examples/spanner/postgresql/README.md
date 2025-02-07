Generate a spanner.key file for a service account with IAM permissions to the Spanner database.

```sh
gcloud iam service-accounts keys create spanner.key --iam-account SERVICE_ACCOUNT_EMAIL
```

Start the PGAdapter.

```sh
docker run -d \
--name pgadapter \
-p 5432:5432 \
-v ${PWD}/spanner.key:/spanner.key:ro \
-e GOOGLE_APPLICATION_CREDENTIALS=/spanner.key \
gcr.io/cloud-spanner-pg-adapter/pgadapter -p PROJECT -i INSTANCE \
-x
```

Run drk (not that the uid:pwd are not used by Spanner; these can be set to anything).

```sh
go run drk.go \
--driver pgx \
--url "postgres://uid:pwd@localhost:5432/DATABASE?sslmode=disable" \
--config examples/spanner/postgresql/drk.yaml
```
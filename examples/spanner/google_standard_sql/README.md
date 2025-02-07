Generate a spanner.key file

```sh
gcloud iam service-accounts keys create spanner.key --iam-account SERVICE_ACCOUNT_EMAIL
```

Run drk

```sh
drk \
--driver spanner \
--url projects/PROJECT/instances/INSTANCE/databases/DATABASE \
--config examples/spanner/google_standard_sql/drk.yaml
```
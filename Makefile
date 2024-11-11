ecommerce_example:
	cockroach sql \
		--host localhost \
		--insecure \
		-f examples/ecommerce/create.sql

	cockroach sql \
		--host localhost \
		--insecure \
		-f examples/ecommerce/populate.sql
	
	go run drk.go \
		--config "examples/ecommerce/drk.yaml" \
		--url "postgres://root@localhost:26257?sslmode=disable"

payments_example:
	cockroach sql \
		--host localhost \
		--insecure \
		-f examples/payments/create.sql

	cockroach sql \
		--host localhost \
		--insecure \
		-f examples/payments/populate.sql
	
	go run drk.go \
		--config "examples/payments/drk.yaml" \
		--url "postgres://root@localhost:26257?sslmode=disable"
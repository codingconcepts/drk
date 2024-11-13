CREATE TABLE IF NOT EXISTS customer (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "email" STRING NOT NULL,

  UNIQUE("email")
);

CREATE TABLE IF NOT EXISTS account (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "customer_id" UUID NOT NULL REFERENCES customer ("id"),
  "balance" DECIMAL NOT NULL
);

CREATE TABLE IF NOT EXISTS transaction_history (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "src_account_id" UUID NOT NULL REFERENCES account ("id"),
  "dst_account_id" UUID NOT NULL REFERENCES account ("id"),
  "amount" DECIMAL NOT NULL,
  "ts" TIMESTAMPTZ NOT NULL DEFAULT now()
) WITH (
  ttl = 'on',
  ttl_expiration_expression = $$(ts + '1 hour')$$,
  ttl_job_cron = '* * * * *',
  ttl_delete_rate_limit = 1000
);

CREATE OR REPLACE FUNCTION open_account(email_in STRING, balance_in DECIMAL) RETURNS UUID AS $$
DECLARE
  customer_id UUID;
  account_id UUID;
BEGIN

  INSERT INTO customer (email)
  VALUES (email_in)
  RETURNING id INTO customer_id;

  INSERT INTO account (customer_id, balance)
  VALUES (customer_id, balance_in)
  RETURNING id INTO account_id;

  RETURN account_id;

END;
$$ LANGUAGE PLPGSQL;
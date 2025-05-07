CREATE TABLE IF NOT EXISTS account (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "balance" DECIMAL NOT NULL
);

ALTER TABLE account SPLIT AT
  SELECT rpad(to_hex(prefix::INT), 32, '0')::UUID
  FROM generate_series(0, 16) AS prefix;

CREATE OR REPLACE FUNCTION open_account(balance_in DECIMAL) RETURNS UUID AS $$
DECLARE
  account_id UUID;
BEGIN

  INSERT INTO account (balance)
  VALUES (balance_in)
  RETURNING id INTO account_id;

  RETURN account_id;

END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE PROCEDURE make_transfer(from_id UUID, to_id UUID, amount DECIMAL) AS $$
BEGIN

  UPDATE account SET
    balance = CASE 
                WHEN id = to_id THEN balance + amount
                WHEN id = from_id THEN balance - amount
              END
  WHERE id IN (to_id, from_id);

END;
$$ LANGUAGE PLPGSQL;


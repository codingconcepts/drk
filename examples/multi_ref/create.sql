CREATE TABLE IF NOT EXISTS account (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "balance" DECIMAL NOT NULL
);
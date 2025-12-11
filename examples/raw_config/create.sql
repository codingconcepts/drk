CREATE TABLE IF NOT EXISTS measurement (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "value" DECIMAL NOT NULL,
  "ts" TIMESTAMPTZ NOT NULL DEFAULT now()
);
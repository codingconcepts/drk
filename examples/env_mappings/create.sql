CREATE TABLE IF NOT EXISTS t (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  region STRING NOT NULL
);
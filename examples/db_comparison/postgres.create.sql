CREATE TABLE IF NOT EXISTS t (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  val STRING NOT NULL
);

INSERT INTO t (val)
SELECT 
  LEFT(sha256(random()::TEXT), 16)
FROM generate_series(1, 1000);
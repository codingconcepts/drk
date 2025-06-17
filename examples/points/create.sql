CREATE TYPE gender AS ENUM ('male', 'female', 'trans-male', 'trans-female', 'non-binary');

CREATE TABLE IF NOT EXISTS shopper (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "email" STRING NOT NULL,
  "gender" gender,
  "date_of_birth" DATE,
  "location" GEOMETRY,

  UNIQUE("email")
);
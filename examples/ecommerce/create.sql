CREATE TABLE shopper (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "email" STRING NOT NULL,

  UNIQUE("email")
);

CREATE TABLE product (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "name" STRING NOT NULL
);

CREATE TABLE basket (
  "shopper_id" UUID NOT NULL REFERENCES shopper("id"),
  "product_id" UUID NOT NULL REFERENCES product("id"),
  "quantity" INT NOT NULL DEFAULT 1,

  PRIMARY KEY ("shopper_id", "product_id")
);
CREATE TABLE shopper (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "email" STRING NOT NULL,

  UNIQUE("email")
);

CREATE TABLE product (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "name" STRING NOT NULL,
  "price" DECIMAL NOT NULL
);

CREATE TABLE basket (
  "shopper_id" UUID NOT NULL REFERENCES shopper("id"),
  "product_id" UUID NOT NULL REFERENCES product("id"),
  "quantity" INT NOT NULL DEFAULT 1,

  PRIMARY KEY ("shopper_id", "product_id")
);

CREATE TABLE purchase (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "shopper_id" UUID NOT NULL REFERENCES shopper("id"),
  "total" DECIMAL NOT NULL,
  "ts" TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE purchase_item (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "purchase_id" UUID NOT NULL REFERENCES purchase("id"),
  "product_id" UUID NOT NULL REFERENCES product("id"),
  "quantity" INT NOT NULL
);

CREATE OR REPLACE FUNCTION checkout(shopper_id_in UUID) RETURNS UUID AS $$
DECLARE
  purchase_id UUID;
BEGIN
  -- Create purchase.
  INSERT INTO purchase (shopper_id, total)
    SELECT 
      b.shopper_id,
      SUM(b.quantity * p.price) as total
    FROM basket b
    JOIN product p ON b.product_id = p.id
    WHERE b.shopper_id = shopper_id_in
    GROUP BY b.shopper_id
  RETURNING id INTO purchase_id;

  -- Create purchase lines.
  INSERT INTO purchase_item (purchase_id, product_id, quantity)
    SELECT
      purchase_id, 
      b.product_id,
      b.quantity
    FROM basket b
    WHERE b.shopper_id = shopper_id_in;

  -- Clear basket.
  DELETE FROM basket
  WHERE shopper_id = shopper_id_in;

  RETURN purchase_id;

END;
$$ LANGUAGE PLPGSQL;
CREATE TABLE IF NOT EXISTS shopper (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email STRING NOT NULL
);

CREATE TABLE IF NOT EXISTS product (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name STRING NOT NULL,
  price DECIMAL NOT NULL
);

CREATE TABLE IF NOT EXISTS basket (
  shopper_id UUID NOT NULL REFERENCES shopper (id),
  product_id UUID NOT NULL REFERENCES product (id),
  quantity INT NOT NULL
);
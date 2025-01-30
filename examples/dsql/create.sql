CREATE TABLE IF NOT EXISTS shopper (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS product (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  price DECIMAL NOT NULL
);

CREATE TABLE IF NOT EXISTS basket (
  shopper_id UUID NOT NULL,
  product_id UUID NOT NULL,
  quantity INT NOT NULL,

  PRIMARY KEY (shopper_id, product_id)
);
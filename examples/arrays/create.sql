CREATE TABLE shopper (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email STRING NOT NULL,
  favourite_products STRING[]
);
CREATE TABLE product (
  id STRING(36) DEFAULT (GENERATE_UUID()),
  name STRING(MAX),
  price FLOAT64
) PRIMARY KEY (id);
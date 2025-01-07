SET sqlblanklines on;

-- Product

CREATE TABLE product (
  id NUMBER PRIMARY KEY,
  name VARCHAR(25) NOT NULL,
  price DECIMAL(15, 2) NOT NULL
);

CREATE SEQUENCE product_seq START WITH 1 INCREMENT BY 1;

CREATE OR REPLACE TRIGGER product_set_id 
BEFORE INSERT ON product 
FOR EACH ROW
BEGIN
  SELECT product_seq.NEXTVAL
  INTO   :new.id
  FROM   dual;
END;

-- Customer

CREATE TABLE customer (
  id NUMBER PRIMARY KEY,
  email VARCHAR(255) NOT NULL
);

CREATE SEQUENCE customer_seq START WITH 1 INCREMENT BY 1;

CREATE OR REPLACE TRIGGER customer_set_id 
BEFORE INSERT ON customer 
FOR EACH ROW
BEGIN
  SELECT customer_seq.NEXTVAL
  INTO   :new.id
  FROM   dual;
END;

-- Basket

CREATE TABLE basket (
  customer_id NUMBER NOT NULL,
  product_id NUMBER NOT NULL,
  quantity NUMBER NOT NULL,

  CONSTRAINT pk_basket PRIMARY KEY (customer_id, product_id),
  CONSTRAINT fk_customer FOREIGN KEY (customer_id) REFERENCES customer (id),
  CONSTRAINT fk_product FOREIGN KEY (product_id) REFERENCES product (id)
);

-- Purchase

CREATE TABLE purchase (
  id NUMBER PRIMARY KEY,
  customer_id NUMBER NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT fk_purchase_customer FOREIGN KEY (customer_id) REFERENCES customer (id)
);

CREATE SEQUENCE purchase_seq START WITH 1 INCREMENT BY 1;

CREATE OR REPLACE TRIGGER purchase_set_id 
BEFORE INSERT ON purchase 
FOR EACH ROW
BEGIN
  SELECT purchase_seq.NEXTVAL
  INTO   :new.id
  FROM   dual;
END;

/
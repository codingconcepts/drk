CREATE TABLE t (
  id INT NOT NULL AUTO_INCREMENT,
  val varchar(255) NOT NULL,
  PRIMARY KEY (id)
);

INSERT INTO t (val)
WITH RECURSIVE numbers AS (
  SELECT 1 AS n
  UNION ALL
  SELECT n + 1 FROM numbers WHERE n < 1000
)
SELECT 
  LEFT(SHA2(RAND(), 256), 16)
FROM numbers;

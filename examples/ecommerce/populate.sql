WITH RECURSIVE chars AS (
  SELECT generate_series(1, 1000) AS row_num, '' AS name, 1 AS pos
  UNION ALL
  SELECT row_num,
         name || chr(CASE WHEN random() < 0.5 
                    THEN trunc(random() * 26 + 65)::INT
                    ELSE trunc(random() * 26 + 97)::INT
                    END),
         pos + 1
  FROM chars
  WHERE pos <= 10
)
INSERT INTO product (name, price)
SELECT 
  name,
  ROUND(CAST(random() * 99 + 1 AS DECIMAL), 2) AS price
FROM (
  SELECT DISTINCT ON (row_num) name
  FROM chars
  WHERE pos > 10
) final;
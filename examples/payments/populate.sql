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
  WHERE pos <= 8
)
SELECT open_account (name || '@example.com', ROUND(CAST(random() * 10000 AS DECIMAL), 2))
FROM (
  SELECT DISTINCT ON (row_num) name
  FROM chars
  WHERE pos > 8
);
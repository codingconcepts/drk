INSERT INTO account (balance)
SELECT random() * 10000
FROM generate_series(1, 1000);
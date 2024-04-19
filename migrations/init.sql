CREATE TABLE IF NOT EXISTS deductions (
  id SERIAL NOT NULL,
  slug VARCHAR NOT NULL,
	"name" VARCHAR NOT NULL,
  amount DECIMAL(10,2) NOT NULL,
  minAmount DECIMAL(10,2) DEFAULT 0,
  maxAmount DECIMAL(10,2),
	CONSTRAINT deductions_pk PRIMARY KEY (id),
	CONSTRAINT deductions_slug_unique UNIQUE (slug)
);

COMMENT ON COLUMN deductions.amount IS 'limit deduction amount in system if set to 0 mean no limit';
COMMENT ON COLUMN deductions.minAmount IS 'lowest amount that allow admin setup to deduction if set to 0 mean no limit';
COMMENT ON COLUMN deductions.maxAmount IS 'highest amount that allow admin setup to deduction if set to 0 mean no limit';

CREATE UNIQUE INDEX IF NOT EXISTS 
  deductions_slug_idx 
ON deductions (slug);

INSERT INTO 
  deductions (slug, "name", amount, minAmount, maxAmount)
VALUES
  ('k-receipt', 'kReceipt', 50000, 0, 100000),
  ('personal','personalDeduction', 60000, 10000, 100000),
  ('donation', 'Donation', 0, 0, 0);

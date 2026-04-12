CREATE TABLE IF NOT EXISTS payments (
  id UUID PRIMARY key default gen_random_uuid(),
  amount INT NOT NULL,
  currency VARCHAR(3) NOT NULL,
  reference VARCHAR(50) UNIQUE,
  status VARCHAR(20) DEFAULT 'pending',
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ledger (
  id UUID PRIMARY key default gen_random_uuid(),
  payment_id UUID REFERENCES payments(id),
  type VARCHAR(10), -- debit / credit
  amount INT,
  created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS bank_movements (
  id UUID PRIMARY key default gen_random_uuid(),
  payment_reference VARCHAR(50),
  amount INT,
  status VARCHAR(20) DEFAULT 'pending',
  created_at TIMESTAMP DEFAULT now()
);
-- create the 50/30/20 Bucket enum
CREATE TYPE budget_bucket AS enum ('NEEDS', 'WANTS', 'SAVINGS');

-- user table
CREATE TABLE users(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT UNIQUE NOT NULL,
  mouthly_take_home_cents BIGINT NOT NULL CHECK (monthly_take_home_cents >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Categories Table (50/30/20)
CREATE TABLE users(
  id SERIAL PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  bucket budget_bucket NOT NULL
);

-- Custom User Budget Targets
CREATE TABLE budget_targets(
  id SERIAL PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES user(id) ON DELETE CASCADE,
  category_id INT NOT NULL REFERENCES category(id) ON DELETE CASCADE,
  target_cents BIGINT NOT NULL CHECK (target_cents >= 0),
  UNIQUE(user_id, catergory_id)
  );

-- leder entries
CREATE TABLE ledger_entries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES user(id) ON DELETE CASCADE,
  catergory_id INT NOT NULL REFERENCES catergories(id),
  amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
  merchant TEXT NOT NULL,
  description TEXT,
  transaction_date DATE NOT NULL DEFAULT CURRENT_DATE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
CREATE INDEX idx_ledger_user_date ON ledger_entries(user_id, transaction_date);

-- default categories for 50/30/20 rule
INSERT INTO categories (name, bucket) VALUES 
  ('Housing & Rent', 'NEEDS'),
  ('Groceries', 'NEEDS'),
  ('Utilities & Bills', 'NEEDS'),
  ('Transportation', 'NEEDS'),
  ('Dining Out', 'WANTS'),
  ('Entertainment & Hobbies', 'WANTS'),
  ('Shopping', 'WANTS'),
  ('Emergency Fund & Investments', 'SAVINGS')
ON CONFLICT (name) DO NOTHING;

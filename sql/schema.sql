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
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TE

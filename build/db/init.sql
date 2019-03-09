
-- Final account state.
CREATE TABLE account (
	id serial NOT NULL,
	name text,
	currency text,
	balance double precision,
	PRIMARY KEY(id),
	CONSTRAINT account_unique UNIQUE(currency, name));

-- Set of atomic actions applied to accounts.
CREATE TABLE trans (
  id serial NOT NULL,
	time timestamp NOT NULL,
	author text NOT NULL,
	PRIMARY KEY(id));

-- Actions applied to accounts.
CREATE TABLE action (
  account integer NOT NULL REFERENCES account(id) ON DELETE RESTRICT,
	trans integer NOT NULL REFERENCES trans(id) ON DELETE CASCADE,
  volume double precision NOT NULL,
	PRIMARY KEY(account, trans));
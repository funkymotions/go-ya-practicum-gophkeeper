CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  username VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE block_types (
	id serial4 NOT NULL,
	type_name varchar NOT NULL,
	description varchar NOT NULL,
	CONSTRAINT block_types_pk PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS blocks (
  id SERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  user_id INT NOT NULL REFERENCES users(id),
  type_id INT NOT NULL REFERENCES block_types(id),
  data BYTEA NOT NULL,
  salt BYTEA NOT NULL,
  nonce BYTEA NOT NULL,
  profile VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_blocks_user_id ON blocks(user_id);

INSERT INTO block_types (type_name, description) VALUES
  ('text', 'Raw text data block'),
  ('file', 'File data block'),
  ('credentials', 'Credentials data block'),
  ('bank_card', 'Bank card data block');

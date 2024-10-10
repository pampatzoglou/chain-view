CREATE EXTENSION IF NOT EXISTS "pgcrypto" WITH SCHEMA public;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;

CREATE TABLE IF NOT EXISTS addresses (
    id SERIAL PRIMARY KEY,
    address VARCHAR(42) NOT NULL UNIQUE,
    current_balance NUMERIC(38, 18) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    transaction_hash VARCHAR(66) NOT NULL UNIQUE,
    from_address VARCHAR(42) NOT NULL REFERENCES addresses(address),
    to_address VARCHAR(42) NOT NULL REFERENCES addresses(address),
    value NUMERIC(38, 18) NOT NULL,
    gas_price NUMERIC(38, 18) NOT NULL,
    gas_used NUMERIC NOT NULL,
    block_number BIGINT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS gas_fees (
    id SERIAL PRIMARY KEY,
    gas_price NUMERIC(38, 18) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS balances (
    id SERIAL PRIMARY KEY,
    address_id INT NOT NULL REFERENCES addresses(id),
    balance NUMERIC(38, 18) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE OR REPLACE VIEW current_balances AS
SELECT 
    a.address, 
    a.current_balance
FROM 
    addresses a;

CREATE OR REPLACE VIEW recent_transactions AS
SELECT 
    t.transaction_hash, 
    t.from_address, 
    t.to_address, 
    t.value, 
    t.gas_price, 
    t.gas_used, 
    t.block_number, 
    t.timestamp
FROM 
    transactions t
ORDER BY 
    t.timestamp DESC
LIMIT 100;  -- Adjust the limit as necessary

CREATE OR REPLACE VIEW gas_fee_trends AS
SELECT 
    DATE_TRUNC('minute', timestamp) AS minute,
    AVG(gas_price) AS avg_gas_price,
    MAX(gas_price) AS max_gas_price,
    MIN(gas_price) AS min_gas_price
FROM 
    gas_fees
GROUP BY 
    minute
ORDER BY 
    minute DESC;

CREATE OR REPLACE VIEW address_balance_changes AS
SELECT 
    b.address_id, 
    a.address, 
    b.balance, 
    b.timestamp
FROM 
    balances b
JOIN 
    addresses a ON b.address_id = a.id
ORDER BY 
    b.timestamp DESC;

CREATE INDEX IF NOT EXISTS idx_transactions_from_address ON transactions(from_address);
CREATE INDEX IF NOT EXISTS idx_transactions_to_address ON transactions(to_address);
CREATE INDEX IF NOT EXISTS idx_gas_fees_timestamp ON gas_fees(timestamp);

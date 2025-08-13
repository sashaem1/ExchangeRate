CREATE TABLE IF NOT EXISTS exchange_rates (
    id SERIAL PRIMARY KEY,
    BaseCurrency VARCHAR(3) NOT NULL,
    TargetCurrency VARCHAR(3) NOT NULL,
    rate FLOAT NOT NULL,
    updated_at DATE DEFAULT CURRENT_DATE,
    CONSTRAINT unique_exchange_date UNIQUE (BaseCurrency, TargetCurrency, updated_at)
);

CREATE TABLE IF NOT EXISTS exchange_rates_log (
    id SERIAL PRIMARY KEY,
    action_name VARCHAR(4) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS api_keys (
	key TEXT PRIMARY KEY
);

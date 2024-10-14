CREATE TYPE loan_status AS ENUM ('open', 'close');

CREATE TABLE loans (
    id SERIAL PRIMARY KEY,
    customer_id INT REFERENCES customers(id),
    amount NUMERIC(12, 2) NOT NULL,
    total_amount NUMERIC(12, 2) NOT NULL,
    status loan_status DEFAULT 'open', -- Enum with 'open' and 'close', default is 'open'
    term_weeks INT NOT NULL,
    rates NUMERIC(5, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

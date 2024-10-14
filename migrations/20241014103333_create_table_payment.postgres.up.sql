-- Create enum type for payment status
CREATE TYPE payment_status AS ENUM ('scheduled', 'outstanding', 'paid');

-- Create payments table
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    loan_id INT REFERENCES loans(id),
    week INT NOT NULL,
    amount NUMERIC(12, 2) NOT NULL,
    due_date DATE NOT NULL,
    status payment_status DEFAULT 'scheduled', -- Enum with default value 'scheduled'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

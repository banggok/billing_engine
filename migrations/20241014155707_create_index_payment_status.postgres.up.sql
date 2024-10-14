-- 20241014123045_add_index_to_status_in_payments_table.up.sql

-- Add index to the 'status' column of 'payments' table (without CONCURRENTLY)
CREATE INDEX IF NOT EXISTS idx_payment_status ON payments (status);

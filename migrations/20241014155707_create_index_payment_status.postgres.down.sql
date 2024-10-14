-- 20241014123046_add_index_to_status_in_payments_table.down.sql

-- Drop the index on the 'status' column
DROP INDEX IF EXISTS idx_payment_status;

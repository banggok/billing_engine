-- 20241014123045_add_pending_to_payment_status_enum.up.sql

DO $$ 
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_type WHERE typname = 'payment_status'
    ) THEN
        -- Add new status 'pending' to the enum type 'payment_status'
        ALTER TYPE payment_status ADD VALUE IF NOT EXISTS 'pending';
    END IF;
END $$;

-- 20241014123045_add_pending_to_payment_status_enum.down.sql

DO $$ 
BEGIN
    -- Step 1: Create a new temporary enum without the 'pending' value
    CREATE TYPE payment_status_new AS ENUM ('scheduled', 'outstanding', 'paid');

    -- Step 2: Remove the default value from the 'status' column
    ALTER TABLE payments 
    ALTER COLUMN status DROP DEFAULT;

    -- Step 3: Alter the 'payments' table to use the new enum type
    ALTER TABLE payments 
    ALTER COLUMN status TYPE payment_status_new USING status::text::payment_status_new;

    -- Step 4: Re-apply the default value for the 'status' column
    ALTER TABLE payments 
    ALTER COLUMN status SET DEFAULT 'scheduled';

    -- Step 5: Drop the old enum type
    DROP TYPE payment_status;

    -- Step 6: Rename the new enum type to the original name
    ALTER TYPE payment_status_new RENAME TO payment_status;
END $$;

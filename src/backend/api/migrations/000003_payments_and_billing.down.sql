-- Drop in reverse dependency order
DROP INDEX IF EXISTS idx_receipt_access_tokens_payment_id;
DROP TABLE IF EXISTS receipt_access_tokens;

DROP INDEX IF EXISTS idx_payment_transaction_attachments_txn_id;
DROP TABLE IF EXISTS payment_transaction_attachments;

DROP INDEX IF EXISTS idx_payment_transactions_payment_id;
DROP TABLE IF EXISTS payment_transactions;

DROP TABLE IF EXISTS payment_receipt_sequences;

DROP INDEX IF EXISTS idx_billing_run_items_run;
DROP INDEX IF EXISTS idx_billing_runs_org_property;
DROP TABLE IF EXISTS billing_run_items;
DROP TABLE IF EXISTS billing_runs;

DROP INDEX IF EXISTS idx_electricity_bill_run_items_run;
DROP INDEX IF EXISTS idx_electricity_bill_runs_org_building;
DROP TABLE IF EXISTS electricity_bill_run_items;
DROP TABLE IF EXISTS electricity_bill_runs;

DROP INDEX IF EXISTS idx_payments_org_status_month_year;
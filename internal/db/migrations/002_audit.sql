-- +goose Up
CREATE TABLE IF NOT EXISTS audit_log (
    id BIGSERIAL PRIMARY KEY,
    table_name TEXT NOT NULL,
    op TEXT NOT NULL,
    row_id TEXT,
    old_data JSONB,
    new_data JSONB,
    at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION audit_row_change() RETURNS trigger AS $audit_fn$
BEGIN
    IF TG_OP = 'DELETE' THEN
        INSERT INTO audit_log (table_name, op, row_id, old_data)
        VALUES (TG_TABLE_NAME, 'DELETE', OLD.id::text, to_jsonb(OLD));
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_log (table_name, op, row_id, old_data, new_data)
        VALUES (TG_TABLE_NAME, 'UPDATE', NEW.id::text, to_jsonb(OLD), to_jsonb(NEW));
        RETURN NEW;
    ELSIF TG_OP = 'INSERT' THEN
        INSERT INTO audit_log (table_name, op, row_id, new_data)
        VALUES (TG_TABLE_NAME, 'INSERT', NEW.id::text, to_jsonb(NEW));
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$audit_fn$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS audit_items ON inventory_items;
CREATE TRIGGER audit_items
AFTER INSERT OR UPDATE OR DELETE ON inventory_items
FOR EACH ROW EXECUTE FUNCTION audit_row_change();

DROP TRIGGER IF EXISTS audit_tx ON stock_transactions;
CREATE TRIGGER audit_tx
AFTER INSERT OR UPDATE OR DELETE ON stock_transactions
FOR EACH ROW EXECUTE FUNCTION audit_row_change();

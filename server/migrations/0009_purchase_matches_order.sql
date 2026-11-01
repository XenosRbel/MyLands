-- 0009_purchase_matches_order (ML-B07): a 'purchase' ledger row must match its payment order.
-- Defect found by TestPurchaseCreditMustMatchOrder: 0002 required only a payment_order_id and one purchase per order.
-- A purchase could credit an unconfirmed order, another wallet's order, or a different amount.
-- Additive: one BEFORE INSERT trigger. Existing rows are not rescanned.

CREATE FUNCTION premium_ledger_purchase_matches_order() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE o payment_orders%ROWTYPE;
BEGIN
    -- A NULL order is rejected by the CHECK in 0002.
    IF NEW.reason <> 'purchase' OR NEW.payment_order_id IS NULL THEN RETURN NEW; END IF;
    -- Confirmed is terminal, so the status cannot change after this check.
    SELECT * INTO o FROM payment_orders WHERE id = NEW.payment_order_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'payment order % not found', NEW.payment_order_id USING ERRCODE = 'foreign_key_violation';
    END IF;
    IF o.wallet_id <> NEW.wallet_id OR o.status <> 'confirmed' OR o.crystals <> NEW.delta THEN
        RAISE EXCEPTION 'purchase must match a confirmed order of the same wallet and amount (order %, status %)', o.id, o.status
            USING ERRCODE = 'check_violation';
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER premium_ledger_purchase_matches_order BEFORE INSERT ON premium_ledger
    FOR EACH ROW EXECUTE FUNCTION premium_ledger_purchase_matches_order();

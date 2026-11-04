-- name: CreateInvoice :one
INSERT INTO invoices (id, owner_id, invoice_number, amount, status, payment_ref)
    VALUES (@id, @owner_id, @invoice_number, @amount, @status, @payment_ref)
RETURNING
    id, owner_id, invoice_number, amount, status, payment_ref, created_at;

-- name: GetInvoiceByNumber :one
SELECT
    id,
    owner_id,
    invoice_number,
    amount,
    status,
    payment_ref,
    created_at
FROM
    invoices
WHERE
    invoice_number = @invoice_number;


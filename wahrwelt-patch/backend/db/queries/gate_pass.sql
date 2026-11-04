-- name: CreateGatePass :one
INSERT INTO gate_passes (id, owner_id, gate_operator_id, gate_pass_number, state, gate_word)
    VALUES (@id, @owner_id, @gate_operator_id, @gate_pass_number, @state, @gate_word)
RETURNING
    id, owner_id, gate_operator_id, gate_pass_number, state, gate_word, created_at;

-- name: GetGatePassByNumber :one
SELECT
    id,
    owner_id,
    gate_operator_id,
    gate_pass_number,
    state,
    gate_word,
    created_at
FROM
    gate_passes
WHERE
    gate_pass_number = @gate_pass_number;

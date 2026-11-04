-- name: CreateContract :one
INSERT INTO service_contracts (id, owner_id, contract_number, crew_name, service, crew_phrase, starts_at, ends_at)
    VALUES (@id, @owner_id, @contract_number, @crew_name, @service, @crew_phrase, @starts_at, @ends_at)
RETURNING
    id, owner_id, contract_number, crew_name, service, crew_phrase, starts_at, ends_at, created_at;

-- name: GetContractByNumber :one
SELECT
    id,
    owner_id,
    contract_number,
    crew_name,
    service,
    crew_phrase,
    starts_at,
    ends_at,
    created_at
FROM
    service_contracts
WHERE
    contract_number = @contract_number;


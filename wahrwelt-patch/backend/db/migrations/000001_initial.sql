-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY,
    username text NOT NULL,
    password_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT users_username_key UNIQUE (username),
    CONSTRAINT users_username_format CHECK (username ~ '^[a-z0-9_.-]{3,32}$')
);

INSERT INTO users (id, username, password_hash)
    VALUES
        ('00000000-0000-4000-8000-000000000001', 'admin', '$2b$12$QHHhzVaXt3hOWH.WJTSCGe.plfK3u6pmOAhqdHH122JUnKmJug4K6'),
        ('00000000-0000-4000-8000-000000000002', 'carrier', '$2b$12$998Um9kUJd.tn1jpb64MlOqAtvyhql0f/F9qa18lUt78PO/pz9kuu'),
        ('00000000-0000-4000-8000-000000000003', 'inspector', '$2b$12$nobMioBNHndrN5GiPR46TORr.e7/fhEgVoXqBmgbl1MTnjDnTCaSG');

CREATE TABLE cases (
    id uuid PRIMARY KEY,
    owner_id uuid NOT NULL REFERENCES users (id),
    kind text NOT NULL CHECK (kind IN ('import', 'export', 'service')),
    origin text NOT NULL CHECK (length(origin) BETWEEN 1 AND 120),
    destination text NOT NULL CHECK (length(destination) BETWEEN 1 AND 120),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT cases_id_owner_key UNIQUE (id, owner_id)
);

CREATE TABLE route_slots (
    id uuid PRIMARY KEY,
    case_id uuid NOT NULL,
    owner_id uuid NOT NULL,
    carrier_id uuid NOT NULL REFERENCES users (id),
    inspector_id uuid NOT NULL REFERENCES users (id),
    slot_code text NOT NULL CHECK (length(slot_code) BETWEEN 1 AND 120),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT route_slots_case_owner_fkey FOREIGN KEY (case_id, owner_id) REFERENCES cases (id, owner_id),
    CONSTRAINT route_slots_distinct_approvers CHECK (carrier_id <> inspector_id)
);

CREATE INDEX route_slots_case_id_idx ON route_slots (case_id);

CREATE INDEX route_slots_owner_id_idx ON route_slots (owner_id);

CREATE TABLE manifests (
    id uuid PRIMARY KEY,
    owner_id uuid NOT NULL REFERENCES users (id),
    manifest_number text NOT NULL UNIQUE,
    origin text NOT NULL CHECK (length(origin) BETWEEN 1 AND 120),
    destination text NOT NULL CHECK (length(destination) BETWEEN 1 AND 120),
    status text NOT NULL CHECK (status IN ('received', 'in_review', 'released')),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT manifests_id_owner_key UNIQUE (id, owner_id),
    CONSTRAINT manifests_number_format CHECK (manifest_number ~ '^MNF-[0-9a-f-]{36}$')
);

CREATE TABLE manifest_revisions (
    id uuid PRIMARY KEY,
    manifest_id uuid NOT NULL,
    owner_id uuid NOT NULL,
    status text NOT NULL CHECK (status IN ('received', 'in_review', 'released')),
    sealed_route text NOT NULL CHECK (length(sealed_route) BETWEEN 1 AND 240),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT manifest_revisions_manifest_owner_fkey FOREIGN KEY (manifest_id, owner_id) REFERENCES manifests (id, owner_id)
);

CREATE INDEX manifest_revisions_manifest_id_created_idx ON manifest_revisions (manifest_id, created_at);

CREATE TABLE certificates (
    id uuid PRIMARY KEY,
    owner_id uuid NOT NULL REFERENCES users (id),
    serial text NOT NULL UNIQUE,
    origin_phrase text NOT NULL CHECK (length(origin_phrase) BETWEEN 1 AND 240),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT certificates_serial_format CHECK (serial ~ '^CRT-[0-9a-f-]{36}$')
);

CREATE INDEX certificates_owner_id_idx ON certificates (owner_id);

CREATE TABLE invoices (
    id uuid PRIMARY KEY,
    owner_id uuid NOT NULL REFERENCES users (id),
    invoice_number text NOT NULL UNIQUE,
    amount bigint NOT NULL CHECK (amount > 0),
    status text NOT NULL CHECK (status IN ('pending', 'paid')),
    payment_ref text NOT NULL CHECK (length(payment_ref) BETWEEN 1 AND 240),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT invoices_number_format CHECK (invoice_number ~ '^INV-[0-9a-f-]{36}$')
);

CREATE INDEX invoices_owner_id_idx ON invoices (owner_id);

CREATE TABLE inspections (
    id uuid PRIMARY KEY,
    case_id uuid NOT NULL,
    owner_id uuid NOT NULL REFERENCES users (id),
    inspector_id uuid NOT NULL REFERENCES users (id),
    inspection_number text NOT NULL UNIQUE,
    inspector_note text NOT NULL CHECK (length(inspector_note) BETWEEN 1 AND 500),
    note_visibility text NOT NULL DEFAULT 'private' CHECK (note_visibility IN ('private', 'public')),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT inspections_case_owner_fkey FOREIGN KEY (case_id, owner_id) REFERENCES cases (id, owner_id),
    CONSTRAINT inspections_distinct_participants CHECK (owner_id <> inspector_id),
    CONSTRAINT inspections_number_format CHECK (inspection_number ~ '^INSP-[0-9a-f-]{36}$')
);

CREATE INDEX inspections_owner_id_idx ON inspections (owner_id);

CREATE INDEX inspections_inspector_id_idx ON inspections (inspector_id);

CREATE TABLE inspection_comments (
    id uuid PRIMARY KEY,
    inspection_id uuid NOT NULL REFERENCES inspections (id) ON DELETE CASCADE,
    author_id uuid NOT NULL REFERENCES users (id),
    text text NOT NULL CHECK (length(text) BETWEEN 1 AND 500),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX inspection_comments_inspection_id_created_idx ON inspection_comments (inspection_id, created_at, id);

CREATE TABLE export_licenses (
    id uuid PRIMARY KEY,
    owner_id uuid NOT NULL REFERENCES users (id),
    license_number text NOT NULL UNIQUE,
    status text NOT NULL CHECK (status IN ('active', 'revoked')),
    release_pin text NOT NULL CHECK (length(release_pin) BETWEEN 1 AND 240),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT export_licenses_number_format CHECK (license_number ~ '^LIC-[0-9a-f-]{36}$')
);

CREATE INDEX export_licenses_owner_id_idx ON export_licenses (owner_id);

CREATE TABLE service_contracts (
    id uuid PRIMARY KEY,
    owner_id uuid NOT NULL REFERENCES users (id),
    contract_number text NOT NULL UNIQUE,
    crew_name text NOT NULL CHECK (length(crew_name) BETWEEN 1 AND 120),
    service text NOT NULL CHECK (length(service) BETWEEN 1 AND 240),
    crew_phrase text NOT NULL CHECK (length(crew_phrase) BETWEEN 1 AND 240),
    starts_at timestamptz NOT NULL,
    ends_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT service_contracts_number_format CHECK (contract_number ~ '^CTR-[0-9a-f-]{36}$'),
    CONSTRAINT service_contracts_time_order CHECK (ends_at > starts_at)
);

CREATE INDEX service_contracts_owner_id_idx ON service_contracts (owner_id);

CREATE TABLE quotas (
    id uuid PRIMARY KEY,
    owner_id uuid NOT NULL REFERENCES users (id),
    quota_number text NOT NULL UNIQUE,
    quota_code text NOT NULL CHECK (length(quota_code) BETWEEN 1 AND 120),
    volume bigint NOT NULL CHECK (volume BETWEEN 1 AND 1000000000),
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT quotas_number_format CHECK (quota_number ~ '^QUO-[0-9a-f-]{36}$')
);

CREATE INDEX quotas_owner_id_idx ON quotas (owner_id);

CREATE TABLE archive_waybills (
    id uuid PRIMARY KEY,
    owner_id uuid NOT NULL REFERENCES users (id),
    waybill_number text NOT NULL UNIQUE,
    archive_code text NOT NULL CHECK (length(archive_code) BETWEEN 1 AND 240),
    state text NOT NULL DEFAULT 'ready' CHECK (state = 'ready'),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT archive_waybills_number_format CHECK (waybill_number ~ '^AWB-[0-9a-f-]{36}$')
);

CREATE INDEX archive_waybills_owner_id_idx ON archive_waybills (owner_id);

CREATE TABLE previews (
    id uuid PRIMARY KEY,
    owner_id uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX previews_owner_id_idx ON previews (owner_id);

CREATE TABLE gate_passes (
    id uuid PRIMARY KEY,
    owner_id uuid NOT NULL REFERENCES users (id),
    gate_operator_id uuid NOT NULL REFERENCES users (id),
    gate_pass_number text NOT NULL UNIQUE,
    state text NOT NULL CHECK (state IN ('issued', 'closed')),
    gate_word text NOT NULL CHECK (length(gate_word) BETWEEN 1 AND 240),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT gate_passes_distinct_participants CHECK (owner_id <> gate_operator_id),
    CONSTRAINT gate_passes_number_format CHECK (gate_pass_number ~ '^GATE-[0-9a-f-]{36}$')
);

CREATE INDEX gate_passes_owner_id_idx ON gate_passes (owner_id);

CREATE INDEX gate_passes_operator_id_idx ON gate_passes (gate_operator_id);

CREATE ROLE wahrwelt_invoice_search LOGIN NOINHERIT PASSWORD 'wahrwelt';

GRANT USAGE ON SCHEMA public TO wahrwelt_invoice_search;

GRANT SELECT ON invoices TO wahrwelt_invoice_search;

-- +goose Down
DROP TABLE gate_passes;

DROP TABLE previews;

DROP TABLE archive_waybills;

DROP TABLE quotas;

DROP TABLE service_contracts;

DROP TABLE export_licenses;

DROP TABLE inspection_comments;

DROP TABLE inspections;

DROP TABLE invoices;

DROP TABLE certificates;

DROP TABLE manifest_revisions;

DROP TABLE manifests;

DROP TABLE route_slots;

DROP TABLE cases;

DROP TABLE users;

REVOKE ALL PRIVILEGES ON SCHEMA public FROM wahrwelt_invoice_search;

DROP ROLE wahrwelt_invoice_search;

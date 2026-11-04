-- name: CreateInspection :one
INSERT INTO inspections (id, case_id, owner_id, inspector_id, inspection_number, inspector_note, note_visibility)
    VALUES (@id, @case_id, @owner_id, @inspector_id, @inspection_number, @inspector_note, @note_visibility)
RETURNING
    id, case_id, owner_id, inspector_id, inspection_number, inspector_note, note_visibility, created_at;

-- name: GetInspectionByNumber :one
SELECT
    id,
    case_id,
    owner_id,
    inspector_id,
    inspection_number,
    inspector_note,
    note_visibility,
    created_at
FROM
    inspections
WHERE
    inspection_number = @inspection_number;

-- name: ListInspectionComments :many
SELECT
    id,
    inspection_id,
    author_id,
    text,
    created_at
FROM
    inspection_comments
WHERE
    inspection_id = @inspection_id
ORDER BY
    created_at,
    id;

-- name: CreateInspectionComment :one
INSERT INTO inspection_comments (id, inspection_id, author_id, text)
    VALUES (@id, @inspection_id, @author_id, @text)
RETURNING
    id, inspection_id, author_id, text, created_at;

-- name: SetInspectionNoteVisibility :exec
UPDATE
    inspections
SET
    note_visibility = @note_visibility
WHERE
    id = @id;

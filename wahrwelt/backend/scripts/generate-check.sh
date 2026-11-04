#!/bin/sh
set -eu

backend_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
work_dir=$(mktemp -d "${TMPDIR:-/tmp}/wahrwelt-generate.XXXXXX")
trap 'rm -rf -- "$work_dir"' EXIT HUP INT TERM

for path in api codegen config db internal go.mod go.sum
do
	cp -a "$backend_dir/$path" "$work_dir/$path"
done
(
	cd "$work_dir"
	go tool ogen -clean -package api -target internal/adapter/primary/http/api -config codegen/ogen.yml api/openapi.yml
	rm -f internal/adapter/secondary/postgres/sqlc/*.sql.go
	go tool sqlc generate -f codegen/sqlc.yml
	go tool wire gen ./internal/bootstrap
)

for generated_path in \
	internal/adapter/primary/http/api \
	internal/adapter/secondary/postgres/sqlc \
	internal/bootstrap/wire_gen.go
do
	if ! diff -ru "$backend_dir/$generated_path" "$work_dir/$generated_path"; then
		printf 'generated output differs: %s\n' "$generated_path" >&2
		exit 1
	fi
done

printf 'generated files are up to date\n'

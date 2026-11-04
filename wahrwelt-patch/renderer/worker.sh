#!/bin/sh
set -u

preview_root=/var/lib/wahrwelt/preview
mkdir -p "$preview_root"
printf 'ready\n' > "$preview_root/.ready"

set_state() {
	printf '%s\n' "$2" > "$1/state.tmp"
	mv "$1/state.tmp" "$1/state"
}

for working in "$preview_root"/*/working; do
	[ -f "$working" ] || continue
	job=${working%/working}
	mv "$working" "$job/request" 2>/dev/null || true
	dir=${job##*/}
	set_state "$job" queued
done

while :; do
	processed=0
	for request in "$preview_root"/*/request; do
		[ -f "$request" ] || continue
		job=${request%/request}
		dir=${job##*/}
		case "$dir" in
			????????-????-????-????-????????????) ;;
			*) continue ;;
		esac
		hex_id=$(printf '%s' "$dir" | tr -d '-')
		[ "${#hex_id}" -eq 32 ] || continue
		case "$hex_id" in
			*[!0-9a-f]*) continue ;;
		esac
		mv "$request" "$job/working" 2>/dev/null || continue
		processed=1
		set_state "$job" processing
		rm -f "$job/result.png" "$job/stdout.raw" "$job/stderr.raw" "$job/stdout.txt" "$job/stderr.txt"
		(
			ulimit -f 8192
			timeout 8 gs -q -dSAFER -dBATCH -dNOPAUSE -dFirstPage=1 -dLastPage=1 -dFIXEDMEDIA -r72 -sDEVICE=png16m -sOutputFile="$job/result.png" "$job/input.eps"
		) > "$job/stdout.raw" 2> "$job/stderr.raw"
		status=$?
		head -c 8192 "$job/stdout.raw" > "$job/stdout.txt" 2>/dev/null || :
		head -c 8192 "$job/stderr.raw" > "$job/stderr.txt" 2>/dev/null || :
		rm -f "$job/stdout.raw" "$job/stderr.raw"
		if [ "$status" -eq 0 ] && [ -f "$job/result.png" ] && [ "$(wc -c < "$job/result.png")" -le 4194304 ]; then
			set_state "$job" complete
		else
			rm -f "$job/result.png"
			set_state "$job" failed
		fi
		rm -f "$job/working"
	done
	if [ "$processed" -eq 0 ]; then
		sleep 0.25
	fi
done

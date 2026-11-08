function __wahrwelt_make_dir
    test -f Makefile; and test -f docker-compose.yml; and test -f backend/Makefile
end

function __wahrwelt_make_targets
    __wahrwelt_make_dir; or return
    awk '/^[a-z][a-z0-9-]*:.*##[[:space:]]*/ {
        target = $0
        sub(/:.*/, "", target)
        description = $0
        sub(/^.*##[[:space:]]*/, "", description)
        print target "\t" description
    }' Makefile
end

complete -C 'make ' >/dev/null
complete -c make -a '(__wahrwelt_make_targets)'

FROM gcc:13.5.0-bookworm@sha256:20eec427045c47d0aad1bd28c81d435916c327cd992240aa78350f755d04d04a AS build

ARG HTTP_PROXY
ARG HTTPS_PROXY
ARG NO_PROXY

RUN curl --fail --location --retry 2 \
        --output /tmp/ghostscript-10.03.0.tar.xz \
        https://github.com/ArtifexSoftware/ghostpdl-downloads/releases/download/gs10030/ghostscript-10.03.0.tar.xz \
    && echo '74d7da586eefbf3fb3c085bb4d2b73ee667fd0ae9cff73ece613c77c443148b847086f91802124baf91ac35cdc83be07199799378def92c8fbd96ced7d0ee9a8  /tmp/ghostscript-10.03.0.tar.xz' \
        | sha512sum -c - \
    && tar -xf /tmp/ghostscript-10.03.0.tar.xz -C /tmp

WORKDIR /tmp/ghostscript-10.03.0

RUN CFLAGS='-O2 -std=gnu17 -fPIE' LDFLAGS='-pie' \
    ./configure \
        --prefix=/opt/ghostscript \
        --disable-cups \
        --disable-gtk \
        --without-tesseract \
        --with-drivers=PNG,uniprint \
    && make -j2 \
    && make install

RUN set -eu; \
    mkdir -p /opt/ghostscript/lib; \
    ldd /opt/ghostscript/bin/gs > /tmp/gs-libraries; \
    if grep -q 'not found' /tmp/gs-libraries; then \
        cat /tmp/gs-libraries; \
        exit 1; \
    fi; \
    awk '{ for (i = 1; i <= NF; i++) if ($i ~ /^\//) print $i }' /tmp/gs-libraries \
        | sort -u \
        | while IFS= read -r path; do \
            name=${path##*/}; \
            case "$name" in \
                libc.so.*|libm.so.*|libpthread.so.*|libdl.so.*) ;; \
                librt.so.*|libresolv.so.*|libutil.so.*|libanl.so.*|ld-linux*.so.*) ;; \
                *) cp -L "$path" "/opt/ghostscript/lib/$name" ;; \
            esac; \
        done

FROM debian:12.15-slim@sha256:734a72b3d6142a71993b4bc580ca01d8ca687f77a2c422d96a961d3376b179bc

RUN mkdir -p /var/lib/wahrwelt/archive /var/lib/wahrwelt/preview \
    && chown -R 10001:10001 /var/lib/wahrwelt
COPY --from=build /opt/ghostscript /opt/ghostscript
COPY worker.sh /usr/local/bin/renderer
RUN chmod 0755 /usr/local/bin/renderer
ENV PATH=/opt/ghostscript/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
    LD_LIBRARY_PATH=/opt/ghostscript/lib
WORKDIR /var/lib/wahrwelt
USER 10001:10001
HEALTHCHECK --interval=2s --timeout=1s --start-period=3s --retries=20 \
    CMD test -f /var/lib/wahrwelt/preview/.ready
ENTRYPOINT ["/usr/local/bin/renderer"]

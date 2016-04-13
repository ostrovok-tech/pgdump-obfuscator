FROM alpine:edge
MAINTAINER George Kutsurua <g.kutsurua@gmail.com>

COPY . /pgdump-obfuscator

RUN apk --update-cache add go &&\
    cd /pgdump-obfuscator &&\
    go build . &&\
    mv /pgdump-obfuscator/pgdump-obfuscator /usr/sbin/pgdump-obfuscator &&\
    apk del go --force && \
    rm -rf /pgdump-obfuscator /var/cache/apk/*

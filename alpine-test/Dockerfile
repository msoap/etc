# docker build -t alpine-test .
FROM alpine:latest

RUN apk --no-cache \
    add bash less vim mc ca-certificates wget curl man openssl && \
    update-ca-certificates

ADD .bashrc /root/.bashrc

ENTRYPOINT bash

# docker build -t alpine-test .
FROM alpine:latest

RUN apk update && \
    apk add bash less vim mc curl man

ADD .bashrc /root/.bashrc

ENTRYPOINT bash

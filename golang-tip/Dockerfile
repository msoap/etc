FROM golang:latest

RUN cd / \
    && git clone https://go.googlesource.com/go go-src \
    && cd /go-src/src \
    && GOROOT_BOOTSTRAP=/usr/local/go ./all.bash || true \
    && cd /usr/local/go/bin \
    && ln -fs /go-src/bin/* ./

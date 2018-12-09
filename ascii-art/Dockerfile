# docker build -t msoap/ascii-art .
FROM golang as builder

ENV FIGLET=figlet-2.2.5
ENV COWSAY=cowsay-3.03

# install figlet
RUN wget ftp://ftp.figlet.org/pub/figlet/program/unix/${FIGLET}.tar.gz -O /tmp/${FIGLET}.tar.gz \
    && chdir /tmp \
    && tar -xzf /tmp/${FIGLET}.tar.gz \
    && chdir /tmp/${FIGLET} \
    && make install

# install cowsay
RUN wget https://web.archive.org/web/20120225123719/http://www.nog.net/~tony/warez/${COWSAY}.tar.gz -O /tmp/${COWSAY}.tar.gz \
    && chdir /tmp \
    && tar -xzf /tmp/${COWSAY}.tar.gz \
    && chdir /tmp/${COWSAY} \
    && sh install.sh

# install shell2http, new-cowsay
ENV GOPATH=/go_app
RUN go get github.com/msoap/shell2http \
    && go get github.com/Code-Hex/Neo-cowsay/cmd/cowsay \
    && go get github.com/Code-Hex/Neo-cowsay/cmd/cowthink \
    && mv /go_app/bin/shell2http /usr/local/bin/shell2http \
    && mv /go_app/bin/cowsay /usr/local/bin/neo-cowsay \
    && mv /go_app/bin/cowthink /usr/local/bin/neo-cowthink

# install endoh1 (http://www.ioccc.org/2012/endoh1/hint.html)
RUN chdir /tmp \
    && wget -r -l 1 http://www.ioccc.org/2012/endoh1/ \
    && chdir ./www.ioccc.org/2012/endoh1 \
    && make endoh1 \
    && make endoh1_color

# create archive with installed applications
RUN bash -c 'tar -czf /tmp/apps_dist.tar.gz /usr/local/bin/{shell2http,chkfont,figlet,figlist,showfigfonts,cowsay,cowthink,neo-cowsay,neo-cowthink} /usr/local/share/{figlet,cows,man} /tmp/www.ioccc.org/2012/endoh1/{endoh1,endoh1_color} /tmp/www.ioccc.org/2012/endoh1/*.c /tmp/www.ioccc.org/2012/endoh1/*.txt'

# -------
FROM debian:stretch

COPY --from=builder /tmp/apps_dist.tar.gz /tmp/

RUN chdir /usr \
    && tar -xzf /tmp/apps_dist.tar.gz --strip-components 1 \
    && rm /tmp/apps_dist.tar.gz \
    && mv /usr/www.ioccc.org/2012/endoh1/endoh1 /usr/www.ioccc.org/2012/endoh1/endoh1_color /usr/local/bin/ \
    && mkdir -p /usr/local/share/endoh1 \
    && mv /usr/www.ioccc.org/2012/endoh1/*.c /usr/www.ioccc.org/2012/endoh1/*.txt /usr/local/share/endoh1/ \
    && rm -rf /usr/www.ioccc.org/

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        man \
        less \
        perl \
    && rm -rf /var/lib/apt/lists/*

ADD server.sh /usr/local/bin/server.sh
CMD server.sh

EXPOSE 8080

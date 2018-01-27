# Docker image for ruby fpm gem

[![Docker Pulls](https://img.shields.io/docker/pulls/msoap/ruby-fpm.svg?maxAge=3600)](https://hub.docker.com/r/msoap/ruby-fpm) [![](https://images.microbadger.com/badges/image/msoap/ruby-fpm.svg)](https://microbadger.com/images/msoap/ruby-fpm)

## Install

    docker pull msoap/ruby-fpm

## Usage

    # create binary:
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o app_name
    
    # create deb-package with them:
	docker run --rm -v $PWD:/app -w /app msoap/ruby-fpm \
		fpm -s dir -t deb --force --name app_name -v 1.33 \
			--license="$(head -1 LICENSE)" \
			--url=https://app_name.io \
			--description="app description" \
			--maintainer="app maintainer" \
			--category=network \
			./app_name=/usr/bin/ \
			./app_name.1=/usr/share/man/man1/ \
			LICENSE=/usr/share/doc/app_name/copyright \
			README.md=/usr/share/doc/app_name/

## Links

  * [Gem source code](https://github.com/jordansissel/fpm)
  * [Source code](https://github.com/msoap/etc/blob/master/fpm-docker)

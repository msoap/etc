# Docker image for ruby ronn gem

[![Docker Pulls](https://img.shields.io/docker/pulls/msoap/ruby-ronn.svg?maxAge=3600)](https://hub.docker.com/r/msoap/ruby-ronn)
[![](https://images.microbadger.com/badges/image/msoap/ruby-ronn.svg)](https://microbadger.com/images/msoap/ruby-ronn)

## Install

    docker pull msoap/ruby-ronn

## Usage:

    # create man page from markdown file:
    docker run --rm -v $PWD:/app -w /app msoap/ruby-ronn ronn app_name.md

## Source code

[github.com/rtomayko/ronn](https://github.com/rtomayko/ronn)

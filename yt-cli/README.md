Command line client for Yandex Translate API
--------------------------------------------

Install
=======

    go get -u github.com/msoap/etc/yt-cli

Usage
=====

    export YT_KEY=***     # get it from https://translate.yandex.ru/developers/keys
    yt-cli "english text" # translate to russian
    yt-cli "russian text" # translate to english

Telegram bot
============

Build own bot Docker image:

    docker build -t yt-bot .
    docker run -d --rm --name yt-bot --env TB_TOKEN=$TB_TOKEN --env YT_KEY=$YT_KEY -v $PWD:/db yt-bot

Use exists image:

    # use current dir for save users DB (/db in container)
    # export TB_TOKEN=*** # get it from https://core.telegram.org/bots#6-botfather
    docker run -d --rm --name yt-bot --env TB_TOKEN=$TB_TOKEN --env YT_KEY=$YT_KEY -v $PWD:/db msoap/yt-bot

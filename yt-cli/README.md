Command line client for Yandex Translate API
--------------------------------------------

[![Docker Pulls](https://img.shields.io/docker/pulls/msoap/yt-bot.svg?maxAge=3600)](https://hub.docker.com/r/msoap/yt-bot) [![](https://images.microbadger.com/badges/image/msoap/yt-bot.svg)](https://microbadger.com/images/msoap/yt-bot)

Install
=======

    GO111MODULE=off go get -u github.com/msoap/etc/yt-cli
    
    # or from docker
    docker pull msoap/yt-bot

Usage
=====

    export YT_KEY=***     # get it from https://translate.yandex.ru/developers/keys
    yt-cli "english text" # translate to russian
    yt-cli "привет"       # translate to english
    echo some text for translate | yt-cli # translate STDIN

Telegram bot
============

Build and run own bot Docker image:

    docker build -t yt-bot .
    docker run -d --rm --name yt-bot --env TB_TOKEN=$TB_TOKEN --env YT_KEY=$YT_KEY -v $PWD:/db yt-bot

Or use exists Docker image:

    # use current dir for save users DB (/db in container)
    export TB_TOKEN=*** # get it from https://core.telegram.org/bots#6-botfather
    export YT_KEY=***   # see above
    docker run -d --rm --name yt-bot --env TB_TOKEN=$TB_TOKEN --env YT_KEY=$YT_KEY -v $PWD:/db msoap/yt-bot

Links
=====

  * [Source code](https://github.com/msoap/etc/tree/master/yt-cli)
  * [API](https://tech.yandex.ru/translate/)
  * [API statistic](https://translate.yandex.ru/developers/stat)

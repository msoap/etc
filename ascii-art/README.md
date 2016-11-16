```
    _    ____   ____ ___ ___              _         _
   / \  / ___| / ___|_ _|_ _|            / \   _ __| |_
  / _ \ \___ \| |    | | | |   _____    / _ \ | '__| __|
 / ___ \ ___) | |___ | | | |  |_____|  / ___ \| |  | |_
/_/   \_\____/ \____|___|___|         /_/   \_\_|   \__|
```

[![Docker Pulls](https://img.shields.io/docker/pulls/msoap/ascii-art.svg?maxAge=3600)](https://hub.docker.com/r/msoap/ascii-art/)

## Get:

    docker pull msoap/ascii-art

## Build image:

    rocker build

## cowsay:

    docker run --rm msoap/ascii-art cowsay 'Hello'
     _______
    < Hello >
     -------
            \   ^__^
             \  (oo)\_______
                (__)\       )\/\
                    ||----w |
                    ||     ||
                    
    # man:
    docker run -it --rm msoap/ascii-art man cowsay

## neo-cowsay:

    docker run -it --rm msoap/ascii-art neo-cowsay --rainbow Hello
    docker run -it --rm msoap/ascii-art neo-cowthink --aurora Hello

<img width="607" alt="screen shot 2016-11-16 at 10 40 32 pm" src="https://cloud.githubusercontent.com/assets/844117/20362773/ce109964-ac4d-11e6-96b0-b93bf798f17a.png">

## figlet:

    docker run --rm msoap/ascii-art figlet 'Hello'
     _   _      _ _
    | | | | ___| | | ___
    | |_| |/ _ \ | |/ _ \
    |  _  |  __/ | | (_) |
    |_| |_|\___|_|_|\___/

    # man:
    docker run -it --rm msoap/ascii-art man figlet

# http-server with cowsay and figlet:

    docker run -it --rm -p 8080:8080 msoap/ascii-art

# Links

  * [cowsay source](https://web.archive.org/web/20111224053105/http://www.nog.net/~tony/warez/cowsay.shtml)
  * [Neo-cowsay](https://github.com/Code-Hex/Neo-cowsay)
  * [figlet](http://www.figlet.org)

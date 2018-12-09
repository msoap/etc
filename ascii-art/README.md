```
    _    ____   ____ ___ ___              _         _
   / \  / ___| / ___|_ _|_ _|            / \   _ __| |_
  / _ \ \___ \| |    | | | |   _____    / _ \ | '__| __|
 / ___ \ ___) | |___ | | | |  |_____|  / ___ \| |  | |_
/_/   \_\____/ \____|___|___|         /_/   \_\_|   \__|
```

[![Docker Pulls](https://img.shields.io/docker/pulls/msoap/ascii-art.svg?maxAge=3600)](https://hub.docker.com/r/msoap/ascii-art/)
[![](https://images.microbadger.com/badges/image/msoap/ascii-art.svg)](https://microbadger.com/images/msoap/ascii-art)

## Get:

    docker pull msoap/ascii-art

## Build image:

    docker build -t ascii-art .

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

# endoh1

    # get files list
    docker run --rm -it msoap/ascii-art sh -c 'ls /usr/local/share/endoh1/*'
    # show ASCII fluid simulation
    docker run --rm -it msoap/ascii-art sh -c 'endoh1_color < /usr/local/share/endoh1/tanada.txt'

<img width="634" alt="screen shot 2016-11-27 at 12 36 42 am" src="https://cloud.githubusercontent.com/assets/844117/20644069/1e444536-b43a-11e6-8dc0-aa9f53cea03a.png">

# http-server with cowsay and figlet:

    docker run -it --rm -p 8080:8080 msoap/ascii-art

# Links

  * [cowsay source](https://web.archive.org/web/20111224053105/http://www.nog.net/~tony/warez/cowsay.shtml)
  * [Neo-cowsay](https://github.com/Code-Hex/Neo-cowsay)
  * [figlet](http://www.figlet.org)
  * [Most complex ASCII fluid](http://www.ioccc.org/2012/endoh1/hint.html) / [youtube](https://www.youtube.com/watch?v=QMYfkOtYYlg)

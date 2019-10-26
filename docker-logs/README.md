docker-log
----------

Utility for show docker logs from multiple containers in "follow" mode (like `docker-compose logs -f`).

Install
=======

    go get -u github.com/msoap/etc/docker-logs

Usage
=====

    docker-logs

### Features

  * [x] Show logs from all containers
  * [x] Color output
  * [x] Auto attach to new containers
  * [ ] Option for exclude some containers
  * [ ] Option for highlight STDERR logs

# Docker container for latest Go compiler (tip)

### Build:

    curl https://raw.githubusercontent.com/msoap/etc/master/golang-tip/Dockerfile > Dockerfile
    docker build -t golang:tip .

### Run:

    docker run --rm -v $PWD:/app -w /app golang:tip go version

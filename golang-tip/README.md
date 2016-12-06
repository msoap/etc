# Docker container for latest Go compiler (tip)

###Build:

    docker build -t golang:tip .

###Run:

    docker run --rm -v $PWD:/app -w /app golang:tip go version

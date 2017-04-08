#!/bin/bash

# Golang: build binaries for Linux/OS X/Windows x amd64/386/arm/arm64
#
# usage:
#   golang-cross-build.sh program_name [path]
#
#   golang-cross-build.sh program_name
#   golang-cross-build.sh program_name ./
#   golang-cross-build.sh program_name ./cmd/name
#
# Version get from last git tag
#
# For build *.deb need install fpm:
#   gem install --no-ri --no-rdoc fpm
#
# https://github.com/msoap/etc/tree/master/golang-cross-build/

# build_one_arch $GOOS $GOARCH
build_one_arch()
{
    export GOOS=$1
    export GOARCH=$2
    APP_NAME_EXE=$APP_NAME
    echo build: $GOOS/$GOARCH
    go get -d -t ./...

    if [ $GOOS == windows ]
    then
        APP_NAME_EXE=${APP_NAME}.exe
        go build -ldflags="-w -s" -o $APP_NAME_EXE $SRC_PATH
        zip_name="$APP_NAME-$VERSION.$GOOS.$GOARCH.zip"
        zip -9 $zip_name $APP_NAME_EXE README.md LICENSE
    else
        go build -ldflags="-w -s" -o $APP_NAME_EXE $SRC_PATH
        zip_name="$APP_NAME-$VERSION.$GOOS.$GOARCH.tar.gz"
        tar -czf $zip_name $APP_NAME_EXE README.md LICENSE $(ls $APP_NAME.1 2>/dev/null)

        # build deb package (need install fpm)
        if [[ $(which fpm) ]] && [[ $GOOS == linux ]] && [[ $GOARCH == amd64 ]]; then
            # or with docker: docker run -it --rm -v $PWD:/app -w /app ruby-fpm fpm ...
            fpm -s dir -t deb --force \
                --name "$APP_NAME" \
                -v "$VERSION" \
                --license="$(head -1 LICENSE)" \
                --maintainer="$APP_MAINTAINER" \
                ./$APP_NAME=/usr/bin/ \
                ./$APP_NAME.1=/usr/share/man/man1/ \
                LICENSE=/usr/share/doc/$APP_NAME/copyright \
                README.md=/usr/share/doc/$APP_NAME/ && \
            echo "$(ls *.deb) $(cat *.deb | shasum -a 256 | awk '{print $1}')" >> $APP_NAME.shasum
        fi
    fi

    echo "$zip_name/$APP_NAME_EXE $(cat $APP_NAME_EXE | shasum -a 256 | awk '{print $1}')" >> $APP_NAME.shasum
    rm $APP_NAME_EXE
}

VERSION=$(git tag --sort=version:refname | tail -1)
VERSION=${VERSION:-0.1}

APP_NAME=$1
APP_MAINTAINER=$(git show HEAD | awk '$1 == "Author:" {print $2 " " $3 " " $4}')

if [ -z $APP_NAME ]
then
    echo "Need name: $0 name"
    exit 1
fi

SRC_PATH=${2:-./}

> "$APP_NAME.shasum"

for GOOS in linux darwin windows
do
    for GOARCH in amd64 386
    do
        build_one_arch $GOOS $GOARCH
    done
done

# ARM
GOARM=6 build_one_arch linux arm
build_one_arch linux arm64

# SHA sums
cat "$APP_NAME.shasum"

# Homebrew sha256 of zips
echo
echo "Homebrew packages sha256 sums:"
shasum -a 256 *.darwin.*.tar.gz

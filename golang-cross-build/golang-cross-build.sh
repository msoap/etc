#!/bin/sh

# Golang: build binaries for Linux/OS X/Windows x amd64/386/arm
#
# usage:
#   golang-cross-build.sh program_name
#
# version get from last git tag
#
# https://github.com/msoap/etc/tree/master/golang-cross-build/

# build_one_arch $name $bin_name $GOOS $GOARCH
build_one_arch()
{
    name=$1
    bin_name=$2
    export GOOS=$3
    export GOARCH=$4
    echo build: $GOOS/$GOARCH
    go build -ldflags="-w" -o $2

    zip_name="$name-$VERSION.$GOOS.$GOARCH.zip"
    zip -9 $zip_name $bin_name README.md LICENSE

    echo "$zip_name/$bin_name $(cat $bin_name | shasum | awk '{print $1}')" >> $name.shasum
    rm $bin_name
}

VERSION=$(git tag 2>/dev/null | grep -E '^[0-9]+' | tail -1)
VERSION=${VERSION:-0.1}

name=$1
if [ -z $name ]
then
    echo "Need name: $0 name"
    exit 1
fi

> "$name.shasum"

for GOOS in linux darwin windows
do
    for GOARCH in amd64 386
    do
        if [ $GOOS == windows ]
        then
            bin_name="$name.exe"
        else
            bin_name=$name
        fi

        build_one_arch $name $bin_name $GOOS $GOARCH
    done
done

# ARM
GOARM=6 build_one_arch $name $name linux arm
build_one_arch $name $name linux arm64

# SHA sums
cat "$name.shasum"

# Homebrew sha256 of zips
echo
echo "Homebrew packages sha256 sums:"
shasum -a 256 *.darwin.*.zip

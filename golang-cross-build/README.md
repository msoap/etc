Go cross compile
----------------

## Usage:

      golang-cross-build.sh [program_name [path]]
    
      golang-cross-build.sh
      golang-cross-build.sh program_name ./
      golang-cross-build.sh program_name ./cmd/name

Version gets from last git tag

For build *.deb need install fpm: `gem install --no-ri --no-rdoc fpm`

## Setup build and auto-deploy to Github releases:

add to your `.travis.ci`

```yaml
env:
  global:
    - APP_NAME=app_binary_name
    - CURRENT_GO_VERSION="1.8"

go:
  - 1.8.x
  - 1.9.x
  - master

before_deploy:
  - curl -SL https://raw.githubusercontent.com/msoap/etc/master/golang-cross-build/golang-cross-build.sh > $GOPATH/bin/golang-cross-build.sh
  - chmod 700 $GOPATH/bin/golang-cross-build.sh
  - gem install --no-ri --no-rdoc fpm
  - golang-cross-build.sh $APP_NAME
  - ls -l *.zip *.tar.gz *.deb

deploy:
  provider: releases
  api_key:
    secure: xxxxxxx
  file_glob: "true"
  file:
    - "*.zip"
    - "*.tar.gz"
    - "*.deb"
  skip_cleanup: true
  on:
    tags: true
    branch: master
    condition: $TRAVIS_GO_VERSION =~ $CURRENT_GO_VERSION
    repo: github_user/$APP_NAME
```

get api_key for GH releases: `travis setup releases --force` (see [GitHub Releases setup](https://docs.travis-ci.com/user/deployment/releases/))

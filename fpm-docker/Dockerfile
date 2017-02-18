# docker build -t ruby-fpm .
FROM ruby:alpine

RUN apk add --no-cache build-base tar
RUN gem install --no-ri --no-rdoc fpm

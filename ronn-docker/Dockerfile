# docker build -t msoap/ruby-ronn .
FROM ruby:alpine

ENV RUBYOPT="-KU -E utf-8:utf-8"
RUN apk add --no-cache build-base groff
RUN gem install ronn

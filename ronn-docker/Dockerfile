FROM ruby:latest

ENV RUBYOPT="-KU -E utf-8:utf-8"
RUN apt-get update && apt-get install groff -y
RUN gem install ronn

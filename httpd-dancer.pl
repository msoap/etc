#!/usr/bin/perl

use warnings;
use strict;

use Dancer;

get '/' => sub {
    return "Hello World from perl with Dancer";
};

dance();

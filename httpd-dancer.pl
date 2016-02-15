#!/usr/bin/perl

use warnings;
use strict;

use Dancer;

get '/' => sub {
    return "Hello World from perl with Dancer/012345678901234567890123456789012345678901234567890123456789///////";
};

dance();

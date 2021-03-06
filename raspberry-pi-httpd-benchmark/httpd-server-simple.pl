#!/usr/bin/perl

use warnings;
use strict;

package SimpleWebServer;

use HTTP::Server::Simple::CGI;
use base qw(HTTP::Server::Simple::CGI);

sub handle_request {
    my $self = shift;
    my $cgi = shift;
    print "HTTP/1.0 200 OK\r\n";    
    print $cgi->header, "\n";
    print "Hello from perl with HTTP::Server::Simple//012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789";
}

package main;

SimpleWebServer->new(8080)->run();

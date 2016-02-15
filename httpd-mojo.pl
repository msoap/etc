#!/usr/bin/perl

use warnings;
use strict;

use Mojolicious::Lite;

get '/' => sub {
    my $self = shift;
    $self->render(text => "Hello world from perl with Mojolicious/0123456789012345678901234567890123456789//////");
};

app->secrets(['ZdeceVW4f&32S*3dF_21'])->start('daemon', listen_address => 'http://*:8080');

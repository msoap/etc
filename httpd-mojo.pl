#!/usr/bin/perl

use warnings;
use strict;

use Mojolicious::Lite;

get '/' => sub {
    my $self = shift;
    $self->render(text => "Hello world from Mojolicious");
};

app->secret('ZdeceVW4f&32S*3dF_21')->start('daemon', listen_address => 'http://*:8080');

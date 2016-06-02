#!/usr/bin/perl

=head1 description

Install dependencies:
    cpan YAML JSON

Usage:
    curl -s http://host/v1/api | ./api_json2swagger_yaml.pl

=cut

use warnings;
use strict;

use YAML;
use YAML::Node;
use JSON qw/from_json/;

# -----------------------------------------------------------------------------
sub main {
    my $json = join("", <STDIN>);
    my $api = from_json($json);
    my $swagger = to_swagger($api);
    print Dump($swagger);
}

# -----------------------------------------------------------------------------
sub to_swagger {
    my $api_item = shift;

    my $swagger;
    if (ref($api_item) eq 'HASH') {
        # object
        $swagger = {type => "object", properties => {}};
        for my $key (keys %$api_item) {
            $swagger->{properties}->{$key} = to_swagger($api_item->{$key});
        }
        $swagger = YAML::Node->new($swagger);
        ynode($swagger)->keys(['type', 'properties']);
    } elsif (ref($api_item) eq 'ARRAY') {
        # array (all items with one type)
        $swagger = {type => "array", items => to_swagger($api_item->[0])};
        $swagger = YAML::Node->new($swagger);
        ynode($swagger)->keys(['type', 'items']);
    } elsif (ref($api_item) eq '' && $api_item =~ /^-?\d+$/) {
        # integer
        $swagger = {type => "integer"};
    } elsif (ref($api_item) eq '' && $api_item =~ /^-?\d+(.\d+)?$/) {
        # float
        $swagger = {type => "float"};
    } elsif (ref($api_item) eq '') {
        # string
        $swagger = {type => "string"};
    }

    return $swagger;
}

# -----------------------------------------------------------------------------
main();

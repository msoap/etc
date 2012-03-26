package E;

=head1 DESCRIPTION

module for perl oneliners

=head1 SYNOPSIS

    perl -ME -e 'p [map {"$_$_"} qw/q w e/]'
    perl -ME -e 'j {%ENV}'
    perl -ME -e 'p {%INC}'
    perl -ME -e 'say encode_base64 read_file "E.pm"'

=cut

use strict;
use warnings;
use feature qw//;

use List::Util      qw/first max maxstr min minstr reduce shuffle sum/;
use List::MoreUtils qw/any all uniq first_index last_index apply/;
use POSIX           qw/ceil floor strftime/;
use MIME::Base64    qw/encode_base64 decode_base64/;
use JSON            qw/from_json to_json/;
use File::Slurp     qw/read_file write_file/;
use LWP::Simple     qw/get head getprint getstore/;
use Data::Dumper    qw/Dumper/;

sub import {
    my $class = shift;
    my $caller = caller;

    feature->import(':5.10', 'unicode_strings');
    $Data::Dumper::Sortkeys = 1;

    no strict 'refs';
    my @func_export = qw/
        first max maxstr min minstr reduce shuffle sum
        any all uniq first_index last_index apply
        ceil floor strftime
        encode_base64 decode_base64
        from_json to_json
        read_file write_file
        get head getprint getstore
        Dumper
    /;

    for my $func (@func_export) {
        *{$caller . "::$func"} = \&$func;
    }

    *{$caller . "::p"} = sub {
        print Dumper(1 == scalar @_ ? $_[0] : \@_);
    };

    *{$caller . "::j"} = sub {
        print to_json(1 == scalar @_ ? ref($_[0]) ? $_[0] : [$_[0]] : \@_, {pretty => 1}) . "\n";
    };

    *{$caller . "::help"} = sub {
        print "functions:\n" . join("\n", map {"  $_"} sort @func_export, qw/help p j/) . "\n";
    };
}

1;

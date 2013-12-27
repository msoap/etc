#!/usr/bin/perl

use warnings;
use strict;

use Data::Dumper;
use utf8;
use open ":std" => ":utf8";

our $CHART_SCALE = 2.5;

# --------------------------------------------------------------------
sub main {
    system("go build test_memory.go");

    my $stat = {};
    for my $type (qw/map_int_key_of_string
                     map_string_key_of_string
                     map_int_key_of_int
                     map_string_key_of_int
                     slice_of_string
                     slice_of_int/
                 )
    {
        for my $nums (qw/500 700 1000 1500/) {
            for my $lang (qw/perl go/) {
                my $exe = $lang eq 'perl'
                   ? 'perl ./test_memory.pl'
                   : $lang eq 'go'
                     ? './test_memory'
                     : die;

                print "$type / $nums / $lang:";
                my $memory = qx/$exe $type $nums | awk '\$1 == "memory:" {print \$2}'/ + 0;
                print " $memory MB\n";

                $stat->{$type}->{$nums}->{$lang} = $memory;
            }
        }
    }
    print "----------------------\n";

    for my $type (sort keys %$stat) {
        printf "%s:\n", {map_int_key_of_string => '{int}->{int} = string'
                         , map_string_key_of_string => '{string}->{string} = string'
                         , map_string_key_of_int => '{string}->{string} = int'
                         , map_int_key_of_int => '{int}->{int} = int'
                         , slice_of_string => '[int]->[int] = string'
                         , slice_of_int => '[int]->[int] = int'
                        }->{$type};
        print "  lang (keys² q-ty): MB of memory\n";
        for my $nums (sort {$b <=> $a} keys %{$stat->{$type}}) {
            printf "  Go   (%4i²): %7s %s\n", $nums, $stat->{$type}->{$nums}->{go},   chr(9632) x int($stat->{$type}->{$nums}->{go}   / $CHART_SCALE);
            printf "  Perl (%4i²): %7s %s\n", $nums, $stat->{$type}->{$nums}->{perl}, chr(9632) x int($stat->{$type}->{$nums}->{perl} / $CHART_SCALE);
            print "\n";
        }
        print "\n";
    }
}

# --------------------------------------------------------------------
main();

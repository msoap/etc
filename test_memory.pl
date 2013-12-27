#!/usr/bin/perl

use warnings;
use strict;

use Data::Dumper;

our $NUMS = 1500;

# --------------------------------------------------------------------
sub test_map_string_key_of_string {
    my $name = shift;
    my $nums = shift;
    print "Start $name\n";

    my $hash = {};

    for (my $i = 0; $i < $nums; $i++) {
        $hash->{"key_" . $i} = {};

        for (my $j = 0; $j < $nums; $j++) {
            $hash->{"key_" . $i}->{"key_j_" . $j} = "string_${i}_${j}";
        }
    }
    print "Finish\n";
}

# --------------------------------------------------------------------
sub test_map_int_key_of_string {
    my $name = shift;
    my $nums = shift;
    print "Start $name\n";

    my $hash = {};

    for (my $i = 0; $i < $nums; $i++) {
        $hash->{$i} = {};

        for (my $j = 0; $j < $nums; $j++) {
            $hash->{$i}->{$j} = "string_${i}_${j}";
        }
    }
    print "Finish\n";
}

# --------------------------------------------------------------------
sub test_slice_of_string {
    my $name = shift;
    my $nums = shift;
    print "Start $name\n";

    my $array = [];
    for (my $i = 0; $i < $nums; $i++) {
        for (my $j = 0; $j < $nums; $j++) {
            $array->[$i]->[$j] = "string_${i}_${j}";
        }
    }

    print "Finish\n";
}

# --------------------------------------------------------------------
sub test_slice_of_int {
    my $name = shift;
    my $nums = shift;
    print "Start $name\n";

    my $array = [];
    for (my $i = 0; $i < $nums; $i++) {
        for (my $j = 0; $j < $nums; $j++) {
            $array->[$i]->[$j] = $i * $nums + $j;
        }
    }

    print "Finish\n";
}

# --------------------------------------------------------------------
sub main {
    my $nums = $ARGV[1] && $ARGV[1] =~ /^\d+$/ && $ARGV[1] > 0 ? $ARGV[1] : $NUMS;
    print "nums: $nums\n";

    my $test_functions = {
        map_int_key_of_string => \&test_map_int_key_of_string
        , map_string_key_of_string => \&test_map_string_key_of_string
        , slice_of_string => \&test_slice_of_string
        , slice_of_int => \&test_slice_of_int
    };

    if ($ARGV[0]
        && $ARGV[0]
        && defined $test_functions->{$ARGV[0]}
       )
    {
        $test_functions->{$ARGV[0]}->($ARGV[0], $nums);
    } else {
        printf("usage: $0 %s [NUMS]\n", join("|", keys %$test_functions));
        return;
    }

    my $memory_kb = qx/ps aux | awk '\$2 == $$ {print \$6}'/;
    printf("memory: %.2f MB\n", $memory_kb / 1024);
}

# --------------------------------------------------------------------
main();

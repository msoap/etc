#!/usr/bin/perl

=head1 NAME

metrika-show-common-stat.pl - show common statistic from Yandex.metrika for last 7 days

use API from: http://api.yandex.com/metrika/

=head1 SYNOPSIS

    metrika-show-common-stat.pl options
      --token=XXXXX       -- OAuth totken, required
      --counter-id=NNNNN  -- counter ID, required
      --days=N            -- show stat for last N days (default is 7 days)
      --short             -- show short stat, only visits and visitors

    or read config from ~/.config/metrika-show-common-stat.cfg

=cut

use strict;
use warnings;
use v5.10.0;

use Getopt::Long;
use LWP::Simple;
use POSIX qw/strftime/;
use JSON;
use Term::ANSIColor;

use utf8;
use open ":std" => ":utf8";

our $BASE_URL = "https://api-metrika.yandex.com/%s.json?id=%s&pretty=1&oauth_token=%s&date1=%s&date2=%s";
our $CONFIG_FILENAME = "$ENV{HOME}/.config/metrika-show-common-stat.cfg";
our $DEFAULT_DAYS = 7;

# ------------------------------------------------------------------------------
sub main {
    my $config = {days => $DEFAULT_DAYS};

    if (-f $CONFIG_FILENAME && -r $CONFIG_FILENAME) {
        open my $FH, '<', $CONFIG_FILENAME or die "Error open file: $!\n";
        for my $row (<$FH>) {
            next if $row =~ /^\s*#/;
            chomp $row;
            my ($k, $v) = split /\s*=\s*/, $row, 2;
            $config->{$k} = $v;
        }
        close $FH;
    }

    GetOptions(
        'token=s' => \$config->{token},
        'counter-id=i' => \$config->{'counter-id'},
        'days=i' => \$config->{days},
        'short' => \$config->{short},
        'help' => sub {
            print "$0 --token=XXXXX --counter-id=NNNNN [--days=N --short --help]\n";
            exit 0;
        }
    );

    die("need --token and --counter-id options\n") unless defined $config->{token} && defined $config->{'counter-id'};

    for my $day (reverse 0 .. $config->{days}) {
        my $date_param = strftime("%Y%m%d", localtime(time() - $day * 3600 * 24));
        my $date = strftime("%Y-%m-%d (%a)", localtime(time() - $day * 3600 * 24));

        my $get_stat = sub($) {
            my $type = shift;
            return get_stat($type, $date_param, $config->{'counter-id'}, $config->{token});
        };

        print colored($date, "yellow") . "\n";

        my $stat = $get_stat->('stat/traffic/summary');
        if (exists $stat->{errors} && $stat->{errors}->[0]->{code}) {
            printf "  -- %s --\n",
                   {'ERR_NO_DATA' => "no visits",
                    'ERR_NO_NET'  => "network priblem",
                    'ERR_JSON'    => "json parsing error",
                   }->{ $stat->{errors}->[0]->{code} }
                   || "error: " . $stat->{errors}->[0]->{code};
            next;
        }

        my ($new_visitors, $visitors, $visits) = (map {$stat->{totals}->{$_}} qw/new_visitors visitors visits/);
        print "  ";
        print format_val('visitors' => $visitors) . ' ';
        print format_val('new visitors' => $new_visitors) . ' ';
        print format_val('visits' => $visits);

        if ($config->{short}) {
            print "\n";
            next;
        }

        # by hours
        my $stat_hourly = $get_stat->('stat/traffic/hourly');
        my $hours = join(', ', map {$_->{hours}} grep {$_->{avg_visits}} @{$stat_hourly->{data}});
        print " " . format_val('hours' => $hours) . "\n";

        # by sources
        my $stat_sources = $get_stat->('stat/sources/summary');
        print "  " . format_val('where from' => join(', ', map {"$_->{name}: " . colored($_->{visits}, 'bold blue')} @{$stat_sources->{data}}));

        # by geo
        my $stat_geo = $get_stat->('stat/geo');
        print " " . format_val('geo' => join(', ', map {"$_->{name}: " . colored($_->{visits}, 'bold blue')} @{$stat_geo->{data}})) . "\n";

        # by os
        my $stat_os = $get_stat->('stat/tech/os');
        print "  " . format_val('OS' => join(', ', map {"$_->{name}: " . colored($_->{visits}, 'bold blue')} @{$stat_os->{data}}));

        # by browsers
        my $stat_browsers = $get_stat->('stat/tech/browsers');
        print " " . format_val('browsers' => join(', ', map {"$_->{name}/$_->{version}: " . colored($_->{visits}, 'bold blue')} @{$stat_browsers->{data}}));

        # by display
        my $stat_display = $get_stat->('stat/tech/display');
        print " " . format_val('screen' => join(', ', map {"$_->{name}: " . colored($_->{visits}, 'bold blue')} @{$stat_display->{data}})) . "\n";
    }
}

# ------------------------------------------------------------------------------
sub get_stat($$$$) {
    my ($type, $date, $counter_id, $token) = @_;

    my $url = sprintf($BASE_URL, $type, $counter_id, $token, $date, $date);
    my $json = get($url);
    return {errors => [{code => 'ERR_NO_NET'}]} unless $json;

    my $stat = eval {from_json($json)};
    return {errors => [{code => 'ERR_JSON'}]} unless ref($stat) eq 'HASH';

    return from_json($json);
}

# ------------------------------------------------------------------------------
sub format_val($$;$) {
    my ($title, $value, $color) = @_;

    $color //= 'green';
    return colored("$title: ", 'green') . $value;
}

# ------------------------------------------------------------------------------
main();

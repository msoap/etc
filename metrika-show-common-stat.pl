#!/usr/bin/perl

=head1 NAME

metrika-show-common-stat.pl - show common statistic from Yandex.metrika for last 7 days

use API from: http://api.yandex.com/metrika/

=head1 SYNOPSIS

    metrika-show-common-stat.pl options
      --token=XXXXX       -- OAuth totken, required
      --counter-id=NNNNN  -- counter ID, required
      --days=N            -- show stat for last N days
      --short             -- show short stat, only visits and visitors

=cut

use strict;
use warnings;

use Getopt::Long;
use LWP::Simple;
use POSIX qw/strftime/;
use JSON;
use Term::ANSIColor;

use utf8;
use open ":std" => ":utf8";

our $base_url = "https://api-metrika.yandex.com/%s.json?id=%s&pretty=1&oauth_token=%s&date1=%s&date2=%s";

# ------------------------------------------------------------------------------
sub main {
    my ($token, $counter_id, $days_ago, $is_short);
    GetOptions(
        'token=s' => \$token,
        'counter-id=i' => \$counter_id,
        'days=i' => \$days_ago,
        'short' => \$is_short,
        'help' => sub {
            print "$0 --token=XXXXX --counter-id=NNNNN [--days=N --short --help]\n";
            exit 0;
        }
    );

    $days_ago //= 7;
    die("need --token and --counter-id options\n") unless defined $token && defined $counter_id;

    for my $day (reverse 0 .. $days_ago) {
        my $date_param = strftime("%Y%m%d", localtime(time() - $day * 3600 * 24));
        my $date = strftime("%Y-%m-%d (%a)", localtime(time() - $day * 3600 * 24));

        my $get_stat = sub($) {
            my $type = shift;
            return get_stat($type, $date_param, $counter_id, $token);
        };

        print colored($date, "yellow") . "\n";

        my $stat = $get_stat->('stat/traffic/summary');
        if (exists $stat->{errors} && $stat->{errors}->[0]->{code} eq 'ERR_NO_DATA') {
            print "  -- no data --\n";
            next;
        }

        my ($new_visitors, $visitors, $visits) = (map {$stat->{totals}->{$_}} qw/new_visitors visitors visits/);
        print "  ";
        print format_val('visitors' => $visitors) . ' ';
        print format_val('new visitors' => $new_visitors) . ' ';
        print format_val('visits' => $visits);

        if ($is_short) {
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

    my $url = sprintf($base_url, $type, $counter_id, $token, $date, $date);
    my $json = get($url);
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

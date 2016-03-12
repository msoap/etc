#!/usr/bin/perl

=head1 DESCRIPTION

https://gist.github.com/msoap/4398036

Download all photos from fotki.yandex.ru album.

Usage:
    yandex-fotki-get-album-fotos.pl http://fotki.yandex.ru/users/user_name/album/12345678/

Install:
    git clone https://gist.github.com/4398036.git
    cp 4398036/yandex-fotki-get-album-fotos.pl ~/bin/yandex-fotki-get-album-fotos.pl
    chmod 744 ~/bin/yandex-fotki-get-album-fotos.pl

=cut

use warnings;
use strict;

use XML::Simple;
use Data::Dumper;

BEGIN {
    $ENV{PERL_LWP_SSL_VERIFY_HOSTNAME} = 0;
    eval "use LWP::Simple";
}

# --------------------------------------------------------------------
sub main {
    my $album_url = $ARGV[0] || die "use: $0 album_url\n";
    my $rss_url = "$album_url/rss2";
    $rss_url =~ s[//rss2$][/rss2];

    my $urls = get_urls_from_rss($rss_url);

    my $i = 1;
    for my $url (@$urls) {
        my $file_name = sprintf "img_%03i.jpg", $i++;
        print "$url -> $file_name\n";
        mirror($url, $file_name);
    }
}

# --------------------------------------------------------------------
sub get_urls_from_rss {
    my $rss_url = shift;

    my $xml = get($rss_url);
    my $rss = XMLin($xml, ForceArray => 1, ForceContent => 1);
    my $urls = [];

    for my $item (@{$rss->{channel}->[0]->{item}}) {
        my $url = $item->{"media:content"}->[0]->{url};
        $url =~ s/_[a-z]+$/_orig/i;
        push @$urls, $url;
    }

    my $next = $rss->{channel}->[0]->{"atom:link"}
                   && $rss->{channel}->[0]->{"atom:link"}->[0]
                   && $rss->{channel}->[0]->{"atom:link"}->[0]->{rel}
                   && $rss->{channel}->[0]->{"atom:link"}->[0]->{rel} eq 'next'
               ? $rss->{channel}->[0]->{"atom:link"}->[0]->{href}
               : undef;

    if ($next) {
        push @$urls, @{get_urls_from_rss($next)};
    }

    return $urls;
}

# --------------------------------------------------------------------
main();

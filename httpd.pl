#!/usr/bin/perl

use warnings;
use strict;

use HTTP::Daemon;
use HTTP::Status;

# --------------------------------------------------------------------
sub main {
    my $httpd = HTTP::Daemon->new(LocalPort => 8080) || die;
    print "URL: ", $httpd->url, "\n";

    while (my $c = $httpd->accept) {
        while (my $r = $c->get_request) {
            if ($r->method eq 'GET' and $r->uri->path eq "/") {
                my $res = HTTP::Response->new(RC_OK);
                $res->content("Hello World from perl with HTTP::Daemon///01234567890123456789012345678901234567890123456789\n");
                $res->content_type('text/plain');
                $c->send_response($res);
            } else {
                $c->send_error(RC_FORBIDDEN)
            }
        }
        $c->close;
    }
}

# --------------------------------------------------------------------
main();

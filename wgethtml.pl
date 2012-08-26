#!/usr/bin/perl

#################################################
#
#  load html with css/js/img and embed it
#
#################################################

use warnings;
use strict;

our $VERSION = 0.80;

use LWP::UserAgent;
use URI::WithBase;
use MIME::Base64 qw(encode_base64);
use POSIX qw(strftime);
use Image::ExifTool;
use Getopt::Std;

our $SAVE_IMG_AS_EXT_FILES_GREATER_THAN = 1_000_000;
our $TIMEOUT = 30;
our $UA="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_2) AppleWebKit/534.52.7 (KHTML, like Gecko) Version/5.1.2 Safari/534.52.7";

# for download JS too : create symlink wgethtml.pl -> wgethtml-js.pl
# or use wgethtml.pl -j
# for append to one html file - use: $0 -a one_file.html
# custom sleep (default = 30 sec): -s sleep_in_seconds
our %options;
getopts('hvwja:s:u:', \%options);

if ($options{h}) {
    print "wgethtml.pl - download html page with all img/css/js and embed all img/css/js into html.\n"
        . "version: $VERSION\n"
        . "usage:\n"
        . "  wgethtml.pl [options] urls_of_html\n"
        . "    -j             download all js (or use symlink wgethtml-js.pl of wgethtml.pl)\n"
        . "    -a file.html   download all urls into one html file\n"
        . "    -s seconds     sleep between each url, default = 30 sec.\n"
        . "    -u user_agent  set custom user agent, default: safari\n"
        . "    -w             print warnings to STDOUT too (default print to ~/.wgethtml/error.log)\n"
        . "    -h             help\n"
        . "    -v             print version\n";
    exit 0;
}

if ($options{v}) {
    print "$VERSION\n";
    exit 0;
}

our $download_js = ($0 =~ /-js/ || $options{j} ? 1 : 0);
$options{s} = (defined $options{s} && $options{s} =~ /^\d+$/) ? $options{s} : 30;
$UA = $options{u} if $options{u};

#................................................
{
    my %www_cache;
    my @black_list = (
        qr[\Qtns-counter.ru]i
      , qr[\Qpink.futurico.ru]i
      , qr[\Qpink.habralab.ru]i
      , qr[\Qhabr.user.madbanner.ru]i
      , qr[\Qgoogle-analytics.com]i
      , qr[\Qcounter.rambler.ru]i
      , qr[\Qspylog.com]i
      , qr[\Qbs.yandex.ru]i
      , qr[\Qmc.yandex.ru]i
      , qr[\Qtop.list.ru]i
      , qr[\Qhotlog.ru]i
    );

sub lwp_load
{
    my $url = shift;
    my $base_url = shift;

    my $abs_url = URI::WithBase->new($url, $base_url)->abs();

    return $www_cache{$abs_url} if $www_cache{$abs_url};

    for my $qr (@black_list) {
        if ($abs_url =~ $qr) {
            $www_cache{$abs_url} = {code => 404
                                  , message => 'url blocked'
                                  , abs_url => 'fakescheme://localhost/' . $abs_url
                                  , content_type => undef};

            return $www_cache{$abs_url};
        }
    }

    my $response = LWP::UserAgent
                   ->new(agent => $UA, timeout => $TIMEOUT)
                   ->request(HTTP::Request->new("GET", $abs_url));

    if ($response->code != 200) {
        $www_cache{$abs_url} = {code => $response->code, message => $response->message, abs_url => $abs_url};
    } else {
        $www_cache{$abs_url} = {code => $response->code
                              , abs_url => $abs_url
                              , content => $response->content()
                              , content_type => ($response->header('Content-Type') || undef)};
    }

    return $www_cache{$abs_url};
}}

#................................................
sub load_and_replace_css
{
    my $html_url = shift;
    my $tag = shift;

    # <link rel="stylesheet" type="text/css" href="http://localhost/wacko/themes/default/css/wakka.css" />

    if ($tag =~ /rel  \s* = \s* ["']* stylesheet/six &&
        #$tag =~ /type \s* = \s* ["']* text\/css/six &&
        $tag =~ /href \s* = \s* ["']* ([^\s"']+) ["']*/six
       )
    {
        my $css_url = $1;
        my $res = lwp_load($css_url, $html_url);
        my $abs_css_url = $res->{abs_url};
        my $css;

        if ($res->{code} == 200) {
            $css = $res->{content};
            $css =~ s/url\((["']*)(.+?)\1\)/'url(' . get_img_data_url($2, $abs_css_url) . ')'/sige;
        } else {
            error_log("LWP error for $abs_css_url: $res->{message}");
            return $tag;
        }

        my $media = '';
        if ($tag =~ /media \s* = \s* ["']* ([^\s"']+) ["']*/six) {
            $media = qq/ media="$1"/;
        }

        return <<CSS;
    <style$media>
      /* css: $abs_css_url */
      $css
    </style>
CSS

    } elsif ($tag =~ /rel  \s* = \s* ["']* shortcut \s+ icon/six &&
             $tag =~ /href \s* = \s* ["']* ([^\s"']+) ["']*/six
            )
    {
        # <link rel="shortcut icon" href="http://habrahabr.ru/favicon.ico">
        my $ico_url = get_img_data_url($1, $html_url);
        return qq|<link rel="shortcut icon" href="$ico_url">|;

    } else {
        return $tag;
    }
}

#................................................
sub load_and_replace_img
{
    my $html_url = shift;
    my $tag = shift;

    if ($tag =~ /src \s* = \s* (["']*) ([^\s"']+) ["']*/six) {

        # <img src="http://habrahabr.ru/dd.png">
        my $quote = $1;
        my $img_url = get_img_data_url($2, $html_url);
        $tag =~ s/src \s* = \s* ["']* [^\s"']+ ["']*/src=${quote}${img_url}${quote}/six;
    }

    return $tag;
}

#................................................
sub load_and_replace_linked_img
{
    my $html_url = shift;
    my $tag = shift;

    if ($tag =~ /href \s* = \s* (["']*) ([^\s"']+) ["']*/six) {

        # <a href="http://habrahabr.ru/dd.png">
        my $quote = $1;
        my $img_href = $2;

        if ($img_href =~ /\.(jpg|jpeg|gif|png|bmp)$/i) {
            my $save_img_size = $SAVE_IMG_AS_EXT_FILES_GREATER_THAN;
            $SAVE_IMG_AS_EXT_FILES_GREATER_THAN = 500_000;
            my $img_url = get_img_data_url($img_href, $html_url);
            $tag =~ s/href \s* = \s* ["']* [^\s"']+ ["']*/href=${quote}${img_url}${quote}/six;
            $SAVE_IMG_AS_EXT_FILES_GREATER_THAN = $save_img_size;
        }
    }

    return $tag;
}

#................................................
sub load_and_replace_js
{
    my $html_url = shift;
    my $tag = shift;

    # <script src="http://habrahabr.ru/js/prototype.js" type="text/javascript"></script>

    if ($tag =~ /src \s* = \s* ["']* ([^\s"']+) ["']*/six) {
        my $js_url = $1;

        if (! $download_js) {
            return qq|<script language="FakeLanguage" src="$js_url" type="text/fakescript">|; # fake lang
        }

        my $res = lwp_load($js_url, $html_url);
        my $abs_js_url = $res->{abs_url};
        my $js;
        if ($res->{code} == 200) {

            $js = $res->{content};
            $js =~ s!</script!</sc" + "ript!igs;

        } else {
            error_log("LWP error for $abs_js_url: $res->{message}");
            return $tag;
        }

        return <<JS;
    <script type="text/javascript">
      // js: $abs_js_url
      $js
    </script>
JS
    } else {
        if (! $download_js) {
            return qq|<script language="FakeLanguage" type="text/fakescript">|; # fake lang
        }
        return $tag;
    }
}

#................................................
sub load_and_replace_css_import
{
    my $html_url = shift;
    my $tag = shift;

    my $added_ext_css = '';

    while ($tag =~ /\@import \s+ (?:url\()? \s* (\S+)/sigx) {
        my $css_url = $1;
        $css_url =~ s/^["']+//i;
        $css_url =~ s/["';\)]+$//i;

        my $res = lwp_load($css_url, $html_url);
        my $abs_css_url = $res->{abs_url};
        my $css;

        if ($res->{code} == 200) {
            $css = $res->{content};
            $css =~ s/url\((["']*)(.+?)\1\)/'url(' . get_img_data_url($2, $abs_css_url) . ')'/sige;
            $added_ext_css .= "\n<style>\n/* css: $abs_css_url */\n$css\n</style>\n";
        } else {
            error_log("LWP error for $abs_css_url: $res->{message}");
            next;
        }
    }

    # resolve images
    $tag = $tag.$added_ext_css;

    return $tag;
}

#................................................
sub get_img_data_url
{
    my $url = shift;
    my $base_url = shift;

    $url =~ s/\s+//sg;
    return $url if $url =~ /^data:/;

    my $img = lwp_load($url, $base_url);
    #print "  get $url";

    my $content = $img->{content};
    if ($content) {

        my $type = $img->{content_type};
        unless ($type) {
            my $exifTool = new Image::ExifTool;
            $exifTool->ExtractInfo(\$content);
            $type = $exifTool->GetValue("MIMEType");

            $type = 'image/x-icon' if (! $type && $url =~ /\.ico$/i);
            $type = 'image/jpeg' if ! $type;
        }

        if (length($content) > $SAVE_IMG_AS_EXT_FILES_GREATER_THAN) {
            my $ext = {'image/jpeg' => 'jpg', 'image/gif' => 'gif', 'image/png' => 'png'}->{$type} || 'jpg';
            mkdir "_ext_images" unless -d "_ext_images";
            my $image_num = 0;
            $image_num++ while -e "_ext_images/img-".strftime("%Y%m%d-%H%M%S", localtime(time))."-$image_num.$ext";
            my $img_local_file_name = "_ext_images/img-".strftime("%Y%m%d-%H%M%S", localtime(time))."-$image_num.$ext";
            open FIMG, ">", $img_local_file_name or die "Error: $!";
            binmode FIMG;
            print FIMG $content;
            close FIMG;
            #print " saved in $img_local_file_name\n";
            return $img_local_file_name;
        }

        unless ($type) {
            error_log("error: not find mime type: $img->{abs_url}");
            return $img->{abs_url};
        }

        my $base64_data = encode_base64($content);
        $base64_data =~ s/\s+//g;

        #print " - ok\n";
        return "data:$type;base64,$base64_data";
    } else {
        error_log("error: not download image ($img->{abs_url})");
        return $img->{abs_url};
    }
}

#................................................
sub update_content_type
{
    my $html = $_[0];
    my $content_type = $_[1] or return;

    # "<meta http-equiv=\"content-type\" content=\"$content_type\">\n"
    if ($$html =~ m/(<meta[^>]+>)/is) {
        my $meta = $1;
        if ($meta =~ m/http-equiv \s* = \s* ['"]* content-type/isx) {
            $$html =~ s/\Q$meta\E/<meta http-equiv="content-type" content="$content_type">/is;
            return;
        }
    }

    # else
    $$html =~ s/(<head[^>]*>)/$1 <meta http-equiv="content-type" content="$content_type">/is
    or
    $$html = qq/<meta http-equiv="content-type" content="$content_type">\n/ . $$html;
}

#................................................
sub error_log
{
    my $msg = shift;
    my $print_stderr = shift;

    unless (-d "$ENV{HOME}/.wgethtml") {
        mkdir "$ENV{HOME}/.wgethtml";
    }

    my $FD;
    open $FD, ">>", "$ENV{HOME}/.wgethtml/error.log" or return;
    print $FD strftime("%Y-%m-%d\t%H:%M:%S", localtime(time)) . "\t$msg\n";
    close $FD;

    if ($options{w} || $print_stderr) {
        print STDERR "$msg\n";
    }
}

#................................................
sub load_inline_img_js_css_and_embed_into_html_file
{
    my $html_url = $_[0];
    my $html = $_[1];

    $$html =~ s/(<link[^>]+>)/load_and_replace_css($html_url, $1)/sige;

    # <style type="text/css"> @import "http://www.terralab.ru/bitrix/php_interface/tl/styles.css"; </style>
    # <style type="text/css"> @import url("http://www.terralab.ru/bitrix/php_interface/tl/styles.css") print; </style>
    $$html =~ s/(<style.+?<\/style>)/load_and_replace_css_import($html_url, $1)/sige;

    # <img src="http://habrahabr.ru/dd.png">
    $$html =~ s/(<img[^>]+>)/load_and_replace_img($html_url, $1)/sige;

    # <a href="http://habrahabr.ru/dd.png">
    $$html =~ s/(<a\s+[^>]+>)/load_and_replace_linked_img($html_url, $1)/sige;

    # <script src="http://habrahabr.ru/js/prototype.js" type="text/javascript"></script>
    $$html =~ s/(<script[^>]*>.*?(<\/\s*script>)?)/load_and_replace_js($html_url, $1)/sige;
}

#................................................
sub main
{
    my $ii = 0;
    for my $url (@ARGV) {

        # between urls
        if ($ii++ && scalar(@ARGV) > 1) {
            sleep $options{s};
        }

        my ($domain) = $url =~ m/^\w+:\/\/([^\/:]+)/;
        unless ($domain) {
            error_log("url is not valid: $url", 'print_stderr');
            next;
        }

        my $lwp = lwp_load($url, $url);
        unless ($lwp->{content}) {
            error_log("dont load $url (code: $lwp->{code} / message: $lwp->{message})", 'print_stderr');
            next;
        }

        print STDERR "$url\n";
        error_log("------------------ $url");

        my $html = "<!-- url: $url -->\n"
                 . "<!-- date: " . strftime("%d.%m.%Y %T", localtime(time)) . " -->\n"
                 . $lwp->{content};

        update_content_type(\$html, $lwp->{content_type});
        load_inline_img_js_css_and_embed_into_html_file($url, \$html);

        my ($file_name, $file_out_mode);

        if ($options{a}) {
            # append all urls to one html file
            $file_name = $options{a};
            $file_out_mode = ">>";
        } else {
            $file_name = $domain . '-' . strftime("%Y%m%d-%H%M%S", localtime(time)) . '.html';
            my $i = 0;
            while (-e $file_name) {
                $file_name = $domain . '-' . strftime("%Y%m%d-%H%M%S", localtime(time)) . "-$i" . '.html';
                $i++;
            }
            $file_out_mode = ">";
        }

        open F, $file_out_mode, $file_name or die "Error: $!";
        print F $html;
        close F;
    }
}

#................................................
main();

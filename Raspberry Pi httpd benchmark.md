Raspberry Pi httpd benchmark
============================

Total RPS
---------

All:

                  Go/fasthttp (10 thr): 937.8 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                        Nginx (10 thr): 826.1 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                        Nginx ( 2 thr): 647.3 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                           Go (10 thr): 642.5 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                  Go/fasthttp ( 2 thr): 544.6 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                           Go ( 2 thr): 392.2 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                      Node.js ( 2 thr): 199.8 ■■■■■■■■■■■■■■■■■■■
                      Node.js (10 thr): 195.1 ■■■■■■■■■■■■■■■■■■■
                       Python ( 2 thr): 122.9 ■■■■■■■■■■■■
                       Python (10 thr):  95.5 ■■■■■■■■■
            Perl/HTTP::Daemon ( 1 thr):  90.4 ■■■■■■■■■
    Perl/HTTP::Server::Simple ( 2 thr):  79.2 ■■■■■■■
                  Perl/Dancer ( 2 thr):  31.0 ■■■
             Perl/Mojolicious ( 2 thr):  28.9 ■■

Run in 10 threads:

               Go/fasthttp (10 thr): 937.8 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                     Nginx (10 thr): 826.1 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                        Go (10 thr): 642.5 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                   Node.js (10 thr): 195.1 ■■■■■■■■■■■■■■■■■■■
                    Python (10 thr):  95.5 ■■■■■■■■■

Run in 2 threads:

                        Nginx ( 2 thr): 647.3 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                  Go/fasthttp ( 2 thr): 544.6 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                           Go ( 2 thr): 392.2 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                      Node.js ( 2 thr): 199.8 ■■■■■■■■■■■■■■■■■■■
                       Python ( 2 thr): 122.9 ■■■■■■■■■■■■
    Perl/HTTP::Server::Simple ( 2 thr):  79.2 ■■■■■■■
                  Perl/Dancer ( 2 thr):  31.0 ■■■
             Perl/Mojolicious ( 2 thr):  28.9 ■■

####Create chart:

For all, 10 and 2 threads:

    cat Raspberry*.md | perl -nlE 'use open ":std" => ":utf8"; BEGIN{my %stat, $lang, $thr} $lang = $1 if /^##(\w.+)$/; $thr = $1 if /^\s+(\d+)\s+threads/; if (/Requests\/sec:\s+(\d+\.\d+)/) {$stat{$lang . sprintf(" (%2d thr)", $thr)} = $1} END {for my $lang (sort {$stat{$b} <=> $stat{$a}} keys %stat) {printf "%35s: %5.1f %s\n", $lang, $stat{$lang}, chr(9632) x int($stat{$lang} / 10)}}'
    cat Raspberry*.md | perl -nlE 'use open ":std" => ":utf8"; BEGIN{my %stat, $lang, $thr} $lang = $1 if /^##(\w.+)$/; $thr = $1 if /^\s+(\d+)\s+threads/; if (/Requests\/sec:\s+(\d+\.\d+)/ && $thr == 10) {$stat{$lang . sprintf(" (%2d thr)", $thr)} = $1} END {for my $lang (sort {$stat{$b} <=> $stat{$a}} keys %stat) {printf "%35s: %5.1f %s\n", $lang, $stat{$lang}, chr(9632) x int($stat{$lang} / 10)}}'
    cat Raspberry*.md | perl -nlE 'use open ":std" => ":utf8"; BEGIN{my %stat, $lang, $thr} $lang = $1 if /^##(\w.+)$/; $thr = $1 if /^\s+(\d+)\s+threads/; if (/Requests\/sec:\s+(\d+\.\d+)/ && $thr == 2) {$stat{$lang . sprintf(" (%2d thr)", $thr)} = $1} END {for my $lang (sort {$stat{$b} <=> $stat{$a}} keys %stat) {printf "%35s: %5.1f %s\n", $lang, $stat{$lang}, chr(9632) x int($stat{$lang} / 10)}}'

####wrk command line

    wrk -d 60 -c 2 -t 2 http://raspberrypi.local:8080/
    wrk -d 60 -c 10 -t 10 http://raspberrypi.local:8080/

version: wrk 4.0.0

HTTP response size is 228 bytes

##Nginx

version: 1.2.1

###wrk result

    Running 1m test @ http://192.168.1.7/test
      2 threads and 2 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency     3.09ms    0.95ms  19.51ms   90.00%
        Req/Sec   325.17     28.44   390.00     89.57%
      38888 requests in 1.00m, 8.45MB read
    Requests/sec:    647.27
    Transfer/sec:    144.09KB

Run in 10 threads:

    Running 1m test @ http://192.168.1.7/test
      10 threads and 10 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency    13.06ms    9.86ms 103.68ms   72.38%
        Req/Sec    82.89     29.38   270.00     74.09%
      49623 requests in 1.00m, 10.79MB read
    Requests/sec:    826.12
    Transfer/sec:    183.90KB

##Go

version: go version go1.6rc2 linux/arm

###run

    go build httpd.go
    ./httpd

###wrk result

    Running 1m test @ http://192.168.1.7:8080/
      2 threads and 2 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency     5.15ms    2.25ms  49.64ms   77.82%
        Req/Sec   197.09     24.45   330.00     77.57%
      23575 requests in 1.00m, 5.13MB read
    Requests/sec:    392.23
    Transfer/sec:     87.33KB

Run in 10 threads:

    Running 1m test @ http://192.168.1.7:8080/
      10 threads and 10 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency    15.84ms    7.68ms 102.41ms   71.94%
        Req/Sec    64.41     18.75   110.00     71.40%
      38611 requests in 1.00m, 8.40MB read
    Requests/sec:    642.51
    Transfer/sec:    143.06KB

##Go/fasthttp

version: go version go1.6rc2 linux/arm

###run

    go build httpd_fast.go
    ./httpd_fast

###wrk result

    Running 1m test @ http://192.168.1.7:8080/
      2 threads and 2 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency     3.70ms    1.30ms  21.67ms   83.68%
        Req/Sec   273.68     18.63   323.00     81.54%
      32729 requests in 1.00m, 7.12MB read
    Requests/sec:    544.58
    Transfer/sec:    121.25KB

Run in 10 threads:

    Running 1m test @ http://192.168.1.7:8080/
      10 threads and 10 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency    11.70ms    8.50ms  97.12ms   85.10%
        Req/Sec    94.14     22.47   180.00     64.05%
      56329 requests in 1.00m, 12.25MB read
    Requests/sec:    937.78
    Transfer/sec:    208.80KB

##Node.js

version: 5.6.0

###wrk result

    Running 1m test @ http://192.168.1.7:8080/
      2 threads and 2 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency    10.01ms    1.53ms  36.74ms   85.47%
        Req/Sec   100.30      7.36   121.00     57.63%
      12007 requests in 1.00m, 2.74MB read
    Requests/sec:    199.76
    Transfer/sec:     46.62KB

Run in 10 threads:

    Running 1m test @ http://192.168.1.7:8080/
      10 threads and 10 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency    51.22ms    3.66ms 141.74ms   96.33%
        Req/Sec    19.28      2.33    30.00     94.80%
      11725 requests in 1.00m, 2.67MB read
    Requests/sec:    195.08
    Transfer/sec:     45.53KB

##Python

version: 2.7.3

###run

    ./httpd.py > /dev/null 2>&1

###wrk result

    Running 1m test @ http://192.168.1.7:8080/
      2 threads and 2 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency    14.15ms    1.12ms  25.76ms   79.60%
        Req/Sec    61.52      4.24    70.00     80.10%
      7377 requests in 1.00m, 1.60MB read
    Requests/sec:    122.87
    Transfer/sec:     27.36KB

Run in 10 threads:

    Running 1m test @ http://192.168.1.7:8080/
      10 threads and 10 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency   147.70ms  280.60ms   1.93s    91.20%
        Req/Sec    16.07      5.75    30.00     68.58%
      5739 requests in 1.00m, 1.25MB read
      Socket errors: connect 0, read 0, write 0, timeout 28
    Requests/sec:     95.52
    Transfer/sec:     21.27KB

##Perl/Mojolicious

version: 6.46

###run

    hypnotoad --foreground ./httpd-mojo.pl

###wrk result

    Running 1m test @ http://192.168.1.7:8080/
      2 threads and 2 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency    69.23ms    7.64ms 174.54ms   87.03%
        Req/Sec    14.10      5.07    20.00     45.06%
      1733 requests in 1.00m, 387.12KB read
    Requests/sec:     28.85
    Transfer/sec:      6.45KB

##Perl/HTTP::Daemon

version: 6.01

    ./httpd.pl

###wrk result

Run in 1 thread:

    Running 1m test @ http://192.168.1.7:8080/
      1 threads and 1 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency    11.05ms    0.91ms  23.85ms   83.68%
        Req/Sec    90.67      5.26   101.00     75.04%
      5426 requests in 1.00m, 1.18MB read
    Requests/sec:     90.40
    Transfer/sec:     20.13KB

##Perl/Dancer

version: 1.3118

    ./httpd-dancer.pl

###wrk result

    Running 1m test @ http://192.168.1.7:3000/
      2 threads and 2 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency    62.11ms    1.59ms  77.28ms   78.48%
        Req/Sec    15.21      5.07    20.00     59.23%
      1863 requests in 1.00m, 414.81KB read
    Requests/sec:     31.02
    Transfer/sec:      6.91KB

##Perl/HTTP::Server::Simple

version: 0.44

    ./httpd-server-simple.pl

###wrk result

    Running 1m test @ http://192.168.1.7:8080/
      2 threads and 2 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency    22.93ms    1.17ms  38.46ms   81.79%
        Req/Sec    39.58      2.72    60.00     93.37%
      4762 requests in 1.00m, 1.04MB read
    Requests/sec:     79.23
    Transfer/sec:     17.64KB

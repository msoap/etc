Raspberry Pi httpd benchmark
============================

Total RPS
---------

All:

                   Go/fasthttp (10 thr): 930.0 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                         Nginx (10 thr): 826.1 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                         Nginx ( 2 thr): 647.3 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                            Go (10 thr): 606.3 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                   Go/fasthttp ( 2 thr): 530.5 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                            Go ( 2 thr): 377.4 ■■■■■■■■■■■■■■■■■■■■■■■■■
                       Node.js ( 2 thr): 199.8 ■■■■■■■■■■■■■
                         Caddy (10 thr): 199.1 ■■■■■■■■■■■■■
                       Node.js (10 thr): 195.1 ■■■■■■■■■■■■■
                         Caddy ( 2 thr): 177.6 ■■■■■■■■■■■
                        Python ( 2 thr): 122.9 ■■■■■■■■
                        Python (10 thr):  95.5 ■■■■■■
             Perl/HTTP::Daemon ( 1 thr):  90.4 ■■■■■■
     Perl/HTTP::Server::Simple ( 2 thr):  79.2 ■■■■■
                   Perl/Dancer ( 2 thr):  31.0 ■■
              Perl/Mojolicious ( 2 thr):  28.9 ■

Run in 10 threads:

    Go/fasthttp (10 thr): 930.0 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
          Nginx (10 thr): 826.1 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
             Go (10 thr): 606.3 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
          Caddy (10 thr): 199.1 ■■■■■■■■■■■■■
        Node.js (10 thr): 195.1 ■■■■■■■■■■■■■
         Python (10 thr):  95.5 ■■■■■■

Run in 2 threads:

                         Nginx ( 2 thr): 647.3 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                   Go/fasthttp ( 2 thr): 530.5 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                            Go ( 2 thr): 377.4 ■■■■■■■■■■■■■■■■■■■■■■■■■
                       Node.js ( 2 thr): 199.8 ■■■■■■■■■■■■■
                         Caddy ( 2 thr): 177.6 ■■■■■■■■■■■
                        Python ( 2 thr): 122.9 ■■■■■■■■
     Perl/HTTP::Server::Simple ( 2 thr):  79.2 ■■■■■
                   Perl/Dancer ( 2 thr):  31.0 ■■
              Perl/Mojolicious ( 2 thr):  28.9 ■

####Create chart:

For all, 10 and 2 threads:

    cat README.md | perl -nlE 'use open ":std" => ":utf8"; BEGIN{my %stat, $lang, $thr} $lang = $1 if /^##(\w.+)$/; $thr = $1 if /^\s+(\d+)\s+threads/; if (/Requests\/sec:\s+(\d+\.\d+)/) {$stat{$lang . sprintf(" (%2d thr)", $thr)} = $1} END {for my $lang (sort {$stat{$b} <=> $stat{$a}} keys %stat) {printf "%35s: %5.1f %s\n", $lang, $stat{$lang}, chr(9632) x int($stat{$lang} / 15)}}'
    cat README.md | perl -nlE 'use open ":std" => ":utf8"; BEGIN{my %stat, $lang, $thr} $lang = $1 if /^##(\w.+)$/; $thr = $1 if /^\s+(\d+)\s+threads/; if (/Requests\/sec:\s+(\d+\.\d+)/ && $thr == 10) {$stat{$lang . sprintf(" (%2d thr)", $thr)} = $1} END {for my $lang (sort {$stat{$b} <=> $stat{$a}} keys %stat) {printf "%35s: %5.1f %s\n", $lang, $stat{$lang}, chr(9632) x int($stat{$lang} / 15)}}'
    cat README.md | perl -nlE 'use open ":std" => ":utf8"; BEGIN{my %stat, $lang, $thr} $lang = $1 if /^##(\w.+)$/; $thr = $1 if /^\s+(\d+)\s+threads/; if (/Requests\/sec:\s+(\d+\.\d+)/ && $thr == 2) {$stat{$lang . sprintf(" (%2d thr)", $thr)} = $1} END {for my $lang (sort {$stat{$b} <=> $stat{$a}} keys %stat) {printf "%35s: %5.1f %s\n", $lang, $stat{$lang}, chr(9632) x int($stat{$lang} / 15)}}'

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

##Caddy

version: 0.8.1

run:

    caddy -port 8080 browse

###wrk result

    Running 1m test @ http://192.168.1.7:8080/test
      2 threads and 2 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency    12.03ms    8.09ms  89.77ms   81.63%
        Req/Sec    89.04     17.43   141.00     62.31%
      10665 requests in 1.00m, 2.32MB read
    Requests/sec:    177.57
    Transfer/sec:     39.54KB

Run in 10 threads:

    Running 1m test @ http://192.168.1.7:8080/test
      10 threads and 10 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency    50.68ms   21.52ms 201.55ms   72.34%
        Req/Sec    19.75      7.00    49.00     55.73%
      11959 requests in 1.00m, 2.60MB read
    Requests/sec:    199.09
    Transfer/sec:     44.33KB

##Go

version: go version go1.6 linux/arm

###run

    go build httpd.go
    ./httpd

###wrk result

    Running 1m test @ http://192.168.1.7:8080/
      2 threads and 2 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency     5.38ms    2.41ms  47.12ms   78.42%
        Req/Sec   189.64     21.68   303.00     77.49%
      22679 requests in 1.00m, 4.93MB read
    Requests/sec:    377.42
    Transfer/sec:     84.04KB

Run in 10 threads:

    Running 1m test @ http://192.168.1.7:8080/
      10 threads and 10 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency    16.74ms    7.80ms  97.04ms   72.18%
        Req/Sec    60.75     16.59   111.00     69.71%
      36420 requests in 1.00m, 7.92MB read
    Requests/sec:    606.27
    Transfer/sec:    134.99KB

##Go/fasthttp

version: go version go1.6 linux/arm

###run

    go build httpd_fast.go
    ./httpd_fast

###wrk result

    Running 1m test @ http://192.168.1.7:8080/
      2 threads and 2 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency     3.86ms    1.85ms  47.12ms   92.12%
        Req/Sec   266.46     27.37   323.00     86.45%
      31856 requests in 1.00m, 6.93MB read
    Requests/sec:    530.54
    Transfer/sec:    118.13KB

Run in 10 threads:

    Running 1m test @ http://192.168.1.7:8080/
      10 threads and 10 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency    11.85ms    8.77ms  96.43ms   85.72%
        Req/Sec    93.40     23.24   190.00     62.97%
      55891 requests in 1.00m, 12.15MB read
    Requests/sec:    929.98
    Transfer/sec:    207.07KB

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

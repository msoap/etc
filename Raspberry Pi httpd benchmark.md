Raspberry Pi httpd benchmark
============================

Total RPS
---------

                     Nginx: 173.8 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                   Node.js:  98.6 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                        Go:  81.2 ■■■■■■■■■■■■■■■■■■■■■■■
                    Python:  78.6 ■■■■■■■■■■■■■■■■■■■■■■
    Perl with HTTP::Daemon:  62.6 ■■■■■■■■■■■■■■■■■
     Perl with Mojolicious:   8.4 ■■

####Create chart:

    cat Raspberry*.md | perl -nlE 'use open ":std" => ":utf8"; BEGIN{my %stat} $lang = $1 if /^##(\w.+)$/; if (/Requests per second:\s+(\d+\.\d+)/) {$stat{$lang} = $1} END {for my $lang (sort {$stat{$b} <=> $stat{$a}} keys %stat) {printf "%22s: %5.1f %s\n", $lang, $stat{$lang}, chr(9632) x int($stat{$lang} / 3.5)}}'

####ab command line

    ab -r -n 1000 http://raspberrypi.local:8080/

##Nginx

With one small html file

###ab result

    Server Software:        nginx/1.2.1
    Server Hostname:        192.168.1.7
    Server Port:            80

    Document Path:          /
    Document Length:        114 bytes

    Concurrency Level:      1
    Time taken for tests:   5.755 seconds
    Complete requests:      1000
    Failed requests:        0
    Write errors:           0
    Total transferred:      324000 bytes
    HTML transferred:       114000 bytes
    Requests per second:    173.76 [#/sec] (mean)
    Time per request:       5.755 [ms] (mean)
    Time per request:       5.755 [ms] (mean, across all concurrent requests)
    Transfer rate:          54.98 [Kbytes/sec] received

    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        1    2   1.1      2      17
    Processing:     2    3   1.3      3      16
    Waiting:        2    3   1.3      3      16
    Total:          4    6   1.8      5      24

    Percentage of the requests served within a certain time (ms)
      50%      5
      66%      5
      75%      6
      80%      6
      90%      7
      95%      9
      98%     11
      99%     14
     100%     24 (longest request)

##Go

###run

    export GOARM=5
    go build httpd.go
    ./httpd

###ab result

    Server Software:
    Server Hostname:        192.168.1.7
    Server Port:            8080

    Document Path:          /
    Document Length:        19 bytes

    Concurrency Level:      1
    Time taken for tests:   12.312 seconds
    Complete requests:      1000
    Failed requests:        0
    Write errors:           0
    Total transferred:      116000 bytes
    HTML transferred:       19000 bytes
    Requests per second:    81.22 [#/sec] (mean)
    Time per request:       12.312 [ms] (mean)
    Time per request:       12.312 [ms] (mean, across all concurrent requests)
    Transfer rate:          9.20 [Kbytes/sec] received

    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        1    2   2.8      2      63
    Processing:     4   10  29.0      5     279
    Waiting:        4   10  29.0      5     279
    Total:          6   12  29.1      7     281

    Percentage of the requests served within a certain time (ms)
      50%      7
      66%      7
      75%      8
      80%      8
      90%     10
      95%     14
      98%    189
      99%    199
     100%    281 (longest request)

##Perl with Mojolicious

###run

    hypnotoad ./httpd-mojo.pl

###ab result

    Server Software:        Mojolicious
    Server Hostname:        192.168.1.7
    Server Port:            8080

    Document Path:          /
    Document Length:        28 bytes

    Concurrency Level:      1
    Time taken for tests:   118.832 seconds
    Complete requests:      1000
    Failed requests:        0
    Write errors:           0
    Total transferred:      190000 bytes
    HTML transferred:       28000 bytes
    Requests per second:    8.42 [#/sec] (mean)
    Time per request:       118.832 [ms] (mean)
    Time per request:       118.832 [ms] (mean, across all concurrent requests)
    Transfer rate:          1.56 [Kbytes/sec] received

    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        2    2   2.8      2      63
    Processing:   104  116   9.2    114     204
    Waiting:      102  114   8.3    112     202
    Total:        106  119   9.7    116     209

    Percentage of the requests served within a certain time (ms)
      50%    116
      66%    121
      75%    125
      80%    126
      90%    128
      95%    132
      98%    138
      99%    157
     100%    209 (longest request)

##Perl with HTTP::Daemon

###ab result

    Server Software:        libwww-perl-daemon/6.01
    Server Hostname:        192.168.1.7
    Server Port:            8080

    Document Path:          /
    Document Length:        35 bytes

    Concurrency Level:      1
    Time taken for tests:   15.980 seconds
    Complete requests:      1000
    Failed requests:        0
    Write errors:           0
    Total transferred:      170000 bytes
    HTML transferred:       35000 bytes
    Requests per second:    62.58 [#/sec] (mean)
    Time per request:       15.980 [ms] (mean)
    Time per request:       15.980 [ms] (mean, across all concurrent requests)
    Transfer rate:          10.39 [Kbytes/sec] received

    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        2    2   3.6      2      94
    Processing:    11   14   4.6     12     133
    Waiting:        9   10   4.3     10     128
    Total:         13   16   5.8     14     136

    Percentage of the requests served within a certain time (ms)
      50%     14
      66%     15
      75%     16
      80%     16
      90%     18
      95%     21
      98%     26
      99%     28
     100%    136 (longest request)

##Node.js

v0.10.16

###ab result

    Server Software:
    Server Hostname:        192.168.1.7
    Server Port:            8080

    Document Path:          /
    Document Length:        24 bytes

    Concurrency Level:      1
    Time taken for tests:   10.139 seconds
    Complete requests:      1000
    Failed requests:        0
    Write errors:           0
    Total transferred:      125000 bytes
    HTML transferred:       24000 bytes
    Requests per second:    98.63 [#/sec] (mean)
    Time per request:       10.139 [ms] (mean)
    Time per request:       10.139 [ms] (mean, across all concurrent requests)
    Transfer rate:          12.04 [Kbytes/sec] received

    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        2    3   6.8      2     165
    Processing:     6    7   2.1      7      42
    Waiting:        5    6   2.0      6      41
    Total:          8   10   7.3      9     175

    Percentage of the requests served within a certain time (ms)
      50%      9
      66%      9
      75%     10
      80%     11
      90%     13
      95%     15
      98%     19
      99%     20
     100%    175 (longest request)

##Python

###ab result

    Server Software:        BaseHTTP/0.3
    Server Hostname:        192.168.1.7
    Server Port:            8080

    Document Path:          /
    Document Length:        17 bytes

    Concurrency Level:      1
    Time taken for tests:   12.729 seconds
    Complete requests:      1000
    Failed requests:        0
    Write errors:           0
    Total transferred:      134000 bytes
    HTML transferred:       17000 bytes
    Requests per second:    78.56 [#/sec] (mean)
    Time per request:       12.729 [ms] (mean)
    Time per request:       12.729 [ms] (mean, across all concurrent requests)
    Transfer rate:          10.28 [Kbytes/sec] received

    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        1    2   2.6      2      61
    Processing:     8   10   2.8      9      63
    Waiting:        6    8   2.0      7      24
    Total:         10   13   3.7     11      70

    Percentage of the requests served within a certain time (ms)
      50%     11
      66%     12
      75%     13
      80%     14
      90%     15
      95%     17
      98%     21
      99%     24
     100%     70 (longest request)

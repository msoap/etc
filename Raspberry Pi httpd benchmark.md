Raspberry Pi httpd benchmark
============================

Total RPS
---------

                        Nginx: 173.8 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                           Go: 143.1 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                      Node.js:  98.6 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                       Python:  78.6 ■■■■■■■■■■■■■■■■■■■■■■
    Perl/HTTP::Server::Simple:  67.0 ■■■■■■■■■■■■■■■■■■■
            Perl/HTTP::Daemon:  62.6 ■■■■■■■■■■■■■■■■■
                  Perl/Dancer:  19.1 ■■■■■
             Perl/Mojolicious:  10.4 ■■

####Create chart:

    cat Raspberry*.md | perl -nlE 'use open ":std" => ":utf8"; BEGIN{my %stat} $lang = $1 if /^##(\w.+)$/; if (/Requests per second:\s+(\d+\.\d+)/) {$stat{$lang} = $1} END {for my $lang (sort {$stat{$b} <=> $stat{$a}} keys %stat) {printf "%27s: %5.1f %s\n", $lang, $stat{$lang}, chr(9632) x int($stat{$lang} / 3.5)}}'

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

go version go1.2 linux/arm

###run

    go build httpd.go
    ./httpd

###ab result

    Server Software:
    Server Hostname:        192.168.1.7
    Server Port:            8080

    Document Path:          /
    Document Length:        19 bytes

    Concurrency Level:      1
    Time taken for tests:   6.990 seconds
    Complete requests:      1000
    Failed requests:        0
    Write errors:           0
    Total transferred:      136000 bytes
    HTML transferred:       19000 bytes
    Requests per second:    143.07 [#/sec] (mean)
    Time per request:       6.990 [ms] (mean)
    Time per request:       6.990 [ms] (mean, across all concurrent requests)
    Transfer rate:          19.00 [Kbytes/sec] received

    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        1    2   4.1      2     103
    Processing:     3    5   1.6      4      14
    Waiting:        3    4   1.3      4      13
    Total:          4    7   4.5      6     110

    Percentage of the requests served within a certain time (ms)
      50%      6
      66%      8
      75%      8
      80%      8
      90%      9
      95%     10
      98%     12
      99%     14
     100%    110 (longest request)

##Perl/Mojolicious

###run

    hypnotoad ./httpd-mojo.pl

###ab result

    Server Software:        Mojolicious
    Server Hostname:        192.168.1.7
    Server Port:            8080

    Document Path:          /
    Document Length:        38 bytes

    Concurrency Level:      1
    Time taken for tests:   96.270 seconds
    Complete requests:      1000
    Failed requests:        0
    Write errors:           0
    Total transferred:      200000 bytes
    HTML transferred:       38000 bytes
    Requests per second:    10.39 [#/sec] (mean)
    Time per request:       96.270 [ms] (mean)
    Time per request:       96.270 [ms] (mean, across all concurrent requests)
    Transfer rate:          2.03 [Kbytes/sec] received

    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        2    3   4.7      2     124
    Processing:    90   93   4.1     93     194
    Waiting:       89   92   2.4     91     115
    Total:         92   96   6.3     95     220

    Percentage of the requests served within a certain time (ms)
      50%     95
      66%     96
      75%     97
      80%     97
      90%     99
      95%    102
      98%    105
      99%    111
     100%    220 (longest request)

##Perl/HTTP::Daemon

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


##Perl/Dancer

###ab result

     Server Software:        Perl
     Server Hostname:        192.168.1.7
     Server Port:            8080

     Document Path:          /
     Document Length:        0 bytes

     Concurrency Level:      1
     Time taken for tests:   52.489 seconds
     Complete requests:      1000
     Failed requests:        1496
        (Connect: 0, Receive: 497, Length: 999, Exceptions: 0)
     Write errors:           503
     Total transferred:      158841 bytes
     HTML transferred:       32967 bytes
     Requests per second:    19.05 [#/sec] (mean)
     Time per request:       52.489 [ms] (mean)
     Time per request:       52.489 [ms] (mean, across all concurrent requests)
     Transfer rate:          2.96 [Kbytes/sec] received

     Connection Times (ms)
                   min  mean[+/-sd] median   max
     Connect:       41   52  41.2     50    1294
     Processing:     0    0   0.0      0       0
     Waiting:        0    0   0.0      0       0
     Total:         41   52  41.2     50    1294

     Percentage of the requests served within a certain time (ms)
       50%     50
       66%     51
       75%     52
       80%     53
       90%     55
       95%     56
       98%     59
       99%    159
      100%   1294 (longest request)

##Perl/HTTP::Server::Simple

###ab result

     Server Software:        Simple
     Server Hostname:        192.168.1.7
     Server Port:            8080

     Document Path:          /
     Document Length:        42 bytes

     Concurrency Level:      1
     Time taken for tests:   14.937 seconds
     Complete requests:      1000
     Failed requests:        0
     Write errors:           0
     Total transferred:      106000 bytes
     HTML transferred:       42000 bytes
     Requests per second:    66.95 [#/sec] (mean)
     Time per request:       14.937 [ms] (mean)
     Time per request:       14.937 [ms] (mean, across all concurrent requests)
     Transfer rate:          6.93 [Kbytes/sec] received

     Connection Times (ms)
                   min  mean[+/-sd] median   max
     Connect:        1    2   3.5      2      90
     Processing:    11   13   0.6     12      17
     Waiting:       11   12   0.6     12      17
     Total:         13   15   3.6     14     104
     WARNING: The median and mean for the processing time are not within a normal deviation
             These results are probably not that reliable.

     Percentage of the requests served within a certain time (ms)
       50%     14
       66%     15
       75%     15
       80%     15
       90%     16
       95%     16
       98%     17
       99%     18
      100%    104 (longest request)

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

RPS on Macbook Air
------------------

                           Go: 3828 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                      Node.js: 3126 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
            Perl/HTTP::Daemon: 1859 ■■■■■■■■■■■■■■■■■■■■■■■■■■
                       Python: 1565 ■■■■■■■■■■■■■■■■■■■■■■
    Perl/HTTP::Server::Simple: 1442 ■■■■■■■■■■■■■■■■■■■■
                  Perl/Dancer:  452 ■■■■■■
             Perl/Mojolicious:  303 ■■■■

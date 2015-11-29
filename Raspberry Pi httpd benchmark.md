Raspberry Pi httpd benchmark
============================

Total RPS
---------

                        Nginx: 351.5 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                  Go/fasthttp: 341.9 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                           Go: 247.2 ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
                      Node.js: 126.0 ■■■■■■■■■■■■■■■■■■■■■■■■■
            Perl/HTTP::Daemon:  66.9 ■■■■■■■■■■■■■
    Perl/HTTP::Server::Simple:  66.4 ■■■■■■■■■■■■■
                       Python:  59.5 ■■■■■■■■■■■
                  Perl/Dancer:  28.2 ■■■■■
             Perl/Mojolicious:  17.6 ■■■

####Create chart:

    cat Raspberry*.md | perl -nlE 'use open ":std" => ":utf8"; BEGIN{my %stat} $lang = $1 if /^##(\w.+)$/; if (/Requests per second:\s+(\d+\.\d+)/) {$stat{$lang} = $1} END {for my $lang (sort {$stat{$b} <=> $stat{$a}} keys %stat) {printf "%27s: %5.1f %s\n", $lang, $stat{$lang}, chr(9632) x int($stat{$lang} / 5)}}'

####ab command line

    ab -r -c 10 -n 1000 http://raspberrypi.local:8080/

##Nginx

With one small html file

###ab result

    Server Software:        nginx/1.2.1
    Server Hostname:        192.168.1.7
    Server Port:            80

    Document Path:          /
    Document Length:        228 bytes

    Concurrency Level:      10
    Time taken for tests:   2.845 seconds
    Complete requests:      1000
    Failed requests:        0
    Total transferred:      438000 bytes
    HTML transferred:       228000 bytes
    Requests per second:    351.49 [#/sec] (mean)
    Time per request:       28.450 [ms] (mean)
    Time per request:       2.845 [ms] (mean, across all concurrent requests)
    Transfer rate:          150.34 [Kbytes/sec] received

    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        2    3   1.4      3      13
    Processing:     6   25   7.7     26      61
    Waiting:        6   25   7.7     26      61
    Total:          9   28   7.7     28      65

    Percentage of the requests served within a certain time (ms)
      50%     28
      66%     30
      75%     31
      80%     32
      90%     38
      95%     43
      98%     48
      99%     54
     100%     65 (longest request)

##Go

go version go1.5 linux/arm

###run

    go build httpd.go
    ./httpd

###ab result

    Server Software:
    Server Hostname:        192.168.1.7
    Server Port:            8080

    Document Path:          /
    Document Length:        19 bytes

    Concurrency Level:      10
    Time taken for tests:   4.045 seconds
    Complete requests:      1000
    Failed requests:        0
    Total transferred:      136000 bytes
    HTML transferred:       19000 bytes
    Requests per second:    247.20 [#/sec] (mean)
    Time per request:       40.453 [ms] (mean)
    Time per request:       4.045 [ms] (mean, across all concurrent requests)
    Transfer rate:          32.83 [Kbytes/sec] received

    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        2    2   0.7      2       8
    Processing:     4   38  10.9     41      65
    Waiting:        4   37  10.8     41      64
    Total:          7   40  10.9     43      68

    Percentage of the requests served within a certain time (ms)
      50%     43
      66%     44
      75%     45
      80%     45
      90%     46
      95%     47
      98%     50
      99%     54
     100%     68 (longest request)

##Go/fasthttp

go version go1.5 linux/arm

###run

    go build httpd_fast.go
    ./httpd_fast

###ab result

    Server Software:        fasthttp
    Server Hostname:        192.168.1.7
    Server Port:            8080

    Document Path:          /
    Document Length:        19 bytes

    Concurrency Level:      10
    Time taken for tests:   2.925 seconds
    Complete requests:      1000
    Failed requests:        0
    Total transferred:      139000 bytes
    HTML transferred:       19000 bytes
    Requests per second:    341.85 [#/sec] (mean)
    Time per request:       29.253 [ms] (mean)
    Time per request:       2.925 [ms] (mean, across all concurrent requests)
    Transfer rate:          46.40 [Kbytes/sec] received

    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        2    3   1.2      3       9
    Processing:     3   26   8.1     28      71
    Waiting:        3   25   7.7     27      49
    Total:          5   29   8.3     31      74

    Percentage of the requests served within a certain time (ms)
      50%     31
      66%     32
      75%     33
      80%     33
      90%     34
      95%     36
      98%     38
      99%     47
     100%     74 (longest request)
 
##Perl/Mojolicious

Version: 6.32

###run

    hypnotoad ./httpd-mojo.pl

###ab result

    Server Software:        Mojolicious
    Server Hostname:        192.168.1.7
    Server Port:            8080

    Document Path:          /
    Document Length:        38 bytes

    Concurrency Level:      10
    Time taken for tests:   56.730 seconds
    Complete requests:      1000
    Failed requests:        0
    Total transferred:      181000 bytes
    HTML transferred:       38000 bytes
    Requests per second:    17.63 [#/sec] (mean)
    Time per request:       567.305 [ms] (mean)
    Time per request:       56.730 [ms] (mean, across all concurrent requests)
    Transfer rate:          3.12 [Kbytes/sec] received

    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        2    2   0.7      2      17
    Processing:   144  564 302.1    516    2276
    Waiting:      142  554 302.1    501    2225
    Total:        147  566 302.1    517    2278

    Percentage of the requests served within a certain time (ms)
      50%    517
      66%    653
      75%    737
      80%    806
      90%    967
      95%   1107
      98%   1358
      99%   1464
     100%   2278 (longest request)

##Perl/HTTP::Daemon

###ab result

    Server Software:        libwww-perl-daemon/6.01
    Server Hostname:        192.168.1.7
    Server Port:            8080

    Document Path:          /
    Document Length:        40 bytes

    Concurrency Level:      1
    Time taken for tests:   14.958 seconds
    Complete requests:      1000
    Failed requests:        0
    Total transferred:      175000 bytes
    HTML transferred:       40000 bytes
    Requests per second:    66.86 [#/sec] (mean)
    Time per request:       14.958 [ms] (mean)
    Time per request:       14.958 [ms] (mean, across all concurrent requests)
    Transfer rate:          11.43 [Kbytes/sec] received

    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        2    2   0.7      2       7
    Processing:    11   13   2.9     12      95
    Waiting:        8    9   0.7      9      18
    Total:         13   15   3.1     14     100

    Percentage of the requests served within a certain time (ms)
      50%     14
      66%     15
      75%     15
      80%     15
      90%     16
      95%     17
      98%     18
      99%     19
     100%    100 (longest request)

##Perl/Dancer

###ab result

    Server Software:        Perl
    Server Hostname:        192.168.1.7
    Server Port:            3000

    Document Path:          /
    Document Length:        33 bytes

    Concurrency Level:      10
    Time taken for tests:   35.496 seconds
    Complete requests:      1000
    Failed requests:        0
    Total transferred:      159000 bytes
    HTML transferred:       33000 bytes
    Requests per second:    28.17 [#/sec] (mean)
    Time per request:       354.959 [ms] (mean)
    Time per request:       35.496 [ms] (mean, across all concurrent requests)
    Transfer rate:          4.37 [Kbytes/sec] received

    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        2    5  23.1      2     360
    Processing:    46  349  27.0    347     609
    Waiting:       46  349  27.0    347     609
    Total:        298  353  21.1    350     613

    Percentage of the requests served within a certain time (ms)
      50%    350
      66%    351
      75%    351
      80%    352
      90%    352
      95%    354
      98%    424
      99%    477
     100%    613 (longest request)

##Perl/HTTP::Server::Simple

###ab result

    Server Software:        Simple
    Server Hostname:        192.168.1.7
    Server Port:            8080

    Document Path:          /
    Document Length:        42 bytes

    Concurrency Level:      10
    Time taken for tests:   15.067 seconds
    Complete requests:      1000
    Failed requests:        0
    Total transferred:      106000 bytes
    HTML transferred:       42000 bytes
    Requests per second:    66.37 [#/sec] (mean)
    Time per request:       150.670 [ms] (mean)
    Time per request:       15.067 [ms] (mean, across all concurrent requests)
    Transfer rate:          6.87 [Kbytes/sec] received

    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        2    2   1.5      2      35
    Processing:   103  148  98.9    138    1192
    Waiting:      103  147  98.9    137    1191
    Total:        125  150  99.1    140    1197

    Percentage of the requests served within a certain time (ms)
      50%    140
      66%    140
      75%    141
      80%    141
      90%    142
      95%    143
      98%    145
      99%   1072
     100%   1197 (longest request)

##Node.js

v0.10.22

###ab result

    Server Software:
    Server Hostname:        192.168.1.7
    Server Port:            8080

    Document Path:          /
    Document Length:        24 bytes

    Concurrency Level:      10
    Time taken for tests:   7.938 seconds
    Complete requests:      1000
    Failed requests:        0
    Total transferred:      125000 bytes
    HTML transferred:       24000 bytes
    Requests per second:    125.97 [#/sec] (mean)
    Time per request:       79.385 [ms] (mean)
    Time per request:       7.938 [ms] (mean, across all concurrent requests)
    Transfer rate:          15.38 [Kbytes/sec] received

    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        1    2   0.6      2       8
    Processing:    23   77  12.7     75     195
    Waiting:       22   76  12.7     74     194
    Total:         25   79  12.6     77     197

    Percentage of the requests served within a certain time (ms)
      50%     77
      66%     78
      75%     78
      80%     79
      90%     80
      95%     87
      98%    111
      99%    175
     100%    197 (longest request)

##Python

###ab result

    Server Software:        BaseHTTP/0.3
    Server Hostname:        192.168.1.7
    Server Port:            8080

    Document Path:          /
    Document Length:        17 bytes

    Concurrency Level:      10
    Time taken for tests:   16.818 seconds
    Complete requests:      1000
    Failed requests:        0
    Total transferred:      134000 bytes
    HTML transferred:       17000 bytes
    Requests per second:    59.46 [#/sec] (mean)
    Time per request:       168.184 [ms] (mean)
    Time per request:       16.818 [ms] (mean, across all concurrent requests)
    Transfer rate:          7.78 [Kbytes/sec] received

    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        2    3  11.4      2     273
    Processing:    21  112 640.4     67   16295
    Waiting:       18  109 640.4     65   16293
    Total:         23  115 644.5     70   16432

    Percentage of the requests served within a certain time (ms)
      50%     70
      66%     70
      75%     71
      80%     71
      90%     72
      95%     72
      98%    262
      99%    548
     100%  16432 (longest request)

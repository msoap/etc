mp3-fix-id3v2
-------------

Fix empty id3 tags in mp3 files.

Accept FS structure:

    /.../Artist/Year - Album Name/Num. Title.mp3

example:

    /home/Music/Artist/2018 - Album Name/01. Some Song.mp3

Usage
=====

    mp3-fix-id3v2 [-dry-run -artist=... -album=... -year=... -title=...] *.mp3

options:

      -dry-run    : only output new tags
      -artist=... : set artist name
      -album=...  : set album name
      -title=...  : set title
      -year=...   : set year

Install
=======

    go get -u github.com/msoap/etc/mp3-fix-id3v2

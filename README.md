## jch-metadata

`jch-metadata` is a CLI tool to retrieve metadata from media file.

### Build

Run the following command:

```
go build ./cmd/jch-metadata
```

To build for Windows (64-bit), run the following command:

```
GOOS=windows GOARCH=amd64 go build ./cmd/jch-metadata
```

### Run

To display metadata for a file, run the following command:

```
$ jch-metadata -f test1.mkv

Info
====
Filename    : 
Date        : 2010-08-21 14:23:03 +0700 WIB
Title       : 
Muxing App  : libebml2 v0.10.0 + libmatroska2 v0.10.1
Writing App : mkclean 0.5.5 ru from libebml v1.0.0 + libmatroska v1.0.0 + mkvmerge v4.1.1 ('Bouncin' Back') built on Jul  3 2010 22:54:08

Track 1
=========
Name     : 
Type     : video
Language : und

Track 2
=========
Name     : 
Type     : audio
Language : und

Tag
===
Name        : TITLE
Target Type : 
Language    : 
Value       : Big Buck Bunny - test 1
```

To remove metadata for a file, run the following command:

```
$ jch-metadata -f test1.mkv -a clear
Removing all values from Info elements...
Metadata cleared
```
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
$ jch-metadata -f test2.mkv

Opening file test1.mkv
File type is Mkv (Matroska)

Info
Filename     : 
Date         : 2010-08-21 14:23:03 +0700 WIB
Title        : 
Muxing App   : libebml2 v0.10.0 + libmatroska2 v0.10.1
Writing App  : mkclean 0.5.5 ru from libebml v1.0.0 + libmatroska v1.0.0 + mkvmerge v4.1.1 ('Bouncin' Back') built on Jul  3 2010 22:54:08

Track 1
Name         : 
Type         : video
Language     : und

Track 2
Name         : 
Type         : audio
Language     : und

Tag
Name         : TITLE
Target Type  : 
Language     : 
Value        : Big Buck Bunny - test 1
```

For container format like Matroska that supports multiple attachments, `jch-metadata` will also perform inspection on the attachments:

```
Opening file test2.mkv
File type is Mkv (Matroska)

Info
Filename     : 
Date         : 
Title        : Big Buck Bunny - test 1
Muxing App   : Lavf58.45.100
Writing App  : Lavf58.45.100

Track 1
Name         : 
Type         : video
Language     : und

Track 2
Name         : 
Type         : audio
Language     : und

Attachment
Name         : test1.flac
Media Type   : image/png
Description  : 

  File type is FLAC

  Vendor String: reference libFLAC 1.3.2 20170101
  User Comments
   Comment=Processed by SoX
  

Attachment
Name         : test1.png
Media Type   : image/png
Description  : 

  File type is PNG

  Software     : gnome-screenshot
  Creation Time: 2023-05-20T02:56:29+0700

Tag
Name         : COMMENT
Target Type  : 
Language     : 
Value        : Matroska Validation File1, basic MPEG4.2 and MP3 with only SimpleBlock

Tag
Name         : DURATION
Target Type  : 
Language     : 
Value        : 00:01:27.333000000

Tag
Name         : DURATION
Target Type  : 
Language     : 
Value        : 00:01:27.336000000
```


To remove metadata for a file, run the following command:

```
$ jch-metadata -f test1.mkv -a clear
Removing all values from Info elements...
Metadata cleared
```
# dvpl-go | A package for Go to work with DVPL compression format.

## Overview
World of Tanks Blitz and Tanks Blitz use a custom compression format named DVPL that is actually a LZ4 compression, usually of a level 2 with the exception of level 0 for .tex files, with the addition of a special footer. This package is made to work with this format, namely to convert files in and out of DVPL format.

## What's inside?
There are one struct and three functions available:
```go

type DVPLFooterData struct {
	OriginalSize   uint32 // The original size of the content.
	CompressedSize uint32 // The compressed size of the content without footer.
	CRC32          uint32 // The CRC32 hash sum of the content without footer.
	CompressType   uint32 // The compression type used. Usually is 2, with the exception of 0 for .tex files.
}

// Read a DVPL footer from the given byte array and return the DVPLFooterData struct type,
// or return an error if no valid DVPL footer is present in the given byte array.
func ReadDVPLFooterData(buf []byte) (DVPLFooterData, error)

// Compress the given byte array using HC_2 compression, or no compression if noCompression is specified,
// add a DVPL footer and return the resulting byte array, or return an error if the compression algorithm fails.
func CompressDVPL(buf []byte, noCompression bool) ([]byte, error)

// Decompress the given byte array and return the decompressed byte array,
// or return an error if the given byte array is not a valid DVPL.
func DecompressDVPL(buf []byte) ([]byte, error)

```

## Install
Assuming you have the go toolchain installed

```
go get github.com/Endg4meZer0/dvpl-go
```

## Usage
```go

file, err := os.ReadFile(path)
if err != nil {
  fmt.Fprintf(os.Stderr, "An unknown error occured when trying to read %s. An issue with permissions?", path)
  return
}

var convertedFile []byte
var newPath string

if decompressMode {
  convertedFile, err = dvpl.DecompressDVPL(file)
  var hasSuffix bool
  newPath, hasSuffix = strings.CutSuffix(path, ".dvpl")
  if !hasSuffix {
    newPath = path + ".nodvpl"
  }
} else {
  convertedFile, err = dvpl.CompressDVPL(file, strings.HasSuffix(path, ".tex"))
  newPath = path + ".dvpl"
}

if err != nil {
  fmt.Fprintf(os.Stderr, "An error occured during the conversion of the file %s:\n%s", path, err.Error())
  return
}
os.WriteFile(newPath, convertedFile, 0777)

```

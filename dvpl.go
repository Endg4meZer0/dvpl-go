package dvpl

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"

	"github.com/pierrec/lz4/v4"
)

type DVPLFooterData struct {
	OriginalSize   uint32 // The original size of the content.
	CompressedSize uint32 // The compressed size of the content without footer.
	CRC32          uint32 // The CRC32 hash sum of the content without footer.
	CompressType   uint32 // The compression type used. Usually is 2, with the exception of 0 for .tex files.
}

// Read a DVPL footer from the given byte array. If no valid DVPL footer is present, return error.
func ReadDVPLFooterData(buf []byte) (DVPLFooterData, error) {
	footerBuf := buf[len(buf)-20:]
	dvplTypeBuf := footerBuf[len(footerBuf)-4:]
	if string(dvplTypeBuf) != "DVPL" {
		return DVPLFooterData{}, fmt.Errorf("invalid filetype in footer data: expected \"DVPL\", received \"%s\"", dvplTypeBuf)
	}

	footerData := DVPLFooterData{
		OriginalSize:   binary.LittleEndian.Uint32(footerBuf[0:4]),
		CompressedSize: binary.LittleEndian.Uint32(footerBuf[4:8]),
		CRC32:          binary.LittleEndian.Uint32(footerBuf[8:12]),
		CompressType:   binary.LittleEndian.Uint32(footerBuf[12:16]),
	}

	return footerData, nil
}

// Convert the DVPLFooterData struct into byte array ready to be appended to the resulting file.
func toDVPLFooter(fd DVPLFooterData) []byte {
	res := make([]byte, 0, 20)
	res = binary.LittleEndian.AppendUint32(res, fd.OriginalSize)
	res = binary.LittleEndian.AppendUint32(res, fd.CompressedSize)
	res = binary.LittleEndian.AppendUint32(res, fd.CRC32)
	res = binary.LittleEndian.AppendUint32(res, fd.CompressType)
	res = append(res, []byte("DVPL")...)
	return res
}

// Decompress the given byte array and return the decompressed byte array, or return an error if the given byte array is not a valid DVPL.
func DecompressDVPL(buf []byte) ([]byte, error) {
	footerData, err := ReadDVPLFooterData(buf)
	if err != nil {
		return nil, err
	}
	targetBuf := buf[0 : len(buf)-20]
	if targetBufLen := len(targetBuf); targetBufLen != int(footerData.CompressedSize) {
		return nil, fmt.Errorf("the real compressed size of the file mismatches with the footer data: found %d in footer, received %d as real size", int(footerData.CompressedSize), targetBufLen)
	}
	if hashSum := crc32.ChecksumIEEE(targetBuf); hashSum != footerData.CRC32 {
		return nil, fmt.Errorf("the CRC32 hash sum mismatches with the footer data: found %d in footer, received %d after hashing real data", int(footerData.CRC32), hashSum)
	}

	if footerData.CompressType == 0 && footerData.OriginalSize != footerData.CompressedSize {
		return nil, fmt.Errorf("compression type is level 0 but footer data shows a mismatch between original and compressed sizes")
	} else if footerData.CompressType == 0 {
		return targetBuf, nil
	} else if footerData.CompressType <= 2 {
		uncompressedBuf := make([]byte, footerData.OriginalSize)
		_, err := lz4.UncompressBlock(targetBuf, uncompressedBuf)
		if err != nil {
			return nil, fmt.Errorf("failed to uncompress the buffer. More:\n" + err.Error())
		}

		i := len(uncompressedBuf) - 1
		for ; uncompressedBuf[i] == '\x00'; i-- {
		}

		return uncompressedBuf[:i+1], nil
	} else {
		return nil, fmt.Errorf("not a DVPL level of compression. The maximum level of compression used in DVPL is 2, but met %d", footerData.CompressType)
	}
}

// Compress the given byte array using level 2 compression, or level 0 compression if noCompression is true,
// add a DVPL footer and return the resulting byte array, or return an error if the compression algorithm fails.
func CompressDVPL(buf []byte, noCompression bool) ([]byte, error) {
	compressionType := 2
	compressor := lz4.CompressorHC{Level: lz4.Level2}
	if noCompression {
		compressionType = 0
		compressor.Level = lz4.Fast
	}
	leastCompressableSize := lz4.CompressBlockBound(len(buf))
	compressedBuf := make([]byte, leastCompressableSize)
	i, err := compressor.CompressBlock(buf, compressedBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to compress the buffer. More:\n" + err.Error())
	}

	readyBuf := append(make([]byte, 0, i+20), compressedBuf[:i]...)
	readyBuf = append(readyBuf, toDVPLFooter(DVPLFooterData{
		OriginalSize:   uint32(len(buf)),
		CompressedSize: uint32(i),
		CRC32:          crc32.ChecksumIEEE(compressedBuf[:i]),
		CompressType:   uint32(compressionType),
	})...)

	return readyBuf, nil
}

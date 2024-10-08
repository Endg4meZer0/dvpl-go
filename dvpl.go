package dvpl

import (
	"bytes"
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

func ReadDVPLFooterData(buf []byte) (DVPLFooterData, error) {
	footerBuf := buf[len(buf)-20:]
	dvplTypeBuf := footerBuf[len(footerBuf)-4:]
	if string(dvplTypeBuf) != "DVPL" {
		return DVPLFooterData{}, DVPLConverterError("Invalid filetype in footer data.")
	}

	footerData := DVPLFooterData{
		OriginalSize:   binary.LittleEndian.Uint32(footerBuf[0:4]),
		CompressedSize: binary.LittleEndian.Uint32(footerBuf[4:8]),
		CRC32:          binary.LittleEndian.Uint32(footerBuf[8:12]),
		CompressType:   binary.LittleEndian.Uint32(footerBuf[12:16]),
	}

	return footerData, nil
}

func toDVPLFooter(fd DVPLFooterData) []byte {
	res := make([]byte, 0, 20)
	res = binary.LittleEndian.AppendUint32(
		binary.LittleEndian.AppendUint32(
			binary.LittleEndian.AppendUint32(
				binary.LittleEndian.AppendUint32(
					res, fd.OriginalSize),
				fd.CompressedSize),
			fd.CRC32),
		fd.CompressType)
	res = append(res, []byte("DVPL")...)
	return res
}

func UncompressDVPL(buf []byte) ([]byte, error) {
	footerData, err := ReadDVPLFooterData(buf)
	if err != nil {
		return nil, err
	}
	targetBuf := buf[0 : len(buf)-20]
	if len(targetBuf) != int(footerData.CompressedSize) {
		return nil, DVPLConverterError("Compressed size mismatch with the footer data.")
	}
	if crc32.ChecksumIEEE(targetBuf) != footerData.CRC32 {
		return nil, DVPLConverterError("CRC32 hash sum mismatch with the footer data.")
	}

	if footerData.CompressType == 0 && footerData.OriginalSize != footerData.CompressedSize {
		return nil, DVPLConverterError("Compression type is 0 but the original size mismatches compressed size.")
	} else if footerData.CompressType == 0 {
		return targetBuf, nil
	} else if footerData.CompressType <= 2 {
		uncompressedBuf := make([]byte, footerData.OriginalSize)
		_, err := lz4.UncompressBlock(targetBuf, uncompressedBuf)
		if err != nil {
			return nil, DVPLConverterError("Failed to uncompress the buffer. More:\n" + err.Error())
		}

		i := len(uncompressedBuf) - 1
		for ; uncompressedBuf[i] == '\x00'; i-- {
		}

		return uncompressedBuf[:i+1], nil
	} else {
		return nil, DVPLConverterError(fmt.Sprintf("Not a DVPL level of compression. The maximum level of compression used in DVPL is 2, but met %d.", footerData.CompressType))
	}
}

func CompressDVPL(buf []byte, isTex bool) ([]byte, error) {
	compressionType := 2
	compressor := lz4.CompressorHC{Level: lz4.Level2}
	if isTex {
		compressionType = 0
		compressor.Level = lz4.Fast
	}
	leastCompressableSize := lz4.CompressBlockBound(len(buf))
	if leastCompressableSize > len(buf) {
		zeros := bytes.Repeat([]byte{0}, leastCompressableSize-len(buf))
		buf = append(buf, zeros...)
	}
	compressedBuf := make([]byte, len(buf))
	i, err := compressor.CompressBlock(buf, compressedBuf)
	if err != nil {
		return nil, DVPLConverterError("Failed to compress the buffer. More:\n" + err.Error())
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

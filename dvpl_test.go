package dvpl

import (
	"strings"
	"testing"
)

func TestReadDVPLFooterData(t *testing.T) {
	buffer := toDVPLFooter(DVPLFooterData{
		OriginalSize:   125,
		CompressedSize: 42,
		CRC32:          6122,
		CompressType:   2,
	})
	fd, err := ReadDVPLFooterData(buffer)
	if err != nil {
		t.Error(err)
	}
	if fd.OriginalSize != 125 {
		t.Errorf("Given original size = 125, received %d.", fd.OriginalSize)
	}
	if fd.CompressedSize != 42 {
		t.Errorf("Given compressed size = 42, received %d.", fd.CompressedSize)
	}
	if fd.CRC32 != 6122 {
		t.Errorf("Given CRC32 = 6122, received %d.", fd.CRC32)
	}
	if fd.CompressType != 2 {
		t.Errorf("Given compression type = 2, received %d.", fd.CompressType)
	}
}

func TestCompressDVPL(t *testing.T) {
	buffer := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nunc et nisl massa.")
	res, err := CompressDVPL(buffer, false)
	if err != nil {
		t.Errorf("Encountered an internal converter error:\n%s", err)
	}

	if strings.Compare(string(res), "\xf4>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nunc et nisl massa.\x00\x01\x00p\x00\x00\x00\x00\x00\x00\x00\\\x00\x00\x00Y\x00\x00\x00\x95\x16+&\x02\x00\x00\x00DVPL") != 0 {
		t.Errorf("Given the lorem ipsum input string, received faulty compressed string:\n%s", string(res))
	}
}

func TestUncompressDVPL(t *testing.T) {
	buffer := []byte("\xf4>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nunc et nisl massa.\x00\x01\x00p\x00\x00\x00\x00\x00\x00\x00\\\x00\x00\x00Y\x00\x00\x00\x95\x16+&\x02\x00\x00\x00DVPL")
	res, err := UncompressDVPL(buffer)
	if err != nil {
		t.Error(err)
	}

	if strings.Compare(string(res), "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nunc et nisl massa.") != 0 {
		t.Errorf(`Given compressed lorem ipsum string, received faulty uncompressed string: %s`, res)
	}
}

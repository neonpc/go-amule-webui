package ec

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"unicode/utf8"
)

type TagType uint8

const (
	TagTypeUnknown  TagType = 0
	TagTypeCustom   TagType = 1
	TagTypeUint8    TagType = 2
	TagTypeUint16   TagType = 3
	TagTypeUint32   TagType = 4
	TagTypeUint64   TagType = 5
	TagTypeString   TagType = 6
	TagTypeDouble   TagType = 7
	TagTypeIPV4     TagType = 8
	TagTypeHash16   TagType = 9
	TagTypeUint128  TagType = 10
)

type TagValue struct {
	Type TagType
	Data interface{}
}

func (v TagValue) String() string {
	switch val := v.Data.(type) {
	case string:
		return val
	case uint8:
		return fmt.Sprintf("%d", val)
	case uint16:
		return fmt.Sprintf("%d", val)
	case uint32:
		return fmt.Sprintf("%d", val)
	case uint64:
		return fmt.Sprintf("%d", val)
	case [16]byte:
		return fmt.Sprintf("%x", val)
	case net.IP:
		return val.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

type DetailLevel uint32

const (
	DetailCMD    DetailLevel = 0
	DetailWEB    DetailLevel = 1
	DetailFULL   DetailLevel = 2
	DetailUPDATE DetailLevel = 3
)

const (
	ProtocolVersion uint32 = 0x0204
)

func readUint8(r io.Reader) (uint8, error) {
	var v uint8
	return v, binary.Read(r, binary.BigEndian, &v)
}

func readUint16(r io.Reader) (uint16, error) {
	var v uint16
	return v, binary.Read(r, binary.BigEndian, &v)
}

func readUint32(r io.Reader) (uint32, error) {
	var v uint32
	return v, binary.Read(r, binary.BigEndian, &v)
}

func readUint64(r io.Reader) (uint64, error) {
	var v uint64
	return v, binary.Read(r, binary.BigEndian, &v)
}

func readBytes(r io.Reader, n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := io.ReadFull(r, buf)
	return buf, err
}

func readString(r io.Reader) (string, error) {
	var buf []byte
	tmp := make([]byte, 1)
	for {
		_, err := io.ReadFull(r, tmp)
		if err != nil {
			return "", err
		}
		if tmp[0] == 0 {
			break
		}
		buf = append(buf, tmp[0])
	}
	return string(buf), nil
}

func readNum(r io.Reader, utf8 bool, bitSize int) (uint64, error) {
	if utf8 {
		cp, err := readUTF8Rune(r)
		if err != nil {
			return 0, err
		}
		return uint64(cp), nil
	}
	switch bitSize {
	case 8:
		v, err := readUint8(r)
		return uint64(v), err
	case 16:
		v, err := readUint16(r)
		return uint64(v), err
	case 32:
		v, err := readUint32(r)
		return uint64(v), err
	case 64:
		v, err := readUint64(r)
		return uint64(v), err
	}
	return 0, fmt.Errorf("unsupported bit size: %d", bitSize)
}

func readUTF8Rune(reader io.Reader) (rune, error) {
	var buf [6]byte
	if _, err := io.ReadFull(reader, buf[:1]); err != nil {
		return 0, err
	}

	contBytes, err := utf8MBCount(buf[0])
	if err != nil {
		return 0, err
	}

	if contBytes > 0 {
		if _, err := io.ReadFull(reader, buf[1:1+contBytes]); err != nil {
			return 0, err
		}
	}

	if contBytes == 0 {
		return rune(buf[0]), nil
	}

	cp, _ := utf8.DecodeRune(buf[:1+contBytes])
	return cp, nil
}

func utf8MBCount(lead byte) (int, error) {
	switch {
	case lead < 0x80:
		return 0, nil
	case lead < 0xC0:
		return 0, fmt.Errorf("unexpected continuation byte: 0x%02x", lead)
	case lead < 0xE0:
		return 1, nil
	case lead < 0xF0:
		return 2, nil
	case lead < 0xF8:
		return 3, nil
	default:
		return 0, fmt.Errorf("invalid UTF-8 leading byte: 0x%02x", lead)
	}
}

func utf8EncodedLen(v uint64) int {
	switch {
	case v < 0x80:
		return 1
	case v < 0x800:
		return 2
	case v < 0x10000:
		return 3
	case v < 0x200000:
		return 4
	default:
		return 5
	}
}

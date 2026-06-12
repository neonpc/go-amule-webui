package ec

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
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

package ec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type Tag struct {
	Name     TagName
	Type     TagType
	Data     interface{}
	Children []Tag
}

func (t *Tag) String() string {
	if len(t.Children) > 0 {
		return fmt.Sprintf("{%s (%d children)}", t.Name, len(t.Children))
	}
	return fmt.Sprintf("{%s: %v}", t.Name, t.Data)
}

func (t *Tag) ChildByName(name TagName) *Tag {
	for i := range t.Children {
		if t.Children[i].Name == name {
			return &t.Children[i]
		}
	}
	return nil
}

func (t *Tag) ChildrenByName(name TagName) []Tag {
	var result []Tag
	for i := range t.Children {
		if t.Children[i].Name == name {
			result = append(result, t.Children[i])
		}
	}
	return result
}

func (t *Tag) StringValue() string {
	if s, ok := t.Data.(string); ok {
		return s
	}
	return ""
}

func (t *Tag) UintValue() uint64 {
	switch v := t.Data.(type) {
	case uint8:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint32:
		return uint64(v)
	case uint64:
		return v
	}
	return 0
}

func (t *Tag) HashValue() [16]byte {
	switch v := t.Data.(type) {
	case [16]byte:
		return v
	case []byte:
		if len(v) >= 16 {
			var h [16]byte
			copy(h[:], v[:16])
			return h
		}
	}
	return [16]byte{}
}

func writeTag(w io.Writer, tag *Tag) error {
	nameVal := uint16(tag.Name) << 1
	hasChildren := len(tag.Children) > 0
	if hasChildren {
		nameVal |= 1
	}

	if err := binary.Write(w, binary.BigEndian, nameVal); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, tag.Type); err != nil {
		return err
	}

	var dataBuf []byte
	var subCount uint16

	if hasChildren {
		subCount = uint16(len(tag.Children))
	}

	switch tag.Type {
	case TagTypeUint8:
		dataBuf = make([]byte, 1)
		dataBuf[0] = byte(tag.Data.(uint8))
	case TagTypeUint16:
		dataBuf = make([]byte, 2)
		binary.BigEndian.PutUint16(dataBuf, tag.Data.(uint16))
	case TagTypeUint32:
		dataBuf = make([]byte, 4)
		binary.BigEndian.PutUint32(dataBuf, tag.Data.(uint32))
	case TagTypeUint64:
		dataBuf = make([]byte, 8)
		binary.BigEndian.PutUint64(dataBuf, tag.Data.(uint64))
	case TagTypeString:
		s := tag.Data.(string)
		dataBuf = append([]byte(s), 0)
	case TagTypeHash16:
		var h [16]byte
		switch v := tag.Data.(type) {
		case [16]byte:
			h = v
		case []byte:
			copy(h[:], v)
		}
		dataBuf = h[:]
	case TagTypeIPV4:
		addr := tag.Data.(string)
		parts := strings.Split(addr, ":")
		ip := net.ParseIP(parts[0]).To4()
		p, _ := strconv.ParseUint(parts[1], 10, 16)
		port := uint16(p)
		dataBuf = make([]byte, 6)
		copy(dataBuf, ip)
		binary.BigEndian.PutUint16(dataBuf[4:], port)
	case TagTypeCustom:
		dataBuf = tag.Data.([]byte)
	default:
		dataBuf = []byte{}
	}

	totalLen := len(dataBuf)
	if hasChildren {
		for _, child := range tag.Children {
			totalLen += tagEncodedLen(&child)
		}
	}

	if err := binary.Write(w, binary.BigEndian, uint32(totalLen)); err != nil {
		return err
	}

	if hasChildren {
		if err := binary.Write(w, binary.BigEndian, subCount); err != nil {
			return err
		}
		for i := range tag.Children {
			if err := writeTag(w, &tag.Children[i]); err != nil {
				return err
			}
		}
	}

	if len(dataBuf) > 0 {
		if _, err := w.Write(dataBuf); err != nil {
			return err
		}
	}

	return nil
}

func tagEncodedLen(tag *Tag) int {
	nameLen := 2
	typeLen := 1
	lenFieldLen := 4
	base := nameLen + typeLen + lenFieldLen

	var dataLen int
	switch tag.Type {
	case TagTypeUint8:
		dataLen = 1
	case TagTypeUint16:
		dataLen = 2
	case TagTypeUint32:
		dataLen = 4
	case TagTypeUint64:
		dataLen = 8
	case TagTypeString:
		dataLen = len(tag.Data.(string)) + 1
	case TagTypeHash16:
		dataLen = 16
	case TagTypeIPV4:
		dataLen = 6
	case TagTypeCustom:
		dataLen = len(tag.Data.([]byte))
	}

	if len(tag.Children) > 0 {
		dataLen += 2
		for _, child := range tag.Children {
			dataLen += tagEncodedLen(&child)
		}
	}

	return base + dataLen
}

func readTag(r *bytes.Reader, useUTF8 bool) (Tag, error) {
	var tag Tag

	nameRaw, err := readNum(r, useUTF8, 16)
	if err != nil {
		return tag, err
	}

	hasChildren := (nameRaw & 1) != 0
	tag.Name = TagName(nameRaw >> 1)

	tagType, err := readNum(r, useUTF8, 8)
	if err != nil {
		return tag, err
	}
	tag.Type = TagType(tagType)

	tagLen, err := readNum(r, useUTF8, 32)
	if err != nil {
		return tag, err
	}

	if hasChildren {
		subCount, err := readNum(r, useUTF8, 16)
		if err != nil {
			return tag, err
		}
		for i := uint64(0); i < subCount; i++ {
			child, err := readTag(r, useUTF8)
			if err != nil {
				break
			}
			tag.Children = append(tag.Children, child)
		}
	}

	// Compute children's fixed-size wire length (matching C++ GetTagLen)
	// so that dataLen = tagLen - childrenLen gives the correct raw data size.
	var childrenLen uint64
	for i := range tag.Children {
		childrenLen += uint64(tagWireDataLen(&tag.Children[i]))
		childrenLen += 2 + 1 + 4 // sizeof(ec_tagname_t) + sizeof(ec_tagtype_t) + sizeof(ec_taglen_t)
		if len(tag.Children[i].Children) > 0 {
			childrenLen += 2 // sizeof(uint16) for subCount
		}
	}

	dataRemaining := int64(tagLen) - int64(childrenLen)
	if dataRemaining < 0 {
		dataRemaining = 0
	}
	if dataRemaining > int64(r.Len()) {
		dataRemaining = int64(r.Len())
	}

	if dataRemaining > 0 {
		dataBuf := make([]byte, dataRemaining)
		io.ReadFull(r, dataBuf)

		switch tag.Type {
		case TagTypeUint8:
			if len(dataBuf) >= 1 {
				tag.Data = dataBuf[0]
			}
		case TagTypeUint16:
			if len(dataBuf) >= 2 {
				tag.Data = binary.BigEndian.Uint16(dataBuf[:2])
			}
		case TagTypeUint32:
			if len(dataBuf) >= 4 {
				tag.Data = binary.BigEndian.Uint32(dataBuf[:4])
			}
		case TagTypeUint64:
			if len(dataBuf) >= 8 {
				tag.Data = binary.BigEndian.Uint64(dataBuf[:8])
			}
		case TagTypeString:
			if len(dataBuf) > 0 {
				tag.Data = string(dataBuf[:len(dataBuf)-1])
			}
		case TagTypeHash16:
			if len(dataBuf) >= 16 {
				var h [16]byte
				copy(h[:], dataBuf[:16])
				tag.Data = h
			}
		case TagTypeIPV4:
			if len(dataBuf) >= 6 {
				ip := net.IP(dataBuf[:4])
				port := binary.BigEndian.Uint16(dataBuf[4:6])
				tag.Data = fmt.Sprintf("%s:%d", ip, port)
			}
		case TagTypeCustom:
			tag.Data = dataBuf
		default:
			tag.Data = dataBuf
		}
	}

	return tag, nil
}

// tagWireDataLen computes the C++ GetTagLen() equivalent:
// data + children (excluding own header).
func tagWireDataLen(tag *Tag) int {
	var length int
	switch tag.Type {
	case TagTypeUint8:
		length = 1
	case TagTypeUint16:
		length = 2
	case TagTypeUint32:
		length = 4
	case TagTypeUint64:
		length = 8
	case TagTypeString:
		if s, ok := tag.Data.(string); ok {
			length = len(s) + 1
		}
	case TagTypeHash16:
		length = 16
	case TagTypeIPV4:
		length = 6
	case TagTypeUint128:
		length = 16
	default:
		if buf, ok := tag.Data.([]byte); ok {
			length = len(buf)
		}
	}
	for i := range tag.Children {
		length += tagWireDataLen(&tag.Children[i])
		length += 2 + 1 + 4
		if len(tag.Children[i].Children) > 0 {
			length += 2
		}
	}
	return length
}

func newStringTag(name TagName, val string) Tag {
	return Tag{Name: name, Type: TagTypeString, Data: val}
}

func newUint8Tag(name TagName, val uint8) Tag {
	return Tag{Name: name, Type: TagTypeUint8, Data: val}
}

func newUint16Tag(name TagName, val uint16) Tag {
	return Tag{Name: name, Type: TagTypeUint16, Data: val}
}

func newUint32Tag(name TagName, val uint32) Tag {
	return Tag{Name: name, Type: TagTypeUint32, Data: val}
}

func newUint64Tag(name TagName, val uint64) Tag {
	return Tag{Name: name, Type: TagTypeUint64, Data: val}
}

func newContainerTag(name TagName, children ...Tag) Tag {
	return Tag{Name: name, Type: TagTypeCustom, Children: children}
}

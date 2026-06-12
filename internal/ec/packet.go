package ec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type PacketFlags uint32

const (
	FlagZLIB          PacketFlags = 0x01
	FlagUTF8Numbers   PacketFlags = 0x02
	FlagHasID         PacketFlags = 0x04
	FlagLargeTagCount PacketFlags = 0x10
	FlagAlways1       PacketFlags = 0x20
)

type Packet struct {
	OpCode OpCode
	Tags   []Tag
	ID     uint32
}

func (p *Packet) String() string {
	return fmt.Sprintf("{%s: %d tags}", p.OpCode, len(p.Tags))
}

func (p *Packet) TagByName(name TagName) *Tag {
	for i := range p.Tags {
		if p.Tags[i].Name == name {
			return &p.Tags[i]
		}
	}
	return nil
}

func (p *Packet) ChildrenByName(name TagName) []Tag {
	var result []Tag
	for i := range p.Tags {
		if p.Tags[i].Name == name {
			result = append(result, p.Tags[i].Children...)
		}
	}
	return result
}

func (p *Packet) Encode() ([]byte, error) {
	var appBuf bytes.Buffer

	if err := binary.Write(&appBuf, binary.BigEndian, uint8(p.OpCode)); err != nil {
		return nil, err
	}

	tagCount := len(p.Tags)
	if tagCount > 0xFFFF {
		if err := binary.Write(&appBuf, binary.BigEndian, uint16(0xFFFF)); err != nil {
			return nil, err
		}
		if err := binary.Write(&appBuf, binary.BigEndian, uint32(tagCount)); err != nil {
			return nil, err
		}
	} else {
		if err := binary.Write(&appBuf, binary.BigEndian, uint16(tagCount)); err != nil {
			return nil, err
		}
	}

	for i := range p.Tags {
		if err := writeTag(&appBuf, &p.Tags[i]); err != nil {
			return nil, err
		}
	}

	appData := appBuf.Bytes()

	flags := PacketFlags(FlagAlways1)

	var headerBuf bytes.Buffer
	if err := binary.Write(&headerBuf, binary.LittleEndian, uint32(flags)); err != nil {
		return nil, err
	}
	if err := binary.Write(&headerBuf, binary.LittleEndian, uint32(len(appData))); err != nil {
		return nil, err
	}

	return append(headerBuf.Bytes(), appData...), nil
}

func DecodePacket(data []byte) (*Packet, error) {
	r := bytes.NewReader(data)

	flagsRaw, err := readUint32LE(r)
	if err != nil {
		return nil, fmt.Errorf("read flags: %w", err)
	}
	flags := PacketFlags(flagsRaw)

	appLen, err := readUint32LE(r)
	if err != nil {
		return nil, fmt.Errorf("read app length: %w", err)
	}

	appData := make([]byte, appLen)
	if _, err := io.ReadFull(r, appData); err != nil {
		return nil, fmt.Errorf("read app data: %w", err)
	}

	if flags&FlagZLIB != 0 {
		return nil, fmt.Errorf("zlib compression not implemented")
	}

	ar := bytes.NewReader(appData)

	opRaw, err := readUint8(ar)
	if err != nil {
		return nil, fmt.Errorf("read opcode: %w", err)
	}
	packet := &Packet{OpCode: OpCode(opRaw)}

	tagCountRaw, err := readUint16(ar)
	if err != nil {
		return nil, fmt.Errorf("read tag count: %w", err)
	}
	tagCount := uint32(tagCountRaw)

	if tagCountRaw == 0xFFFF && flags&FlagLargeTagCount != 0 {
		extCount, err := readUint32(ar)
		if err != nil {
			return nil, fmt.Errorf("read extended tag count: %w", err)
		}
		tagCount = extCount
	}

	for i := uint32(0); i < tagCount; i++ {
		tag, err := readTag(ar)
		if err != nil {
			return nil, fmt.Errorf("read tag %d: %w", i, err)
		}
		packet.Tags = append(packet.Tags, tag)
	}

	return packet, nil
}

func readUint32LE(r io.Reader) (uint32, error) {
	var v uint32
	return v, binary.Read(r, binary.LittleEndian, &v)
}

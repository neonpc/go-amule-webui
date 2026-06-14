package ec

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Conn struct {
	conn     net.Conn
	reader   *bufio.Reader
	mu       sync.Mutex
	password string
	host     string
	port     int
	version  string
	canUTF8  bool
	timeout  time.Duration
}

func Dial(host string, port int, password string, timeout time.Duration) (*Conn, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	c := &Conn{
		conn:     conn,
		reader:   bufio.NewReaderSize(conn, 65536),
		password: password,
		host:     host,
		port:     port,
		timeout:  timeout,
	}
	return c, nil
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) Authenticate() error {
	req := buildAuthRequest()
	data, err := req.Encode()
	if err != nil {
		return fmt.Errorf("encode auth: %w", err)
	}

	c.mu.Lock()
	if _, err := c.conn.Write(data); err != nil {
		c.mu.Unlock()
		return fmt.Errorf("write auth: %w", err)
	}
	c.mu.Unlock()

	resp, err := c.ReadPacket()
	if err != nil {
		return fmt.Errorf("read auth response: %w", err)
	}

	switch resp.OpCode {
	case OpAuthOK:
		if tag := resp.TagByName(TagServerVersion); tag != nil {
			c.version = tag.StringValue()
		}
		return nil
	case OpAuthSalt:
		saltTag := resp.TagByName(TagPasswdSalt)
		if saltTag == nil {
			return fmt.Errorf("auth salt without salt data")
		}
		saltVal, ok := saltTag.Data.(uint64)
		if !ok {
			return fmt.Errorf("auth salt has unexpected type: %T", saltTag.Data)
		}
		saltResp := buildSaltAuthResponse(c.password, saltVal)
		data, err := saltResp.Encode()
		if err != nil {
			return fmt.Errorf("encode salt response: %w", err)
		}
		c.mu.Lock()
		if _, err := c.conn.Write(data); err != nil {
			c.mu.Unlock()
			return fmt.Errorf("write salt response: %w", err)
		}
		c.mu.Unlock()

		finalResp, err := c.ReadPacket()
		if err != nil {
			return fmt.Errorf("read final auth: %w", err)
		}
		if finalResp.OpCode == OpAuthOK {
			return nil
		}
		return fmt.Errorf("auth failed after salt challenge")
	case OpAuthFail:
		return fmt.Errorf("authentication failed: wrong password")
	default:
		return fmt.Errorf("unexpected auth response: %s", resp.OpCode)
	}
}

func (c *Conn) SendRequest(packet *Packet) (*Packet, error) {
	data, err := packet.Encode()
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}

	c.mu.Lock()
	if _, err := c.conn.Write(data); err != nil {
		c.mu.Unlock()
		return nil, fmt.Errorf("write: %w", err)
	}
	c.mu.Unlock()

	return c.ReadPacket()
}

func (c *Conn) ReadPacket() (*Packet, error) {
	_ = c.conn.SetReadDeadline(time.Now().Add(c.timeout))
	flagsBuf := make([]byte, 4)
	if _, err := io.ReadFull(c.reader, flagsBuf); err != nil {
		return nil, fmt.Errorf("read flags: %w", err)
	}

	flags := binary.BigEndian.Uint32(flagsBuf)

	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(c.reader, lenBuf); err != nil {
		return nil, fmt.Errorf("read length: %w", err)
	}
	appLen := binary.BigEndian.Uint32(lenBuf)

	appData := make([]byte, appLen)
	if _, err := io.ReadFull(c.reader, appData); err != nil {
		return nil, fmt.Errorf("read app data: %w", err)
	}

	fullPacket := append(flagsBuf, lenBuf...)
	fullPacket = append(fullPacket, appData...)

	packet, err := DecodePacket(fullPacket)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	c.canUTF8 = c.canUTF8 || (PacketFlags(flags)&FlagUTF8Numbers != 0)
	return packet, nil
}

func (c *Conn) Version() string {
	return c.version
}

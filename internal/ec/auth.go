package ec

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

func MD5Hash(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

func buildAuthRequest() *Packet {
	return &Packet{
		OpCode: OpAuthReq,
		Tags: []Tag{
			newUint32Tag(TagProtocolVersion, ProtocolVersion),
			newStringTag(TagClientName, "go-amule-webui"),
			newStringTag(TagClientVersion, "1.0.0"),
			newUint8Tag(TagCanUTF8Numbers, 1),
		},
	}
}

func buildSaltAuthResponse(password string, saltVal uint64) *Packet {
	storedHash := MD5Hash(password)
	saltHex := fmt.Sprintf("%X", saltVal)
	saltHash := MD5Hash(saltHex)
	finalHashHex := MD5Hash(storedHash + saltHash)

	hashBytes, _ := hex.DecodeString(finalHashHex)
	var h [16]byte
	copy(h[:], hashBytes)

	return &Packet{
		OpCode: OpAuthPasswd,
		Tags: []Tag{
			{Name: TagPasswdHash, Type: TagTypeHash16, Data: h},
		},
	}
}

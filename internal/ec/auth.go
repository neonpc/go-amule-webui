package ec

import (
	"crypto/md5"
	"fmt"
)

func MD5Hash(pwd string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(pwd)))
}

func buildAuthRequest(password string) *Packet {
	hashed := MD5Hash(password)

	return &Packet{
		OpCode: OpAuthReq,
		Tags: []Tag{
			newUint32Tag(TagProtocolVersion, ProtocolVersion),
			newStringTag(TagPasswdHash, hashed),
			newStringTag(TagClientName, "go-amule-webui"),
			newStringTag(TagClientVersion, "1.0.0"),
			newUint8Tag(TagCanZLib, 0),
			newUint8Tag(TagCanUTF8Numbers, 0),
		},
	}
}

func buildSaltAuthResponse(password, salt string) *Packet {
	hashed := MD5Hash(password + salt)

	return &Packet{
		OpCode: OpAuthPasswd,
		Tags: []Tag{
			newStringTag(TagPasswdHash, hashed),
		},
	}
}

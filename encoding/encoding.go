package encoding

import (
	"bytes"
	"encoding/base64"

	"github.com/vmihailenco/msgpack"
)

func decodeSessionBase64() {

}

// EncodeBase64 takes a byte array and base 64 encodes it
func EncodeBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

func decodeSessionMsgPack() {

}

// EncodeMsgPack performs message pack encryption
// Currently this takes a map[string]interface{} parameter because we only
// want to message pack encode JSON objects
func EncodeMsgPack(data map[string]interface{}) ([]byte, error) {
	var encoded []byte
	encBuf := bytes.NewBuffer(encoded)
	enc := msgpack.NewEncoder(encBuf)

	if err := enc.Encode(data); err != nil {
		return nil, err
	}

	return encBuf.Bytes(), nil
}

package encoding

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"

	"github.com/vmihailenco/msgpack"
)

//DecodeBase64 takes a base64-encoded string and decodes it to a []byte.
func DecodeBase64(base64Encoded string) ([]byte, error) {
	base64Decoded, err := base64.StdEncoding.DecodeString(base64Encoded)

	return base64Decoded, err
}

// EncodeBase64 takes a byte array and base 64 encodes it
func EncodeBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

//DecodeMsgPack takes a msgpack'd []byte and decodes it to json.
func DecodeMsgPack(msgpackEncoded []byte) (map[string]interface{}, error) {
	var decoded map[string]interface{}

	dec := msgpack.NewDecoder(bytes.NewBuffer(msgpackEncoded))
	err := dec.Decode(&decoded)

	return decoded, err
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

//GenerateSha1Sum generates a sha1 sum for a given []byte.
func GenerateSha1Sum(sum []byte) [20]byte {
	return sha1.Sum(sum)
}

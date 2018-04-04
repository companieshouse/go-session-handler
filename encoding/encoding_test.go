package encoding

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack"
)

// ------------------- Routes Through DecodeBase64() -------------------

// TestDecodeBase64 - Verify no errors are thrown when DecodeBase64 is called on
// a base64 encoded string
func TestDecodeBase64(t *testing.T) {
	assert := assert.New(t)

	encoder := New()

	test := []byte("Hello, world!")

	encodedString := base64.StdEncoding.EncodeToString(test)

	decoded, err := encoder.DecodeBase64(encodedString)

	assert.Nil(err)
	assert.Equal(test, decoded)
}

// ------------------- Routes Through EncodeBase64() -------------------

// TestEncodeBase64 - Verify no errors are thrown when we base64 decode a string
// which had previously been encoded
func TestEncodeBase64(t *testing.T) {
	assert := assert.New(t)

	encoder := New()

	test := []byte("Hello, world!")

	encodedString := encoder.EncodeBase64(test)

	decoded, err := base64.StdEncoding.DecodeString(encodedString)

	assert.Nil(err)
	assert.Equal(test, decoded)
}

// ------------------- Routes Through DecodeMsgPack() -------------------

// TestDecodeMsgPack - Verify no errors are thrown when DecodeMsgPack is called
// on previously message pack encoded data
func TestDecodeMsgPack(t *testing.T) {
	assert := assert.New(t)

	encoder := New()

	test := map[string]interface{}{"test": "hello, world!"}

	var encoded []byte
	encBuf := bytes.NewBuffer(encoded)
	enc := msgpack.NewEncoder(encBuf)
	enc.Encode(test)
	encodedBytes := encBuf.Bytes()

	decoded, err := encoder.DecodeMsgPack(encodedBytes)

	assert.Nil(err)
	assert.Equal(test, decoded)
}

// ------------------- Routes Through EncodeMsgPack() -------------------

// TestEncodeMsgPack - Verify no errors are thrown when EncodeMsgPack is called
// and subsequently decoded
func TestEncodeMsgPack(t *testing.T) {
	assert := assert.New(t)

	encoder := New()

	test := map[string]interface{}{"test": "hello, world!"}

	encodedBytes, err := encoder.EncodeMsgPack(test)

	assert.Nil(err)

	var decoded map[string]interface{}
	dec := msgpack.NewDecoder(bytes.NewBuffer(encodedBytes))
	err = dec.Decode(&decoded)

	assert.Nil(err)
	assert.Equal(test, decoded)
}

package encoding

import (
	"bytes"
	"encoding/base64"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/vmihailenco/msgpack"
)

// ------------------- Routes Through DecodeBase64() -------------------

// TestDecodeBase64 - Verify no errors are thrown when DecodeBase64 is called on
// a base64 encoded string
func TestUnitDecodeBase64(t *testing.T) {

	Convey("Given I base64 encode a test byte array", t, func() {

		test := []byte("foo")
		encodedString := base64.StdEncoding.EncodeToString(test)

		Convey("When I call DecodeBase64 on the result", func() {

			decoded, err := DecodeBase64(encodedString)

			Convey("Then I expect the decoded byte array to be returned, with no errors", func() {

				So(err, ShouldBeNil)
				So(string(test), ShouldEqual, string(decoded))
			})
		})
	})
}

// ------------------- Routes Through EncodeBase64() -------------------

// TestEncodeBase64 - Verify no errors are thrown when we base64 decode a string
// which had previously been encoded
func TestUnitEncodeBase64(t *testing.T) {

	Convey("Given I call EncodeBase64 on a byte array", t, func() {

		test := []byte("foo")
		encodedString := EncodeBase64(test)

		Convey("When I manually base64 decode the result", func() {

			decoded, err := base64.StdEncoding.DecodeString(encodedString)

			Convey("Then I expect the decoded byte array to be returned, with no errors", func() {

				So(err, ShouldBeNil)
				So(string(test), ShouldEqual, string(decoded))
			})
		})
	})
}

// ------------------- Routes Through DecodeMsgPack() -------------------

// TestDecodeMsgPack - Verify no errors are thrown when DecodeMsgPack is called
// on previously message pack encoded data
func TestUnitDecodeMsgPack(t *testing.T) {

	Convey("Given I message pack encode some JSON data", t, func() {

		test := map[string]interface{}{"foo": "bar"}

		var encoded []byte
		encBuf := bytes.NewBuffer(encoded)
		enc := msgpack.NewEncoder(encBuf)
		_ = enc.Encode(test)
		encodedBytes := encBuf.Bytes()

		Convey("When I call DecodeMsgPack on the result", func() {

			decoded, err := DecodeMsgPack(encodedBytes)

			Convey("Then I expect the JSON to be returned, with no errors", func() {

				So(err, ShouldBeNil)
				So(test["foo"], ShouldEqual, decoded["foo"])
			})
		})
	})
}

// ------------------- Routes Through EncodeMsgPack() -------------------

// TestEncodeMsgPack - Verify no errors are thrown when EncodeMsgPack is called
// and subsequently decoded
func TestUnitEncodeMsgPack(t *testing.T) {

	Convey("Given I call EncodeMsgPack on some JSON data", t, func() {

		test := map[string]interface{}{"foo": "bar"}

		encodedBytes, err := EncodeMsgPack(test)

		Convey("Then no errors should be returned", func() {

			So(err, ShouldBeNil)

			Convey("When I manually messagepack decode the result", func() {

				var decoded map[string]interface{}
				dec := msgpack.NewDecoder(bytes.NewBuffer(encodedBytes))
				err = dec.Decode(&decoded)

				Convey("Then I expect the JSON to be returned, with no errors", func() {

					So(err, ShouldBeNil)
					So(test["foo"], ShouldEqual, decoded["foo"])
				})
			})
		})
	})
}

func TestUnitGenerateSha1Sum(t *testing.T) {
	Convey("test GenerateSha1Sum", t, func() {
		sum := GenerateSha1Sum([]byte("test"))
		So(sum, ShouldEqual, [20]uint8{169, 74, 143, 229, 204, 177, 155, 166, 28, 76, 8, 115, 211, 145, 233, 135, 152, 47, 187, 211})
	})
}

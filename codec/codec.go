// Package codec implements codecs for marshaling/unmarshaling of native Go types into byte slices
package codec

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"errors"

	"gopkg.in/mgo.v2/bson"
)

// Codec defines an interface for encoding/decoding Go values to bytes
type Codec struct {
	Marshal   func(interface{}) ([]byte, error)
	Unmarshal func([]byte, interface{}) error
}

var (
	// Gob is a Codec that uses the gob package
	Gob = Codec{gobMarshal, gobUnmarshal}
	// JSON is a Codec that uses the json package.
	JSON = Codec{json.Marshal, json.Unmarshal}
	// XML is a codec that uses the xml pacakge
	XML = Codec{xml.Marshal, xml.Unmarshal}
	// BSON is a codec that used the labix.org/v2/mgo/bson pacakge
	BSON = Codec{bson.Marshal, bson.Unmarshal}
	// ErrTestCodec is a codec that returns errors, used for testing other packages.
	ErrTestCodec = Codec{testMarshalErr, testUnmarshalErr}
)

func gobMarshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gobUnmarshal(data []byte, v interface{}) error {
	return gob.NewDecoder(bytes.NewBuffer(data)).Decode(v)
}

// testMarshalErr is a marshal func that always returns an error. Used for testing.
func testMarshalErr(v interface{}) ([]byte, error) {
	return nil, errors.New("test error")
}

// testUnmarshalErr is a unmarshal func that always returns an error. Used for testing.
func testUnmarshalErr(data []byte, v interface{}) error {
	return errors.New("test error")
}

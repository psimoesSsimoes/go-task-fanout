package app

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
)

type util struct{}

// Util is a collection of utiliy tools.
var Util = new(util)

// GenerateUniqueID generates a base64 hash of the provided strings, best used as a deterministic ID generator.
func (*util) GenerateUniqueID(data ...string) string {
	hash := sha1.New()

	for _, str := range data {
		hash.Write([]byte(str))
	}

	return base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
}

// Assert implementation for quick panics.
func (*util) Assert(val bool, msg string) {
	if !val {
		panic(msg)
	}
}

// LoadFile reads the entire contents of a file into a byte slice.
func (*util) LoadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// MustSerialize (to JSON) panics on serialization error.
func (*util) MustSerialize(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return data
}

// ShouldSerialize (to JSON) returns "SERIALIZATION_ERROR" on serialization error.
func (*util) ShouldSerialize(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		return []byte("SERIALIZATION_ERROR")
	}

	return data
}

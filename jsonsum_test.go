package jsonsum

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/blake2b"
	"hash"
	"hash/crc32"
	"io"
	"strings"
	"testing"
)

func sumCRC32(r io.Reader) (uint32, error) {
	if digest, err := Sum(r, func() hash.Hash { return crc32.NewIEEE() }); err != nil {
		return 0, err
	} else {
		return digest.(hash.Hash32).Sum32(), nil
	}
}

func TestJsonSum_CRC32(t *testing.T) {
	tests := []struct {
		name string
		j    []string
		sum  uint32
	}{
		// strings
		{name: "empty string", j: []string{`""`}, sum: 2018536706},
		{name: "short string", j: []string{`"hi"`}, sum: 3572422966},
		{name: "emoji string", j: []string{`"🐉🐺"`}, sum: 1230036448},
		{name: "tab string", j: []string{"\"\\t\"", "\"\\u0009\""}, sum: 3529881795},
		{name: "noop escaped slash", j: []string{`"/"`, `"\/"`}, sum: 3541612589},
		// numbers
		{name: "zero", j: []string{"0", "0.0", "0e0", "0e1", "0e100", "0.0e100", "-0", "-0.0", "-0.0e1"}, sum: 2102537547},
		{name: "one", j: []string{"1", "1.0", "10.0e-1", "0.1e1", "10000000000000000000000000000000000000000e-40"}, sum: 2089830268},
		// true/false/null
		{name: "true", j: []string{`true`}, sum: 2238339752},
		{name: "false", j: []string{`false`}, sum: 1993550816},
		{name: "null", j: []string{`null`}, sum: 2013832146},
		// objects
		{name: "empty object", j: []string{`{}`}, sum: 3711965057},
		{name: "trivial object", j: []string{`{"1":1,"2":2}`}, sum: 593357170},
		{name: "object with reordered keys", j: []string{`{"hi":1,"ho":2}`, `{"ho":2,"hi":1}`}, sum: 1308407541},
		// arrays
		{name: "empty array", j: []string{`[]`}, sum: 223132457},
		{name: "array", j: []string{`[[],[2]]`}, sum: 2077149373},
		{name: "nested array", j: []string{`[[[2]]]`}, sum: 3539679565},
		{name: "array with two equal objects", j: []string{`[{"1":1},{"1":1}]`}, sum: 2843930034},
		{name: "array with three equal objects", j: []string{`[{"1":1},{"1":1},{"1":1}]`}, sum: 12555174},
		// nested objects and arrays
		{name: "values in array", j: []string{`[{},"ho",{"hi":2}]`}, sum: 2520557453},
		{name: "object in array", j: []string{`[{"ho":{"hi":2}}]`}, sum: 4250229204},
		// CRC32 has issues with repeated subkeys, which cancel each other out
		{name: "CRC32 gotcha", j: []string{`{"a":{},"b":{}}`, `{"a":{"x":{}},"b":{"x":{}}}`}, sum: 3508198686},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, input := range tt.j {
				sum, err := sumCRC32(strings.NewReader(input))
				if err != nil {
					t.Errorf("JsonSum_CRC32(%q) error = %v", input, err)
					return
				}
				if sum != tt.sum {
					t.Errorf("JsonSum_CRC32(%q) sum = %v, want %v", input, sum, tt.sum)
				}
			}
		})
	}
}

func TestJsonSum_Blake2b256(t *testing.T) {
	tests := []struct {
		name string
		j    string
		sum  string // hex
	}{
		// The CRC32 gotcha does not apply to Blake2b
		{name: "CRC32 gotch 1", j: `{"a":{},"b":{}}`, sum: "95a5ac5dc6415aa04b89433359b701002c7595a1fb09d33d7df283c270551768"},
		{name: "CRC32 gotch 2", j: `{"a":{"x":{}},"b":{"x":{}}}`, sum: "d32d6916a7aab27f8aaa096663093b6d25c6e4aa26d2c9a46c8c876c5daf1d9d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			digest, err := Sum(strings.NewReader(tt.j), func() hash.Hash {
				digest, _ := blake2b.New256(nil)
				return digest
			})
			if err != nil {
				t.Errorf("JsonSum_Blake2b256(%q) error = %v", tt.j, err)
				return
			}
			sum := hex.EncodeToString(digest.Sum(nil))
			if sum != tt.sum {
				t.Errorf("JsonSum_Blake2b256(%q) sum = %v, want %v", tt.j, sum, tt.sum)
			}
		})
	}
}

func TestJsonSum_InvalidInput(t *testing.T) {
	tests := []struct {
		name string
		j    string
		err  string
	}{
		{name: "missing closing bracket", j: `{"test":"hi"`, err: "EOF"},
		{name: "missing object value", j: `{"test":}`, err: "invalid character"},
		{name: "duplicated key", j: `{"1":1,"2":2,"1":1}`, err: "duplicated object key"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sumCRC32(strings.NewReader(tt.j))
			if err == nil {
				t.Errorf("JsonSum(%v) expected error, but returned none", tt.j)
			} else if !strings.Contains(err.Error(), tt.err) {
				t.Errorf("JsonSum(%v) error = %q did not contain %q", tt.j, err.Error(), tt.err)
			}
		})
	}
}

type DebugHash struct {
	hash.Hash
	Name  string
	Input bytes.Buffer
}

func (d *DebugHash) DebugToken(t json.Token) {
	fmt.Printf("%v ", t)
}

func (d *DebugHash) Write(p []byte) (n int, err error) {
	d.Input.Write(p)
	return d.Hash.Write(p)
}

func (d *DebugHash) Reset() {
	d.Input.Reset()
	d.Hash.Reset()
}

func (d *DebugHash) Dump() string {
	return hex.EncodeToString(d.Input.Bytes())
}

func (d *DebugHash) Sum(b []byte) []byte {
	sum := d.Hash.Sum(b)
	fmt.Printf("/* %v(%v) = %v */ ", d.Name, d.Dump(), hex.EncodeToString(sum))
	return sum
}

func TestJsonSum_Gotcha(t *testing.T) {
	tests := []struct {
		name string
		j    string
	}{
		{name: "CRC32 gotch 1", j: `{"a":{},"b":{}}`},
		{name: "CRC32 gotch 2", j: `{"a":{"x":{}},"b":{"x":{}}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Sum(strings.NewReader(tt.j), func() hash.Hash {
				return &DebugHash{Name: "CRC32", Hash: crc32.NewIEEE()}
			})
			if err != nil {
				t.Errorf("JsonSum(%v) error = %v", tt.j, err)
				return
			}
			resultDH := result.(*DebugHash)
			resultDH.Sum(nil) // for debug writer
		})
	}
}

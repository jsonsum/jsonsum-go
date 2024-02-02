package jsonsum

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestJsonSum_Fixtures(t *testing.T) {
	type TestCase struct {
		Name   string   `json:"name"`
		Inputs []string `json:"inputs"`
		SHA256 string   `json:"sha256"`
	}
	testData, err := os.ReadFile("testdata/testdata.json")
	if err != nil {
		t.Fatalf("os.ReadFile(testdata.json) error = %v", err)
	}
	var testCases []TestCase
	if err := json.Unmarshal(testData, &testCases); err != nil {
		t.Fatalf("json.Unmarshal(testdata.json) error = %v", err)
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			for _, input := range tc.Inputs {
				if expected, err := hex.DecodeString(tc.SHA256); err != nil {
					t.Fatalf("hex.DecodeString(%q) error = %v", tc.SHA256, err)
				} else if result, err := Sum(strings.NewReader(input), sha256.New); err != nil {
					t.Errorf("JsonSum(%q) error = %v", input, err)
				} else if actual := result.Sum(nil); !bytes.Equal(actual, expected) {
					t.Errorf("JsonSum(%q) result = %v, want %v", input, hex.EncodeToString(actual), tc.SHA256)
				}
			}
		})
	}
}

func TestJsonSum_Invalid(t *testing.T) {
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
			_, err := Sum(strings.NewReader(tt.j), sha256.New)
			if err == nil {
				t.Errorf("JsonSum(%v) expected error, but returned none", tt.j)
			} else if !strings.Contains(err.Error(), tt.err) {
				t.Errorf("JsonSum(%v) error = %q did not contain %q", tt.j, err.Error(), tt.err)
			}
		})
	}
}

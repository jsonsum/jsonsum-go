package jsonsum

import (
	"strings"
	"testing"
)

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, input := range tt.j {
				sum, err := CRC32(strings.NewReader(input))
				if err != nil {
					t.Errorf("JsonSum(%q) error = %v", input, err)
					return
				}
				if sum != tt.sum {
					t.Errorf("JsonSum(%q) sum = %v, want %v", input, sum, tt.sum)
				}
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
			_, err := CRC32(strings.NewReader(tt.j))
			if err == nil {
				t.Errorf("JsonSum(%q) expected error, but returned none", tt.j)
			} else if !strings.Contains(err.Error(), tt.err) {
				t.Errorf("JsonSum(%q) error = %q did not contain %q", tt.j, err.Error(), tt.err)
			}
		})
	}
}

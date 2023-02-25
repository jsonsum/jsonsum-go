package jsonsum

import (
	"strings"
	"testing"
)

func TestJsonSum_Equality(t *testing.T) {
	tests := []struct {
		name  string
		jsonA string
		jsonB string
		equal bool
	}{
		{name: "Different key order gives the same checksum", jsonA: `{"hi":1,"ho":2}`, jsonB: `{"ho":2,"hi":1}`, equal: true},
		{name: "Different array nesting gives different checksums", jsonA: `[[],[2]]`, jsonB: `[[[2]]]`, equal: false},
		{name: "Different object nesting gives different checksums", jsonA: `[{},"ho",{"hi":2}]`, jsonB: `[{"ho":{"hi":2}}]`, equal: false},
		{name: "Encoding of numbers should not matter", jsonA: `2`, jsonB: `2.0`, equal: true},
		{name: "Repeated fields in arrays should not cancel", jsonA: `[{"1":1},{"1":1},{"1":1}]`, jsonB: `[{"1":1}]`, equal: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sumA, err := JsonSum(strings.NewReader(tt.jsonA))
			if err != nil {
				t.Errorf("JsonSum(A) error = %v", err)
				return
			}
			sumB, err := JsonSum(strings.NewReader(tt.jsonB))
			if err != nil {
				t.Errorf("JsonSum(B) error = %v", err)
				return
			}
			if (sumA == sumB) != tt.equal {
				t.Errorf("JsonSum(A) = %v, JsonSum(B) = %v, equal = %v", sumA, sumB, tt.equal)
			}
		})
	}
}

func TestJsonSum_Checksum(t *testing.T) {
	tests := []struct {
		jsonStr string
		sum     uint32
	}{
		{jsonStr: `"hi"`, sum: 43166820},
		{jsonStr: `{"hi":1,"ho":2}`, sum: 353985019},
		{jsonStr: `[[],[2]]`, sum: 1127135515},
		{jsonStr: `[[[2]]]`, sum: 483837260},
		{jsonStr: `[{},"ho",{"hi":2}]`, sum: 714017136},
		{jsonStr: `[{"ho":{"hi":2}}]`, sum: 2387720476},
		{jsonStr: `[{"hi":{"hi":2}}]`, sum: 3894056877},
		{jsonStr: `{"1":1,"2":2}`, sum: 3705610668},
		{jsonStr: `[{"1":1},{"1":1}]`, sum: 3782046115},
	}
	for _, tt := range tests {
		t.Run(tt.jsonStr, func(t *testing.T) {
			sum, err := JsonSum(strings.NewReader(tt.jsonStr))
			if err != nil {
				t.Errorf("JsonSum() error = %v", err)
				return
			}
			if sum != tt.sum {
				t.Errorf("JsonSum() sum = %v, want %v", sum, tt.sum)
			}
		})
	}
}

package jsonsum

import (
	"bytes"
	"testing"
)

func TestNormalizeNum(t *testing.T) {
	tests := []struct {
		norm string
		nums []string
	}{
		{norm: "0e0", nums: []string{"0", "000", "0.0", "0e0", "0E0", "0E+0", "0E-0", "0e1", "0e100", "0.0e100", "-0", "-0.0", "-00.00", "-0.0e1"}},
		{norm: "1e0", nums: []string{"1", "001", "1.0", "1e0", "1E0", "1E+0", "1E-0", "1.00", "10.0e-1", "0.1e1", "10000000000000000000000000000000000000000e-40"}},
		{norm: "-1e0", nums: []string{"-1", "-001", "-1.0", "-1e0", "-1E0", "-1E+0", "-1E-0", "-1.00", "-10.0e-1", "-0.1e1", "-10000000000000000000000000000000000000000e-40"}},
		{norm: "2e4", nums: []string{"20E3"}},
		{norm: "3003003e-3", nums: []string{"0003.00300300e+003"}},
	}
	for _, tt := range tests {
		t.Run(tt.norm, func(t *testing.T) {
			for _, input := range tt.nums {
				buf := bytes.NewBuffer(nil)
				if err := normalizeNumber(buf, input); err != nil {
					t.Errorf("normalizeNum(%q) error = %v", input, err)
					return
				}
				if result := buf.String(); result != tt.norm {
					t.Errorf("normalizeNum(%q) = %v, want %v", input, result, tt.norm)
				}
			}
		})
	}
}

package jsonsum

import (
	"fmt"
	"io"
	"strconv"
)

// normalizeNumber normalizes a number in scientific notation by removing the decimal point,
// adjusting the exponent accordingly, stripping unnecessary zeroes, and removing fluff when the
// entire value is zero.
func normalizeNumber(w io.Writer, s string) error {
	isZero := true

	// strip leading minus sign and zeroes
	isNegative := s[0] == '-'
	if isNegative {
		s = s[1:]
	}
	for len(s) > 0 && s[0] == '0' {
		s = s[1:]
	}

	// look ahead into decimal parts to write minus sign and single leading zero
	// TODO: we can skip this since we're not writing the decimal point anyway!
	inFractionalPart := false
	zeroIntegral := true
	zeroFractional := true
	for _, c := range s {
		if c == '.' {
			inFractionalPart = true
			continue
		}
		if c == 'e' || c == 'E' {
			break
		}
		if '1' <= c && c <= '9' {
			if inFractionalPart {
				zeroFractional = false
			} else {
				zeroIntegral = false
			}
		}
	}
	if isNegative && (!zeroIntegral || !zeroFractional) {
		w.Write([]byte{'-'})
	}
	if zeroIntegral && zeroFractional {
		w.Write([]byte{'0'})
	}

	// normalize integral part
	var exp int64
	for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
		if s[0] == '0' {
			exp++
		} else {
			for i := int64(0); i < exp; i++ {
				w.Write([]byte{'0'})
			}
			isZero = false
			exp = 0
			w.Write([]byte{s[0]})
		}
		s = s[1:]
	}

	// normalize fractional part
	var leadingZeroes int64
	resetExponent := false
	if len(s) > 0 && s[0] == '.' {
		s = s[1:]
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			if s[0] == '0' {
				leadingZeroes++
			} else {
				if !resetExponent {
					for i := int64(0); i < exp; i++ {
						w.Write([]byte{'0'})
					}
					resetExponent = true
					exp = 0
				}
				for i := int64(0); i < leadingZeroes; i++ {
					w.Write([]byte{'0'})
				}
				w.Write([]byte{s[0]})
				isZero = false
				exp = exp - leadingZeroes - 1
				leadingZeroes = 0
			}
			s = s[1:]
		}
	}
	if len(s) > 0 && (s[0] == 'e' || s[0] == 'E') {
		s = s[1:]
		if s[0] == '+' {
			s = s[1:]
		}
		inExp, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("could not parse number exponent %q: %w", s, err)
		}
		exp += inExp
		for len(s) > 0 && ('0' <= s[0] && s[0] <= '9' || s[0] == '-' || s[0] == '+') {
			s = s[1:]
		}
	}
	w.Write([]byte{'e'})
	if isZero {
		exp = 0
	} else if exp < 0 {
		w.Write([]byte{'-'})
		exp *= -1
	}
	w.Write([]byte(strconv.FormatInt(exp, 10)))
	if len(s) != 0 {
		return fmt.Errorf("unexpected remainder %q", s)
	}

	return nil
}

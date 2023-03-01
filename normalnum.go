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
	a := 0
	n := len(s)

	// strip leading minus sign and zeroes
	writeMinus := s[a] == '-'
	if writeMinus {
		a++
	}
	for n > a && s[a] == '0' {
		a++
	}

	// normalize integral part
	var exp int64
	for n > a && '0' <= s[a] && s[a] <= '9' {
		if s[a] == '0' {
			exp++
		} else {
			if writeMinus {
				w.Write([]byte{'-'})
				writeMinus = false
			}
			for i := int64(0); i < exp; i++ {
				w.Write([]byte{'0'})
			}
			isZero = false
			exp = 0
			w.Write([]byte{s[a]})
		}
		a++
	}

	// normalize fractional part
	var leadingZeroes int64
	resetExponent := false
	if n > a && s[a] == '.' {
		a++
		for n > a && '0' <= s[a] && s[a] <= '9' {
			if s[a] == '0' {
				leadingZeroes++
			} else {
				if writeMinus {
					w.Write([]byte{'-'})
					writeMinus = false
				}
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
				w.Write([]byte{s[a]})
				isZero = false
				exp = exp - leadingZeroes - 1
				leadingZeroes = 0
			}
			a++
		}
	}
	if n > a && (s[a] == 'e' || s[a] == 'E') {
		a++
		if s[a] == '+' {
			a++
		}
		inExp, err := strconv.ParseInt(s[a:], 10, 64)
		if err != nil {
			return fmt.Errorf("could not parse number exponent %q: %w", s[a:], err)
		}
		exp += inExp
		for n > a && ('0' <= s[a] && s[a] <= '9' || s[a] == '-' || s[a] == '+') {
			a++
		}
	}
	if isZero {
		w.Write([]byte("0e0"))
	} else {
		w.Write([]byte{'e'})
		if exp < 0 {
			w.Write([]byte{'-'})
			exp *= -1
		}
		w.Write([]byte(strconv.FormatInt(exp, 10)))
	}
	if n != a {
		return fmt.Errorf("unexpected remainder %q", s[a:])
	}
	return nil
}

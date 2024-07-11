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
	a := 0
	n := len(s)
	z := true

	// strip leading minus sign and zeroes
	writeMinus := s[a] == '-'
	if writeMinus {
		a++
	}
	for a < n && s[a] == '0' {
		a++
	}

	// normalize integral part
	var e int64
	for a < n && '0' <= s[a] && s[a] <= '9' {
		if s[a] == '0' {
			e++
		} else {
			if writeMinus {
				w.Write([]byte{'-'})
				writeMinus = false
			}
			for i := int64(0); i < e; i++ {
				w.Write([]byte{'0'})
			}
			z = false
			e = 0
			w.Write([]byte{s[a]})
		}
		a++
	}

	// normalize fractional part
	if a < n && s[a] == '.' {
		zeroes := int64(0)
		resetExponent := false
		a++
		for a < n && '0' <= s[a] && s[a] <= '9' {
			if s[a] == '0' {
				zeroes++
			} else {
				if writeMinus {
					w.Write([]byte{'-'})
					writeMinus = false
				}
				if !resetExponent {
					for i := int64(0); i < e; i++ {
						w.Write([]byte{'0'})
					}
					resetExponent = true
					e = 0
				}
				for i := int64(0); i < zeroes && !z; i++ {
					w.Write([]byte{'0'})
				}
				w.Write([]byte{s[a]})
				z = false
				e = e - zeroes - 1
				zeroes = 0
			}
			a++
		}
	}
	if a < n && (s[a] == 'e' || s[a] == 'E') {
		a++
		if s[a] == '+' {
			a++
		}
		inExp, err := strconv.ParseInt(s[a:], 10, 64)
		if err != nil {
			return fmt.Errorf("could not parse number exponent %q: %w", s[a:], err)
		}
		e += inExp
		for a < n && ('0' <= s[a] && s[a] <= '9' || s[a] == '-' || s[a] == '+') {
			a++
		}
	}
	if z {
		w.Write([]byte("0e0"))
	} else {
		w.Write([]byte{'e'})
		w.Write([]byte(strconv.FormatInt(e, 10)))
	}
	if n != a {
		return fmt.Errorf("unexpected remainder %q", s[a:])
	}
	return nil
}

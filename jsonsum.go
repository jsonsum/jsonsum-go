package jsonsum

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"math"
)

func jsonStrSum(crc io.Writer, s string) {
	crc.Write([]byte{byte('s')})
	binary.Write(crc, binary.BigEndian, uint32(len(s))) // FIXME: teeeechnically json strings can longer than 4GiB :3
	binary.Write(crc, binary.BigEndian, uint32(0))      // FIXME: superfluous write, would write 64-bit len instead
	crc.Write([]byte(s))
}

func jsonObjSum(dec *json.Decoder) (sum uint32, err error) {
	var t json.Token
	keysSeen := make(map[string]bool)
	for t, err = dec.Token(); err == nil; t, err = dec.Token() {
		if delim, ok := t.(json.Delim); ok && delim == '}' {
			return
		}
		key, ok := t.(string)
		if !ok {
			return 0, fmt.Errorf("expected string key for object at byte %d", dec.InputOffset())
		}
		if keysSeen[key] {
			return 0, fmt.Errorf("duplicated object key %q at byte %d", key, dec.InputOffset())
		}
		keysSeen[key] = true
		crc := crc32.NewIEEE()
		jsonStrSum(crc, key)
		if err := jsonValSum(dec, crc); err != nil {
			return 0, fmt.Errorf("failed to compute object value sum: %w", err)
		} else {
			sum ^= crc.Sum32()
		}
	}
	return 0, fmt.Errorf("read from JSON token stream for object failed at byte %d: %w", dec.InputOffset(), err)
}

func jsonValSum(dec *json.Decoder, crc hash.Hash32) error {
	var err error
	var t json.Token
	arr := 0
	for t, err = dec.Token(); err == nil; t, err = dec.Token() {
		switch v := t.(type) {
		case json.Delim:
			switch v {
			case '[':
				crc.Write([]byte{byte('[')})
				arr++
			case ']':
				crc.Write([]byte{byte(']')})
				arr--
			case '{':
				// XXX: needs depth check, malicious JSON could overflow the stack
				if sum, err := jsonObjSum(dec); err != nil {
					return err
				} else {
					crc.Write([]byte{byte('o')})
					binary.Write(crc, binary.BigEndian, sum)
					binary.Write(crc, binary.BigEndian, uint32(0)) // FIXME: superfluous write like with string len
				}
			default:
				panic("unexpected json.Delim: " + string(v))
			}
		case bool:
			if v {
				crc.Write([]byte{byte('t')})
			} else {
				crc.Write([]byte{byte('f')})
			}
		case float64:
			// XXX: probably does not handle quirky numbers well; investigate dec.UseNumber()
			crc.Write([]byte{byte('i')})
			binary.Write(crc, binary.BigEndian, math.Float64bits(v))
		case string:
			jsonStrSum(crc, v)
		case nil:
			crc.Write([]byte{byte('n')})
		}
		if arr == 0 {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("read from JSON token stream failed at byte %d: %w", dec.InputOffset(), err)
	}
	return nil
}

func JsonSum(r io.Reader) (uint32, error) {
	dec := json.NewDecoder(r)
	crc := crc32.NewIEEE()

	if err := jsonValSum(dec, crc); err != nil {
		return 0, fmt.Errorf("error computing JSON value sum: %w", err)
	}

	if _, err := dec.Token(); err == io.EOF {
		return crc.Sum32(), nil
	} else if err != nil {
		return 0, fmt.Errorf("error reading final JSON token at byte %d: %w", dec.InputOffset(), err)
	} else {
		return 0, fmt.Errorf("extraneous data after end of first JSON value at byte %d", dec.InputOffset())
	}
}

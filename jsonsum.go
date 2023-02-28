package jsonsum

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
)

const DefaultMaxDepth = 64

type Config struct {
	MaxDepth int
	Digest   func() hash.Hash
}

func (c Config) withCheckedDefaults() Config {
	if c.MaxDepth == 0 {
		c.MaxDepth = DefaultMaxDepth
	} else if c.MaxDepth < 0 {
		panic("jsonsum.Config MaxDepth cannot be negative")
	}

	if c.Digest == nil {
		panic("jsonsum.Config Digest function is required")
	}

	return c
}

type state struct {
	Config
	nDigest  int
	arrDepth int
	objDepth int
	buf      []byte
}

var ErrDepthLimitExceeded = errors.New("JSON nesting depth limit exceeded")

func (j *state) depth() int {
	return j.arrDepth + j.objDepth
}

func (j *state) jsonStrSum(dig hash.Hash, s string) {
	strDig := j.Digest()
	strDig.Write([]byte(s))
	dig.Write([]byte{'s'})
	dig.Write(strDig.Sum(j.buf))
}

func (j *state) jsonNumSum(dig hash.Hash, num json.Number) error {
	dig.Write([]byte{'i'})
	if err := normalizeNumber(dig, num.String()); err != nil {
		return fmt.Errorf("could not normalize number %q: %w", num.String(), err)
	}
	return nil
}

func (j *state) jsonObjSum(dig hash.Hash, dec *json.Decoder) error {
	var err error
	var t json.Token
	sum := make([]byte, j.nDigest)
	keysSeen := make(map[string]struct{})

	for t, err = dec.Token(); err == nil; t, err = dec.Token() {
		if delim, ok := t.(json.Delim); ok && delim == '}' {
			dig.Write([]byte{'o'})
			dig.Write(sum)
			return nil
		}
		key, ok := t.(string)
		if !ok {
			return fmt.Errorf("expected string key for object at byte %d", dec.InputOffset())
		}
		if _, seen := keysSeen[key]; seen {
			return fmt.Errorf("duplicated object key %q at byte %d", key, dec.InputOffset())
		}
		keysSeen[key] = struct{}{}
		pairDig := j.Digest()
		j.jsonStrSum(pairDig, key)
		if err := j.jsonValSum(pairDig, dec); err != nil {
			return fmt.Errorf("failed to compute object value sum: %w", err)
		}
		pairSum := pairDig.Sum(j.buf)
		for i := 0; i < j.nDigest; i++ {
			sum[i] ^= pairSum[i]
		}
	}

	return fmt.Errorf("read from JSON token stream for object failed at byte %d: %w", dec.InputOffset(), err)
}

func (j *state) jsonValSum(dig hash.Hash, dec *json.Decoder) error {
	var err error
	var t json.Token
	localDepth := 0
	for t, err = dec.Token(); err == nil; t, err = dec.Token() {
		switch v := t.(type) {
		case json.Delim:
			switch v {
			case '[':
				if j.depth() == j.MaxDepth {
					return ErrDepthLimitExceeded
				}
				dig.Write([]byte{byte('[')})
				j.arrDepth++
				localDepth++
			case ']':
				dig.Write([]byte{byte(']')})
				j.arrDepth--
				localDepth--
			case '{':
				if j.depth() == j.MaxDepth {
					return ErrDepthLimitExceeded
				}
				j.objDepth++
				localDepth++
				err = j.jsonObjSum(dig, dec)
				j.objDepth--
				localDepth--
				if err != nil {
					return err
				}
			default:
				panic("unexpected json.Delim: " + string(v))
			}
		case json.Number:
			if err := j.jsonNumSum(dig, v); err != nil {
				return fmt.Errorf("cannot parse number at byte %d: %w", dec.InputOffset(), err)
			}
		case bool:
			if v {
				dig.Write([]byte{byte('t')})
			} else {
				dig.Write([]byte{byte('f')})
			}
		case string:
			j.jsonStrSum(dig, v)
		case nil:
			dig.Write([]byte{byte('n')})
		default:
			panic(fmt.Sprintf("unexpected JSON token type: %T", v))
		}
		if localDepth == 0 {
			break // to process first root value only
		}
	}
	if err != nil {
		return fmt.Errorf("read from JSON token stream failed at byte %d: %w", dec.InputOffset(), err)
	}
	return nil
}

func Sum(r io.Reader, config Config) (hash.Hash, error) {
	sum := &state{Config: config.withCheckedDefaults()}
	dec := json.NewDecoder(r)
	dec.UseNumber()
	dig := sum.Digest()
	sum.nDigest = dig.Size()
	sum.buf = make([]byte, 0, sum.nDigest) // to reduce allocation overhead

	if err := sum.jsonValSum(dig, dec); err != nil {
		return nil, fmt.Errorf("error computing JSON value sum: %w", err)
	}

	if _, err := dec.Token(); err == io.EOF {
		return dig, nil
	} else if err != nil {
		return nil, fmt.Errorf("error reading final JSON token at byte %d: %w", dec.InputOffset(), err)
	} else {
		return nil, fmt.Errorf("extraneous data after end of first JSON value at byte %d", dec.InputOffset())
	}
}

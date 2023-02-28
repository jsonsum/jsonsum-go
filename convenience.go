package jsonsum

import (
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"hash/crc32"
	"io"
)

func CRC32(r io.Reader) (uint32, error) {
	if crc, err := Sum(r, Config{
		Digest: func() hash.Hash { return crc32.NewIEEE() },
	}); err != nil {
		return 0, err
	} else {
		return crc.(hash.Hash32).Sum32(), nil
	}
}

func sumWithHash(r io.Reader, buf []byte, makeHash func() hash.Hash) ([]byte, error) {
	cfg := Config{Digest: makeHash}
	if result, err := Sum(r, cfg); err != nil {
		return nil, err
	} else {
		return result.Sum(buf), nil
	}
}

func SHA256(r io.Reader, buf []byte) ([]byte, error) {
	return sumWithHash(r, buf, sha256.New)
}

func SHA512(r io.Reader, buf []byte) ([]byte, error) {
	return sumWithHash(r, buf, sha512.New)
}

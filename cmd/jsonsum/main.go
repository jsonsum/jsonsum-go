package main

import (
	"bufio"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/mologie/jsonsum-go"
	"github.com/pborman/getopt/v2"
	"hash"
	"hash/crc32"
	"io"
	"os"
	"runtime/pprof"
)

func main() {
	algo := getopt.StringLong("algorithm", 'a', "crc32", "crc32 (default), sha256, sha512)")
	cpuprofile := getopt.StringLong("cpuprofile", 0, "", "write cpu profile to file")
	infile := getopt.StringLong("input", 'i', "", "input file, reads from stdin if omitted")
	getopt.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating pprof file %q: %v", *cpuprofile, err)
			os.Exit(1)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "error starting CPU profiler: %v", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	var digestFunc func() hash.Hash
	switch *algo {
	case "crc32":
		digestFunc = func() hash.Hash { return crc32.NewIEEE() }
	case "sha256":
		digestFunc = func() hash.Hash { return sha256.New() }
	case "sha512":
		digestFunc = func() hash.Hash { return sha512.New() }
	default:
		fmt.Fprintf(os.Stderr, "error: no such algorithm: %v", *algo)
		os.Exit(1)
	}

	// buffer size chosen experimentally with 200 MiB JSON file on an Apple M1
	const bufSize = 1024 * 1024 * 8 // 8 MiB
	var input io.Reader
	if *infile == "" {
		input = bufio.NewReaderSize(os.Stdin, bufSize)
	} else {
		f, err := os.Open(*infile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening input file %q: %v", *infile, *algo)
			os.Exit(1)
		}
		input = bufio.NewReaderSize(f, bufSize)
		defer f.Close()
	}
	result, err := jsonsum.Sum(input, jsonsum.Config{Digest: digestFunc})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error calculating sum: %v\n", err)
		os.Exit(1)
	}
	out := os.Stdout
	hex.NewEncoder(out).Write(result.Sum(nil))
	out.Write([]byte{'\n'})
}

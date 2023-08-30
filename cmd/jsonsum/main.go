package main

import (
	"bufio"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/mologie/jsonsum-go"
	"github.com/pborman/getopt/v2"
	"golang.org/x/crypto/blake2b"
	"hash"
	"hash/crc32"
	"io"
	"os"
	"runtime/pprof"
	"strings"
)

func main() {
	algo := getopt.StringLong("algorithm", 'a', "blake2b-256", "blake2b-32/.../256/512, sha256, sha512, crc32")
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
	switch {
	case strings.HasPrefix(*algo, "blake2b-"):
		var size int
		if _, err := fmt.Sscanf(*algo, "blake2b-%d", &size); err != nil {
			fmt.Fprintf(os.Stderr, "error parsing blake2b size: %v", err)
			os.Exit(1)
		}
		if size%8 != 0 {
			fmt.Fprintf(os.Stderr, "error: blake2b size must be a multiple of 8")
			os.Exit(1)
		}
		if size < 8 {
			fmt.Fprintf(os.Stderr, "error: blake2b size must be >= 8")
			os.Exit(1)
		}
		if size > 512 {
			fmt.Fprintf(os.Stderr, "error: blake2b size must be <= 512")
			os.Exit(1)
		}
		digestFunc = func() hash.Hash {
			h, err := blake2b.New(size/8, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error creating blake2b-%d hash: %v", size, err)
				os.Exit(1)
			}
			return h
		}
	case *algo == "sha256":
		// TBD how this fares against length extensions attacks
		digestFunc = func() hash.Hash { return sha256.New() }
	case *algo == "sha512":
		digestFunc = func() hash.Hash { return sha512.New() }
	case *algo == "crc32":
		// bad choice with repeating substructures, see tests
		digestFunc = func() hash.Hash { return crc32.NewIEEE() }
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
	result, err := jsonsum.Sum(input, digestFunc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error calculating sum: %v\n", err)
		os.Exit(1)
	}
	out := os.Stdout
	hex.NewEncoder(out).Write(result.Sum(nil))
	out.Write([]byte{'\n'})
}

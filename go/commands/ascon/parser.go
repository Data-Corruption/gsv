package ascon

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Data-Corruption/stdx/xlog"
)

// TestVector models one KAT case parsed from the .rsp-style file.
// All fields are hex-encoded, uppercase, with no "0x" prefix.
type TestVector struct {
	Cnt        int
	Key, Nonce string // kept for potential KAT sanity checks
	PT, AD, CT string // hex payloads
}

func (v TestVector) isZero() bool {
	return v.Key == "" && v.Nonce == "" && v.PT == "" && v.AD == "" && v.CT == ""
}

// chunkHexTo128b splits an uppercase hex string into 128-bit (32 hex) chunks.
// Only the final chunk is right-padded with '0' to 32 hex chars.
// Input may be empty; output will be empty unless caller wants to synthesize an "empty" chunk.
func chunkHexTo128b(s string) []string {
	s = strings.ToUpper(strings.TrimSpace(s))
	var out []string
	for i := 0; i < len(s); i += 32 {
		end := i + 32
		if end > len(s) {
			chunk := s[i:]
			if len(chunk) > 0 {
				chunk = chunk + strings.Repeat("0", 32-len(chunk))
				out = append(out, chunk)
			}
			break
		}
		out = append(out, s[i:end])
	}
	return out
}

// padBytesTo128b returns how many BYTES (0..16) of MSB padding are needed
// to make s a multiple of 128 bits. Empty string is defined as 16.
func padBytesTo128b(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 16
	}
	// hex chars -> bytes/2
	rem := len(s) % 32
	if rem == 0 {
		return 0
	}
	return (32 - rem) // hex chars padded
	// caller will divide by 2 when formatting as bytes
}

// parse reads a .rsp-style file and returns a slice of TestVector structs.
func parse(ctx context.Context, path string) ([]TestVector, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var v TestVector
	var out []TestVector
	sc := bufio.NewScanner(f)
	// allow very long KAT lines (default token limit is ~64KiB)
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	ln := 0
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			xlog.Debugf(ctx, "skipping empty line %d", ln)
			ln++
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("parse:%d: expected key=value, got %q", ln, line)
		}
		tag := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch tag {
		case "Count":
			// flush previous vector if it has content
			if !v.isZero() {
				out = append(out, v)
			}
			v = TestVector{}
			v.Cnt, _ = strconv.Atoi(val)
		case "Key":
			v.Key = val
		case "Nonce":
			v.Nonce = val
		case "PT":
			v.PT = val
		case "AD":
			v.AD = val
		case "CT":
			v.CT = val
		}
		ln++
	}
	xlog.Debugf(ctx, "scanned %d lines", ln)
	// push last one
	if !v.isZero() {
		out = append(out, v)
	}
	return out, sc.Err()
}

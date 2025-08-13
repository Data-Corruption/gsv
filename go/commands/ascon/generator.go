package ascon

import (
	"context"
	"fmt"
	"strings"
)

// ByteSwapMode controls in-chunk byte reordering of the *data* bytes.
type ByteSwapMode int

const (
	SwapNone ByteSwapMode = iota
	SwapDataReverse
)

const (
	bytesPerRow = 16
	hexPerByte  = 2
	hexPerRow   = bytesPerRow * hexPerByte
)

// applyByteSwap returns a 32-hex-char (128-bit) chunk whose first `usedBytes`
// are optionally byte-reversed depending on `mode`. Only the *data* portion
// (the first usedBytes) is reversed; the right-side padding remains in place.
// If `chunk` is malformed (wrong length), the function returns an all-zero chunk.
// Callers should normalize hex case (e.g., uppercase) at parse boundaries.
func applyByteSwap(chunk string, usedBytes int, mode ByteSwapMode) string {
	// chunk must be exactly 32 hex chars (128 bits)
	if len(chunk) != hexPerRow {
		return strings.Repeat("0", hexPerRow)
	}

	// clamp usedBytes to [0, 16]
	if usedBytes < 0 {
		usedBytes = 0
	} else if usedBytes > bytesPerRow {
		usedBytes = bytesPerRow
	}

	// no-op cases
	if mode == SwapNone || usedBytes == 0 {
		return chunk
	}

	// split into data and right-side padding
	dataHex := chunk[:usedBytes*hexPerByte]
	padHex := chunk[usedBytes*hexPerByte:]

	switch mode {
	case SwapDataReverse: // reverse the order of bytes in dataHex (pairs of hex characters)
		rev := make([]byte, len(dataHex))
		// i walks bytes; j walks the destination index in hex chars.
		for i, j := 0, 0; i < usedBytes; i, j = i+1, j+2 {
			srcStart := (usedBytes - 1 - i) * hexPerByte
			copy(rev[j:j+2], dataHex[srcStart:srcStart+2])
		}
		return string(rev) + padHex
	default: // unknown mode: be conservative and pass-through.
		return chunk
	}
}

// Emit 135b rows: { empty(1), last(1), pad_bytes(5), data(128) }
func emitVecLines(sb *strings.Builder, kind string, cnt int, field string, cap3 bool, sm ByteSwapMode) {
	var chunks []string
	empty := len(strings.TrimSpace(field)) == 0
	if empty {
		chunks = []string{strings.Repeat("0", 32)}
	} else {
		chunks = chunkHexTo128b(field)
		if cap3 && len(chunks) > 3 {
			chunks = chunks[:3]
		}
	}
	pbAll := padBytesTo128b(field) / 2 // convert hex chars to BYTES
	for idx, c := range chunks {
		// pad bytes field is set to the final chunk's pad value for ALL chunks
		// (or 16 if empty)
		emptyBit := "1'b0"
		if empty {
			emptyBit = "1'b1"
		}
		lastBit := "1'b1"
		if idx != len(chunks)-1 {
			lastBit = "1'b0"
		}
		pb := 0
		if empty {
			pb = 16
		} else {
			// Use the final chunk's pad value for ALL chunks of this vector
			pb = pbAll
		}

		usedBytes := 16
		if empty {
			usedBytes = 0
		} else if idx == len(chunks)-1 {
			usedBytes = 16 - pb
		}
		c = applyByteSwap(c, usedBytes, sm)

		index := fmt.Sprintf("13'h%X_%X", cnt, idx)
		line := fmt.Sprintf("        %-6s: %s_text = {%s, %s, 5'd%d, 128'h%s};\n",
			index, kind, emptyBit, lastBit, pb, c)
		sb.WriteString(line)
	}
}

func generate(ctx context.Context, vs []TestVector, sm ByteSwapMode) string {
	var ptCase, adCase, ctCase strings.Builder

	for _, v := range vs {
		// PT/AD/CT
		emitVecLines(&ptCase, "plain", v.Cnt, v.PT, false, sm)
		emitVecLines(&adCase, "associated", v.Cnt, v.AD, false, sm)
		emitVecLines(&ctCase, "cipher", v.Cnt, v.CT, true, sm)
	}

	const tplPfx = `/**
 *  Module: ascon_test_rom
 *  Auto-generated.
 **/
module ascon_test_rom (
    // key/nonce ROM removed (constant elsewhere)
    input  [12:0] plain_index,
    input  [12:0] associated_index,
    input  [12:0] cipher_index,
    output logic [134:0] plain_text,
    output logic [134:0] associated_text,
    output logic [134:0] cipher_text
);

// ---- PT (Plain) ------------------------------------------------------------
always_comb begin : plain_text_mux
    case (plain_index)
`
	const tplMid2 = `        default: plain_text = 135'd0;
    endcase
end

// ---- AD (Associated) -------------------------------------------------------
always_comb begin : associated_text_mux
    case (associated_index)
`
	const tplMid3 = `        default: associated_text = 135'd0;
    endcase
end

// ---- CT (Cipher) -----------------------------------------------------------
always_comb begin : cipher_text_mux
    case (cipher_index)
`
	const tplMid4 = `        default: cipher_text = 135'd0;
    endcase
end

endmodule : ascon_test_rom
`
	var out strings.Builder
	out.WriteString(tplPfx)
	out.WriteString(ptCase.String())
	out.WriteString(tplMid2)
	out.WriteString(adCase.String())
	out.WriteString(tplMid3)
	out.WriteString(ctCase.String())
	out.WriteString(tplMid4)

	return out.String()
}

package engine

import (
	"bytes"
	"context"
	"testing"

	"github.com/oferchen/hclalign/config"
)

func FuzzProcessReader(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		// Limit excessively large inputs to avoid OOMs during fuzzing.
		const maxBytes = 1 << 12
		if len(data) > maxBytes {
			t.Skip()
		}

		cfg := &config.Config{Mode: config.ModeWrite, Stdout: true}

		var out bytes.Buffer
		_, err := ProcessReader(context.Background(), bytes.NewReader(data), &out, cfg)
		if err != nil {
			return
		}
		styled := out.Bytes()

		var out2 bytes.Buffer
		changed, err := ProcessReader(context.Background(), bytes.NewReader(styled), &out2, cfg)
		if err != nil {
			t.Fatalf("second process: %v", err)
		}
		if changed {
			t.Fatalf("processing is not idempotent")
		}
		if !bytes.Equal(styled, out2.Bytes()) {
			t.Fatalf("round-trip mismatch")
		}
	})
}

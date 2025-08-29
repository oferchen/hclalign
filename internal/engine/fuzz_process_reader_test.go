// internal/engine/fuzz_process_reader_test.go
package engine

import (
	"bytes"
	"context"
	"testing"

	"github.com/oferchen/hclalign/config"
)

func FuzzProcessReader(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {

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


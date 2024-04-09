// main_test.go

package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain_Execute(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectedError  string
	}{
		{
			name:           "ValidArgs",
			args:           []string{"testdata/example.hcl"},
			expectedOutput: "Executing with target: testdata/example.hcl\n",
			expectedError:  "",
		},
		{
			name:           "MissingArgs",
			args:           []string{},
			expectedOutput: "",
			expectedError:  "Error: accept 1 arg(s), received 0\n",
		},
		{
			name:           "MultipleArgs",
			args:           []string{"file1.hcl", "file2.hcl"},
			expectedOutput: "",
			expectedError:  "Error: accept 1 arg(s), received 2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Redirect stdout to capture output
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Restore stdout when done
			defer func() {
				os.Stdout = old
			}()

			// Execute main with test args
			os.Args = append([]string{"hcl_align"}, tt.args...)
			main()

			// Close pipe and read captured output
			w.Close()
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			actualOutput := buf.String()

			// Assert output and error
			assert.Equal(t, tt.expectedOutput, actualOutput)
			if tt.expectedError != "" {
				assert.Contains(t, actualOutput, tt.expectedError)
			}
		})
	}
}

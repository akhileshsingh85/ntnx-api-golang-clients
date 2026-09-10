package client

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestSetLogOutput(t *testing.T) {
	apiClient := NewApiClient()
	var output bytes.Buffer

	apiClient.SetLogOutput(&output)
	apiClient.logger.Info("before setup")
	apiClient.setupClient()
	apiClient.logger.Info("after setup")

	logs := output.String()
	if !strings.Contains(logs, "before setup") || !strings.Contains(logs, "after setup") {
		t.Fatalf("custom log output did not survive client setup: %q", logs)
	}

	lengthBeforeDiscard := output.Len()
	apiClient.SetLogOutput(io.Discard)
	apiClient.setupClient()
	apiClient.logger.Info("discarded")
	if output.Len() != lengthBeforeDiscard {
		t.Fatalf("discarded log was written to the previous output: %q", output.String())
	}
}

func TestSetLogOutputNilRestoresDefault(t *testing.T) {
	apiClient := NewApiClient()
	apiClient.SetLogOutput(io.Discard)
	apiClient.SetLogOutput(nil)

	if apiClient.logOutput != nil {
		t.Fatal("expected nil custom log output")
	}
	if apiClient.logger.Out != os.Stderr {
		t.Fatalf("expected default stderr output, got %T", apiClient.logger.Out)
	}
}

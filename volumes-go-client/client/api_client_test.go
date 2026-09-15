package client

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
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

func TestNegotiateVersionHandlesHTTPErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	host, portText, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}

	apiClient := NewApiClient()
	apiClient.Scheme = "http"
	apiClient.Host = host
	apiClient.Port = port
	apiClient.SetLogOutput(io.Discard)
	apiClient.NegotiateVersion(nil)

	if !apiClient.negotiationCompleted {
		t.Fatal("expected failed version negotiation to fall back without another OPTIONS request")
	}
	if apiClient.negotiatedVersion != "" {
		t.Fatalf("expected generated API version fallback, got %q", apiClient.negotiatedVersion)
	}
}

func TestCallApiReturnsErrorForTextResponse(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	host, portText, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}

	apiClient := NewApiClient()
	apiClient.Scheme = "http"
	apiClient.Host = host
	apiClient.Port = port
	apiClient.SetLogOutput(io.Discard)
	path := "/missing"
	_, err = apiClient.callApiInternal(context.Background(), &path, http.MethodGet, nil, url.Values{}, map[string]string{}, url.Values{}, nil, nil, nil)
	if err == nil {
		t.Fatal("expected a non-2xx text response to return an error")
	}
}

package har

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

func gzipBody(t *testing.T, plain []byte) []byte {
	t.Helper()
	var output bytes.Buffer
	writer := gzip.NewWriter(&output)
	if _, err := writer.Write(plain); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func brotliBody(t *testing.T, plain []byte) []byte {
	t.Helper()
	var output bytes.Buffer
	writer := brotli.NewWriter(&output)
	if _, err := writer.Write(plain); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func zstdBody(t *testing.T, plain []byte) []byte {
	t.Helper()
	var output bytes.Buffer
	writer, err := zstd.NewWriter(&output, zstd.WithEncoderConcurrency(1))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write(plain); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func TestDecompressBodySupportsReqableEncodings(t *testing.T) {
	plain := []byte("hello reqable")
	cases := []struct {
		name     string
		encoding string
		body     []byte
	}{
		{"identity", "identity", plain},
		{"gzip", "gzip", gzipBody(t, plain)},
		{"x-gzip", "x-gzip", gzipBody(t, plain)},
		{"brotli", "br", brotliBody(t, plain)},
		{"zstd", "zstd", zstdBody(t, plain)},
		{"zstandard", "zstandard", zstdBody(t, plain)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DecompressBody(tc.body, tc.encoding, 1024)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, plain) {
				t.Fatalf("got %q, want %q", got, plain)
			}
		})
	}
}

func TestDecompressBodyRejectsUnsupportedAndOversized(t *testing.T) {
	if _, err := DecompressBody([]byte("x"), "deflate", 10); !errors.Is(err, ErrUnsupportedContentEncoding) {
		t.Fatalf("got %v, want unsupported encoding error", err)
	}
	if _, err := DecompressBody([]byte("12345"), "identity", 4); !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("got %v, want size error", err)
	}
	if _, err := DecompressBody(gzipBody(t, bytes.Repeat([]byte("x"), 16)), "gzip", 8); !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("got %v, want size error", err)
	}
}

func TestParseFallsBackToPlainJSONWhenHeaderIsWrong(t *testing.T) {
	plain := []byte(`{"log":{"entries":[]}}`)
	document, err := Parse(plain, "gzip", 1024)
	if err != nil {
		t.Fatal(err)
	}
	if len(Entries(document)) != 0 {
		t.Fatalf("got entries %v", Entries(document))
	}
	if _, err := Parse(zstdBody(t, plain), "gzip", 1024); err == nil {
		t.Fatal("expected non-JSON body with incorrect header to fail")
	}
}

func TestCandidateBodiesPreferResponseThenRequest(t *testing.T) {
	response := []byte{0x00, 0x01, 0xff}
	entry := map[string]any{
		"request": map[string]any{
			"url":      "https://example.test/user/42/mysekai",
			"postData": map[string]any{"text": "request"},
		},
		"response": map[string]any{
			"content": map[string]any{
				"encoding": "base64",
				"text":     base64.StdEncoding.EncodeToString(response),
			},
		},
	}
	bodies, err := CandidateBodies(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(bodies) != 2 || !bytes.Equal(bodies[0], response) || string(bodies[1]) != "request" {
		t.Fatalf("unexpected candidates: %#v", bodies)
	}
	if got := RequestURL(entry); got != "https://example.test/user/42/mysekai" {
		t.Fatalf("got URL %q", got)
	}
}

func TestEntriesAndContentToBytesHandleMalformedValues(t *testing.T) {
	if entries := Entries(map[string]any{}); len(entries) != 0 {
		t.Fatalf("got %#v", entries)
	}
	if body, present, err := ContentToBytes(map[string]any{"text": 42}); err != nil || present || body != nil {
		t.Fatalf("got body=%v present=%v err=%v", body, present, err)
	}
	if _, _, err := ContentToBytes(map[string]any{"encoding": "base64", "text": "%%%"}); err == nil {
		t.Fatal("expected invalid base64 error")
	}

	document := map[string]any{"log": map[string]any{"entries": []any{map[string]any{"request": map[string]any{}}}}}
	encoded, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(encoded, "", 1024)
	if err != nil || len(Entries(parsed)) != 1 {
		t.Fatalf("parsed=%v err=%v", parsed, err)
	}
}

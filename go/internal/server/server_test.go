package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
	"github.com/mouse233/MySekaiMapper/go/internal/service"
	"github.com/vmihailenco/msgpack/v5"
)

type submittedArchive struct {
	data   []byte
	userID string
	taskID string
}

type recordingSubmitter struct {
	mu    sync.Mutex
	items []submittedArchive
	err   error
}

func (s *recordingSubmitter) Submit(_ context.Context, data []byte, userID, taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	s.items = append(s.items, submittedArchive{data: append([]byte{}, data...), userID: userID, taskID: taskID})
	return nil
}

func (s *recordingSubmitter) snapshot() []submittedArchive {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]submittedArchive{}, s.items...)
}

func newTestHandler(t *testing.T, config Config, submitter *recordingSubmitter) *Handler {
	t.Helper()
	chunks, err := service.NewChunkStore(t.TempDir(), config.MaxTotalSize)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := New(config, chunks, submitter)
	if err != nil {
		t.Fatal(err)
	}
	handler.newTaskID = func() (string, error) { return "reporttask01", nil }
	return handler
}

func baseConfig() Config {
	return Config{
		MaxTotalSize:  1024,
		MaxChunkSize:  1024,
		MaxChunks:     10,
		ReportEnabled: true,
		ReportPath:    "/reqable/report",
		ReportMaxSize: 1024,
	}
}

func TestNewRejectsReportPathCollision(t *testing.T) {
	config := baseConfig()
	config.ReportPath = "/uploadMySekai"
	chunks, err := service.NewChunkStore(t.TempDir(), config.MaxTotalSize)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := New(config, chunks, &recordingSubmitter{}); err == nil {
		t.Fatal("expected route collision error")
	}
}

func TestChunkUploadMergesAndSubmits(t *testing.T) {
	submitter := &recordingSubmitter{}
	handler := newTestHandler(t, baseConfig(), submitter)
	parts := [][]byte{[]byte("part-0"), []byte("part-1"), []byte("part-2")}
	for index, part := range parts {
		request := httptest.NewRequest(http.MethodPost, "/uploadMySekai", bytes.NewReader(part))
		request.Header.Set("X-Upload-Id", "regression01")
		request.Header.Set("X-Chunk-Index", string(rune('0'+index)))
		request.Header.Set("X-Total-Chunks", "3")
		request.Header.Set("X-Original-Url", "https://example.test/user/888/mysekai")
		response := httptest.NewRecorder()
		handler.Routes().ServeHTTP(response, request)
		if response.Code != http.StatusOK || response.Body.String() != "OK" {
			t.Fatalf("chunk %d: status=%d body=%q", index, response.Code, response.Body.String())
		}
	}
	items := submitter.snapshot()
	if len(items) != 1 || string(items[0].data) != "part-0part-1part-2" || items[0].userID != "888" || items[0].taskID != "regression01" {
		t.Fatalf("submissions=%#v", items)
	}
}

func TestChunkUploadRetainsCompletedPartsWhenSubmitFails(t *testing.T) {
	config := baseConfig()
	chunks, err := service.NewChunkStore(t.TempDir(), config.MaxTotalSize)
	if err != nil {
		t.Fatal(err)
	}
	submitter := &recordingSubmitter{err: errors.New("temporary storage failure")}
	handler, err := New(config, chunks, submitter)
	if err != nil {
		t.Fatal(err)
	}
	for index, body := range []string{"left", "right"} {
		request := httptest.NewRequest(http.MethodPost, "/uploadMySekai", bytes.NewBufferString(body))
		request.Header.Set("X-Upload-Id", "retryable")
		request.Header.Set("X-Chunk-Index", string(rune('0'+index)))
		request.Header.Set("X-Total-Chunks", "2")
		response := httptest.NewRecorder()
		handler.Routes().ServeHTTP(response, request)
		if index == 0 && response.Code != http.StatusOK {
			t.Fatalf("incomplete chunk status=%d", response.Code)
		}
		if index == 1 && response.Code != http.StatusInternalServerError {
			t.Fatalf("failed submit status=%d", response.Code)
		}
	}
	if _, err := os.Stat(filepath.Join(chunks.Root, "retryable", "chunk_0")); err != nil {
		t.Fatalf("first chunk not retained: %v", err)
	}
	submitter.err = nil
	request := httptest.NewRequest(http.MethodPost, "/uploadMySekai", bytes.NewBufferString("right"))
	request.Header.Set("X-Upload-Id", "retryable")
	request.Header.Set("X-Chunk-Index", "1")
	request.Header.Set("X-Total-Chunks", "2")
	response := httptest.NewRecorder()
	handler.Routes().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("retry status=%d body=%q", response.Code, response.Body.String())
	}
	items := submitter.snapshot()
	if len(items) != 1 || string(items[0].data) != "leftright" {
		t.Fatalf("submissions=%#v", items)
	}
	if _, err := os.Stat(filepath.Join(chunks.Root, "retryable")); !os.IsNotExist(err) {
		t.Fatalf("completed chunks remain: %v", err)
	}
}

func TestChunkUploadRejectsBadParametersAndSize(t *testing.T) {
	submitter := &recordingSubmitter{}
	config := baseConfig()
	config.MaxChunkSize = 4
	handler := newTestHandler(t, config, submitter)

	cases := []struct {
		name    string
		headers map[string]string
		body    string
		status  int
	}{
		{"missing headers", map[string]string{}, "x", 422},
		{"invalid integer", map[string]string{"X-Upload-Id": "good", "X-Chunk-Index": "nope", "X-Total-Chunks": "1"}, "x", 422},
		{"bad id", map[string]string{"X-Upload-Id": "bad id", "X-Chunk-Index": "0", "X-Total-Chunks": "1"}, "x", 400},
		{"bad total", map[string]string{"X-Upload-Id": "good", "X-Chunk-Index": "0", "X-Total-Chunks": "11"}, "x", 400},
		{"bad index", map[string]string{"X-Upload-Id": "good", "X-Chunk-Index": "5", "X-Total-Chunks": "3"}, "x", 400},
		{"too large", map[string]string{"X-Upload-Id": "good", "X-Chunk-Index": "0", "X-Total-Chunks": "1"}, "12345", 413},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/uploadMySekai", stringsNewReader(tc.body))
			for key, value := range tc.headers {
				request.Header.Set(key, value)
			}
			response := httptest.NewRecorder()
			handler.Routes().ServeHTTP(response, request)
			if response.Code != tc.status {
				t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
			}
		})
	}
}

func TestReqableReportAcceptsGzipAndFirstValidArchive(t *testing.T) {
	t.Setenv("AES_KEY", "0123456789abcdef")
	t.Setenv("AES_IV", "fedcba9876543210")
	archive := validEncryptedArchive(t)
	document := map[string]any{
		"log": map[string]any{"entries": []any{
			map[string]any{
				"request":  map[string]any{"url": "https://api.test/ping"},
				"response": map[string]any{"content": map[string]any{"text": "{}"}},
			},
			map[string]any{
				"request": map[string]any{"url": "https://api.test/user/42/mysekai"},
				"response": map[string]any{"content": map[string]any{
					"encoding": "base64", "text": base64.StdEncoding.EncodeToString(archive),
				}},
			},
		}},
	}
	body, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	submitter := &recordingSubmitter{}
	handler := newTestHandler(t, baseConfig(), submitter)
	request := httptest.NewRequest(http.MethodPost, "/reqable/report", &compressed)
	request.Header.Set("Content-Encoding", "gzip")
	response := httptest.NewRecorder()
	handler.Routes().ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Body.String() != "ok" {
		t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
	}
	items := submitter.snapshot()
	if len(items) != 1 || !bytes.Equal(items[0].data, archive) || items[0].userID != "42" || !regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`).MatchString(items[0].taskID) {
		t.Fatalf("submissions=%#v", items)
	}
}

func TestReqableReportAcceptsBrotliAndZstd(t *testing.T) {
	t.Setenv("AES_KEY", "0123456789abcdef")
	t.Setenv("AES_IV", "fedcba9876543210")
	archive := validEncryptedArchive(t)
	document := map[string]any{"log": map[string]any{"entries": []any{map[string]any{
		"request": map[string]any{"url": "https://api.test/user/42/mysekai"},
		"response": map[string]any{"content": map[string]any{
			"encoding": "base64", "text": base64.StdEncoding.EncodeToString(archive),
		}},
	}}}}
	plain, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name     string
		encoding string
		compress func(*testing.T, []byte) []byte
	}{
		{
			name: "brotli", encoding: "br",
			compress: func(t *testing.T, input []byte) []byte {
				t.Helper()
				var output bytes.Buffer
				writer := brotli.NewWriter(&output)
				if _, err := writer.Write(input); err != nil {
					t.Fatal(err)
				}
				if err := writer.Close(); err != nil {
					t.Fatal(err)
				}
				return output.Bytes()
			},
		},
		{
			name: "zstd", encoding: "zstd",
			compress: func(t *testing.T, input []byte) []byte {
				t.Helper()
				var output bytes.Buffer
				writer, err := zstd.NewWriter(&output, zstd.WithEncoderConcurrency(1))
				if err != nil {
					t.Fatal(err)
				}
				if _, err := writer.Write(input); err != nil {
					t.Fatal(err)
				}
				if err := writer.Close(); err != nil {
					t.Fatal(err)
				}
				return output.Bytes()
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			submitter := &recordingSubmitter{}
			handler := newTestHandler(t, baseConfig(), submitter)
			request := httptest.NewRequest(http.MethodPost, "/reqable/report", bytes.NewReader(tc.compress(t, plain)))
			request.Header.Set("Content-Encoding", tc.encoding)
			response := httptest.NewRecorder()
			handler.Routes().ServeHTTP(response, request)
			if response.Code != http.StatusOK || response.Body.String() != "ok" || len(submitter.snapshot()) != 1 {
				t.Fatalf("status=%d body=%q submissions=%#v", response.Code, response.Body.String(), submitter.snapshot())
			}
		})
	}
}

func TestReqableReportHonorsDisabledTokenAndNoArchiveCases(t *testing.T) {
	submitter := &recordingSubmitter{}
	config := baseConfig()
	config.ReportEnabled = false
	handler := newTestHandler(t, config, submitter)
	request := httptest.NewRequest(http.MethodPost, "/reqable/report", bytes.NewBufferString("{}"))
	response := httptest.NewRecorder()
	handler.Routes().ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("disabled status=%d", response.Code)
	}

	config = baseConfig()
	config.ReportToken = "secret"
	handler = newTestHandler(t, config, submitter)
	request = httptest.NewRequest(http.MethodPost, "/reqable/report", bytes.NewBufferString("{}"))
	response = httptest.NewRecorder()
	handler.Routes().ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("token status=%d", response.Code)
	}

	config = baseConfig()
	handler = newTestHandler(t, config, submitter)
	request = httptest.NewRequest(http.MethodPost, "/reqable/report", bytes.NewBufferString(`{"log":{"entries":[]}}`))
	response = httptest.NewRecorder()
	handler.Routes().ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Body.String() != "ok" || len(submitter.snapshot()) != 0 {
		t.Fatalf("no archive status=%d body=%q calls=%v", response.Code, response.Body.String(), submitter.snapshot())
	}
}

func TestUserIDFromURL(t *testing.T) {
	if got := UserIDFromURL("https://example.test/user/123/mysekai"); got != "123" {
		t.Fatalf("got %q", got)
	}
	if got := UserIDFromURL("https://example.test/user/not-a-number"); got != "unknown" {
		t.Fatalf("got %q", got)
	}
}

func validEncryptedArchive(t *testing.T) []byte {
	t.Helper()
	plain, err := msgpack.Marshal(map[string]any{"updatedResources": map[string]any{"userMysekaiHarvestMaps": []any{}}})
	if err != nil {
		t.Fatal(err)
	}
	key := []byte("0123456789abcdef")
	iv := []byte("fedcba9876543210")
	pad := aes.BlockSize - len(plain)%aes.BlockSize
	padded := append(append([]byte{}, plain...), bytes.Repeat([]byte{byte(pad)}, pad)...)
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	output := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(output, padded)
	return output
}

func stringsNewReader(value string) *bytes.Reader {
	return bytes.NewReader([]byte(value))
}

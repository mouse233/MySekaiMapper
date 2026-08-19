// Package har parses Reqable HAR reports without retaining raw report bodies.
package har

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

var (
	ErrBodyTooLarge               = errors.New("HAR body exceeds configured limit")
	ErrUnsupportedContentEncoding = errors.New("unsupported Content-Encoding")
)

// DecompressBody decodes a request body according to Content-Encoding and
// bounds both compressed identity input and decompressed output.
func DecompressBody(raw []byte, contentEncoding string, maxSize int64) ([]byte, error) {
	if maxSize <= 0 {
		return nil, fmt.Errorf("invalid maximum body size %d", maxSize)
	}
	encoding := strings.ToLower(strings.TrimSpace(contentEncoding))
	switch encoding {
	case "", "identity":
		if int64(len(raw)) > maxSize {
			return nil, ErrBodyTooLarge
		}
		return raw, nil
	case "gzip", "x-gzip":
		reader, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return readLimited(reader, maxSize)
	case "br":
		return readLimited(brotli.NewReader(bytes.NewReader(raw)), maxSize)
	case "zstd", "zstandard":
		reader, err := zstd.NewReader(bytes.NewReader(raw), zstd.WithDecoderConcurrency(1))
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return readLimited(reader, maxSize)
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedContentEncoding, contentEncoding)
	}
}

func readLimited(reader io.Reader, maxSize int64) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(reader, maxSize+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxSize {
		return nil, ErrBodyTooLarge
	}
	return body, nil
}

// Parse decodes a HAR JSON object. Reqable occasionally sends an uncompressed
// JSON body with a compression header; when decompression fails, this preserves
// a JSON-only fallback for that malformed-client behavior.
func Parse(raw []byte, contentEncoding string, maxSize int64) (map[string]any, error) {
	body, err := DecompressBody(raw, contentEncoding, maxSize)
	if err != nil {
		trimmed := bytes.TrimSpace(raw)
		if !bytes.HasPrefix(trimmed, []byte("{")) && !bytes.HasPrefix(trimmed, []byte("[")) {
			return nil, err
		}
		if int64(len(raw)) > maxSize {
			return nil, ErrBodyTooLarge
		}
		body = raw
	}
	var document map[string]any
	if err := json.Unmarshal(body, &document); err != nil {
		return nil, fmt.Errorf("decode HAR JSON: %w", err)
	}
	return document, nil
}

// Entries returns log.entries, treating malformed or missing data as empty.
func Entries(document map[string]any) []map[string]any {
	log, ok := object(document["log"])
	if !ok {
		return nil
	}
	values, ok := array(log["entries"])
	if !ok {
		return nil
	}
	entries := make([]map[string]any, 0, len(values))
	for _, value := range values {
		if entry, ok := object(value); ok {
			entries = append(entries, entry)
		}
	}
	return entries
}

// CandidateBodies returns non-empty response content before request postData,
// matching the MySekai response-first extraction rule.
func CandidateBodies(entry map[string]any) ([][]byte, error) {
	candidates := make([][]byte, 0, 2)
	if response, ok := object(entry["response"]); ok {
		body, present, err := ContentToBytes(response["content"])
		if err != nil {
			return nil, err
		}
		if present && len(body) > 0 {
			candidates = append(candidates, body)
		}
	}
	if request, ok := object(entry["request"]); ok {
		body, present, err := ContentToBytes(request["postData"])
		if err != nil {
			return nil, err
		}
		if present && len(body) > 0 {
			candidates = append(candidates, body)
		}
	}
	return candidates, nil
}

// ContentToBytes restores a HAR response.content or request.postData body.
func ContentToBytes(value any) ([]byte, bool, error) {
	content, ok := object(value)
	if !ok {
		return nil, false, nil
	}
	text, exists := content["text"]
	if !exists || text == nil {
		return nil, false, nil
	}
	var body []byte
	switch typed := text.(type) {
	case string:
		body = []byte(typed)
	case []byte:
		body = typed
	default:
		return nil, false, nil
	}
	encoding, _ := content["encoding"].(string)
	if encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(string(body))
		if err != nil {
			return nil, false, fmt.Errorf("decode HAR base64 content: %w", err)
		}
		return decoded, true, nil
	}
	return body, true, nil
}

// RequestURL extracts a HAR entry request URL or returns an empty string.
func RequestURL(entry map[string]any) string {
	request, ok := object(entry["request"])
	if !ok {
		return ""
	}
	url, _ := request["url"].(string)
	return url
}

func object(value any) (map[string]any, bool) {
	result, ok := value.(map[string]any)
	return result, ok
}

func array(value any) ([]any, bool) {
	result, ok := value.([]any)
	return result, ok
}

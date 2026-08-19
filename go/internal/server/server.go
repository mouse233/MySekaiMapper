// Package server implements the capture-client and Reqable HTTP contracts.
package server

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/mouse233/MySekaiMapper/go/internal/har"
	"github.com/mouse233/MySekaiMapper/go/internal/mapper"
	"github.com/mouse233/MySekaiMapper/go/internal/service"
)

const (
	DefaultMaxChunks = 10
)

var uploadIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
var userIDPattern = regexp.MustCompile(`/user/(\d+)`)

// Config mirrors the existing Python endpoint limits and report configuration.
type Config struct {
	MaxTotalSize  int64
	MaxChunkSize  int64
	MaxChunks     int
	ReportEnabled bool
	ReportPath    string
	ReportMaxSize int64
	ReportToken   string
}

func ConfigFromSettings(settings service.Settings) Config {
	return Config{
		MaxTotalSize:  mapper.MaxArchiveSize,
		MaxChunkSize:  mapper.MaxArchiveSize,
		MaxChunks:     DefaultMaxChunks,
		ReportEnabled: settings.ReportEnabled,
		ReportPath:    settings.ReportPath,
		ReportMaxSize: settings.ReportMaxSize,
		ReportToken:   settings.ReportToken,
	}
}

// Submitter persists a complete encrypted archive and launches its background
// processing pipeline. It is small enough to substitute in HTTP contract tests.
type Submitter interface {
	Submit(ctx context.Context, data []byte, userID, taskID string) error
}

type chunkAdder interface {
	Add(uploadID string, index, total int, data []byte) ([]byte, bool, error)
	MarkSubmitted(uploadID string) error
	Complete(uploadID string) error
}

// Handler exposes both POST endpoints without logging archive bodies or tokens.
type Handler struct {
	config    Config
	chunks    chunkAdder
	submitter Submitter
	newTaskID func() (string, error)
	uploadMu  sync.Mutex

	// Logf receives operational messages without archive bodies, URLs, or tokens.
	// It is optional so library users and tests can stay silent.
	Logf func(format string, args ...any)
}

func New(config Config, chunks chunkAdder, submitter Submitter) (*Handler, error) {
	if config.MaxTotalSize <= 0 || config.MaxChunkSize <= 0 || config.MaxChunks < 1 {
		return nil, fmt.Errorf("invalid upload limits")
	}
	if config.ReportMaxSize <= 0 {
		return nil, fmt.Errorf("invalid report body limit")
	}
	if config.ReportPath == "" {
		config.ReportPath = "/reqable/report"
	}
	if !strings.HasPrefix(config.ReportPath, "/") {
		config.ReportPath = "/" + config.ReportPath
	}
	if config.ReportPath == "/uploadMySekai" {
		return nil, fmt.Errorf("report path conflicts with upload endpoint")
	}
	if chunks == nil || submitter == nil {
		return nil, fmt.Errorf("chunk store and submitter are required")
	}
	return &Handler{config: config, chunks: chunks, submitter: submitter, newTaskID: randomTaskID}, nil
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/uploadMySekai", h.uploadChunk)
	mux.HandleFunc(h.config.ReportPath, h.reqableReport)
	return mux
}

func (h *Handler) uploadChunk(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		response.Header().Set("Allow", http.MethodPost)
		http.Error(response, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !hasHeader(request, "X-Upload-Id") || !hasHeader(request, "X-Chunk-Index") || !hasHeader(request, "X-Total-Chunks") {
		http.Error(response, "Missing required upload headers", http.StatusUnprocessableEntity)
		return
	}
	uploadID := request.Header.Get("X-Upload-Id")
	if !uploadIDPattern.MatchString(uploadID) {
		http.Error(response, "Invalid upload id", http.StatusBadRequest)
		return
	}
	index, indexOK := headerInt(request, "X-Chunk-Index")
	total, totalOK := headerInt(request, "X-Total-Chunks")
	if !indexOK || !totalOK {
		http.Error(response, "Invalid integer upload headers", http.StatusUnprocessableEntity)
		return
	}
	if total < 1 || total > h.config.MaxChunks || index < 0 || index >= total {
		http.Error(response, "Invalid chunk parameters", http.StatusBadRequest)
		return
	}
	playerID := UserIDFromURL(request.Header.Get("X-Original-Url"))
	body, err := readBody(request, h.config.MaxChunkSize)
	if errors.Is(err, har.ErrBodyTooLarge) {
		http.Error(response, "Chunk too large", http.StatusRequestEntityTooLarge)
		return
	}
	if err != nil {
		http.Error(response, "Invalid request body", http.StatusBadRequest)
		return
	}
	h.logf("[UPLOAD] received task=%s player_id=%s chunk=%d/%d bytes=%d", uploadID, playerID, index+1, total, len(body))
	// Serialize the short merge/submit handoff so simultaneous retries for the
	// same upload ID cannot launch duplicate background jobs before marking it.
	h.uploadMu.Lock()
	defer h.uploadMu.Unlock()
	merged, complete, err := h.chunks.Add(uploadID, index, total, body)
	if errors.Is(err, service.ErrChunkAlreadySubmitted) {
		// The raw archive was already saved; retry only the private cleanup and
		// acknowledge without launching a duplicate background task.
		_ = h.chunks.Complete(uploadID)
		writePlain(response, http.StatusOK, "OK")
		return
	}
	if err != nil {
		writeChunkError(response, err)
		return
	}
	if complete {
		if err := h.submitter.Submit(request.Context(), merged, playerID, uploadID); err != nil {
			// Keep chunks on disk so a retry of the last chunk can submit the
			// same merged archive after a transient storage failure.
			writeSubmitError(response, err)
			return
		}
		h.logf("[UPLOAD] accepted task=%s player_id=%s bytes=%d", uploadID, playerID, len(merged))
		if err := h.chunks.MarkSubmitted(uploadID); err != nil {
			http.Error(response, "Unable to finalize completed upload", http.StatusInternalServerError)
			return
		}
		// Cleanup is best-effort: the archive is already safely persisted and the
		// submitted marker suppresses duplicate work if the client retries.
		_ = h.chunks.Complete(uploadID)
	}
	writePlain(response, http.StatusOK, "OK")
}

func (h *Handler) reqableReport(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		response.Header().Set("Allow", http.MethodPost)
		http.Error(response, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.config.ReportEnabled {
		http.Error(response, "Report server disabled", http.StatusNotFound)
		return
	}
	if h.config.ReportToken != "" && subtle.ConstantTimeCompare([]byte(request.Header.Get("X-Report-Token")), []byte(h.config.ReportToken)) != 1 {
		http.Error(response, "Invalid report token", http.StatusUnauthorized)
		return
	}
	raw, err := readBody(request, h.config.ReportMaxSize)
	if errors.Is(err, har.ErrBodyTooLarge) {
		http.Error(response, "Report body too large", http.StatusRequestEntityTooLarge)
		return
	}
	if err != nil {
		http.Error(response, "Invalid request body", http.StatusBadRequest)
		return
	}
	h.logf("[REPORT] received bytes=%d", len(raw))
	document, err := har.Parse(raw, request.Header.Get("Content-Encoding"), h.config.ReportMaxSize)
	if errors.Is(err, har.ErrBodyTooLarge) {
		http.Error(response, "Report body too large", http.StatusRequestEntityTooLarge)
		return
	}
	if err != nil {
		http.Error(response, "Invalid HAR body", http.StatusBadRequest)
		return
	}

	for _, entry := range har.Entries(document) {
		candidates, candidateErr := har.CandidateBodies(entry)
		if candidateErr != nil {
			continue
		}
		for _, candidate := range candidates {
			if int64(len(candidate)) > h.config.MaxTotalSize {
				continue
			}
			valid, validationErr := mapper.LooksLikeArchive(candidate)
			if errors.Is(validationErr, mapper.ErrAESNotConfigured) {
				http.Error(response, "AES keys not configured", http.StatusInternalServerError)
				return
			}
			if validationErr != nil || !valid {
				continue
			}
			taskID, taskErr := h.newTaskID()
			if taskErr != nil {
				http.Error(response, "Unable to create task", http.StatusInternalServerError)
				return
			}
			playerID := UserIDFromURL(har.RequestURL(entry))
			if err := h.submitter.Submit(request.Context(), candidate, playerID, taskID); err != nil {
				writeSubmitError(response, err)
				return
			}
			h.logf("[REPORT] accepted task=%s player_id=%s bytes=%d", taskID, playerID, len(candidate))
			writePlain(response, http.StatusOK, "ok")
			return
		}
	}
	// A report may contain unrelated API sessions. Reqable does not retry, so a
	// syntactically valid report with no archive is still acknowledged.
	h.logf("[REPORT] no MySekai archive found")
	writePlain(response, http.StatusOK, "ok")
}

func (h *Handler) logf(format string, args ...any) {
	if h.Logf != nil {
		h.Logf(format, args...)
	}
}

// UserIDFromURL follows the same /user/<digits> extraction rule as Python.
func UserIDFromURL(originalURL string) string {
	matches := userIDPattern.FindStringSubmatch(originalURL)
	if len(matches) == 2 {
		return matches[1]
	}
	return "unknown"
}

func hasHeader(request *http.Request, name string) bool {
	return len(request.Header.Values(name)) > 0
}

func headerInt(request *http.Request, name string) (int, bool) {
	value := request.Header.Get(name)
	parsed, err := strconv.Atoi(value)
	return parsed, err == nil
}

func readBody(request *http.Request, maxSize int64) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(request.Body, maxSize+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxSize {
		return nil, har.ErrBodyTooLarge
	}
	return body, nil
}

func writeChunkError(response http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrChunkTotalTooLarge):
		http.Error(response, "Total file too large", http.StatusRequestEntityTooLarge)
	case errors.Is(err, service.ErrInvalidChunk), errors.Is(err, service.ErrChunkTotalMismatch):
		http.Error(response, "Invalid chunk parameters", http.StatusBadRequest)
	default:
		http.Error(response, "Unable to store chunk", http.StatusInternalServerError)
	}
}

func writeSubmitError(response http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrSubmitterClosed) {
		http.Error(response, "Service is shutting down", http.StatusServiceUnavailable)
		return
	}
	http.Error(response, "Unable to save archive", http.StatusInternalServerError)
}

func writePlain(response http.ResponseWriter, status int, body string) {
	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response.WriteHeader(status)
	_, _ = io.WriteString(response, body)
}

func randomTaskID() (string, error) {
	buffer := make([]byte, 6)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

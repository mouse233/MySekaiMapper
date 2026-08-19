package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrChunkTotalTooLarge    = errors.New("combined upload exceeds configured limit")
	ErrChunkTotalMismatch    = errors.New("upload id was reused with a different total chunk count")
	ErrChunkAlreadySubmitted = errors.New("merged upload was already submitted")
	ErrInvalidChunk          = errors.New("invalid chunk parameters")
)

var (
	validTaskID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
	validUserID = regexp.MustCompile(`^\d+$`)
	chunkName   = regexp.MustCompile(`^chunk_(\d+)$`)
)

// Store owns service runtime directories. It never treats a user-supplied ID
// as a path component without validation.
type Store struct {
	RawDir     string
	TmpDir     string
	ArchiveDir string
	LatestDir  string

	rawMu     sync.Mutex
	archiveMu sync.Mutex
	latestMu  sync.Mutex
}

func NewStore(settings Settings) (*Store, error) {
	store := &Store{
		RawDir:     settings.RawDir,
		TmpDir:     settings.TmpDir,
		ArchiveDir: settings.ArchiveDir,
		LatestDir:  settings.LatestDir,
	}
	for _, directory := range []struct {
		path string
		mode os.FileMode
	}{
		{store.RawDir, 0o700},
		{store.TmpDir, 0o700},
		{store.ArchiveDir, 0o755},
		{store.LatestDir, 0o755},
	} {
		if err := os.MkdirAll(directory.path, directory.mode); err != nil {
			return nil, fmt.Errorf("create runtime directory %s: %w", directory.path, err)
		}
		if err := os.Chmod(directory.path, directory.mode); err != nil {
			return nil, fmt.Errorf("set runtime directory permissions for %s: %w", directory.path, err)
		}
	}
	return store, nil
}

// SaveRaw persists the encrypted archive with owner-only permissions before it
// is processed asynchronously.
func (s *Store) SaveRaw(data []byte, userID, taskID string) (string, error) {
	s.rawMu.Lock()
	defer s.rawMu.Unlock()
	if int64(len(data)) > 1*1024*1024 {
		return "", ErrChunkTotalTooLarge
	}
	userID = normalizeUserID(userID)
	if !validTaskID.MatchString(taskID) {
		return "", fmt.Errorf("invalid task id")
	}
	filename := fmt.Sprintf("mysekai_%s_%s.bin", userID, taskID)
	path := filepath.Join(s.RawDir, filename)
	for fileExists(path) {
		suffix, err := randomID(4)
		if err != nil {
			return "", err
		}
		path = filepath.Join(s.RawDir, fmt.Sprintf("mysekai_%s_%s_%s.bin", userID, taskID, suffix))
	}
	if err := atomicWrite(path, data, 0o600); err != nil {
		return "", fmt.Errorf("save raw archive: %w", err)
	}
	return path, nil
}

func normalizeUserID(value string) string {
	if validUserID.MatchString(value) {
		return value
	}
	return "unknown"
}

// NewJobOutput creates a private work directory. The cleanup function removes
// only that generated job directory after archive creation completes.
func (s *Store) NewJobOutput(taskID string) (string, func(), error) {
	if !validTaskID.MatchString(taskID) {
		return "", nil, fmt.Errorf("invalid task id")
	}
	id, err := randomID(8)
	if err != nil {
		return "", nil, err
	}
	base := filepath.Join(s.TmpDir, "jobs", taskID+"-"+id)
	output := filepath.Join(base, "output")
	if err := os.MkdirAll(output, 0o700); err != nil {
		return "", nil, fmt.Errorf("create job output directory: %w", err)
	}
	return output, func() { _ = os.RemoveAll(base) }, nil
}

// Archive copies one completed job output to its stable user/timestamp path.
func (s *Store) Archive(outputDir, userID string, now time.Time) (string, error) {
	s.archiveMu.Lock()
	defer s.archiveMu.Unlock()
	userID = normalizeUserID(userID)
	base := filepath.Join(s.ArchiveDir, "by-id", userID)
	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", fmt.Errorf("create archive user directory: %w", err)
	}
	name := now.Format("20060102_150405")
	target := filepath.Join(base, name)
	for index := 1; fileExists(target); index++ {
		target = filepath.Join(base, fmt.Sprintf("%s_%02d", name, index))
	}
	stagingID, err := randomID(6)
	if err != nil {
		return "", err
	}
	staging := target + ".staging-" + stagingID
	defer os.RemoveAll(staging)
	if err := copyDirectory(outputDir, staging); err != nil {
		return "", fmt.Errorf("archive generated output: %w", err)
	}
	if err := os.Rename(staging, target); err != nil {
		return "", fmt.Errorf("publish archive output: %w", err)
	}
	return target, nil
}

// PromoteLatest atomically swaps data/latest after a complete archive is ready.
// A mutex prevents two background jobs from exposing a partially copied latest
// directory to notification and static-file consumers.
func (s *Store) PromoteLatest(sourceDir string) error {
	s.latestMu.Lock()
	defer s.latestMu.Unlock()

	id, err := randomID(6)
	if err != nil {
		return err
	}
	staging := s.LatestDir + ".staging-" + id
	backup := s.LatestDir + ".backup-" + id
	defer os.RemoveAll(staging)
	defer os.RemoveAll(backup)

	if err := copyDirectory(sourceDir, staging); err != nil {
		return err
	}
	if fileExists(s.LatestDir) {
		if err := os.Rename(s.LatestDir, backup); err != nil {
			return fmt.Errorf("move previous latest output: %w", err)
		}
	}
	if err := os.Rename(staging, s.LatestDir); err != nil {
		if fileExists(backup) {
			_ = os.Rename(backup, s.LatestDir)
		}
		return fmt.Errorf("publish latest output: %w", err)
	}
	return nil
}

// ChunkStore persists bounded chunk uploads so the process does not retain
// request bodies longer than needed. All operations are serialized to preserve
// consistency for concurrent requests using the same upload ID.
type ChunkStore struct {
	Root         string
	MaxTotalSize int64
	mu           sync.Mutex
}

func NewChunkStore(root string, maxTotalSize int64) (*ChunkStore, error) {
	if maxTotalSize <= 0 {
		return nil, fmt.Errorf("invalid total upload limit %d", maxTotalSize)
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("create chunk directory: %w", err)
	}
	return &ChunkStore{Root: root, MaxTotalSize: maxTotalSize}, nil
}

// Add writes a chunk and returns a merged archive only after every index in
// [0,total) exists. The returned data is at most MaxTotalSize bytes.
func (s *ChunkStore) Add(uploadID string, index, total int, data []byte) ([]byte, bool, error) {
	if !validTaskID.MatchString(uploadID) || total < 1 || total > 10 || index < 0 || index >= total {
		return nil, false, ErrInvalidChunk
	}
	if int64(len(data)) > s.MaxTotalSize {
		return nil, false, ErrChunkTotalTooLarge
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	directory := filepath.Join(s.Root, uploadID)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, false, err
	}
	if _, err := os.Stat(filepath.Join(directory, ".submitted")); err == nil {
		return nil, false, ErrChunkAlreadySubmitted
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, false, err
	}
	metaPath := filepath.Join(directory, ".total")
	if existing, err := os.ReadFile(metaPath); err == nil {
		expected, parseErr := strconv.Atoi(strings.TrimSpace(string(existing)))
		if parseErr != nil || expected != total {
			return nil, false, ErrChunkTotalMismatch
		}
	} else if errors.Is(err, os.ErrNotExist) {
		if err := atomicWrite(metaPath, []byte(strconv.Itoa(total)), 0o600); err != nil {
			return nil, false, err
		}
	} else {
		return nil, false, err
	}

	target := filepath.Join(directory, fmt.Sprintf("chunk_%d", index))
	currentSize, err := chunkSize(directory)
	if err != nil {
		return nil, false, err
	}
	if info, err := os.Stat(target); err == nil && info.Mode().IsRegular() {
		currentSize -= info.Size()
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, false, err
	}
	if currentSize+int64(len(data)) > s.MaxTotalSize {
		return nil, false, ErrChunkTotalTooLarge
	}
	if err := atomicWrite(target, data, 0o600); err != nil {
		return nil, false, err
	}

	merged := make([]byte, 0, currentSize+int64(len(data)))
	for chunkIndex := 0; chunkIndex < total; chunkIndex++ {
		path := filepath.Join(directory, fmt.Sprintf("chunk_%d", chunkIndex))
		part, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}
		if err != nil {
			return nil, false, err
		}
		if int64(len(merged)+len(part)) > s.MaxTotalSize {
			return nil, false, ErrChunkTotalTooLarge
		}
		merged = append(merged, part...)
	}
	// Keep the chunks until the caller confirms the merged archive was safely
	// handed to its submitter. This makes a transient save/queue failure retryable.
	return merged, true, nil
}

// MarkSubmitted makes successful handoff idempotent if a later cleanup fails.
func (s *ChunkStore) MarkSubmitted(uploadID string) error {
	if !validTaskID.MatchString(uploadID) {
		return ErrInvalidChunk
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	directory := filepath.Join(s.Root, uploadID)
	if _, err := os.Stat(directory); err != nil {
		return err
	}
	return atomicWrite(filepath.Join(directory, ".submitted"), []byte("1"), 0o600)
}

// Complete removes a merged upload only after its raw archive was persisted.
func (s *ChunkStore) Complete(uploadID string) error {
	if !validTaskID.MatchString(uploadID) {
		return ErrInvalidChunk
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return os.RemoveAll(filepath.Join(s.Root, uploadID))
}

func chunkSize(directory string) (int64, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return 0, err
	}
	var size int64
	for _, entry := range entries {
		if !chunkName.MatchString(entry.Name()) || entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return 0, err
		}
		if info.Mode().IsRegular() {
			size += info.Size()
		}
	}
	return size, nil
}

func atomicWrite(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tempPath := file.Name()
	defer os.Remove(tempPath)
	defer file.Close()
	if err := file.Chmod(mode); err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

func copyDirectory(source, destination string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("source %s is not a directory", source)
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return err
	}
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		return copyFile(path, target)
	})
}

func copyFile(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	output, err := os.CreateTemp(filepath.Dir(destination), "."+filepath.Base(destination)+".*.tmp")
	if err != nil {
		return err
	}
	tempPath := output.Name()
	defer os.Remove(tempPath)
	defer output.Close()
	if err := output.Chmod(0o644); err != nil {
		return err
	}
	if _, err := io.Copy(output, input); err != nil {
		return err
	}
	if err := output.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, destination)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func randomID(bytesLength int) (string, error) {
	buffer := make([]byte, bytesLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

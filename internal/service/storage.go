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
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrChunkTotalTooLarge    = errors.New("combined upload exceeds configured limit")
	ErrChunkTotalMismatch    = errors.New("upload id was reused with a different total chunk count")
	ErrChunkAlreadySubmitted = errors.New("merged upload was already submitted")
	ErrChunkStoreFull        = errors.New("temporary upload storage quota exceeded")
	ErrInvalidChunk          = errors.New("invalid chunk parameters")
)

var (
	validTaskID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
	validUserID = regexp.MustCompile(`^\d+$`)
	chunkName   = regexp.MustCompile(`^chunk_(\d+)$`)
)

// MaxFailedRawArchives bounds how many unprocessable raw archives are kept for
// debugging. Older failures are deleted so bad input cannot fill the disk.
const MaxFailedRawArchives = 20

// Store owns service runtime directories. It never treats a user-supplied ID
// as a path component without validation.
//
// Accepted raw archives are written to PendingDir first. A successfully
// processed archive is moved to RawDir (the historical location), a failed one
// to FailedDir. Anything left in PendingDir after a crash is recovered on the
// next start.
type Store struct {
	RawDir     string
	PendingDir string
	FailedDir  string
	TmpDir     string
	ArchiveDir string
	LatestDir  string

	rawMu           sync.Mutex
	uploadQueueDir  string
	lastUploadPrune time.Time
	archiveMu       sync.Mutex
	latestMu        sync.Mutex
}

func NewStore(settings Settings) (*Store, error) {
	store := &Store{
		RawDir:     settings.RawDir,
		PendingDir: filepath.Join(settings.RawDir, "pending"),
		FailedDir:  filepath.Join(settings.RawDir, "failed"),
		TmpDir:     settings.TmpDir,
		ArchiveDir: settings.ArchiveDir,
		LatestDir:  settings.LatestDir,
	}
	store.uploadQueueDir = filepath.Join(store.PendingDir, ".uploads")
	for _, directory := range []struct {
		path string
		mode os.FileMode
	}{
		{store.RawDir, 0o700},
		{store.PendingDir, 0o700},
		{store.uploadQueueDir, 0o700},
		{store.FailedDir, 0o700},
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

// SaveRaw persists the encrypted archive with owner-only permissions in
// PendingDir before it is processed asynchronously.
func (s *Store) SaveRaw(data []byte, userID, taskID string) (string, error) {
	s.rawMu.Lock()
	defer s.rawMu.Unlock()
	name, err := rawArchiveName(data, userID, taskID)
	if err != nil {
		return "", err
	}
	filename := name
	// Keep names unique across pending and processed archives so a finished
	// archive never overwrites an earlier one when it is moved to RawDir.
	for fileExists(filepath.Join(s.PendingDir, filename)) || fileExists(filepath.Join(s.RawDir, filename)) {
		suffix, err := randomID(4)
		if err != nil {
			return "", err
		}
		filename = strings.TrimSuffix(name, ".bin") + "_" + suffix + ".bin"
	}
	path := filepath.Join(s.PendingDir, filename)
	if err := atomicWrite(path, data, 0o600); err != nil {
		return "", fmt.Errorf("save raw archive: %w", err)
	}
	return path, nil
}

func rawArchiveName(data []byte, userID, taskID string) (string, error) {
	if int64(len(data)) > 1*1024*1024 {
		return "", ErrChunkTotalTooLarge
	}
	if !validTaskID.MatchString(taskID) {
		return "", fmt.Errorf("invalid task id")
	}
	return fmt.Sprintf("mysekai_%s_%s.bin", normalizeUserID(userID), taskID), nil
}

// PendingArchive is a raw archive that was accepted but not yet processed.
type PendingArchive struct {
	Path   string
	UserID string
	TaskID string
}

type pendingFile struct {
	archive PendingArchive
	modTime time.Time
}

// PendingRaw lists archives left in PendingDir, oldest first. They were
// accepted by a previous process that stopped before finishing them.
func (s *Store) PendingRaw() ([]PendingArchive, error) {
	s.rawMu.Lock()
	defer s.rawMu.Unlock()
	files, err := appendPendingFiles(nil, s.PendingDir)
	if err != nil {
		return nil, err
	}
	uploads, err := os.ReadDir(s.uploadQueueDir)
	if err != nil {
		return nil, err
	}
	s.pruneUploadQueueLocked(uploads, time.Now())
	for _, upload := range uploads {
		if !upload.IsDir() || strings.HasPrefix(upload.Name(), uploadStagingPrefix) {
			continue
		}
		files, err = appendPendingFiles(files, filepath.Join(s.uploadQueueDir, upload.Name()))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}
	sort.SliceStable(files, func(i, j int) bool { return files[i].modTime.Before(files[j].modTime) })
	result := make([]PendingArchive, 0, len(files))
	for _, file := range files {
		result = append(result, file.archive)
	}
	return result, nil
}

func appendPendingFiles(files []pendingFile, directory string) ([]pendingFile, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return files, err
	}
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}
		userID, taskID, ok := parseRawName(entry.Name())
		if !ok {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return files, err
		}
		files = append(files, pendingFile{
			archive: PendingArchive{Path: filepath.Join(directory, entry.Name()), UserID: userID, TaskID: taskID},
			modTime: info.ModTime(),
		})
	}
	return files, nil
}

// parseRawName recovers routing metadata from mysekai_<user>_<task>[_<suffix>].bin.
// The user part never contains an underscore (digits or "unknown").
func parseRawName(name string) (string, string, bool) {
	if !strings.HasPrefix(name, "mysekai_") || !strings.HasSuffix(name, ".bin") {
		return "", "", false
	}
	userID, taskID, ok := strings.Cut(strings.TrimSuffix(strings.TrimPrefix(name, "mysekai_"), ".bin"), "_")
	if !ok || normalizeUserID(userID) != userID {
		return "", "", false
	}
	if !validTaskID.MatchString(taskID) {
		// A collision suffix can push a 64-character task ID over the limit.
		// The task ID only labels logs and notifications, so trim the suffix.
		if index := strings.LastIndex(taskID, "_"); index > 0 && validTaskID.MatchString(taskID[:index]) {
			taskID = taskID[:index]
		} else {
			return "", "", false
		}
	}
	return userID, taskID, true
}

// FinishRaw moves a processed archive out of PendingDir: to RawDir when
// processing succeeded, or to FailedDir otherwise. FailedDir keeps only the
// newest MaxFailedRawArchives files.
func (s *Store) FinishRaw(path string, succeeded bool) error {
	s.rawMu.Lock()
	defer s.rawMu.Unlock()
	directory := filepath.Dir(path)
	if directory != filepath.Clean(s.PendingDir) && filepath.Dir(directory) != filepath.Clean(s.uploadQueueDir) {
		return fmt.Errorf("raw archive is not in the pending directory")
	}
	destinationDir := s.RawDir
	if !succeeded {
		destinationDir = s.FailedDir
	}
	name := filepath.Base(path)
	destination := filepath.Join(destinationDir, name)
	for fileExists(destination) {
		suffix, err := randomID(4)
		if err != nil {
			return err
		}
		destination = filepath.Join(destinationDir, strings.TrimSuffix(name, ".bin")+"_"+suffix+".bin")
	}
	// For chunk uploads, the now-empty directory remains as the receipt.
	// Moving the archive updates its directory's modification time.
	if err := os.Rename(path, destination); err != nil {
		return fmt.Errorf("move processed raw archive: %w", err)
	}
	if !succeeded {
		return pruneOldestFiles(s.FailedDir, MaxFailedRawArchives)
	}
	return nil
}

// pruneOldestFiles keeps the newest keep regular files in directory.
func pruneOldestFiles(directory string, keep int) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	type fileInfo struct {
		path    string
		modTime time.Time
	}
	files := make([]fileInfo, 0, len(entries))
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, fileInfo{path: filepath.Join(directory, entry.Name()), modTime: info.ModTime()})
	}
	if len(files) <= keep {
		return nil
	}
	sort.SliceStable(files, func(i, j int) bool { return files[i].modTime.After(files[j].modTime) })
	var errorsSeen []error
	for _, file := range files[keep:] {
		if err := os.Remove(file.path); err != nil && !errors.Is(err, os.ErrNotExist) {
			errorsSeen = append(errorsSeen, err)
		}
	}
	return errors.Join(errorsSeen...)
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

const (
	// StaleUploadAge is how long an incomplete chunk upload may sit idle before
	// its directory is removed.
	StaleUploadAge = time.Hour
	// SubmittedRetention is how long a completed upload ID is remembered, so a
	// client retry of the last chunk is acknowledged instead of reprocessed.
	SubmittedRetention = 24 * time.Hour
	pruneInterval      = 5 * time.Minute
	maxPendingUploads  = 256
	maxPendingBytes    = 64 * 1024 * 1024
	// submittedDirName identifies receipts written by earlier versions.
	submittedDirName = ".submitted"
)

// ChunkStore persists bounded chunk uploads so the process does not retain
// request bodies longer than needed. All operations are serialized to preserve
// consistency for concurrent requests using the same upload ID.
type ChunkStore struct {
	Root           string
	MaxTotalSize   int64
	maxUploads     int
	maxBytes       int64
	mu             sync.Mutex
	lastPrune      time.Time
	pendingUploads int
	pendingBytes   int64
}

func NewChunkStore(root string, maxTotalSize int64) (*ChunkStore, error) {
	if maxTotalSize <= 0 {
		return nil, fmt.Errorf("invalid total upload limit %d", maxTotalSize)
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("create chunk directory: %w", err)
	}
	store := &ChunkStore{
		Root:         root,
		MaxTotalSize: maxTotalSize,
		maxUploads:   maxPendingUploads,
		maxBytes:     maxPendingBytes,
	}
	if err := store.scanUsageLocked(); err != nil {
		return nil, fmt.Errorf("measure temporary upload storage: %w", err)
	}
	return store, nil
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

	if now := time.Now(); now.Sub(s.lastPrune) >= pruneInterval {
		// Best-effort housekeeping; a failure must not reject a valid chunk.
		_ = s.pruneLocked(now)
	}

	directory := filepath.Join(s.Root, uploadID)
	// Check the marker before creating the directory, so a retry of an upload
	// that was already submitted and cleaned up leaves nothing behind.
	for _, marker := range []string{s.submittedPath(uploadID), filepath.Join(directory, ".submitted")} {
		if _, err := os.Stat(marker); err == nil {
			return nil, false, ErrChunkAlreadySubmitted
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, false, err
		}
	}
	metaPath := filepath.Join(directory, ".total")
	existing, metaErr := os.ReadFile(metaPath)
	if metaErr == nil {
		expected, parseErr := strconv.Atoi(strings.TrimSpace(string(existing)))
		if parseErr != nil || expected != total {
			return nil, false, ErrChunkTotalMismatch
		}
	} else if !errors.Is(metaErr, os.ErrNotExist) {
		return nil, false, metaErr
	}

	target := filepath.Join(directory, fmt.Sprintf("chunk_%d", index))
	currentSize, err := chunkSize(directory)
	newUpload := errors.Is(err, os.ErrNotExist)
	if err != nil && !newUpload {
		return nil, false, err
	}
	addedBytes := int64(len(data))
	if info, err := os.Stat(target); err == nil && info.Mode().IsRegular() {
		addedBytes -= info.Size()
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, false, err
	}
	mergedSize := currentSize + addedBytes
	if mergedSize > s.MaxTotalSize {
		return nil, false, ErrChunkTotalTooLarge
	}
	if (newUpload && s.pendingUploads >= s.maxUploads) || (addedBytes > 0 && addedBytes > s.maxBytes-s.pendingBytes) {
		return nil, false, ErrChunkStoreFull
	}
	if newUpload {
		if err := os.Mkdir(directory, 0o700); err != nil {
			return nil, false, err
		}
		s.pendingUploads++
	}
	if metaErr != nil {
		if err := atomicWrite(metaPath, []byte(strconv.Itoa(total)), 0o600); err != nil {
			return nil, false, err
		}
	}
	if err := atomicWrite(target, data, 0o600); err != nil {
		return nil, false, err
	}
	s.pendingBytes += addedBytes

	merged := make([]byte, 0, mergedSize)
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

// Complete removes a merged upload's chunks and releases their quota.
func (s *ChunkStore) Complete(uploadID string) error {
	if !validTaskID.MatchString(uploadID) {
		return ErrInvalidChunk
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	directory := filepath.Join(s.Root, uploadID)
	size, err := chunkSize(directory)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := os.RemoveAll(directory); err != nil {
		// Some files may have been removed before the error.
		return errors.Join(err, s.scanUsageLocked())
	}
	s.pendingUploads--
	s.pendingBytes -= size
	return nil
}

// PruneStale removes incomplete uploads idle for longer than StaleUploadAge
// and submitted records older than SubmittedRetention.
func (s *ChunkStore) PruneStale(now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pruneLocked(now)
}

func (s *ChunkStore) pruneLocked(now time.Time) error {
	s.lastPrune = now
	var errorsSeen []error
	entries, err := os.ReadDir(s.Root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() || !validTaskID.MatchString(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if now.Sub(info.ModTime()) > StaleUploadAge {
			if err := os.RemoveAll(filepath.Join(s.Root, entry.Name())); err != nil {
				errorsSeen = append(errorsSeen, err)
			}
		}
	}
	markers, err := os.ReadDir(filepath.Join(s.Root, submittedDirName))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		errorsSeen = append(errorsSeen, err)
	}
	for _, marker := range markers {
		info, err := marker.Info()
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		if now.Sub(info.ModTime()) > SubmittedRetention {
			if err := os.Remove(filepath.Join(s.Root, submittedDirName, marker.Name())); err != nil && !errors.Is(err, os.ErrNotExist) {
				errorsSeen = append(errorsSeen, err)
			}
		}
	}
	return errors.Join(append(errorsSeen, s.scanUsageLocked())...)
}

func (s *ChunkStore) scanUsageLocked() error {
	entries, err := os.ReadDir(s.Root)
	if err != nil {
		return err
	}
	var uploads int
	var bytes int64
	for _, entry := range entries {
		if !entry.IsDir() || !validTaskID.MatchString(entry.Name()) {
			continue
		}
		size, err := chunkSize(filepath.Join(s.Root, entry.Name()))
		if err != nil {
			return err
		}
		uploads++
		bytes += size
	}
	s.pendingUploads, s.pendingBytes = uploads, bytes
	return nil
}

func (s *ChunkStore) submittedPath(uploadID string) string {
	return filepath.Join(s.Root, submittedDirName, uploadID)
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

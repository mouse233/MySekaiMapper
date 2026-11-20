package service

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const uploadStagingPrefix = ".incoming-"

// UploadSubmitted reports whether the queue has accepted this upload ID.
func (s *Store) UploadSubmitted(uploadID string) (bool, error) {
	if !validTaskID.MatchString(uploadID) {
		return false, ErrInvalidChunk
	}
	s.rawMu.Lock()
	defer s.rawMu.Unlock()
	return s.uploadSubmittedLocked(s.uploadPath(uploadID))
}

// SaveUpload publishes the raw archive and its upload ID together. An empty
// path with a nil error means the upload ID was already accepted.
func (s *Store) SaveUpload(data []byte, userID, uploadID string) (string, error) {
	name, err := rawArchiveName(data, userID, uploadID)
	if err != nil {
		return "", err
	}
	s.rawMu.Lock()
	defer s.rawMu.Unlock()

	destination := s.uploadPath(uploadID)
	submitted, err := s.uploadSubmittedLocked(destination)
	if err != nil || submitted {
		return "", err
	}
	staging, err := os.MkdirTemp(s.uploadQueueDir, uploadStagingPrefix)
	if err != nil {
		return "", fmt.Errorf("stage upload: %w", err)
	}
	defer os.RemoveAll(staging)
	if err := atomicWrite(filepath.Join(staging, name), data, 0o600); err != nil {
		return "", fmt.Errorf("write staged upload: %w", err)
	}
	if err := os.Rename(staging, destination); err != nil {
		return "", fmt.Errorf("publish upload: %w", err)
	}
	return filepath.Join(destination, name), nil
}

func (s *Store) uploadPath(uploadID string) string {
	// Encode IDs so their case remains distinct even on case-insensitive filesystems.
	return filepath.Join(s.uploadQueueDir, hex.EncodeToString([]byte(uploadID)))
}

func (s *Store) uploadSubmittedLocked(path string) (bool, error) {
	if now := time.Now(); now.Sub(s.lastUploadPrune) >= pruneInterval {
		if uploads, err := os.ReadDir(s.uploadQueueDir); err == nil {
			s.pruneUploadQueueLocked(uploads, now)
		}
	}
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.IsDir() {
		return false, fmt.Errorf("upload receipt is not a directory")
	}
	return true, nil
}

func (s *Store) pruneUploadQueueLocked(uploads []os.DirEntry, now time.Time) {
	s.lastUploadPrune = now
	for _, upload := range uploads {
		if !upload.IsDir() {
			continue
		}
		info, err := upload.Info()
		if err != nil {
			continue
		}
		age := now.Sub(info.ModTime())
		path := filepath.Join(s.uploadQueueDir, upload.Name())
		if strings.HasPrefix(upload.Name(), uploadStagingPrefix) {
			if age > StaleUploadAge {
				_ = os.RemoveAll(path)
			}
		} else if age > SubmittedRetention {
			// Remove only empty receipts. A directory holding an accepted
			// archive remains in the queue regardless of its age.
			_ = os.Remove(path)
		}
	}
}

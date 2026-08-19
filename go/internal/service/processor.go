package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mouse233/MySekaiMapper/go/internal/mapper"
)

// OutputNotifier is intentionally small so service tests can use a fully local
// recorder instead of real Telegram or Bark endpoints.
type OutputNotifier interface {
	Notify(ctx context.Context, outputDir, taskID, playerID, imageBase string) error
}

// Processor converts one already-saved encrypted archive into maps, an archive
// directory, latest output, and optional notifications.
type Processor interface {
	Process(ctx context.Context, rawPath, userID, taskID string) error
}

// GenerationProcessor is the Go equivalent of Python's generate -> archive ->
// notify portion of the background pipeline.
type GenerationProcessor struct {
	Store         *Store
	ResourceCSV   string
	FontFile      string
	BarkImageBase string
	Notifier      OutputNotifier
	Now           func() time.Time
	Logf          func(format string, args ...any)
}

func NewGenerationProcessor(settings Settings, store *Store, notifier OutputNotifier) *GenerationProcessor {
	return &GenerationProcessor{
		Store:         store,
		ResourceCSV:   settings.ResourceCSV,
		FontFile:      settings.FontFile,
		BarkImageBase: settings.BarkImageBase,
		Notifier:      notifier,
		Now:           time.Now,
	}
}

func (p *GenerationProcessor) Process(ctx context.Context, rawPath, userID, taskID string) error {
	if p.Store == nil {
		return fmt.Errorf("processor store is not configured")
	}
	if !validTaskID.MatchString(taskID) {
		return fmt.Errorf("invalid task id")
	}

	started := time.Now()
	p.logf("[JOB] start task=%s", taskID)
	drops, err := mapper.ReadDrops(rawPath)
	if err != nil {
		return err
	}
	p.logf("[PARSE] complete task=%s drops=%d", taskID, len(drops))

	outputDir, cleanup, err := p.Store.NewJobOutput(taskID)
	if err != nil {
		return err
	}
	defer cleanup()
	result, err := mapper.Generate(drops, p.ResourceCSV, p.FontFile, outputDir)
	if err != nil {
		return err
	}
	p.logf("[RENDER] complete task=%s maps=%d", taskID, len(result.MapFiles))

	now := time.Now()
	if p.Now != nil {
		now = p.Now()
	}
	archiveDir, err := p.Store.Archive(outputDir, userID, now)
	if err != nil {
		return err
	}
	if err := p.Store.PromoteLatest(archiveDir); err != nil {
		return err
	}
	p.logf("[ARCHIVE] complete task=%s", taskID)

	if p.Notifier != nil {
		imageBase := archiveImageBase(p.BarkImageBase, normalizeUserID(userID), archiveDir)
		if err := p.Notifier.Notify(ctx, archiveDir, taskID, normalizeUserID(userID), imageBase); err != nil {
			p.logf("[NOTIFY] failed task=%s: %v", taskID, err)
		} else {
			p.logf("[NOTIFY] dispatch complete task=%s", taskID)
		}
	} else {
		p.logf("[NOTIFY] skipped task=%s", taskID)
	}
	p.logf("[DONE] task=%s elapsed=%s", taskID, time.Since(started).Round(time.Millisecond))
	return nil
}

func (p *GenerationProcessor) logf(format string, args ...any) {
	if p.Logf != nil {
		p.Logf(format, args...)
	}
}

func archiveImageBase(base, userID, archiveDir string) string {
	if strings.TrimSpace(base) == "" {
		return ""
	}
	return strings.TrimRight(base, "/") + "/archive/by-id/" + url.PathEscape(userID) + "/" + url.PathEscape(filepath.Base(archiveDir))
}

// ErrSubmitterClosed prevents callers from accepting an archive after service
// shutdown has begun.
var ErrSubmitterClosed = errors.New("background submitter is closed")

type queuedJob struct {
	path   string
	userID string
	taskID string
}

// AsyncSubmitter saves every accepted encrypted archive before placing only its
// small metadata record into a worker-managed pending queue. Worker count is
// bounded, while busy render workers never cause an accepted upload to be
// discarded from the in-process queue.
type AsyncSubmitter struct {
	Store     *Store
	Processor Processor
	Logf      func(format string, args ...any)
	Workers   int

	startOnce sync.Once
	mu        sync.Mutex
	ready     *sync.Cond
	pending   []queuedJob
	closed    bool
	workersWG sync.WaitGroup
}

func (s *AsyncSubmitter) Submit(_ context.Context, data []byte, userID, taskID string) error {
	if s.Store == nil || s.Processor == nil {
		return fmt.Errorf("submitter is not configured")
	}
	s.startWorkers()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrSubmitterClosed
	}
	// Hold the queue lock through SaveRaw so Close cannot accept a file that
	// would never be scheduled for processing.
	path, err := s.Store.SaveRaw(data, userID, taskID)
	if err != nil {
		return err
	}
	s.pending = append(s.pending, queuedJob{path: path, userID: userID, taskID: taskID})
	pending := len(s.pending)
	s.ready.Signal()
	if s.Logf != nil {
		s.Logf("[QUEUE] accepted task=%s bytes=%d pending=%d", taskID, len(data), pending)
	}
	return nil
}

// Close stops accepting work and waits for already accepted jobs to finish.
func (s *AsyncSubmitter) Close() {
	s.startWorkers()
	s.mu.Lock()
	s.closed = true
	s.ready.Broadcast()
	s.mu.Unlock()
	s.workersWG.Wait()
}

func (s *AsyncSubmitter) startWorkers() {
	s.startOnce.Do(func() {
		s.ready = sync.NewCond(&s.mu)
		workers := s.Workers
		if workers < 1 {
			workers = 1
		}
		s.workersWG.Add(workers)
		for index := 0; index < workers; index++ {
			go s.runWorker()
		}
	})
}

func (s *AsyncSubmitter) runWorker() {
	defer s.workersWG.Done()
	for {
		job, ok := s.nextJob()
		if !ok {
			return
		}
		s.processJob(job)
	}
}

func (s *AsyncSubmitter) nextJob() (queuedJob, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for len(s.pending) == 0 && !s.closed {
		s.ready.Wait()
	}
	if len(s.pending) == 0 {
		return queuedJob{}, false
	}
	job := s.pending[0]
	s.pending[0] = queuedJob{}
	s.pending = s.pending[1:]
	return job, true
}

func (s *AsyncSubmitter) processJob(job queuedJob) {
	defer func() {
		if recovered := recover(); recovered != nil && s.Logf != nil {
			s.Logf("[ERROR] processing panicked task=%s", job.taskID)
		}
	}()
	if err := s.Processor.Process(context.Background(), job.path, job.userID, job.taskID); err != nil && s.Logf != nil {
		s.Logf("[ERROR] processing failed task=%s: %v", job.taskID, err)
	}
}

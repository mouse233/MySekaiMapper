package service

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

func testSettings(t *testing.T) Settings {
	t.Helper()
	root := filepath.Join("..", "..", "..")
	dataDir := t.TempDir()
	return Settings{
		DataDir:     dataDir,
		RawDir:      filepath.Join(dataDir, "raw_mysekai"),
		TmpDir:      filepath.Join(dataDir, "tmp"),
		ArchiveDir:  filepath.Join(dataDir, "archive"),
		LatestDir:   filepath.Join(dataDir, "latest"),
		ResourceCSV: filepath.Join(root, "assets", "resourceId.csv"),
		FontFile:    filepath.Join(root, "assets", "NotoSansSC-Regular.ttf"),
	}
}

func TestChunkStoreMergesOutOfOrderAndCleansUp(t *testing.T) {
	store, err := NewChunkStore(t.TempDir(), 1024)
	if err != nil {
		t.Fatal(err)
	}
	if merged, complete, err := store.Add("job_1", 1, 3, []byte("middle")); err != nil || complete || merged != nil {
		t.Fatalf("first add: merged=%q complete=%v err=%v", merged, complete, err)
	}
	if _, complete, err := store.Add("job_1", 2, 3, []byte("end")); err != nil || complete {
		t.Fatalf("second add complete=%v err=%v", complete, err)
	}
	merged, complete, err := store.Add("job_1", 0, 3, []byte("start-"))
	if err != nil || !complete || string(merged) != "start-middleend" {
		t.Fatalf("merged=%q complete=%v err=%v", merged, complete, err)
	}
	if _, err := os.Stat(filepath.Join(store.Root, "job_1")); err != nil {
		t.Fatalf("retryable chunk directory missing: %v", err)
	}
	if err := store.MarkSubmitted("job_1"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.Add("job_1", 0, 3, []byte("start-")); err != ErrChunkAlreadySubmitted {
		t.Fatalf("got %v, want submitted marker", err)
	}
	if err := store.Complete("job_1"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(store.Root, "job_1")); !os.IsNotExist(err) {
		t.Fatalf("chunk directory remains after completion: %v", err)
	}
}

func TestChunkStoreRejectsMismatchedAndOversizedData(t *testing.T) {
	store, err := NewChunkStore(t.TempDir(), 4)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.Add("same", 0, 2, []byte("a")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.Add("same", 1, 3, []byte("b")); err != ErrChunkTotalMismatch {
		t.Fatalf("got %v", err)
	}
	if _, _, err := store.Add("large", 0, 1, []byte("12345")); err != ErrChunkTotalTooLarge {
		t.Fatalf("got %v", err)
	}
}

type blockingProcessor struct {
	started  chan string
	release  <-chan struct{}
	finished chan string
}

func (p *blockingProcessor) Process(_ context.Context, _ string, _ string, taskID string) error {
	p.started <- taskID
	<-p.release
	p.finished <- taskID
	return nil
}

func TestAsyncSubmitterQueuesAndDrainsJobs(t *testing.T) {
	store, err := NewStore(testSettings(t))
	if err != nil {
		t.Fatal(err)
	}
	release := make(chan struct{})
	processor := &blockingProcessor{
		started:  make(chan string, 3),
		release:  release,
		finished: make(chan string, 3),
	}
	submitter := &AsyncSubmitter{Store: store, Processor: processor, Workers: 1}
	if err := submitter.Submit(context.Background(), []byte("one"), "42", "task1"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-processor.started:
	case <-time.After(time.Second):
		t.Fatal("first job did not start")
	}
	for _, taskID := range []string{"task2", "task3"} {
		if err := submitter.Submit(context.Background(), []byte(taskID), "42", taskID); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(store.RawDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("got raw files %v, want all accepted jobs", entries)
	}
	close(release)
	for index := 0; index < 3; index++ {
		select {
		case <-processor.finished:
		case <-time.After(time.Second):
			t.Fatal("queued job did not finish")
		}
	}
	submitter.Close()
	if err := submitter.Submit(context.Background(), []byte("late"), "42", "task4"); err != ErrSubmitterClosed {
		t.Fatalf("got %v, want closed submitter", err)
	}
}

type panicThenProcessor struct {
	mu        sync.Mutex
	calls     int
	completed chan struct{}
}

func (p *panicThenProcessor) Process(_ context.Context, _ string, _ string, _ string) error {
	p.mu.Lock()
	p.calls++
	call := p.calls
	p.mu.Unlock()
	if call == 1 {
		panic("synthetic processor panic")
	}
	close(p.completed)
	return nil
}

func TestAsyncSubmitterWorkerSurvivesProcessorPanic(t *testing.T) {
	store, err := NewStore(testSettings(t))
	if err != nil {
		t.Fatal(err)
	}
	processor := &panicThenProcessor{completed: make(chan struct{})}
	submitter := &AsyncSubmitter{Store: store, Processor: processor, Workers: 1}
	if err := submitter.Submit(context.Background(), []byte("first"), "42", "panic1"); err != nil {
		t.Fatal(err)
	}
	if err := submitter.Submit(context.Background(), []byte("second"), "42", "panic2"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-processor.completed:
	case <-time.After(time.Second):
		t.Fatal("worker did not continue after panic")
	}
	submitter.Close()
}

func TestStoreKeepsSameTaskRawArchivesDistinct(t *testing.T) {
	store, err := NewStore(testSettings(t))
	if err != nil {
		t.Fatal(err)
	}
	first, err := store.SaveRaw([]byte("first"), "42", "same-task")
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.SaveRaw([]byte("second"), "42", "same-task")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatalf("raw paths collided: %q", first)
	}
	firstData, err := os.ReadFile(first)
	if err != nil || string(firstData) != "first" {
		t.Fatalf("first=%q err=%v", firstData, err)
	}
	secondData, err := os.ReadFile(second)
	if err != nil || string(secondData) != "second" {
		t.Fatalf("second=%q err=%v", secondData, err)
	}
}

func TestStoreArchivesAndPromotesLatest(t *testing.T) {
	store, err := NewStore(testSettings(t))
	if err != nil {
		t.Fatal(err)
	}
	output := t.TempDir()
	if err := os.WriteFile(filepath.Join(output, "site_5.png"), []byte("map"), 0o644); err != nil {
		t.Fatal(err)
	}
	archive, err := store.Archive(output, "42", time.Date(2026, 8, 17, 1, 2, 3, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(archive, filepath.Join("42", "20260817_010203")) {
		t.Fatalf("archive=%q", archive)
	}
	if err := store.PromoteLatest(archive); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(store.LatestDir, "site_5.png"))
	if err != nil || string(got) != "map" {
		t.Fatalf("latest=%q err=%v", got, err)
	}
}

type recordedNotifier struct {
	outputDir string
	taskID    string
	playerID  string
	imageBase string
}

func (n *recordedNotifier) Notify(_ context.Context, outputDir, taskID, playerID, imageBase string) error {
	n.outputDir, n.taskID, n.playerID, n.imageBase = outputDir, taskID, playerID, imageBase
	return nil
}

func TestGenerationProcessorCreatesArchiveAndNotifies(t *testing.T) {
	settings := testSettings(t)
	store, err := NewStore(settings)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("AES_KEY", "0123456789abcdef")
	t.Setenv("AES_IV", "fedcba9876543210")
	payload := map[string]any{
		"updatedResources": map[string]any{
			"userMysekaiHarvestMaps": []any{map[string]any{
				"mysekaiSiteId": 5,
				"userMysekaiSiteHarvestResourceDrops": []any{map[string]any{
					"resourceId": 5, "positionX": 1.0, "positionZ": 2.0,
				}},
			}},
		},
	}
	plain, err := msgpack.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	rawPath, err := store.SaveRaw(encryptForServiceTest(t, plain), "42", "task123")
	if err != nil {
		t.Fatal(err)
	}
	notifier := &recordedNotifier{}
	processor := NewGenerationProcessor(settings, store, notifier)
	processor.BarkImageBase = "https://cdn.example"
	processor.Now = func() time.Time { return time.Date(2026, 8, 17, 1, 2, 3, 0, time.UTC) }
	if err := processor.Process(context.Background(), rawPath, "42", "task123"); err != nil {
		t.Fatal(err)
	}
	if notifier.outputDir == "" || notifier.taskID != "task123" || notifier.playerID != "42" {
		t.Fatalf("notifier=%#v", notifier)
	}
	if !strings.Contains(notifier.imageBase, "/archive/by-id/42/20260817_010203") {
		t.Fatalf("image base=%q", notifier.imageBase)
	}
	if _, err := os.Stat(filepath.Join(notifier.outputDir, "site_5.png")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(store.LatestDir, "rare_resources.txt")); err != nil {
		t.Fatal(err)
	}
}

func encryptForServiceTest(t *testing.T, plain []byte) []byte {
	t.Helper()
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

package notify

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestResolveMethodMatchesLegacyAndListForms(t *testing.T) {
	cases := []struct {
		input        any
		wantAliases  []string
		wantTelegram bool
	}{
		{nil, nil, false},
		{"none", nil, false},
		{"telegram", nil, true},
		{"klee+tg", []string{"klee"}, true},
		{"dodoco", []string{"dodoco"}, false},
		{[]any{"telegram", "dodoco"}, []string{"dodoco"}, true},
		{[]string{}, nil, false},
	}
	for _, tc := range cases {
		aliases, telegram := ResolveMethod(tc.input)
		if strings.Join(aliases, ",") != strings.Join(tc.wantAliases, ",") || telegram != tc.wantTelegram {
			t.Errorf("ResolveMethod(%#v) = (%#v, %v), want (%#v, %v)", tc.input, aliases, telegram, tc.wantAliases, tc.wantTelegram)
		}
	}
}

func TestBarkKeyExplicitValueWins(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bark_map.json")
	if err := os.WriteFile(path, []byte(`{"klee":"alias-key"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	notifier := New(Config{BarkMapFile: path})
	if got := notifier.BarkKey("klee", "explicit"); got != "explicit" {
		t.Fatalf("got %q", got)
	}
	if got := notifier.BarkKey("klee", ""); got != "alias-key" {
		t.Fatalf("got %q", got)
	}
	if got := notifier.BarkKey("missing", ""); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestNotifySendsBarkAndTelegramMediaGroup(t *testing.T) {
	dir := t.TempDir()
	outputDir := filepath.Join(dir, "output")
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "rare_resources.txt"), []byte("稀有资源统计\n\n钻石 × 1"), 0o600); err != nil {
		t.Fatal(err)
	}
	for siteID := 5; siteID <= 8; siteID++ {
		if err := os.WriteFile(filepath.Join(outputDir, "site_"+string(rune('0'+siteID))+".png"), []byte("not-a-real-png"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	barkMapPath := filepath.Join(dir, "bark_map.json")
	pushMapPath := filepath.Join(dir, "push_map.json")
	writeJSON(t, barkMapPath, map[string]string{"device": "device-key"})
	writeJSON(t, pushMapPath, map[string]any{"42": []string{"telegram", "device"}})

	var mu sync.Mutex
	paths := make([]string, 0)
	barkImages := make([]string, 0)
	var media string
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		paths = append(paths, request.URL.Path)
		if strings.HasPrefix(request.URL.Path, "/bark/") {
			barkImages = append(barkImages, request.URL.Query().Get("image"))
		}
		if strings.HasSuffix(request.URL.Path, "/sendMediaGroup") {
			if err := request.ParseMultipartForm(1 << 20); err != nil {
				t.Errorf("parse multipart: %v", err)
			}
			media = request.FormValue("media")
			if request.FormValue("chat_id") != "99" {
				t.Errorf("missing chat_id")
			}
		}
		response.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := New(Config{
		BarkMapFile:       barkMapPath,
		PushMapFile:       pushMapPath,
		BarkIcon:          "https://icon.example/icon.png",
		TelegramBotToken:  "token",
		TelegramChatID:    "99",
		BarkAPIBase:       server.URL + "/bark",
		TelegramAPIBase:   server.URL + "/telegram",
		HTTPClient:        server.Client(),
		AllowInsecureHTTP: true,
	})
	if err := notifier.Notify(context.Background(), outputDir, "task1", "42", "https://cdn.example/archive"); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(paths) != 6 {
		t.Fatalf("got paths %v, want 6 requests", paths)
	}
	if !strings.HasPrefix(paths[0], "/bark/device-key/") {
		t.Fatalf("unexpected Bark path %q", paths[0])
	}
	if paths[len(paths)-1] != "/telegram/bottoken/sendMediaGroup" {
		t.Fatalf("unexpected Telegram path %q", paths[len(paths)-1])
	}
	if len(barkImages) != 5 || barkImages[0] != "" || barkImages[1] != "https://cdn.example/archive/site_5.png" {
		t.Fatalf("unexpected Bark image URLs %v", barkImages)
	}
	for _, attachment := range []string{"attach://photo0", "attach://photo3"} {
		if !strings.Contains(media, attachment) {
			t.Fatalf("media %q missing %q", media, attachment)
		}
	}
}

func TestNotifyDefaultsToTelegramForEmptyPlayerRoute(t *testing.T) {
	dir := t.TempDir()
	outputDir := filepath.Join(dir, "output")
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "rare_resources.txt"), []byte("stats"), 0o600); err != nil {
		t.Fatal(err)
	}
	pushMapPath := filepath.Join(dir, "push_map.json")
	writeJSON(t, pushMapPath, map[string]any{"42": ""})
	var method string
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		method = filepath.Base(request.URL.Path)
		response.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	notifier := New(Config{
		PushMapFile:       pushMapPath,
		TelegramBotToken:  "token",
		TelegramChatID:    "99",
		TelegramAPIBase:   server.URL,
		HTTPClient:        server.Client(),
		AllowInsecureHTTP: true,
	})
	if err := notifier.Notify(context.Background(), outputDir, "task1", "42", ""); err != nil {
		t.Fatal(err)
	}
	if method != "sendMessage" {
		t.Fatalf("got method %q", method)
	}
}

func TestNotifyContinuesTelegramWhenBarkFails(t *testing.T) {
	dir := t.TempDir()
	outputDir := filepath.Join(dir, "output")
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "rare_resources.txt"), []byte("stats"), 0o600); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(dir, "bark_map.json"), map[string]string{"device": "key"})
	writeJSON(t, filepath.Join(dir, "push_map.json"), map[string]any{"42": []string{"device", "telegram"}})

	var telegramCalls int
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, "/bark/") {
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		telegramCalls++
		response.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := New(Config{
		BarkMapFile:       filepath.Join(dir, "bark_map.json"),
		PushMapFile:       filepath.Join(dir, "push_map.json"),
		TelegramBotToken:  "token",
		TelegramChatID:    "99",
		BarkAPIBase:       server.URL + "/bark",
		TelegramAPIBase:   server.URL + "/telegram",
		HTTPClient:        server.Client(),
		AllowInsecureHTTP: true,
	})
	if err := notifier.Notify(context.Background(), outputDir, "task1", "42", ""); err == nil {
		t.Fatal("expected Bark failures to be reported")
	}
	if telegramCalls != 1 {
		t.Fatalf("got %d Telegram calls, want 1", telegramCalls)
	}
}

func TestTelegramUsesTextPhotoAndMediaEndpoints(t *testing.T) {
	dir := t.TempDir()
	images := testImages(t, dir, 2)
	methods := make([]string, 0)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		methods = append(methods, filepath.Base(request.URL.Path))
		response.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	notifier := New(Config{
		TelegramBotToken:  "token",
		TelegramChatID:    "99",
		TelegramAPIBase:   server.URL,
		HTTPClient:        server.Client(),
		AllowInsecureHTTP: true,
	})
	if err := notifier.sendTelegram(context.Background(), nil, "caption"); err != nil {
		t.Fatal(err)
	}
	if err := notifier.sendTelegram(context.Background(), images[:1], "caption"); err != nil {
		t.Fatal(err)
	}
	if err := notifier.sendTelegram(context.Background(), images, "caption"); err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(methods, ","), "sendMessage,sendPhoto,sendMediaGroup"; got != want {
		t.Fatalf("got methods %q, want %q", got, want)
	}
}

func TestTelegramContinuesAfterEarlierBatchFailure(t *testing.T) {
	dir := t.TempDir()
	images := testImages(t, dir, 11)
	methods := make([]string, 0)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		method := filepath.Base(request.URL.Path)
		methods = append(methods, method)
		if method == "sendMediaGroup" {
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	notifier := New(Config{
		TelegramBotToken:  "token",
		TelegramChatID:    "99",
		TelegramAPIBase:   server.URL,
		HTTPClient:        server.Client(),
		AllowInsecureHTTP: true,
	})
	if err := notifier.sendTelegram(context.Background(), images, "caption"); err == nil {
		t.Fatal("expected first batch failure")
	}
	if got, want := strings.Join(methods, ","), "sendMediaGroup,sendPhoto"; got != want {
		t.Fatalf("got methods %q, want later batch to continue as %q", got, want)
	}
}

func TestNotifySkipsSymlinkedSiteImages(t *testing.T) {
	dir := t.TempDir()
	outputDir := filepath.Join(dir, "output")
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "rare_resources.txt"), []byte("stats"), 0o600); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "outside.png")
	if err := os.WriteFile(target, []byte("not-for-upload"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(outputDir, "site_5.png")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	var method string
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		method = filepath.Base(request.URL.Path)
		response.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	notifier := New(Config{
		TelegramBotToken:  "token",
		TelegramChatID:    "99",
		TelegramAPIBase:   server.URL,
		HTTPClient:        server.Client(),
		AllowInsecureHTTP: true,
	})
	if err := notifier.Notify(context.Background(), outputDir, "task1", "42", ""); err != nil {
		t.Fatal(err)
	}
	if method != "sendMessage" {
		t.Fatalf("symlink caused image upload via %q", method)
	}
}

func TestTelegramRejectsAPIErrorPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"ok":false,"description":"hidden"}`))
	}))
	defer server.Close()
	notifier := New(Config{
		TelegramBotToken:  "token",
		TelegramChatID:    "99",
		TelegramAPIBase:   server.URL,
		HTTPClient:        server.Client(),
		AllowInsecureHTTP: true,
	})
	if err := notifier.sendTelegram(context.Background(), nil, "caption"); err == nil || err.Error() != "Telegram API rejected request" {
		t.Fatalf("unexpected Telegram result %v", err)
	}
}

func TestNotifierRequiresHTTPSByDefault(t *testing.T) {
	notifier := New(Config{
		TelegramBotToken: "token",
		TelegramChatID:   "99",
		TelegramAPIBase:  "http://localhost:9999",
	})
	if err := notifier.sendTelegram(context.Background(), nil, "caption"); err == nil || err.Error() != "notification endpoint must use HTTPS" {
		t.Fatalf("got insecure endpoint error %v", err)
	}
}

func TestTelegramTransportErrorDoesNotExposeToken(t *testing.T) {
	notifier := New(Config{
		TelegramBotToken: "very-secret-token",
		TelegramChatID:   "99",
		TelegramAPIBase:  "https://telegram.invalid",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("network failed at https://telegram.invalid/botvery-secret-token/sendMessage")
		})},
	})
	err := notifier.sendTelegram(context.Background(), nil, "caption")
	if err == nil || strings.Contains(err.Error(), "very-secret-token") {
		t.Fatalf("unexpected unsafe error %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func testImages(t *testing.T, directory string, count int) []string {
	t.Helper()
	paths := make([]string, 0, count)
	for index := 0; index < count; index++ {
		path := filepath.Join(directory, "image_"+strings.Repeat("x", index+1)+".png")
		if err := os.WriteFile(path, []byte("image"), 0o600); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	return paths
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

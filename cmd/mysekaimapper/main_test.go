package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mouse233/MySekaiMapper/internal/service"
)

func TestNewNotifierUsesAstrBotSettings(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	if err := os.Mkdir(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string]string{
		"push_map.json":    `{"42":["astrbot:me"]}`,
		"astrbot_map.json": `{"me":{"platform_id":"qq_main","type":"private","qq":"123456789"}}`,
	} {
		if err := os.WriteFile(filepath.Join(configDir, name), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var gotUMO string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/send" || r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("incorrect gateway configuration")
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		gotUMO = payload["umo"]
		_, _ = w.Write([]byte(`{"status":"queued","message_id":"test"}`))
	}))
	defer server.Close()
	t.Setenv("MYSK_CONFIG_DIR", configDir)
	t.Setenv("ASTRBOT_PUSH_URL", server.URL)
	t.Setenv("ASTRBOT_PUSH_TOKEN", "test-token")
	t.Setenv("ASTRBOT_ALLOW_INSECURE_HTTP", "1")
	notifier := newNotifier(service.SettingsFromRoot(root))
	if err := notifier.Notify(context.Background(), root, "task", "42", ""); err != nil {
		t.Fatal(err)
	}
	if gotUMO != "qq_main:FriendMessage:123456789" {
		t.Fatalf("got umo %q", gotUMO)
	}
}

func TestResolveRootOverrideKeepsCommandFlags(t *testing.T) {
	repositoryRoot := filepath.Join("..", "..")
	root, args, err := resolveRoot(t.TempDir(), []string{
		"--input", "save.bin", "--root", repositoryRoot, "--env", "local.env",
	})
	if err != nil {
		t.Fatal(err)
	}
	wantRoot, err := filepath.Abs(repositoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	if root != wantRoot {
		t.Fatalf("got root %q, want %q", root, wantRoot)
	}
	wantArgs := []string{"--input", "save.bin", "--env", "local.env"}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("got args %q, want %q", args, wantArgs)
	}
}

func TestResolveRootRejectsMissingValue(t *testing.T) {
	if _, _, err := resolveRoot(t.TempDir(), []string{"--root"}); err == nil {
		t.Fatal("expected --root without a value to fail")
	}
}

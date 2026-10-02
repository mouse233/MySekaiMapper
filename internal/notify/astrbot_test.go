package notify

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAstrBotTargetSession(t *testing.T) {
	for _, tc := range []struct {
		target AstrBotTarget
		want   string
	}{
		{AstrBotTarget{PlatformID: "qq_main", Type: "private", QQ: "123456789"}, "qq_main:FriendMessage:123456789"},
		{AstrBotTarget{PlatformID: "qq_other", Type: "group", QQ: "987654321"}, "qq_other:GroupMessage:987654321"},
		{AstrBotTarget{UMO: "qq_main:GroupMessage:123_456"}, "qq_main:GroupMessage:123_456"},
	} {
		got, err := tc.target.session()
		if err != nil || got != tc.want {
			t.Fatalf("session = %q, %v; want %q", got, err, tc.want)
		}
	}
	for _, target := range []AstrBotTarget{
		{}, {PlatformID: "wrong:id", Type: "private", QQ: "123"},
		{PlatformID: "qq", Type: "other", QQ: "123"},
		{PlatformID: "qq", Type: "private", QQ: "0"},
		{PlatformID: "qq", Type: "private", QQ: "12x"},
		{UMO: "123456789"}, {UMO: "qq:OtherMessage:123"},
		{UMO: "qq:FriendMessage:"}, {UMO: "qq:FriendMessage:123", QQ: "123"},
	} {
		if _, err := target.session(); err == nil {
			t.Fatalf("accepted invalid target %#v", target)
		}
	}
}

func TestAstrBotStringAndLegacyCombinedRoutes(t *testing.T) {
	for _, route := range []any{"astrbot:me", "astrbot:me+tg", []any{"astrbot:me", "device", "telegram"}} {
		bark, astrbot, telegram := resolveRoutes(route)
		if len(astrbot) != 1 || astrbot[0] != "me" {
			t.Fatalf("AstrBot route lost: %v", route)
		}
		if route == "astrbot:me" && (telegram || len(bark) != 0) {
			t.Fatal("AstrBot string selected another channel")
		}
		if route == "astrbot:me+tg" && (!telegram || len(bark) != 0) {
			t.Fatal("legacy +tg combination lost")
		}
	}
}

func astrBotFixture(t *testing.T, routes []string) (string, Config) {
	t.Helper()
	dir := t.TempDir()
	writeJSON(t, filepath.Join(dir, "push_map.json"), map[string]any{"42": routes})
	writeJSON(t, filepath.Join(dir, "astrbot_map.json"), map[string]any{
		"me":      AstrBotTarget{PlatformID: "qq_main", Type: "private", QQ: "123456789"},
		"group":   AstrBotTarget{PlatformID: "qq_main", Type: "group", QQ: "987654321"},
		"invalid": map[string]any{"platform_id": "qq_main", "type": "private", "qq": 123},
	})
	if err := os.WriteFile(filepath.Join(dir, "rare_resources.txt"), []byte("钻石 × 1"), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir, Config{PushMapFile: filepath.Join(dir, "push_map.json"), AstrBotMapFile: filepath.Join(dir, "astrbot_map.json"), AstrBotPushToken: "secret-token", AstrBotAllowInsecureHTTP: true}
}

func queued(response http.ResponseWriter) {
	_, _ = response.Write([]byte(`{"status":"queued","message_id":"id","queue_size":1}`))
}

func TestNotifyAstrBotTextImagesMultipleTargetsAndLegacyRoutes(t *testing.T) {
	dir, config := astrBotFixture(t, []string{"astrbot:me", "astrbot:group", "astrbot:me", "telegram", "device"})
	image, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/l9sAAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"site_5.png", "site_6.png"} {
		if err := os.WriteFile(filepath.Join(dir, name), image, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(dir, "site_5.png"), filepath.Join(dir, "site_7.png")); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(dir, "bark_map.json"), map[string]string{"device": "key"})
	var messages []map[string]string
	var barkCalls, telegramCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/prefix/send":
			if r.Method != "POST" || r.Header.Get("Authorization") != "Bearer secret-token" || r.Header.Get("Content-Type") != "application/json" {
				t.Error("wrong AstrBot request headers")
			}
			var message map[string]string
			if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
				t.Error(err)
			}
			messages = append(messages, message)
			queued(w)
		case strings.HasPrefix(r.URL.Path, "/bark/"):
			barkCalls++
			w.WriteHeader(200)
		case strings.HasPrefix(r.URL.Path, "/telegram/"):
			telegramCalls++
			w.WriteHeader(200)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	config.AstrBotPushURL = server.URL + "/prefix/"
	config.AllowInsecureHTTP = true
	config.BarkAPIBase = server.URL + "/bark"
	config.BarkMapFile = filepath.Join(dir, "bark_map.json")
	config.TelegramAPIBase = server.URL + "/telegram"
	config.TelegramBotToken, config.TelegramChatID = "token", "99"
	var logs []string
	config.Logf = func(format string, args ...any) { logs = append(logs, fmt.Sprintf(format, args...)) }
	if err := New(config).Notify(context.Background(), dir, "task1", "42", ""); err != nil {
		t.Fatal(err)
	}
	if len(messages) != 6 || barkCalls != 3 || telegramCalls != 1 {
		t.Fatalf("AstrBot=%d Bark=%d Telegram=%d", len(messages), barkCalls, telegramCalls)
	}
	for i, message := range messages {
		wantUMO := "qq_main:FriendMessage:123456789"
		if i >= 3 {
			wantUMO = "qq_main:GroupMessage:987654321"
		}
		if message["umo"] != wantUMO {
			t.Errorf("unexpected umo %q", message["umo"])
		}
		if i%3 == 0 {
			if message["message_type"] != "text" || !strings.Contains(message["content"], "Player: 42") || !strings.Contains(message["content"], "钻石 × 1") {
				t.Errorf("unexpected summary %v", message)
			}
		} else if message["message_type"] != "image" || message["content"] != base64.StdEncoding.EncodeToString(image) {
			t.Error("wrong image payload")
		}
	}
	if len(logs) != 6 || !strings.Contains(logs[0], "queued") || strings.Contains(strings.Join(logs, " "), "secret-token") {
		t.Fatalf("unexpected logs %v", logs)
	}
}

func TestAstrBotFailureDoesNotSuppressOtherTargetsOrTelegram(t *testing.T) {
	dir, config := astrBotFixture(t, []string{"astrbot:missing", "astrbot:invalid", "astrbot:me", "astrbot:group", "telegram"})
	var astrBotCalls, telegramCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/send" {
			astrBotCalls++
			if astrBotCalls == 1 {
				w.WriteHeader(403)
				return
			}
			queued(w)
		} else {
			telegramCalls++
			w.WriteHeader(200)
		}
	}))
	defer server.Close()
	config.AstrBotPushURL = server.URL
	config.TelegramAPIBase = server.URL
	config.TelegramBotToken, config.TelegramChatID = "token", "99"
	config.AllowInsecureHTTP = true
	err := New(config).Notify(context.Background(), dir, "task", "42", "")
	if err == nil || !strings.Contains(err.Error(), "missing") || !strings.Contains(err.Error(), "invalid") || !strings.Contains(err.Error(), "403") || astrBotCalls != 2 || telegramCalls != 1 {
		t.Fatalf("err=%v AstrBot=%d Telegram=%d", err, astrBotCalls, telegramCalls)
	}
}

func TestAstrBotRejectsUnacknowledgedResponsesAndRedirects(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
	}{
		{200, ""}, {200, "not JSON"}, {200, `{"status":"sent","message_id":"id"}`},
		{200, `{"status":"queued"}`}, {200, strings.Repeat("x", 64*1024+1)}, {500, "secret-token"}, {307, ""},
	} {
		t.Run(fmt.Sprintf("%d-%d", tc.status, len(tc.body)), func(t *testing.T) {
			var redirectCalls int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/redirect" {
					redirectCalls++
					queued(w)
					return
				}
				w.Header().Set("Location", "/redirect")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			err := New(Config{AstrBotPushToken: "secret-token"}).postAstrBot(context.Background(), server.URL+"/send", "qq:FriendMessage:123", "text", "hi")
			if err == nil || strings.Contains(err.Error(), "secret-token") || redirectCalls != 0 {
				t.Fatalf("err=%v redirected=%d", err, redirectCalls)
			}
		})
	}
}

func TestAstrBotInvalidSettingsAreIsolatedAndDoNotLeakSecrets(t *testing.T) {
	for _, change := range []func(*Config){
		func(c *Config) { c.AstrBotPushURL = "http://localhost:9966"; c.AstrBotAllowInsecureHTTP = false },
		func(c *Config) { c.AstrBotPushURL = "https://user:secret-token@example.com" },
		func(c *Config) { c.AstrBotPushURL = "https://example.com?token=secret-token" },
		func(c *Config) { c.AstrBotPushURL = "ftp://example.com" },
		func(c *Config) { c.AstrBotPushToken = "" },
		func(c *Config) { c.AstrBotMapFile = "" },
	} {
		dir, config := astrBotFixture(t, []string{"astrbot:me", "telegram"})
		var telegramCalls int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { telegramCalls++; w.WriteHeader(200) }))
		config.AstrBotPushURL = "https://example.com"
		config.TelegramAPIBase = server.URL
		config.TelegramBotToken, config.TelegramChatID = "token", "99"
		config.AllowInsecureHTTP = true
		change(&config)
		err := New(config).Notify(context.Background(), dir, "task", "42", "")
		server.Close()
		if err == nil || strings.Contains(err.Error(), "secret-token") || telegramCalls != 1 {
			t.Fatalf("err=%v Telegram=%d", err, telegramCalls)
		}
	}
	config := Config{AstrBotPushURL: "https://example.com", AstrBotPushToken: "secret-token", HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("secret-token https://example.com/private")
	})}}
	err := New(config).postAstrBot(context.Background(), "https://example.com/send", "qq:FriendMessage:123", "text", "hi")
	if err == nil || strings.Contains(err.Error(), "secret-token") || strings.Contains(err.Error(), "example.com") {
		t.Fatalf("leaked transport error %v", err)
	}
}

func TestAstrBotOversizedImageDoesNotSuppressLaterImages(t *testing.T) {
	dir, config := astrBotFixture(t, []string{"astrbot:me"})
	file, err := os.Create(filepath.Join(dir, "site_5.png"))
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(maxAstrBotImageSize + 1); err != nil {
		t.Fatal(err)
	}
	file.Close()
	if err := os.WriteFile(filepath.Join(dir, "site_6.png"), []byte("image"), 0o600); err != nil {
		t.Fatal(err)
	}
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; queued(w) }))
	defer server.Close()
	config.AstrBotPushURL = server.URL
	err = New(config).Notify(context.Background(), dir, "task", "42", "")
	if err == nil || !strings.Contains(err.Error(), "8 MiB") || calls != 2 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
}

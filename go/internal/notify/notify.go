// Package notify implements local-output notifications for Telegram and Bark.
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	maxConfigSize             int64 = 1 * 1024 * 1024
	maxTelegramMultipartSize  int64 = 32 * 1024 * 1024
	barkRequestTimeout              = 10 * time.Second
	telegramTextTimeout             = 10 * time.Second
	telegramPhotoTimeout            = 30 * time.Second
	telegramMediaGroupTimeout       = 60 * time.Second
)

var ErrTelegramPayloadTooLarge = errors.New("Telegram multipart payload exceeds configured limit")

// Config contains notification-only settings. It deliberately has no archive
// payload fields, so this package never needs to read an encrypted save.
type Config struct {
	BarkMapFile       string
	PushMapFile       string
	BarkIcon          string
	BarkImageBase     string
	FallbackImageBase string
	TelegramBotToken  string
	TelegramChatID    string

	BarkAPIBase       string
	TelegramAPIBase   string
	HTTPClient        *http.Client
	AllowInsecureHTTP bool // Only for local test servers or an explicitly trusted proxy.
}

// Notifier routes completed, local map outputs to the configured channels.
type Notifier struct {
	config    Config
	client    *http.Client
	configErr error
}

func New(config Config) *Notifier {
	if config.BarkAPIBase == "" {
		config.BarkAPIBase = "https://api.day.app"
	}
	if config.TelegramAPIBase == "" {
		config.TelegramAPIBase = "https://api.telegram.org"
	}
	client := config.HTTPClient
	if client == nil {
		// Individual Bark/Telegram request methods apply their protocol-specific
		// deadlines; a client-wide deadline would incorrectly cap media groups.
		client = &http.Client{}
	}
	notifier := &Notifier{config: config, client: client}
	if !config.AllowInsecureHTTP && (!isHTTPSURL(config.BarkAPIBase) || !isHTTPSURL(config.TelegramAPIBase)) {
		notifier.configErr = errors.New("notification endpoint must use HTTPS")
	}
	return notifier
}

func isHTTPSURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "https" && parsed.Host != ""
}

// ResolveMethod preserves legacy string and modern JSON-list routing syntax.
func ResolveMethod(value any) (barkAliases []string, sendTelegram bool) {
	if value == nil {
		return nil, false
	}
	selections := make([]string, 0)
	switch typed := value.(type) {
	case []string:
		selections = append(selections, typed...)
	case []any:
		for _, item := range typed {
			if text, ok := item.(string); ok {
				selections = append(selections, text)
			}
		}
	case string:
		if typed == "" || typed == "none" {
			return nil, false
		}
		if typed == "telegram" {
			selections = append(selections, "telegram")
		} else if strings.Contains(typed, "+tg") {
			alias := strings.SplitN(typed, "+tg", 2)[0]
			selections = append(selections, alias, "telegram")
		} else {
			selections = append(selections, typed)
		}
	default:
		return nil, false
	}
	for _, selection := range selections {
		if selection == "telegram" {
			sendTelegram = true
		} else if selection != "" {
			barkAliases = append(barkAliases, selection)
		}
	}
	return barkAliases, sendTelegram
}

// hasPushMethod mirrors Python's `if not method: method = 'telegram'` rule
// for the JSON values accepted by push_map.json.
func hasPushMethod(value any) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case string:
		return typed != ""
	case []string:
		return len(typed) > 0
	case []any:
		return len(typed) > 0
	case bool:
		return typed
	case float64:
		return typed != 0
	case map[string]any:
		return len(typed) > 0
	default:
		return true
	}
}

func (n *Notifier) barkMap() map[string]string {
	mapping := map[string]string{}
	if n.config.BarkMapFile == "" {
		return mapping
	}
	data, err := readSmallFile(n.config.BarkMapFile, maxConfigSize)
	if err != nil {
		return mapping
	}
	if err := json.Unmarshal(data, &mapping); err != nil {
		return map[string]string{}
	}
	return mapping
}

func (n *Notifier) pushMap() map[string]any {
	mapping := map[string]any{}
	if n.config.PushMapFile == "" {
		return mapping
	}
	data, err := readSmallFile(n.config.PushMapFile, maxConfigSize)
	if err != nil {
		return mapping
	}
	if err := json.Unmarshal(data, &mapping); err != nil {
		return map[string]any{}
	}
	return mapping
}

// BarkKey returns an explicit key before falling back to the configured alias.
func (n *Notifier) BarkKey(alias, explicitKey string) string {
	if explicitKey != "" {
		return explicitKey
	}
	if alias == "" {
		return ""
	}
	return n.barkMap()[alias]
}

// Notify sends the rare summary first, then Bark image links and/or Telegram
// images according to the player's configured route. A failed channel does not
// prevent attempts to the other selected channel.
func (n *Notifier) Notify(ctx context.Context, outputDir, taskID, playerID, imageBase string) error {
	if n.configErr != nil {
		return n.configErr
	}
	rareText := readRareText(outputDir)
	method, exists := n.pushMap()[playerID]
	if !exists || !hasPushMethod(method) {
		method = "telegram"
	}
	barkAliases, sendTelegram := ResolveMethod(method)

	errorsSeen := make([]error, 0)
	rareCompact := compactRareText(rareText)
	for _, alias := range barkAliases {
		key := n.BarkKey(alias, "")
		if key == "" {
			errorsSeen = append(errorsSeen, fmt.Errorf("Bark key not configured for alias %q", alias))
			continue
		}
		if err := n.sendBark(ctx, key, rareCompact, "", "", n.config.BarkIcon); err != nil {
			errorsSeen = append(errorsSeen, err)
		}
	}

	base := strings.TrimRight(firstNonEmpty(imageBase, n.config.BarkImageBase, n.config.FallbackImageBase), "/")
	for _, alias := range barkAliases {
		key := n.BarkKey(alias, "")
		if key == "" {
			continue
		}
		for siteID := 5; siteID <= 8; siteID++ {
			imageURL := ""
			if base != "" {
				imageURL = fmt.Sprintf("%s/site_%d.png", base, siteID)
			}
			if err := n.sendBark(ctx, key, fmt.Sprintf("site%d", siteID), "", imageURL, n.config.BarkIcon); err != nil {
				errorsSeen = append(errorsSeen, err)
			}
		}
	}

	if sendTelegram {
		caption := notificationHeader(taskID, playerID) + rareText
		if err := n.sendTelegram(ctx, siteImagePaths(outputDir), caption); err != nil {
			errorsSeen = append(errorsSeen, err)
		}
	}
	return errors.Join(errorsSeen...)
}

func (n *Notifier) sendBark(ctx context.Context, key, title, body, imageURL, iconURL string) error {
	if n.configErr != nil {
		return n.configErr
	}
	endpoint := strings.TrimRight(n.config.BarkAPIBase, "/") + "/" + url.PathEscape(key) + "/" + url.QueryEscape(title)
	query := url.Values{}
	if body != "" {
		query.Set("body", body)
	}
	if imageURL != "" {
		query.Set("image", imageURL)
	}
	if iconURL != "" {
		query.Set("icon", iconURL)
	}
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	timedContext, cancel := context.WithTimeout(ctx, barkRequestTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(timedContext, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("Bark request setup failed")
	}
	response, err := n.client.Do(request)
	if err != nil {
		return fmt.Errorf("Bark request failed")
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64*1024))
	if response.StatusCode/100 != 2 {
		return fmt.Errorf("Bark returned status %d", response.StatusCode)
	}
	return nil
}

func (n *Notifier) sendTelegram(ctx context.Context, imagePaths []string, caption string) error {
	if n.configErr != nil {
		return n.configErr
	}
	if n.config.TelegramBotToken == "" || n.config.TelegramChatID == "" {
		return nil
	}
	if len(imagePaths) == 0 {
		return n.postTelegramText(ctx, caption)
	}
	errorsSeen := make([]error, 0)
	for start := 0; start < len(imagePaths); start += 10 {
		end := min(start+10, len(imagePaths))
		batchCaption := ""
		if start == 0 {
			batchCaption = truncateRunes(caption, 1000)
		}
		if len(imagePaths[start:end]) == 1 {
			if err := n.postTelegramPhoto(ctx, imagePaths[start], batchCaption); err != nil {
				errorsSeen = append(errorsSeen, err)
			}
		} else if err := n.postTelegramMediaGroup(ctx, imagePaths[start:end], batchCaption); err != nil {
			errorsSeen = append(errorsSeen, err)
		}
	}
	return errors.Join(errorsSeen...)
}

func (n *Notifier) telegramEndpoint(method string) string {
	return strings.TrimRight(n.config.TelegramAPIBase, "/") + "/bot" + n.config.TelegramBotToken + "/" + method
}

func (n *Notifier) postTelegramText(ctx context.Context, text string) error {
	form := url.Values{"chat_id": {n.config.TelegramChatID}, "text": {text}}
	timedContext, cancel := context.WithTimeout(ctx, telegramTextTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(timedContext, http.MethodPost, n.telegramEndpoint("sendMessage"), strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("Telegram request setup failed")
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return n.doTelegram(request)
}

func (n *Notifier) postTelegramPhoto(ctx context.Context, imagePath, caption string) error {
	return n.postMultipart(ctx, "sendPhoto", telegramPhotoTimeout, func(writer *multipart.Writer) error {
		if err := writer.WriteField("chat_id", n.config.TelegramChatID); err != nil {
			return err
		}
		if caption != "" {
			if err := writer.WriteField("caption", caption); err != nil {
				return err
			}
		}
		return writeFilePart(writer, "photo", imagePath)
	})
}

func (n *Notifier) postTelegramMediaGroup(ctx context.Context, imagePaths []string, caption string) error {
	media := make([]map[string]string, 0, len(imagePaths))
	for index := range imagePaths {
		item := map[string]string{"type": "photo", "media": fmt.Sprintf("attach://photo%d", index)}
		if index == 0 && caption != "" {
			item["caption"] = caption
		}
		media = append(media, item)
	}
	encodedMedia, err := json.Marshal(media)
	if err != nil {
		return err
	}
	return n.postMultipart(ctx, "sendMediaGroup", telegramMediaGroupTimeout, func(writer *multipart.Writer) error {
		if err := writer.WriteField("chat_id", n.config.TelegramChatID); err != nil {
			return err
		}
		if err := writer.WriteField("media", string(encodedMedia)); err != nil {
			return err
		}
		for index, imagePath := range imagePaths {
			if err := writeFilePart(writer, fmt.Sprintf("photo%d", index), imagePath); err != nil {
				return err
			}
		}
		return nil
	})
}

func (n *Notifier) postMultipart(ctx context.Context, method string, timeout time.Duration, populate func(*multipart.Writer) error) error {
	body := &limitedBuffer{limit: maxTelegramMultipartSize}
	writer := multipart.NewWriter(body)
	if err := populate(writer); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	timedContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(timedContext, http.MethodPost, n.telegramEndpoint(method), bytes.NewReader(body.Bytes()))
	if err != nil {
		return fmt.Errorf("Telegram request setup failed")
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return n.doTelegram(request)
}

func (n *Notifier) doTelegram(request *http.Request) error {
	response, err := n.client.Do(request)
	if err != nil {
		return fmt.Errorf("Telegram request failed")
	}
	defer response.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(response.Body, 64*1024+1))
	if readErr != nil {
		return fmt.Errorf("Telegram response read failed")
	}
	if response.StatusCode/100 != 2 {
		return fmt.Errorf("Telegram returned status %d", response.StatusCode)
	}
	if len(body) <= 64*1024 && len(body) > 0 {
		var result struct {
			OK *bool `json:"ok"`
		}
		if json.Unmarshal(body, &result) == nil && result.OK != nil && !*result.OK {
			return fmt.Errorf("Telegram API rejected request")
		}
	}
	return nil
}

type limitedBuffer struct {
	bytes.Buffer
	limit int64
}

func (b *limitedBuffer) Write(data []byte) (int, error) {
	if int64(b.Len()+len(data)) > b.limit {
		return 0, ErrTelegramPayloadTooLarge
	}
	return b.Buffer.Write(data)
}

func writeFilePart(writer *multipart.Writer, fieldName, path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("notification attachment is not a regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	part, err := writer.CreateFormFile(fieldName, filepath.Base(path))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	return err
}

// siteImagePaths accepts only the renderer's four expected regular files. It
// deliberately uses Lstat so an attacker cannot redirect a notification upload
// through a symlink placed in the output directory.
func siteImagePaths(outputDir string) []string {
	paths := make([]string, 0, 4)
	for siteID := 5; siteID <= 8; siteID++ {
		path := filepath.Join(outputDir, fmt.Sprintf("site_%d.png", siteID))
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		paths = append(paths, path)
	}
	return paths
}

func readRareText(outputDir string) string {
	data, err := readSmallFile(filepath.Join(outputDir, "rare_resources.txt"), maxConfigSize)
	if err != nil {
		return "Mysekai 抓取完成,但未生成稀有资源统计。"
	}
	return string(data)
}

func readSmallFile(path string, maxSize int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxSize+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxSize {
		return nil, fmt.Errorf("file exceeds %d-byte limit", maxSize)
	}
	return data, nil
}

func notificationHeader(taskID, playerID string) string {
	header := "🎮 Mysekai 抓取完成\nTask: " + taskID + "\n"
	if playerID != "" {
		header += "Player: " + playerID + "\n"
	}
	return header + "\n"
}

func compactRareText(text string) string {
	parts := make([]string, 0, 8)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			parts = append(parts, line)
		}
		if len(parts) == 8 {
			break
		}
	}
	if len(parts) == 0 {
		return "稀有资源统计: 无"
	}
	return strings.Join(parts, " | ")
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 || utf8.RuneCountInString(value) <= limit {
		return value
	}
	return string([]rune(value)[:limit])
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}

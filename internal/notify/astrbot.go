package notify

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	maxAstrBotImageSize int64 = 8 * 1024 * 1024
	astrBotTimeout            = 30 * time.Second
)

// AstrBotTarget accepts a full UMO or an explicit OneBot v11 QQ destination.
// QQ identifiers are strings to avoid losing precision in JSON configuration.
type AstrBotTarget struct {
	PlatformID string `json:"platform_id"`
	Type       string `json:"type"`
	QQ         string `json:"qq"`
	UMO        string `json:"umo"`
}

func (target AstrBotTarget) session() (string, error) {
	if target.UMO != "" {
		if target.PlatformID != "" || target.Type != "" || target.QQ != "" {
			return "", errors.New("use either umo or platform_id/type/qq")
		}
		parts := strings.SplitN(target.UMO, ":", 3)
		if len(parts) != 3 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[2]) == "" || (parts[1] != "FriendMessage" && parts[1] != "GroupMessage") {
			return "", errors.New("invalid umo; expected platform_id:FriendMessage/GroupMessage:session_id")
		}
		return target.UMO, nil
	}
	if strings.TrimSpace(target.PlatformID) == "" || strings.ContainsAny(target.PlatformID, ":\r\n") || target.PlatformID != strings.TrimSpace(target.PlatformID) {
		return "", errors.New("platform_id must be a nonempty platform instance ID without colons or surrounding whitespace")
	}
	if target.QQ == "" || strings.Trim(target.QQ, "0") == "" {
		return "", errors.New("qq must be a positive decimal string")
	}
	for _, digit := range target.QQ {
		if digit < '0' || digit > '9' {
			return "", errors.New("qq must be a positive decimal string")
		}
	}
	messageType := ""
	switch target.Type {
	case "private":
		messageType = "FriendMessage"
	case "group":
		messageType = "GroupMessage"
	default:
		return "", errors.New("type must be private or group")
	}
	return target.PlatformID + ":" + messageType + ":" + target.QQ, nil
}

func (n *Notifier) astrBotEndpoint() (string, error) {
	if strings.TrimSpace(n.config.AstrBotPushToken) == "" {
		return "", errors.New("ASTRBOT_PUSH_TOKEN is not configured")
	}
	endpoint, err := url.Parse(n.config.AstrBotPushURL)
	if err != nil || endpoint.Hostname() == "" || endpoint.Opaque != "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return "", errors.New("ASTRBOT_PUSH_URL must be an HTTP(S) base URL without credentials, query, or fragment")
	}
	if endpoint.Scheme != "https" && !(endpoint.Scheme == "http" && n.config.AstrBotAllowInsecureHTTP) {
		return "", errors.New("AstrBot requires HTTPS; set ASTRBOT_ALLOW_INSECURE_HTTP=1 only for a trusted local HTTP service")
	}
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/send"
	endpoint.RawPath = ""
	return endpoint.String(), nil
}

func (n *Notifier) sendAstrBot(ctx context.Context, aliases, imagePaths []string, text string) error {
	endpoint, err := n.astrBotEndpoint()
	if err != nil {
		return err
	}
	data, err := readSmallFile(n.config.AstrBotMapFile, maxConfigSize)
	if err != nil {
		return errors.New("AstrBot recipient configuration could not be read")
	}
	// Decode targets independently so a malformed target cannot suppress others.
	var targets map[string]json.RawMessage
	if json.Unmarshal(data, &targets) != nil {
		return errors.New("AstrBot recipient configuration is not a valid JSON object")
	}
	var errorsSeen []error
	for _, alias := range aliases {
		raw, exists := targets[alias]
		if !exists {
			errorsSeen = append(errorsSeen, fmt.Errorf("AstrBot target not configured for alias %q", alias))
			continue
		}
		var target AstrBotTarget
		if json.Unmarshal(raw, &target) != nil {
			errorsSeen = append(errorsSeen, fmt.Errorf("invalid AstrBot target for alias %q", alias))
			continue
		}
		umo, err := target.session()
		if err != nil {
			errorsSeen = append(errorsSeen, fmt.Errorf("AstrBot alias %q: %w", alias, err))
			continue
		}
		if err := n.postAstrBot(ctx, endpoint, umo, "text", text); err != nil {
			errorsSeen = append(errorsSeen, fmt.Errorf("AstrBot alias %q summary: %w", alias, err))
		}
		for _, path := range imagePaths {
			info, statErr := os.Lstat(path)
			if statErr != nil || !info.Mode().IsRegular() {
				errorsSeen = append(errorsSeen, errors.New("AstrBot image is not a regular file"))
				continue
			}
			image, readErr := readSmallFile(path, maxAstrBotImageSize)
			if readErr != nil {
				errorsSeen = append(errorsSeen, fmt.Errorf("AstrBot image %s could not be read within the 8 MiB limit", filepath.Base(path)))
				continue
			}
			if err := n.postAstrBot(ctx, endpoint, umo, "image", base64.StdEncoding.EncodeToString(image)); err != nil {
				errorsSeen = append(errorsSeen, fmt.Errorf("AstrBot alias %q image %s: %w", alias, filepath.Base(path), err))
			}
		}
	}
	return errors.Join(errorsSeen...)
}

func (n *Notifier) postAstrBot(ctx context.Context, endpoint, umo, messageType, content string) error {
	body, err := json.Marshal(struct {
		UMO         string `json:"umo"`
		MessageType string `json:"message_type"`
		Content     string `json:"content"`
	}{umo, messageType, content})
	if err != nil {
		return errors.New("AstrBot request encoding failed")
	}
	timedContext, cancel := context.WithTimeout(ctx, astrBotTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(timedContext, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return errors.New("AstrBot request setup failed")
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+n.config.AstrBotPushToken)
	// Never forward the token or message to a redirected endpoint.
	client := *n.client
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, err := client.Do(request)
	if err != nil {
		return errors.New("AstrBot request failed")
	}
	defer response.Body.Close()
	if response.StatusCode/100 != 2 {
		return fmt.Errorf("AstrBot returned status %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 64*1024+1))
	if err != nil || len(data) > 64*1024 {
		return errors.New("AstrBot response read failed")
	}
	var result struct {
		Status    string `json:"status"`
		MessageID string `json:"message_id"`
	}
	if json.Unmarshal(data, &result) != nil || result.Status != "queued" || result.MessageID == "" {
		return errors.New("AstrBot did not acknowledge queue acceptance")
	}
	if n.config.Logf != nil {
		n.config.Logf("[ASTRBOT] %s request queued", messageType)
	}
	return nil
}

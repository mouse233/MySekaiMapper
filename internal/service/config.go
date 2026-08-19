package service

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Settings centralizes Go service paths and environment-backed endpoint
// settings. Call mapper.LoadDotEnv before constructing it when .env support is
// desired.
type Settings struct {
	Root       string
	AssetsDir  string
	ConfigDir  string
	DataDir    string
	RawDir     string
	TmpDir     string
	ArchiveDir string
	LatestDir  string

	ResourceCSV string
	FontFile    string

	ReportEnabled bool
	ReportPath    string
	ReportMaxSize int64
	ReportToken   string

	TelegramBotToken  string
	TelegramChatID    string
	BarkIcon          string
	BarkImageBase     string
	FallbackImageBase string
	BarkMapFile       string
	PushMapFile       string
}

func SettingsFromRoot(root string) Settings {
	assetsDir := envPath("MYSK_ASSETS_DIR", filepath.Join(root, "assets"))
	configDir := envPath("MYSK_CONFIG_DIR", filepath.Join(root, "config"))
	dataDir := envPath("MYSK_DATA_DIR", filepath.Join(root, "data"))
	reportPath := strings.TrimSpace(os.Getenv("REPORT_PATH"))
	if reportPath == "" {
		reportPath = "/reqable/report"
	}
	if !strings.HasPrefix(reportPath, "/") {
		reportPath = "/" + reportPath
	}
	return Settings{
		Root:          root,
		AssetsDir:     assetsDir,
		ConfigDir:     configDir,
		DataDir:       dataDir,
		RawDir:        filepath.Join(dataDir, "raw_mysekai"),
		TmpDir:        filepath.Join(dataDir, "tmp"),
		ArchiveDir:    filepath.Join(dataDir, "archive"),
		LatestDir:     filepath.Join(dataDir, "latest"),
		ResourceCSV:   filepath.Join(assetsDir, "resourceId.csv"),
		FontFile:      filepath.Join(assetsDir, "NotoSansSC-Regular.ttf"),
		ReportEnabled: envBool("REPORT_ENABLED", true),
		ReportPath:    reportPath,
		ReportMaxSize: envMegabytes("REPORT_MAX_SIZE", 1),
		ReportToken:   os.Getenv("REPORT_TOKEN"),

		TelegramBotToken:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:    os.Getenv("TELEGRAM_CHAT_ID"),
		BarkIcon:          os.Getenv("BARK_ICON"),
		BarkImageBase:     os.Getenv("BARK_IMAGE_BASE"),
		FallbackImageBase: os.Getenv("FALLBACK_IMAGE_BASE"),
		BarkMapFile:       filepath.Join(configDir, "bark_map.json"),
		PushMapFile:       filepath.Join(configDir, "push_map.json"),
	}
}

func envPath(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func envBool(name string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	if value == "" {
		return fallback
	}
	switch value {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func envMegabytes(name string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback * 1024 * 1024
	}
	megabytes, err := strconv.ParseInt(value, 10, 64)
	const bytesPerMiB int64 = 1024 * 1024
	if err != nil || megabytes <= 0 || megabytes > ((1<<63-1)/bytesPerMiB) {
		return fallback * bytesPerMiB
	}
	return megabytes * bytesPerMiB
}

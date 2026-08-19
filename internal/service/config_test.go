package service

import (
	"path/filepath"
	"testing"
)

func TestSettingsFromRootUsesCompatibleDefaults(t *testing.T) {
	t.Setenv("MYSK_ASSETS_DIR", "")
	t.Setenv("MYSK_CONFIG_DIR", "")
	t.Setenv("MYSK_DATA_DIR", "")
	t.Setenv("REPORT_PATH", "report/custom")
	t.Setenv("REPORT_ENABLED", "0")
	t.Setenv("REPORT_MAX_SIZE", "2")
	settings := SettingsFromRoot("/repo")
	if settings.ReportPath != "/report/custom" || settings.ReportEnabled || settings.ReportMaxSize != 2*1024*1024 {
		t.Fatalf("unexpected settings %#v", settings)
	}
	if settings.RawDir != filepath.Join("/repo", "data", "raw_mysekai") || settings.BarkMapFile != filepath.Join("/repo", "config", "bark_map.json") {
		t.Fatalf("unexpected paths %#v", settings)
	}
}

func TestEnvMegabytesRejectsInvalidAndOverflow(t *testing.T) {
	t.Setenv("TEST_MEGABYTES", "invalid")
	if got := envMegabytes("TEST_MEGABYTES", 3); got != 3*1024*1024 {
		t.Fatalf("invalid value got %d", got)
	}
	t.Setenv("TEST_MEGABYTES", "9223372036854775807")
	if got := envMegabytes("TEST_MEGABYTES", 3); got != 3*1024*1024 {
		t.Fatalf("overflow value got %d", got)
	}
}

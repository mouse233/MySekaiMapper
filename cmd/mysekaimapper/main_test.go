package main

import (
	"path/filepath"
	"reflect"
	"testing"
)

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

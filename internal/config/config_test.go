package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPasswordEncryptedOnSaveAndDecryptedOnLoad(t *testing.T) {
	dir := t.TempDir()
	opts := Options{
		ConfigDir: dir,
		KeyFile:   filepath.Join(dir, "secret.key"),
	}
	cfg := NewWithOptions(opts)
	cfg.Connections = []Host{{
		Name:     "test",
		Host:     "127.0.0.1",
		Port:     22,
		User:     "root",
		Password: "plain-secret",
	}}

	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatalf("read saved config: %v", err)
	}
	if strings.Contains(string(raw), "plain-secret") {
		t.Fatalf("saved config contains plaintext password: %s", raw)
	}
	if !strings.Contains(string(raw), encryptedPasswordPrefix) {
		t.Fatalf("saved config does not contain encrypted password prefix: %s", raw)
	}

	loaded, err := LoadWithOptions(opts)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if got := loaded.Connections[0].Password; got != "plain-secret" {
		t.Fatalf("loaded password = %q, want plaintext", got)
	}
}

func TestCustomThemeDirAppended(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "custom.yaml"), []byte(`
body:
  fgColor: "#ffffff"
  bgColor: "#000000"
`), 0o600); err != nil {
		t.Fatal(err)
	}

	opts := Options{ThemeDirs: []string{dir}}
	if !containsSkin(AvailableSkinsWithOptions(opts), "custom") {
		t.Fatal("expected custom skin in available skins")
	}
	styles := NewStylesWithOptions(opts)
	if err := styles.Load("custom"); err != nil {
		t.Fatalf("load custom skin: %v", err)
	}
	if got := styles.Body().BgColor; got != "#000000" {
		t.Fatalf("custom body bg = %q, want #000000", got)
	}
}

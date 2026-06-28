package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestSFTPPathInputConfirmKeys(t *testing.T) {
	app := newHostFormTestApp(t)
	base := t.TempDir()
	target := filepath.Join(base, "target")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}

	v := &SFTPView{
		app:      app,
		localDir: base,
	}
	v.build()
	handler := v.root.InputHandler()
	if handler == nil {
		t.Fatal("sftp root input handler is nil")
	}

	for _, key := range []tcell.Key{tcell.KeyEnter, tcell.KeyCtrlJ} {
		v.localDir = base
		v.activePane = 0
		v.inputMode = "path"
		v.pathInput = tview.NewInputField()
		v.pathInput.SetText(target)

		runKey(t, func() {
			handler(tcell.NewEventKey(key, 0, tcell.ModNone), nil)
		})

		if v.isInputMode() {
			t.Fatalf("expected path mode to finish for key %v", key)
		}
		if v.localDir != target {
			t.Fatalf("expected local dir %q for key %v, got %q", target, key, v.localDir)
		}
	}
}

func TestSFTPCreateLocalDir(t *testing.T) {
	app := newHostFormTestApp(t)
	base := t.TempDir()
	v := &SFTPView{
		app:      app,
		localDir: base,
	}
	v.build()

	if !v.createLocalDir("nested/new-dir") {
		t.Fatal("expected local mkdir to succeed")
	}

	if info, err := os.Stat(filepath.Join(base, "nested", "new-dir")); err != nil {
		t.Fatal(err)
	} else if !info.IsDir() {
		t.Fatal("expected created path to be a directory")
	}
}

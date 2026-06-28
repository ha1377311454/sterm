package ui

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/helianthus/sterm/internal/config"
)

func TestHostFormInputDoesNotBlockAndEditsField(t *testing.T) {
	app := newHostFormTestApp(t)
	hf := NewHostForm(app, &app.cfg.Connections[0], 0)
	handler := hf.InputHandler()

	runKey(t, func() {
		handler(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone), nil)
	})

	if got := hf.fields[0].value; got != "229x" {
		t.Fatalf("expected edited name, got %q", got)
	}

	runKey(t, func() {
		handler(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), nil)
	})
	runKey(t, func() {
		handler(tcell.NewEventKey(tcell.KeyRune, '1', tcell.ModNone), nil)
	})

	if got := hf.fields[1].value; got != "192.168.5.2291" {
		t.Fatalf("expected edited host, got %q", got)
	}
}

func TestHostFormGlobalCancelDoesNotBlock(t *testing.T) {
	app := newHostFormTestApp(t)
	hf := NewHostForm(app, &app.cfg.Connections[0], 0)
	app.stack.AddAndSwitchToPage("modal", modalLayout(hf, 64, 24), true)

	runKey(t, func() {
		if event := app.onGlobalKey(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone)); event != nil {
			t.Fatalf("expected q to be consumed")
		}
	})

	if app.stack.HasPage("modal") {
		t.Fatal("expected modal to be closed")
	}
}

func TestHostFormButtonsSwitchWithLeftRight(t *testing.T) {
	app := newHostFormTestApp(t)
	hf := NewHostForm(app, &app.cfg.Connections[0], 0)
	handler := hf.InputHandler()
	hf.focus = saveButton

	runKey(t, func() {
		handler(tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone), nil)
	})
	if hf.focus != cancelButton {
		t.Fatalf("expected right to focus cancel, got %d", hf.focus)
	}

	runKey(t, func() {
		handler(tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone), nil)
	})
	if hf.focus != saveButton {
		t.Fatalf("expected left to focus save, got %d", hf.focus)
	}
}

func newHostFormTestApp(t *testing.T) *App {
	t.Helper()
	t.Setenv("HOME", t.TempDir())

	cfg := config.New()
	cfg.Connections = []config.Host{{
		Name:     "229",
		Host:     "192.168.5.229",
		Port:     22,
		User:     "root",
		Password: "secret",
	}}

	app := &App{
		tv:     tview.NewApplication(),
		cfg:    cfg,
		styles: config.NewStyles(),
		stack:  tview.NewPages(),
		pages:  tview.NewPages(),
	}
	app.hostList = NewHostList(app)
	app.stack.AddPage("main", app.hostList.root, true, true)
	app.tv.SetRoot(app.stack, true)
	app.tv.SetInputCapture(app.onGlobalKey)
	return app
}

func runKey(t *testing.T, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("key handler blocked")
	}
}

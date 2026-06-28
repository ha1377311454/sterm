package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/ha1377311454/sterm/internal/config"
)

func TestHostFilterUsesSubstringMatch(t *testing.T) {
	host229 := config.Host{Name: "229", Host: "192.168.5.229", Port: 22, User: "root", Tags: []string{"dev"}}
	host95 := config.Host{Name: "95", Host: "10.128.2.95", Port: 22, User: "tingyun", Tags: []string{"95环境"}}

	if !hostMatchesQuery(host229, "229") {
		t.Fatal("expected 229 host to match query 229")
	}
	if hostMatchesQuery(host95, "229") {
		t.Fatal("expected 95 host not to match query 229")
	}
	if !hostMatchesQuery(host95, "95") {
		t.Fatal("expected 95 host to match query 95")
	}
	if !hostMatchesQuery(host95, "tingyun") {
		t.Fatal("expected user field to be searchable")
	}
}

func TestDeleteConfirmEnterDeletesHost(t *testing.T) {
	app := newDeleteTestApp(t)
	hl := app.hostList
	if len(app.cfg.Connections) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(app.cfg.Connections))
	}

	hl.doDelete()
	if !app.stack.HasPage("confirm") {
		t.Fatal("expected confirm page on stack")
	}

	dlg, ok := app.tv.GetFocus().(*ConfirmDialog)
	if !ok {
		t.Fatalf("expected ConfirmDialog focus, got %T", app.tv.GetFocus())
	}

	runKey(t, func() {
		if handler := dlg.InputHandler(); handler != nil {
			handler(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {
				app.tv.SetFocus(p)
			})
		}
	})

	if app.stack.HasPage("confirm") {
		t.Fatal("expected confirm page to close")
	}
	if len(app.cfg.Connections) != 1 {
		t.Fatalf("expected 1 host after delete, got %d", len(app.cfg.Connections))
	}
	if app.cfg.Connections[0].Name != "keep" {
		t.Fatalf("expected remaining host %q, got %q", "keep", app.cfg.Connections[0].Name)
	}
}

func newDeleteTestApp(t *testing.T) *App {
	t.Helper()
	t.Setenv("HOME", t.TempDir())

	cfg := config.New()
	cfg.Connections = []config.Host{
		{Name: "mac", Host: "192.168.1.1"},
		{Name: "keep", Host: "192.168.1.2"},
	}

	app := &App{
		tv:     tview.NewApplication(),
		cfg:    cfg,
		styles: config.NewStyles(),
		pages:  tview.NewPages(),
		stack:  tview.NewPages(),
	}
	app.hostList = NewHostList(app)
	app.pages.AddPage("hostlist", app.hostList.root, true, true)
	app.rootFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(app.pages, 0, 1, true)
	app.stack.AddPage("main", app.rootFlex, true, true)
	app.tv.SetRoot(app.stack, true)
	app.tv.SetInputCapture(app.onGlobalKey)
	app.hostList.table.Select(1, 0) // mac
	return app
}

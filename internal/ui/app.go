package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/helianthus/sterm/internal/config"
	"github.com/helianthus/sterm/internal/ssh"
)

const appName = "sterm"

// App 是 TUI 应用的根节点。
type App struct {
	tv        *tview.Application
	stack     *tview.Pages // 根级页面：main | modal
	pages     *tview.Pages // 内层页面：hostlist | sftp | confirm | theme
	cfg       *config.Config
	styles    *config.Styles
	hostList  *HostList
	titleBar  *tview.TextView
	statusBar *tview.TextView
	cmdBar    *tview.TextView
	rootFlex  *tview.Flex
	options   config.Options
}

// NewApp 创建并组装完整 TUI。
func NewApp() (*App, error) {
	return NewAppWithOptions(config.Options{})
}

// NewAppWithOptions 使用运行时路径选项创建并组装完整 TUI。
func NewAppWithOptions(opts config.Options) (*App, error) {
	cfg, err := config.LoadWithOptions(opts)
	if err != nil {
		return nil, err
	}
	styles := config.NewStylesWithOptions(opts)
	if cfg.Theme != "" {
		_ = styles.Load(cfg.Theme)
	}
	a := &App{
		tv:      tview.NewApplication(),
		cfg:     cfg,
		styles:  styles,
		options: opts,
	}
	a.rebuild()
	return a, nil
}

// rebuild 重新创建所有组件并设置应用根节点；主题变更后也会调用。
func (a *App) rebuild() {
	st := a.styles
	applyTviewStyles(st)
	a.pages = tview.NewPages()

	a.hostList = NewHostList(a)

	a.titleBar = tview.NewTextView().SetDynamicColors(true)
	a.titleBar.SetBackgroundColor(st.Body().BgColor.Color())
	a.setTitle("")

	a.statusBar = tview.NewTextView().SetDynamicColors(true)
	a.statusBar.SetBackgroundColor(st.Status().BgColor.Color())
	a.statusBar.SetTextColor(st.Status().FgColor.Color())

	a.cmdBar = tview.NewTextView().SetDynamicColors(true)
	a.cmdBar.SetBackgroundColor(st.Status().BgColor.Color())
	a.cmdBar.SetText(cmdBarText())

	a.pages.AddPage("hostlist", a.hostList.root, true, true)

	a.rootFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.titleBar, 1, 0, false).
		AddItem(a.pages, 0, 1, true).
		AddItem(a.statusBar, 1, 0, false).
		AddItem(a.cmdBar, 1, 0, false)
	a.rootFlex.SetBackgroundColor(st.Body().BgColor.Color())

	a.stack = tview.NewPages()
	a.stack.AddPage("main", a.rootFlex, true, true)

	a.tv.SetRoot(a.stack, true)
	a.tv.SetInputCapture(a.onGlobalKey)
	a.tv.SetFocus(a.hostList.table)
}

func (a *App) setTitle(extra string) {
	st := a.styles
	text := fmt.Sprintf("[%s::b] %s [-:-:-]  SSH Connection Manager", st.Frame().Title.FgColor, appName)
	if extra != "" {
		text += "  " + extra
	}
	a.titleBar.SetText(text)
}

func cmdBarText() string {
	return "[yellow]<enter>[-]Connect  [yellow]a[-]Add  [yellow]e[-]Edit  [yellow]d[-]Del  " +
		"[yellow]f[-]SFTP  [yellow]/[-]Search  [yellow]t[-]Theme  [yellow]q[-]Quit"
}

// applyTviewStyles 将 rivo/tview 全局默认样式与当前皮肤同步。
// tview 默认将 ContrastBackgroundColor 设为蓝色，会导致输入框被染色。
func applyTviewStyles(st *config.Styles) {
	bg := st.Body().BgColor.Color()
	formBg := st.Form().BgColor.Color()
	tview.Styles.PrimitiveBackgroundColor = bg
	tview.Styles.ContrastBackgroundColor = formBg
	tview.Styles.MoreContrastBackgroundColor = formBg
	tview.Styles.PrimaryTextColor = st.Body().FgColor.Color()
	tview.Styles.ContrastSecondaryTextColor = st.Form().FieldFgColor.Color()
	tview.Styles.SecondaryTextColor = st.Status().FgColor.Color()
	tview.Styles.BorderColor = st.Frame().Border.FgColor.Color()
	tview.Styles.TitleColor = st.Frame().Title.FgColor.Color()
}

func (a *App) onGlobalKey(event *tcell.EventKey) *tcell.EventKey {
	// 根级模态框（添加/编辑主机表单）。
	if a.stack.HasPage("modal") {
		switch event.Key() {
		case tcell.KeyEscape:
			a.closeFormModal()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q':
				a.closeFormModal()
				return nil
			}
		}
		return event
	}

	front, _ := a.pages.GetFrontPage()

	if event.Key() == tcell.KeyEscape {
		switch front {
		case "theme":
			a.CloseModal("theme")
			return nil
		case "confirm":
			a.closeConfirm()
			return nil
		}
	}

	if front != "hostlist" {
		return event
	}
	switch event.Key() {
	case tcell.KeyCtrlC:
		a.tv.Stop()
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case 'q', 'Q':
			a.tv.Stop()
			return nil
		case 't':
			a.showThemeSelector()
			return nil
		}
	}
	return event
}

// SetStatus 在状态栏显示消息（成功为绿色，错误为红色）。
func (a *App) SetStatus(msg string, isErr bool) {
	go a.tv.QueueUpdateDraw(func() {
		col := a.styles.Status().OkColor
		if isErr {
			col = a.styles.Status().ErrColor
		}
		a.statusBar.SetText(fmt.Sprintf("[%s]%s[-]", col, msg))
	})
}

// ClearStatus 清除状态栏消息。
func (a *App) ClearStatus() {
	go a.tv.QueueUpdateDraw(func() {
		a.statusBar.SetText("")
	})
}

// Connect 暂停 TUI 并运行交互式 SSH 会话。
func (a *App) Connect(idx int) {
	if idx < 0 || idx >= len(a.cfg.Connections) {
		return
	}
	h := a.cfg.Connections[idx]
	opts := ssh.ConnectOptions{
		Host: h.Host, Port: h.Port,
		User: h.User, Password: h.Password,
		KeyPath: h.KeyPath,
	}
	var connErr error
	a.tv.Suspend(func() {
		connErr = ssh.Connect(opts)
	})
	if connErr != nil {
		a.SetStatus(fmt.Sprintf("Connection error: %v", connErr), true)
	}
}

// OpenSFTP 在后台建立 SFTP 连接并打开 SFTP 浏览器。
func (a *App) OpenSFTP(idx int) {
	if idx < 0 || idx >= len(a.cfg.Connections) {
		return
	}
	h := a.cfg.Connections[idx]
	opts := ssh.ConnectOptions{
		Host: h.Host, Port: h.Port,
		User: h.User, Password: h.Password,
		KeyPath: h.KeyPath,
	}
	a.SetStatus(fmt.Sprintf("Connecting SFTP to %s…", h.Host), false)
	go func() {
		sc, err := ssh.NewSFTPClient(opts)
		if err != nil {
			a.SetStatus(fmt.Sprintf("SFTP error: %v", err), true)
			return
		}
		a.tv.QueueUpdateDraw(func() {
			a.ClearStatus()
			sv := NewSFTPView(a, sc)
			a.pages.AddPage("sftp", sv.root, true, true)
			a.pages.SwitchToPage("sftp")
		})
	}()
}

// ShowHostForm 打开添加/编辑表单。editIdx=-1 表示新增主机。
func (a *App) ShowHostForm(editIdx int) {
	applyTviewStyles(a.styles)
	var existing *config.Host
	if editIdx >= 0 && editIdx < len(a.cfg.Connections) {
		h := a.cfg.Connections[editIdx]
		existing = &h
	}
	hf := NewHostForm(a, existing, editIdx)
	a.stack.AddAndSwitchToPage("modal", modalLayout(hf, 64, 24), true)
	a.tv.SetFocus(hf)
}

func (a *App) closeFormModal() {
	if a.stack.HasPage("modal") {
		a.stack.RemovePage("modal")
	}
	a.stack.SwitchToPage("main")
	a.tv.SetFocus(a.hostList.table)
	a.tv.ForceDraw()
}

// CloseModal 移除内层覆盖页面并将焦点返回主机列表。
func (a *App) CloseModal(name string) {
	if a.pages.HasPage(name) {
		a.pages.RemovePage(name)
	}
	a.pages.SwitchToPage("hostlist")
	a.tv.SetFocus(a.hostList.table)
	a.tv.ForceDraw()
}

func (a *App) closeConfirm() {
	if a.pages.HasPage("confirm") {
		a.pages.RemovePage("confirm")
	}
	a.pages.SwitchToPage("hostlist")
	a.tv.SetFocus(a.hostList.table)
	a.tv.ForceDraw()
}

func (a *App) showThemeSelector() {
	v := NewThemeView(a)
	a.pages.AddPage("theme", v.root, true, true)
	a.pages.SwitchToPage("theme")
	a.tv.SetFocus(v.list)
}

// ApplyTheme 加载皮肤并重建 UI。
func (a *App) ApplyTheme(name string) {
	if err := a.styles.Load(name); err != nil {
		a.SetStatus(fmt.Sprintf("Theme error: %v", err), true)
		return
	}
	a.cfg.Theme = name
	_ = a.cfg.Save()
	a.rebuild()
	a.tv.ForceDraw()
}

// Run 启动 TUI 事件循环。
func (a *App) Run() error {
	return a.tv.Run()
}

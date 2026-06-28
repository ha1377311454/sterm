package ui

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/helianthus/sterm/internal/ssh"
)

// SFTPView 是双栏文件浏览器：本地（左）↔ 远程（右）。
type SFTPView struct {
	app         *App
	root        *tview.Flex
	localTable  *tview.Table
	remoteTable *tview.Table
	client      *ssh.SFTPClient
	localDir    string
	remoteDir   string
	localFiles  []localEntry
	remoteFiles []ssh.RemoteFile
	activePane  int // 0=本地，1=远程
	pathInput   *tview.InputField
	inputMode   string // "" | "path" | "mkdir"
}

type localEntry struct {
	name    string
	size    int64
	isDir   bool
	modTime time.Time
}

// NewSFTPView 创建 SFTP 浏览器并开始加载双栏。
func NewSFTPView(app *App, client *ssh.SFTPClient) *SFTPView {
	home, _ := os.UserHomeDir()
	v := &SFTPView{
		app:      app,
		client:   client,
		localDir: home,
	}
	if wd, err := client.WorkingDir(); err == nil {
		v.remoteDir = wd
	} else {
		v.remoteDir = "/"
	}
	v.build()
	v.loadLocal()
	v.loadRemote()
	return v
}

func (v *SFTPView) build() {
	st := v.app.styles

	mkTable := func() *tview.Table {
		t := tview.NewTable()
		t.SetSelectable(true, false)
		t.SetFixed(1, 0)
		t.SetBorder(true)
		t.SetBackgroundColor(st.Table().BgColor.Color())
		t.SetTitleColor(st.Frame().Title.FgColor.Color())
		t.SetSelectedStyle(tcell.StyleDefault.
			Foreground(st.Table().CursorFgColor.Color()).
			Background(st.Table().CursorBgColor.Color()).
			Attributes(tcell.AttrBold))
		return t
	}

	v.localTable = mkTable()
	v.remoteTable = mkTable()

	v.pathInput = tview.NewInputField()
	v.pathInput.SetBackgroundColor(st.Status().BgColor.Color())
	v.pathInput.SetLabelColor(st.Prompt().FilterColor.Color())
	v.pathInput.SetFieldTextColor(st.Prompt().FgColor.Color())
	v.pathInput.SetFieldBackgroundColor(st.Status().BgColor.Color())
	v.pathInput.SetPlaceholder("g path / m mkdir")
	v.pathInput.SetPlaceholderTextColor(tcell.ColorDarkGray)
	v.pathInput.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEnter, tcell.KeyCtrlJ:
			v.confirmPathInput()
			return nil
		case tcell.KeyEscape:
			v.cancelPathInput()
			return nil
		}
		return ev
	})

	cmdBar := tview.NewTextView().SetDynamicColors(true).
		SetText("[yellow]Tab[-]Switch  [yellow]u[-]Upload  [yellow]d[-]Download  " +
			"[yellow]Enter[-]Open dir/Go  [yellow]g[-]Go path  [yellow]m[-]Mkdir  [yellow]Esc[-]Close")
	cmdBar.SetBackgroundColor(st.Status().BgColor.Color())

	panels := tview.NewFlex().
		AddItem(v.localTable, 0, 1, true).
		AddItem(v.remoteTable, 0, 1, false)

	v.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(panels, 0, 1, true).
		AddItem(v.pathInput, 1, 0, false).
		AddItem(cmdBar, 1, 0, false)
	v.root.SetBackgroundColor(st.Body().BgColor.Color())
	v.root.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if !v.isInputMode() {
			return ev
		}
		switch ev.Key() {
		case tcell.KeyEnter, tcell.KeyCtrlJ:
			v.confirmPathInput()
			return nil
		case tcell.KeyEscape:
			v.cancelPathInput()
			return nil
		}
		return ev
	})

	v.setActivePane(0)

	v.localTable.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		return v.handleKey(ev, 0)
	})
	v.remoteTable.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		return v.handleKey(ev, 1)
	})
}

func (v *SFTPView) handleKey(ev *tcell.EventKey, pane int) *tcell.EventKey {
	switch ev.Key() {
	case tcell.KeyTab, tcell.KeyBacktab:
		v.setActivePane(1 - pane)
		return nil
	case tcell.KeyEscape:
		v.close()
		return nil
	case tcell.KeyEnter:
		if pane == 0 {
			v.localEnter()
		} else {
			v.remoteEnter()
		}
		return nil
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'u':
			v.upload()
			return nil
		case 'd':
			v.download()
			return nil
		case 'g':
			v.startPathInput(pane)
			return nil
		case 'm':
			v.startMkdirInput(pane)
			return nil
		}
	}
	return ev
}

func (v *SFTPView) setActivePane(pane int) {
	v.activePane = pane
	st := v.app.styles
	focus := st.Frame().Border.FocusColor.Color()
	blur := st.Frame().Border.FgColor.Color()
	if pane == 0 {
		v.localTable.SetBorderColor(focus)
		v.remoteTable.SetBorderColor(blur)
		if !v.isInputMode() {
			v.app.tv.SetFocus(v.localTable)
		}
	} else {
		v.localTable.SetBorderColor(blur)
		v.remoteTable.SetBorderColor(focus)
		if !v.isInputMode() {
			v.app.tv.SetFocus(v.remoteTable)
		}
	}
}

func (v *SFTPView) startPathInput(pane int) {
	v.inputMode = "path"
	v.setActivePane(pane)
	if pane == 0 {
		v.pathInput.SetLabel(" Local path: ")
		v.pathInput.SetText(v.localDir)
	} else {
		v.pathInput.SetLabel(" Remote path: ")
		v.pathInput.SetText(v.remoteDir)
	}
	v.app.tv.SetFocus(v.pathInput)
	v.app.tv.ForceDraw()
}

func (v *SFTPView) startMkdirInput(pane int) {
	v.inputMode = "mkdir"
	v.setActivePane(pane)
	if pane == 0 {
		v.pathInput.SetLabel(" Local mkdir: ")
	} else {
		v.pathInput.SetLabel(" Remote mkdir: ")
	}
	v.pathInput.SetText("")
	v.pathInput.SetPlaceholder("new directory name or path")
	v.app.tv.SetFocus(v.pathInput)
	v.app.tv.ForceDraw()
}

func (v *SFTPView) confirmPathInput() {
	text := strings.TrimSpace(v.pathInput.GetText())
	if text == "" {
		v.cancelPathInput()
		return
	}

	var ok bool
	switch v.inputMode {
	case "path":
		if v.activePane == 0 {
			ok = v.gotoLocalPath(text)
		} else {
			ok = v.gotoRemotePath(text)
		}
	case "mkdir":
		if v.activePane == 0 {
			ok = v.createLocalDir(text)
		} else {
			ok = v.createRemoteDir(text)
		}
	}
	if !ok {
		v.app.tv.ForceDraw()
		return
	}
	v.finishPathInput()
}

func (v *SFTPView) cancelPathInput() {
	v.finishPathInput()
}

func (v *SFTPView) finishPathInput() {
	v.inputMode = ""
	v.pathInput.SetLabel("")
	v.pathInput.SetText("")
	v.pathInput.SetPlaceholder("g path / m mkdir")
	v.setActivePane(v.activePane)
	v.app.tv.ForceDraw()
}

func (v *SFTPView) gotoLocalPath(input string) bool {
	target := expandLocalPath(input)
	if !filepath.IsAbs(target) {
		target = filepath.Join(v.localDir, target)
	}
	target = filepath.Clean(target)
	info, err := os.Stat(target)
	if err != nil {
		v.app.SetStatus(fmt.Sprintf("Local path error: %v", err), true)
		return false
	}
	if !info.IsDir() {
		v.app.SetStatus(fmt.Sprintf("Local path is not a directory: %s", target), true)
		return false
	}
	v.localDir = target
	v.loadLocal()
	v.app.SetStatus(fmt.Sprintf("Local path: %s", target), false)
	return true
}

func (v *SFTPView) createLocalDir(input string) bool {
	target := expandLocalPath(input)
	if !filepath.IsAbs(target) {
		target = filepath.Join(v.localDir, target)
	}
	target = filepath.Clean(target)
	if err := os.MkdirAll(target, 0o755); err != nil {
		v.app.SetStatus(fmt.Sprintf("Local mkdir error: %v", err), true)
		return false
	}
	v.loadLocal()
	v.app.SetStatus(fmt.Sprintf("Created local dir: %s", target), false)
	return true
}

func (v *SFTPView) createRemoteDir(input string) bool {
	target := input
	if !path.IsAbs(target) {
		target = path.Join(v.remoteDir, target)
	}
	target = path.Clean(target)
	if target == "." || target == "/" {
		v.app.SetStatus("Remote mkdir error: invalid directory name", true)
		return false
	}
	if err := v.client.MkdirAll(target); err != nil {
		v.app.SetStatus(fmt.Sprintf("Remote mkdir error: %v", err), true)
		return false
	}
	v.loadRemote()
	v.app.SetStatus(fmt.Sprintf("Created remote dir: %s", target), false)
	return true
}

func (v *SFTPView) isInputMode() bool {
	return v.inputMode != ""
}

func (v *SFTPView) gotoRemotePath(input string) bool {
	target := input
	if !path.IsAbs(target) {
		target = path.Join(v.remoteDir, target)
	}
	target = path.Clean(target)
	if target == "." {
		target = "/"
	}
	files, err := v.client.ListDir(target)
	if err != nil {
		v.app.SetStatus(fmt.Sprintf("Remote path error: %v", err), true)
		return false
	}
	v.remoteDir = target
	v.remoteFiles = files
	v.renderRemoteTable()
	v.app.SetStatus(fmt.Sprintf("Remote path: %s", target), false)
	return true
}

// ── 导航 ──────────────────────────────────────────────────────────────

func (v *SFTPView) localEnter() {
	row, _ := v.localTable.GetSelection()
	row--
	if row < 0 || row >= len(v.localFiles) {
		return
	}
	f := v.localFiles[row]
	if !f.isDir {
		return
	}
	if f.name == ".." {
		v.localDir = filepath.Dir(v.localDir)
	} else {
		v.localDir = filepath.Join(v.localDir, f.name)
	}
	v.loadLocal()
}

func (v *SFTPView) remoteEnter() {
	row, _ := v.remoteTable.GetSelection()
	row--
	if row < 0 || row >= len(v.remoteFiles) {
		return
	}
	f := v.remoteFiles[row]
	if !f.IsDir {
		return
	}
	if f.Name == ".." {
		v.remoteDir = path.Dir(v.remoteDir)
		if v.remoteDir == "" {
			v.remoteDir = "/"
		}
	} else {
		v.remoteDir = path.Join(v.remoteDir, f.Name)
	}
	v.loadRemote()
}

// ── 传输 ────────────────────────────────────────────────────────────────

func (v *SFTPView) upload() {
	row, _ := v.localTable.GetSelection()
	row--
	if row < 0 || row >= len(v.localFiles) {
		return
	}
	f := v.localFiles[row]
	if f.isDir {
		v.app.SetStatus("Cannot upload a directory", true)
		return
	}
	localPath := filepath.Join(v.localDir, f.name)
	v.app.SetStatus(fmt.Sprintf("Uploading %s…", f.name), false)
	go func() {
		progress := newTransferProgress(v.app, "Uploading", f.name)
		if err := v.client.UploadWithProgress(localPath, v.remoteDir, progress); err != nil {
			v.app.SetStatus(fmt.Sprintf("Upload failed: %v", err), true)
			return
		}
		v.app.SetStatus(fmt.Sprintf("Uploaded: %s", f.name), false)
		v.app.tv.QueueUpdateDraw(v.loadRemote)
	}()
}

func (v *SFTPView) download() {
	row, _ := v.remoteTable.GetSelection()
	row--
	if row < 0 || row >= len(v.remoteFiles) {
		return
	}
	f := v.remoteFiles[row]
	if f.IsDir {
		v.app.SetStatus("Cannot download a directory", true)
		return
	}
	remotePath := path.Join(v.remoteDir, f.Name)
	v.app.SetStatus(fmt.Sprintf("Downloading %s…", f.Name), false)
	go func() {
		progress := newTransferProgress(v.app, "Downloading", f.Name)
		if err := v.client.DownloadWithProgress(remotePath, v.localDir, progress); err != nil {
			v.app.SetStatus(fmt.Sprintf("Download failed: %v", err), true)
			return
		}
		v.app.SetStatus(fmt.Sprintf("Downloaded: %s", f.Name), false)
		v.app.tv.QueueUpdateDraw(v.loadLocal)
	}()
}

// ── 数据加载 ────────────────────────────────────────────────────────────

func (v *SFTPView) loadLocal() {
	entries, err := os.ReadDir(v.localDir)
	if err != nil {
		v.app.SetStatus(fmt.Sprintf("Local read error: %v", err), true)
		return
	}
	v.localFiles = nil
	if v.localDir != "/" {
		v.localFiles = append(v.localFiles, localEntry{name: "..", isDir: true})
	}
	for _, e := range entries {
		info, _ := e.Info()
		var sz int64
		var mod time.Time
		if info != nil {
			sz = info.Size()
			mod = info.ModTime()
		}
		v.localFiles = append(v.localFiles, localEntry{
			name: e.Name(), isDir: e.IsDir(), size: sz, modTime: mod,
		})
	}
	v.renderLocalTable()
}

func (v *SFTPView) loadRemote() {
	files, err := v.client.ListDir(v.remoteDir)
	if err != nil {
		v.app.SetStatus(fmt.Sprintf("Remote read error: %v", err), true)
		return
	}
	v.remoteFiles = files
	v.renderRemoteTable()
}

// ── 渲染 ───────────────────────────────────────────────────────────────

func (v *SFTPView) renderLocalTable() {
	v.localTable.Clear()
	v.localTable.SetTitle(fmt.Sprintf(" Local: %s ", v.localDir))
	v.setFileHeader(v.localTable)
	st := v.app.styles.Table()
	row := tcell.StyleDefault.Foreground(st.FgColor.Color()).Background(st.BgColor.Color())
	for i, f := range v.localFiles {
		size := ""
		mod := ""
		if !f.isDir {
			size = ssh.FormatSize(f.size)
		}
		if !f.modTime.IsZero() {
			mod = f.modTime.Format("2006-01-02 15:04")
		}
		v.setFileRow(v.localTable, i+1, f.name, f.isDir, size, mod, row)
	}
	if len(v.localFiles) > 0 {
		v.localTable.Select(1, 0)
	}
}

func (v *SFTPView) renderRemoteTable() {
	v.remoteTable.Clear()
	v.remoteTable.SetTitle(fmt.Sprintf(" Remote: %s ", v.remoteDir))
	v.setFileHeader(v.remoteTable)
	st := v.app.styles.Table()
	row := tcell.StyleDefault.Foreground(st.FgColor.Color()).Background(st.BgColor.Color())
	for i, f := range v.remoteFiles {
		size := ""
		mod := ""
		if !f.IsDir {
			size = ssh.FormatSize(f.Size)
		}
		if !f.ModTime.IsZero() {
			mod = f.ModTime.Format("2006-01-02 15:04")
		}
		v.setFileRow(v.remoteTable, i+1, f.Name, f.IsDir, size, mod, row)
	}
	if len(v.remoteFiles) > 0 {
		v.remoteTable.Select(1, 0)
	}
}

func (v *SFTPView) setFileHeader(t *tview.Table) {
	st := v.app.styles.Table()
	hdr := tcell.StyleDefault.
		Foreground(st.Header.FgColor.Color()).
		Background(st.Header.BgColor.Color()).
		Attributes(tcell.AttrBold)
	for col, h := range []string{"NAME", "SIZE", "MODIFIED"} {
		t.SetCell(0, col, tview.NewTableCell(" "+h+" ").
			SetStyle(hdr).SetSelectable(false).SetExpansion([]int{4, 1, 2}[col]))
	}
}

func (v *SFTPView) setFileRow(t *tview.Table, row int, name string, isDir bool, size, mod string, base tcell.Style) {
	nameCell := tview.NewTableCell("  " + name + " ").SetStyle(base).SetExpansion(4)
	if isDir {
		nameCell.SetTextColor(tcell.ColorSkyblue).SetAttributes(tcell.AttrBold)
		nameCell.SetText("  " + name + "/")
	}
	t.SetCell(row, 0, nameCell)
	t.SetCell(row, 1, tview.NewTableCell(" "+size+" ").SetStyle(base).SetExpansion(1))
	t.SetCell(row, 2, tview.NewTableCell(" "+mod+" ").SetStyle(base).SetExpansion(2))
}

func expandLocalPath(input string) string {
	if input == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
	}
	if strings.HasPrefix(input, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(input, "~/"))
		}
	}
	return input
}

func newTransferProgress(app *App, action, name string) ssh.ProgressFunc {
	var last time.Time
	return func(done, total int64, elapsed time.Duration) {
		now := time.Now()
		if done < total && !last.IsZero() && now.Sub(last) < 200*time.Millisecond {
			return
		}
		last = now

		speed := float64(done)
		if elapsed > 0 {
			speed = float64(done) / elapsed.Seconds()
		}

		if total > 0 {
			percent := float64(done) * 100 / float64(total)
			app.SetStatus(fmt.Sprintf(
				"%s %s  %.1f%%  %s/%s  %s/s",
				action,
				name,
				percent,
				ssh.FormatSize(done),
				ssh.FormatSize(total),
				ssh.FormatSize(int64(speed)),
			), false)
			return
		}

		app.SetStatus(fmt.Sprintf(
			"%s %s  %s  %s/s",
			action,
			name,
			ssh.FormatSize(done),
			ssh.FormatSize(int64(speed)),
		), false)
	}
}

func (v *SFTPView) close() {
	v.client.Close()
	v.app.CloseModal("sftp")
}

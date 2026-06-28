package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/ha1377311454/sterm/internal/config"
)

// HostList 是主主机列表视图：可筛选的表格。
type HostList struct {
	app      *App
	root     *tview.Flex
	table    *tview.Table
	filter   *tview.InputField
	query    string
	filtered []int // cfg.Connections 中的索引
}

// NewHostList 创建并组装主机列表视图。
func NewHostList(app *App) *HostList {
	hl := &HostList{app: app}
	hl.build()
	return hl
}

func (hl *HostList) build() {
	st := hl.app.styles

	// ── 表格 ──────────────────────────────────────────────────────────────
	hl.table = tview.NewTable()
	hl.table.SetFixed(1, 0) // 固定表头行
	hl.table.SetSelectable(true, false)
	hl.table.SetBorder(true)
	hl.table.SetTitle(" Hosts ")
	hl.table.SetTitleColor(st.Frame().Title.FgColor.Color())
	hl.table.SetBorderColor(st.Frame().Border.FgColor.Color())
	hl.table.SetBackgroundColor(st.Table().BgColor.Color())
	hl.table.SetSelectedStyle(tcell.StyleDefault.
		Foreground(st.Table().CursorFgColor.Color()).
		Background(st.Table().CursorBgColor.Color()).
		Attributes(tcell.AttrBold))

	// ── 筛选栏（按 '/' 激活）───────────────────────────────────────────────
	hl.filter = tview.NewInputField()
	hl.filter.SetLabel("/ ")
	hl.filter.SetLabelColor(tcell.ColorYellow)
	hl.filter.SetBackgroundColor(st.Prompt().BgColor.Color())
	hl.filter.SetFieldBackgroundColor(st.Prompt().BgColor.Color())
	hl.filter.SetFieldTextColor(st.Prompt().FgColor.Color())
	hl.filter.SetFieldStyle(tcell.StyleDefault.
		Foreground(st.Prompt().FgColor.Color()).
		Background(st.Prompt().BgColor.Color()))
	hl.filter.SetPlaceholder("filter…")
	hl.filter.SetPlaceholderTextColor(tcell.ColorDarkGray)
	hl.filter.SetChangedFunc(func(text string) {
		hl.query = text
		hl.refresh()
	})
	hl.filter.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			hl.clearFilter()
		}
		hl.app.tv.SetFocus(hl.table)
	})

	hl.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(hl.table, 0, 1, true).
		AddItem(hl.filter, 1, 0, false)
	hl.root.SetBackgroundColor(st.Body().BgColor.Color())

	hl.refresh()
	hl.table.SetInputCapture(hl.onKey)
}

func (hl *HostList) onKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEnter:
		hl.doConnect()
		return nil
	case tcell.KeyEscape:
		hl.clearFilter()
		return nil
	case tcell.KeyCtrlU:
		hl.clearFilter()
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case '/':
			hl.activateFilter()
			return nil
		case 'a':
			hl.app.ShowHostForm(-1)
			return nil
		case 'e':
			hl.doEdit()
			return nil
		case 'd':
			hl.doDelete()
			return nil
		case 'f':
			hl.doSFTP()
			return nil
		}
	}
	return event
}

func (hl *HostList) activateFilter() {
	hl.app.tv.SetFocus(hl.filter)
}

func (hl *HostList) isFilterFocused() bool {
	return hl.filter.HasFocus()
}

func (hl *HostList) clearFilter() {
	hl.query = ""
	hl.filter.SetText("")
	hl.refresh()
}

// selectedCfgIdx 返回当前高亮行在 cfg.Connections 中的索引，未选中时返回 -1。
func (hl *HostList) selectedCfgIdx() int {
	row, _ := hl.table.GetSelection()
	row-- // 减去表头行
	if row < 0 || row >= len(hl.filtered) {
		return -1
	}
	return hl.filtered[row]
}

func (hl *HostList) doConnect() {
	if idx := hl.selectedCfgIdx(); idx >= 0 {
		hl.app.Connect(idx)
	}
}

func (hl *HostList) doEdit() {
	if idx := hl.selectedCfgIdx(); idx >= 0 {
		hl.app.ShowHostForm(idx)
	}
}

func (hl *HostList) doDelete() {
	idx := hl.selectedCfgIdx()
	if idx < 0 {
		return
	}
	name := hl.app.cfg.Connections[idx].Name
	deleteIdx := idx

	dlg := NewConfirmDialog(hl.app,
		fmt.Sprintf("Delete %s?", name),
		func() { hl.app.deleteHostAt(deleteIdx) },
	)
	hl.app.stack.AddAndSwitchToPage("confirm", modalLayout(dlg, 44, 7), true)
	hl.app.tv.SetFocus(dlg)
}

func (hl *HostList) doSFTP() {
	if idx := hl.selectedCfgIdx(); idx >= 0 {
		hl.app.OpenSFTP(idx)
	}
}

// Reload 重新读取 cfg.Connections 并重绘表格。
func (hl *HostList) Reload() {
	hl.refresh()
}

func (hl *HostList) refresh() {
	hl.buildFiltered()
	hl.renderTable()
	if hl.query != "" {
		hl.table.SetTitle(fmt.Sprintf(" Hosts [yellow][/%s][-] ", hl.query))
	} else {
		hl.table.SetTitle(" Hosts ")
	}
}

func (hl *HostList) buildFiltered() {
	all := hl.app.cfg.Connections
	if hl.query == "" {
		hl.filtered = make([]int, len(all))
		for i := range all {
			hl.filtered[i] = i
		}
		return
	}
	// 在各字段中做不区分大小写的子串匹配
	query := strings.ToLower(strings.TrimSpace(hl.query))
	hl.filtered = hl.filtered[:0]
	for i, h := range all {
		if hostMatchesQuery(h, query) {
			hl.filtered = append(hl.filtered, i)
		}
	}
}

// hostMatchesQuery 判断主机是否在任一字段中包含 query（不区分大小写）。
func hostMatchesQuery(h config.Host, query string) bool {
	if query == "" {
		return true
	}
	for _, s := range []string{
		h.Name, h.Host, h.User, h.TagsStr(), h.Description,
		fmt.Sprintf("%d", h.Port),
	} {
		if strings.Contains(strings.ToLower(s), query) {
			return true
		}
	}
	return false
}

func (hl *HostList) renderTable() {
	hl.table.Clear()
	st := hl.app.styles.Table()

	hdrStyle := tcell.StyleDefault.
		Foreground(st.Header.FgColor.Color()).
		Background(st.Header.BgColor.Color()).
		Attributes(tcell.AttrBold)

	headers := []string{"  NAME", "HOST", "PORT", "USER", "TAGS", "DESCRIPTION"}
	colExpand := []int{2, 2, 1, 1, 2, 3}
	for col, h := range headers {
		cell := tview.NewTableCell(h).
			SetStyle(hdrStyle).
			SetSelectable(false).
			SetExpansion(colExpand[col])
		hl.table.SetCell(0, col, cell)
	}

	rowStyle := tcell.StyleDefault.
		Foreground(st.FgColor.Color()).
		Background(st.BgColor.Color())

	for row, cfgIdx := range hl.filtered {
		h := hl.app.cfg.Connections[cfgIdx]
		vals := []string{
			"  " + h.Name,
			h.Host,
			fmt.Sprintf("%d", h.Port),
			h.User,
			h.TagsStr(),
			h.Description,
		}
		for col, v := range vals {
			cell := tview.NewTableCell(v + " ").
				SetStyle(rowStyle).
				SetExpansion(colExpand[col])
			hl.table.SetCell(row+1, col, cell)
		}
	}

	if len(hl.filtered) > 0 {
		hl.table.Select(1, 0)
	}
}

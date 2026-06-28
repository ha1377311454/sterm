package ui

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/ha1377311454/sterm/internal/config"
)

const (
	formFieldLen = 28
	formFields   = 8
	saveButton   = formFields
	cancelButton = formFields + 1
)

type hostFormField struct {
	label    string
	value    string
	cursor   int
	password bool
	digits   bool
}

// HostForm 是 SSH 连接的模态添加/编辑面板。
type HostForm struct {
	*tview.Box
	app     *App
	editIdx int
	title   string
	fields  []hostFormField
	focus   int
}

// NewHostForm 创建表单。editIdx=-1 表示新增主机。
func NewHostForm(app *App, existing *config.Host, editIdx int) *HostForm {
	hf := &HostForm{
		Box:     tview.NewBox(),
		app:     app,
		editIdx: editIdx,
		title:   " Add Host ",
	}
	if existing != nil {
		hf.title = fmt.Sprintf(" Edit: %s ", existing.Name)
	}
	hf.build(existing)
	return hf
}

func (hf *HostForm) build(existing *config.Host) {
	name, host, port, user, pass, keyPath, desc, tags :=
		"", "", "22", "root", "", "", "", ""
	if existing != nil {
		name = existing.Name
		host = existing.Host
		port = strconv.Itoa(existing.Port)
		user = existing.User
		pass = existing.Password
		keyPath = existing.KeyPath
		desc = existing.Description
		tags = existing.TagsStr()
	}

	hf.fields = []hostFormField{
		{label: "Name:", value: name},
		{label: "Host:", value: host},
		{label: "Port:", value: port, digits: true},
		{label: "User:", value: user},
		{label: "Password:", value: pass, password: true},
		{label: "Key Path:", value: keyPath},
		{label: "Description:", value: desc},
		{label: "Tags:", value: tags},
	}
	for i := range hf.fields {
		hf.fields[i].cursor = utf8.RuneCountInString(hf.fields[i].value)
	}
}

func (hf *HostForm) Draw(screen tcell.Screen) {
	hf.Box.DrawForSubclass(screen, hf)
	x, y, width, height := hf.GetInnerRect()
	if width < 20 || height < 10 {
		return
	}

	st := hf.app.styles
	formSt := st.Form()
	bg := formSt.BgColor.Color()
	fg := formSt.FgColor.Color()
	fieldFg := formSt.FieldFgColor.Color()
	btnBg := formSt.ButtonBgColor.Color()
	btnFg := formSt.ButtonFgColor.Color()
	border := st.Frame().Border.FocusColor.Color()
	title := st.Frame().Title.FgColor.Color()
	focusBg := st.Table().CursorBgColor.Color()
	focusFg := st.Table().CursorFgColor.Color()

	borderStyle := tcell.StyleDefault.Foreground(border).Background(bg)
	plainStyle := tcell.StyleDefault.Foreground(fg).Background(bg)
	fieldStyle := tcell.StyleDefault.Foreground(fieldFg).Background(bg)
	focusStyle := tcell.StyleDefault.Foreground(focusFg).Background(focusBg)
	buttonStyle := tcell.StyleDefault.
		Foreground(btnBg).
		Background(bg)
	buttonFocusStyle := tcell.StyleDefault.
		Foreground(btnFg).
		Background(btnBg).
		Attributes(tcell.AttrBold)

	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			screen.SetContent(x+col, y+row, ' ', nil, plainStyle)
		}
	}
	drawBorder(screen, x, y, width, height, borderStyle)
	tview.Print(screen, hf.title+"  (Esc: cancel)", x+2, y, width-4, tview.AlignCenter, title)

	labelX := x + 3
	valueX := x + 16
	startY := y + 3
	fieldWidth := minInt(formFieldLen, maxInt(8, width-20))

	for i := range hf.fields {
		field := &hf.fields[i]
		rowY := startY + i
		tview.Print(screen, field.label, labelX, rowY, 12, tview.AlignLeft, fg)

		display := field.value
		if field.password {
			display = strings.Repeat("*", utf8.RuneCountInString(field.value))
		}
		display = fitAroundCursor(display, field.cursor, fieldWidth)
		style := fieldStyle
		if hf.focus == i {
			style = focusStyle
			for col := 0; col < fieldWidth; col++ {
				screen.SetContent(valueX+col, rowY, ' ', nil, style)
			}
		}
		tview.Print(screen, display, valueX, rowY, fieldWidth, tview.AlignLeft, fieldFg)
		if hf.focus == i {
			cursorX := valueX + minInt(field.cursor, fieldWidth-1)
			screen.SetContent(cursorX, rowY, '|', nil, style)
		}
	}

	buttonY := startY + len(hf.fields) + 3
	saveLabel := "Save"
	cancelLabel := "Cancel"
	saveWidth := 12
	cancelWidth := 12
	total := saveWidth + 2 + cancelWidth
	buttonX := x + (width-total)/2
	hf.drawButton(screen, buttonX, buttonY, saveWidth, saveLabel, hf.focus == saveButton, buttonStyle, buttonFocusStyle)
	hf.drawButton(screen, buttonX+saveWidth+2, buttonY, cancelWidth, cancelLabel, hf.focus == cancelButton, buttonStyle, buttonFocusStyle)
	screen.HideCursor()
}

func (hf *HostForm) drawButton(screen tcell.Screen, x, y, width int, label string, focused bool, normal, focusedStyle tcell.Style) {
	style := normal
	if focused {
		style = focusedStyle
	}
	for col := 0; col < width; col++ {
		screen.SetContent(x+col, y, ' ', nil, style)
	}
	textX := x + (width-utf8.RuneCountInString(label))/2
	for i, r := range label {
		screen.SetContent(textX+i, y, r, nil, style)
	}
}

func (hf *HostForm) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return hf.WrapInputHandler(func(event *tcell.EventKey, _ func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyEscape:
			hf.app.closeFormModal()
		case tcell.KeyTab, tcell.KeyDown:
			hf.moveFocus(1)
		case tcell.KeyBacktab, tcell.KeyUp:
			hf.moveFocus(-1)
		case tcell.KeyEnter:
			if hf.focus == saveButton {
				hf.save()
				return
			}
			if hf.focus == cancelButton {
				hf.app.closeFormModal()
				return
			}
			hf.moveFocus(1)
		case tcell.KeyLeft:
			if hf.focus >= saveButton {
				hf.moveButtonFocus(-1)
			} else {
				hf.moveCursor(-1)
			}
		case tcell.KeyRight:
			if hf.focus >= saveButton {
				hf.moveButtonFocus(1)
			} else {
				hf.moveCursor(1)
			}
		case tcell.KeyHome:
			hf.setCursor(0)
		case tcell.KeyEnd:
			if hf.focus < len(hf.fields) {
				hf.setCursor(utf8.RuneCountInString(hf.fields[hf.focus].value))
			}
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			hf.backspace()
		case tcell.KeyDelete:
			hf.delete()
		case tcell.KeyRune:
			hf.insertRune(event.Rune())
		}
		hf.app.tv.ForceDraw()
	})
}

func (hf *HostForm) moveFocus(delta int) {
	count := len(hf.fields) + 2
	hf.focus = (hf.focus + delta + count) % count
}

func (hf *HostForm) moveButtonFocus(delta int) {
	if hf.focus < saveButton {
		return
	}
	if delta < 0 {
		hf.focus = saveButton
		return
	}
	hf.focus = cancelButton
}

func (hf *HostForm) moveCursor(delta int) {
	if hf.focus >= len(hf.fields) {
		return
	}
	field := &hf.fields[hf.focus]
	field.cursor = clampInt(field.cursor+delta, 0, utf8.RuneCountInString(field.value))
}

func (hf *HostForm) setCursor(pos int) {
	if hf.focus >= len(hf.fields) {
		return
	}
	field := &hf.fields[hf.focus]
	field.cursor = clampInt(pos, 0, utf8.RuneCountInString(field.value))
}

func (hf *HostForm) insertRune(r rune) {
	if hf.focus >= len(hf.fields) || r == 0 {
		return
	}
	field := &hf.fields[hf.focus]
	if field.digits && (r < '0' || r > '9') {
		return
	}
	runes := []rune(field.value)
	field.cursor = clampInt(field.cursor, 0, len(runes))
	runes = append(runes[:field.cursor], append([]rune{r}, runes[field.cursor:]...)...)
	field.value = string(runes)
	field.cursor++
}

func (hf *HostForm) backspace() {
	if hf.focus >= len(hf.fields) {
		return
	}
	field := &hf.fields[hf.focus]
	runes := []rune(field.value)
	if field.cursor <= 0 || len(runes) == 0 {
		return
	}
	field.cursor = clampInt(field.cursor, 0, len(runes))
	runes = append(runes[:field.cursor-1], runes[field.cursor:]...)
	field.value = string(runes)
	field.cursor--
}

func (hf *HostForm) delete() {
	if hf.focus >= len(hf.fields) {
		return
	}
	field := &hf.fields[hf.focus]
	runes := []rune(field.value)
	if field.cursor < 0 || field.cursor >= len(runes) {
		return
	}
	runes = append(runes[:field.cursor], runes[field.cursor+1:]...)
	field.value = string(runes)
}

func (hf *HostForm) save() {
	name := strings.TrimSpace(hf.fields[0].value)
	host := strings.TrimSpace(hf.fields[1].value)
	portStr := strings.TrimSpace(hf.fields[2].value)
	user := strings.TrimSpace(hf.fields[3].value)
	pass := hf.fields[4].value
	keyPath := strings.TrimSpace(hf.fields[5].value)
	desc := strings.TrimSpace(hf.fields[6].value)
	tagsRaw := strings.TrimSpace(hf.fields[7].value)

	if name == "" || host == "" {
		hf.app.SetStatus("Name and Host are required", true)
		return
	}

	port, _ := strconv.Atoi(portStr)
	if port <= 0 {
		port = 22
	}

	var tagList []string
	for _, t := range strings.Split(tagsRaw, ",") {
		if t = strings.TrimSpace(t); t != "" {
			tagList = append(tagList, t)
		}
	}

	h := config.Host{
		Name: name, Host: host, Port: port,
		User: user, Password: pass, KeyPath: keyPath,
		Description: desc, Tags: tagList,
	}

	if hf.editIdx >= 0 {
		hf.app.cfg.UpdateHost(hf.editIdx, h)
	} else {
		hf.app.cfg.AddHost(h)
	}

	if err := hf.app.cfg.Save(); err != nil {
		hf.app.SetStatus(fmt.Sprintf("Save error: %v", err), true)
		return
	}

	hf.app.hostList.Reload()
	hf.app.closeFormModal()
	hf.app.SetStatus(fmt.Sprintf("Saved: %s", name), false)
}

func drawBorder(screen tcell.Screen, x, y, width, height int, style tcell.Style) {
	for col := 0; col < width; col++ {
		screen.SetContent(x+col, y, tview.BoxDrawingsLightHorizontal, nil, style)
		screen.SetContent(x+col, y+height-1, tview.BoxDrawingsLightHorizontal, nil, style)
	}
	for row := 0; row < height; row++ {
		screen.SetContent(x, y+row, tview.BoxDrawingsLightVertical, nil, style)
		screen.SetContent(x+width-1, y+row, tview.BoxDrawingsLightVertical, nil, style)
	}
	screen.SetContent(x, y, tview.BoxDrawingsLightDownAndRight, nil, style)
	screen.SetContent(x+width-1, y, tview.BoxDrawingsLightDownAndLeft, nil, style)
	screen.SetContent(x, y+height-1, tview.BoxDrawingsLightUpAndRight, nil, style)
	screen.SetContent(x+width-1, y+height-1, tview.BoxDrawingsLightUpAndLeft, nil, style)
}

func fitAroundCursor(text string, cursor, width int) string {
	runes := []rune(text)
	if width <= 0 {
		return ""
	}
	if len(runes) <= width {
		return text
	}
	cursor = clampInt(cursor, 0, len(runes))
	start := 0
	if cursor >= width {
		start = cursor - width + 1
	}
	if start+width > len(runes) {
		start = len(runes) - width
	}
	return string(runes[start : start+width])
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

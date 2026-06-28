package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	confirmDeleteButton = 0
	confirmCancelButton = 1
)

// ConfirmDialog 是带 Delete / Cancel 按钮的确认框。
type ConfirmDialog struct {
	*tview.Box
	app       *App
	title     string
	message   string
	focus     int
	onConfirm func()
}

// NewConfirmDialog 创建确认框。onConfirm 在用户确认后调用。
func NewConfirmDialog(app *App, message string, onConfirm func()) *ConfirmDialog {
	return &ConfirmDialog{
		Box:       tview.NewBox(),
		app:       app,
		title:     " Confirm ",
		message:   message,
		focus:     confirmDeleteButton,
		onConfirm: onConfirm,
	}
}

func (d *ConfirmDialog) Draw(screen tcell.Screen) {
	d.Box.DrawForSubclass(screen, d)
	x, y, width, height := d.GetInnerRect()
	if width < 20 || height < 5 {
		return
	}

	st := d.app.styles
	formSt := st.Form()
	bg := formSt.BgColor.Color()
	fg := formSt.FgColor.Color()
	btnBg := formSt.ButtonBgColor.Color()
	btnFg := formSt.ButtonFgColor.Color()
	border := st.Frame().Border.FocusColor.Color()
	title := st.Frame().Title.FgColor.Color()

	borderStyle := tcell.StyleDefault.Foreground(border).Background(bg)
	plainStyle := tcell.StyleDefault.Foreground(fg).Background(bg)
	buttonStyle := tcell.StyleDefault.Foreground(btnBg).Background(bg)
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
	tview.Print(screen, d.title+"  (Esc: cancel)", x+2, y, width-4, tview.AlignCenter, title)
	tview.Print(screen, d.message, x+2, y+2, width-4, tview.AlignCenter, fg)

	deleteLabel := "Delete"
	cancelLabel := "Cancel"
	btnWidth := 12
	total := btnWidth*2 + 2
	btnX := x + (width-total)/2
	btnY := y + height - 3
	d.drawButton(screen, btnX, btnY, btnWidth, deleteLabel, d.focus == confirmDeleteButton, buttonStyle, buttonFocusStyle)
	d.drawButton(screen, btnX+btnWidth+2, btnY, btnWidth, cancelLabel, d.focus == confirmCancelButton, buttonStyle, buttonFocusStyle)
	screen.HideCursor()
}

func (d *ConfirmDialog) drawButton(screen tcell.Screen, x, y, width int, label string, focused bool, normal, focusedStyle tcell.Style) {
	style := normal
	if focused {
		style = focusedStyle
	}
	for col := 0; col < width; col++ {
		screen.SetContent(x+col, y, ' ', nil, style)
	}
	textX := x + (width-len(label))/2
	for i, r := range label {
		screen.SetContent(textX+i, y, r, nil, style)
	}
}

func (d *ConfirmDialog) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return d.WrapInputHandler(func(event *tcell.EventKey, _ func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyEscape:
			d.app.closeConfirm()
		case tcell.KeyTab, tcell.KeyDown, tcell.KeyRight:
			d.focus = confirmCancelButton
		case tcell.KeyBacktab, tcell.KeyUp, tcell.KeyLeft:
			d.focus = confirmDeleteButton
		case tcell.KeyEnter:
			if d.focus == confirmDeleteButton {
				if d.onConfirm != nil {
					d.onConfirm()
				}
			} else {
				d.app.closeConfirm()
			}
		}
		d.app.tv.ForceDraw()
	})
}

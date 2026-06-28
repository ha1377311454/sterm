package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/helianthus/sterm/internal/config"
)

// ThemeView 展示可选皮肤列表。
type ThemeView struct {
	app  *App
	root tview.Primitive
	list *tview.List
}

// NewThemeView 创建主题选择器。
func NewThemeView(app *App) *ThemeView {
	v := &ThemeView{app: app}
	v.build()
	return v
}

func (v *ThemeView) build() {
	st := v.app.styles

	v.list = tview.NewList()
	v.list.ShowSecondaryText(false)
	v.list.SetBorder(true)
	v.list.SetTitle(" Select Theme ")
	v.list.SetTitleColor(st.Frame().Title.FgColor.Color())
	v.list.SetBorderColor(st.Frame().Border.FocusColor.Color())
	v.list.SetBackgroundColor(st.Form().BgColor.Color())
	v.list.SetMainTextColor(st.Form().FgColor.Color())
	v.list.SetSelectedBackgroundColor(st.Table().CursorBgColor.Color())
	v.list.SetSelectedTextColor(st.Table().CursorFgColor.Color())

	current := v.app.cfg.Theme
	for _, name := range config.AvailableSkinsWithOptions(v.app.options) {
		label := name
		if name == current {
			label += " [yellow](active)[-]"
		}
		skinName := name
		v.list.AddItem(label, "", 0, func() {
			v.app.CloseModal("theme")
			v.app.ApplyTheme(skinName)
		})
	}

	v.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			v.app.CloseModal("theme")
			return nil
		}
		return event
	})

	v.root = centerBox(v.list, 40, 16)
}

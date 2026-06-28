package ui

import "github.com/rivo/tview"

// centerBox 用嵌套 Flex 容器包裹组件，使其在屏幕中央浮动显示。
func centerBox(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, height, 0, true).
				AddItem(nil, 0, 1, false),
			width, 0, true,
		).
		AddItem(nil, 0, 1, false)
}

// modalLayout 在根模态栈的全屏页面上居中显示内容。
func modalLayout(p tview.Primitive, width, height int) tview.Primitive {
	return centerBox(p, width, height)
}

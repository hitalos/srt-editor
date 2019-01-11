package application

import (
	"strings"

	"github.com/rivo/tview"
)

func modal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}

func notfound(done func()) tview.Primitive {
	form := tview.NewForm().AddButton("OK", done)
	form.SetBorder(true).SetTitle(" Not found! ")
	return modal(form, 20, 5)
}

func button(label string) *tview.Button {
	button := tview.NewButton(label)
	button.SetBorder(true)
	return button
}

func formatShortcuts(sc map[string]string) string {
	list := []string{}
	for key, value := range sc {
		list = append(list, "[yellow]"+key+"[white]: "+value)
	}
	return strings.Join(list, "\t")
}

package application

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/hitalos/srt-editor/types"
	"github.com/rivo/tview"
)

const timeFormat = "15:04:05.999"

// Application build UI to manipulate SRT files
type Application struct {
	UI          *tview.Application
	srt         *types.Srt
	defaultView tview.Primitive
}

var searchTerm string

func (app *Application) table() *tview.Table {
	table := tview.NewTable().SetSelectable(true, false)
	table.SetTitle(" " + app.srt.Filename + " ").SetBorder(true)

	columns := []string{"Num", "Start", "End", "Text"}
	for i, t := range columns {
		cell := tview.NewTableCell(t).SetTextColor(tcell.ColorYellow).SetAlign(1)
		cell.NotSelectable = true
		table.SetCell(0, i, cell)
	}

	dateCell := func(t time.Time, row, col int) {
		table.SetCell(row, col, tview.NewTableCell(t.Format(timeFormat)).SetTextColor(tcell.ColorBlue))
	}
	textCell := func(text string, row, col int) {
		text = strings.Replace(text, "\n", "↩️", -1)
		table.SetCell(row, col, tview.NewTableCell(text).SetTextColor(tcell.ColorGreen).SetExpansion(1))
	}
	for i, sub := range app.srt.Subtitles {
		table.SetCellSimple(i+1, 0, fmt.Sprint(sub.Num))
		dateCell(sub.Start, i+1, 1)
		dateCell(sub.End, i+1, 2)
		textCell(sub.Text, i+1, 3)
	}
	table.SetInputCapture(app.tableHandle(table))
	return table
}

func (app *Application) tableHandle(table *tview.Table) func(ev *tcell.EventKey) *tcell.EventKey {
	return func(ev *tcell.EventKey) *tcell.EventKey {
		rowCount := table.GetRowCount()
		switch ev.Name() {
		case "Rune[]]":
			table.Select(rowCount-1, 0)
		case "Rune[[]":
			table.Select(0, 0)
		case "Rune[+]":
			app.shift("+")
		case "Rune[-]":
			app.shift("-")
		case "Rune[/]":
			app.search(table)
		case "Rune[n]":
			app.searchNext(table)
		case "Backspace", "Backspace2", "Delete":
			app.toogleSub(table)
		case "Enter":
			app.editSub(table)
		}
		return ev
	}
}

func (app *Application) editSub(table *tview.Table) {
	row, _ := table.GetSelection()
	text := &app.srt.Subtitles[row-1].Text
	form := tview.NewForm()
	form.AddInputField("Text", strings.Replace(*text, "\n", "\\n", -1), 78, nil, nil).
		AddButton("OK", func() {
			*text = form.GetFormItemByLabel("Text").(*tview.InputField).GetText()
			*text = strings.Replace(*text, "\\n", "\n", -1)
			table.GetCell(row, 3).SetText(strings.Replace(*text, "\n", "↩️", -1))
			app.goToDefault()
		}).
		AddButton("cancel", app.goToDefault).SetTitle(" Edit Subtitle ").SetBorder(true)
	app.UI.SetRoot(modal(form, 80, 7), true)
}

func (app *Application) toogleSub(table *tview.Table) {
	row, _ := table.GetSelection()
	num := app.srt.Subtitles[row-1].Num
	color := tcell.ColorBlack
	app.srt.Subtitles[num-1].Delete = !app.srt.Subtitles[num-1].Delete
	if app.srt.Subtitles[num-1].Delete {
		color = tcell.ColorDarkRed
	}
	for c := 0; c < table.GetColumnCount(); c++ {
		table.GetCell(row, c).SetBackgroundColor(color)
	}
}

func (app *Application) goToDefault() {
	app.UI.SetRoot(app.defaultView, true)
}

func (app Application) shift(direction string) {
	form := tview.NewForm().AddInputField("Duration in milliseconds", "", 5, tview.InputFieldInteger, nil)
	form.AddButton("OK", func() {
		text := form.GetFormItem(0).(*tview.InputField).GetText()
		num, err := strconv.Atoi(direction + text)
		if err != nil {
			num = 0
		}
		app.srt.Shift(time.Duration(num) * time.Millisecond)
		app.goToDefault()
	}).AddButton("cancel", app.goToDefault)
	action := " Add"
	if direction == "-" {
		action = " Subtract"
	}
	form.SetBorder(true).SetTitle(action + " time ")
	app.UI.SetRoot(modal(form, 40, 7), true)
}

func (app Application) search(table *tview.Table) {
	text := tview.NewInputField().SetLabel("Text to search")
	text.SetDoneFunc(func(key tcell.Key) {
		searchTerm = text.GetText()
		line, err := app.srt.Search(searchTerm, 0)
		if err != nil {
			app.UI.SetRoot(notfound(app.goToDefault), true)
		}
		table.Select(line+1, 0)
		app.goToDefault()
	})
	text.SetBorder(true)
	app.UI.SetRoot(modal(text, 40, 3), true)
}

func (app Application) searchNext(table *tview.Table) {
	row, _ := table.GetSelection()
	line, err := app.srt.Search(searchTerm, row)
	if err != nil {
		app.UI.SetRoot(notfound(app.goToDefault), true)
	}
	table.Select(line+1, 0)
}

func (app *Application) saveAndExit() {
	if err := app.srt.Save("output.srt"); err != nil {
		app.UI.Stop()
		fmt.Println(err)
		os.Exit(1)
	}
	app.UI.Stop()
}

func (app *Application) btSave() *tview.Button {
	return button("Save").SetSelectedFunc(app.saveAndExit)
}

func (app *Application) btExit() *tview.Button {
	return button("Exit").SetSelectedFunc(func() { app.UI.Stop() })
}

func (app *Application) setNext(actual, next tview.Primitive) {
	if v, ok := actual.(*tview.Button); ok {
		v.SetBlurFunc(func(key tcell.Key) {
			app.UI.SetFocus(next)
		})
	} else if v, ok := actual.(*tview.Table); ok {
		v.SetDoneFunc(func(key tcell.Key) {
			app.UI.SetFocus(next)
		})
	}
}

// New returns a new Application
func New(srt *types.Srt) Application {
	app := Application{srt: srt}
	app.UI = tview.NewApplication()

	buttonExit := app.btExit()
	buttonSave := app.btSave()
	table := app.table()

	shortcutsView := tview.NewTextView().SetDynamicColors(true)
	list := formatShortcuts(map[string]string{
		"]": "Go to end",
		"[": "Go to start",
		"+": "Add time",
		"-": "Subtract time",
		"/": "Search",
		"n": "Search Next"})
	shortcutsView.SetText(list)

	app.setNext(table, buttonSave)
	app.setNext(buttonSave, buttonExit)
	app.setNext(buttonExit, table)

	app.defaultView = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true).
		AddItem(tview.NewFlex().
			AddItem(shortcutsView, 0, 4, false).
			AddItem(buttonSave, 0, 1, true).
			AddItem(buttonExit, 0, 1, true), 3, 0, true)
	app.goToDefault()
	return app
}

// TODO save as

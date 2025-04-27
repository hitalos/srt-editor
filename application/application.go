package application

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/text/encoding/charmap"
)

type Application struct {
	UI            *tview.Application
	srt           *Srt
	main          *tview.Flex
	table         *tview.Table
	menu          *tview.Flex
	searchTerm    string
	utf8Converted bool
}

func New(srt *Srt) *Application {
	app := &Application{
		UI:   tview.NewApplication().EnableMouse(true),
		srt:  srt,
		main: tview.NewFlex().SetDirection(tview.FlexColumnCSS),
	}
	app.setTable()
	app.setMenu()

	app.main.
		AddItem(app.table, 0, 1, true).
		AddItem(app.menu, 3, 1, false)

	app.goToDefault()

	return app
}

func (app *Application) Run() error {
	return app.UI.Run()
}

func (app *Application) setTable() {
	app.table = tview.NewTable().SetSelectable(true, false)

	app.table.SetTitle(" " + app.srt.Filename + " ").
		SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
			switch ev.Name() {
			case "Rune[+]":
				app.shift("+")
			case "Rune[-]":
				app.shift("-")
			case "Rune[/]":
				app.search()
			case "Rune[n]":
				app.searchNext()
			case "Backspace", "Backspace2", "Delete":
				app.toogleSub()
			case "Enter":
				app.editSub()
			case "Rune[u]":
				app.convert2UTF()
			}

			return ev
		})

	app.table.SetDoneFunc(func(_ tcell.Key) {
		app.UI.SetFocus(app.menu)
	})

	app.setTableLines()
}

func (app *Application) setColumNames() {
	for i, col := range []string{"Num", "Start", "End", "Text"} {
		cell := tview.NewTableCell(col).
			SetTextColor(tview.Styles.PrimaryTextColor).
			SetAlign(tview.AlignCenter)
		cell.NotSelectable = true

		app.table.SetCell(0, i, cell).SetBorder(true)
	}
}

func (app *Application) setTableLines() {
	app.table.Clear()
	app.setColumNames()

	for i, sub := range app.srt.Subtitles {
		num := fmt.Sprintf("%4d", sub.Num)
		text := strings.ReplaceAll(sub.Text, "\n", newLineSymbol)
		start := sub.Start.String()
		end := sub.End.String()

		app.table.SetCellSimple(i+1, 0, num)
		app.table.SetCell(i+1, 1, tview.NewTableCell(start).SetTextColor(tcell.ColorDarkCyan))
		app.table.SetCell(i+1, 2, tview.NewTableCell(end).SetTextColor(tcell.ColorDarkCyan))
		app.table.SetCell(i+1, 3, tview.NewTableCell(text).SetTextColor(tcell.ColorGreen).SetExpansion(1))
	}
}

func (app *Application) setMenu() {
	app.menu = tview.NewFlex().SetDirection(tview.FlexRowCSS)
	shortcutsView := tview.NewTextView().
		SetDynamicColors(true).
		SetText(shortcuts.String())

	btnSave := tview.NewButton("Save").SetSelectedFunc(app.save)
	btnExit := tview.NewButton("Exit").SetSelectedFunc(app.UI.Stop)

	btnSave.SetExitFunc(func(_ tcell.Key) { app.UI.SetFocus(btnExit) }).SetBorder(true)
	btnExit.SetExitFunc(func(_ tcell.Key) { app.UI.SetFocus(app.table) }).SetBorder(true)

	app.menu.
		AddItem(shortcutsView, 0, 4, false).
		AddItem(btnSave, 12, 1, true).
		AddItem(nil, 1, 0, false).
		AddItem(btnExit, 12, 1, false)
}

func (app *Application) save() {
	if err := app.srt.Save(app.srt.Filename); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	app.showMessage("File saved successfully!")

	if err := app.srt.Load(app.srt.Filename); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	app.setTableLines()
}

func (app *Application) convert2UTF() {
	if app.utf8Converted {
		return
	}

	dec := charmap.ISO8859_1.NewDecoder()
	for i, s := range app.srt.Subtitles {
		if s, err := dec.String(s.Text); err == nil {
			app.srt.Subtitles[i].Text = s
			continue
		}
		app.srt.Subtitles[i].Text = s.Text
	}

	app.setTableLines()
	app.utf8Converted = true

	app.showMessage("Subtitles converted to UTF-8!")
}

func (app *Application) shift(direction string) {
	label := "Duration in milliseconds"
	form := tview.NewForm().AddInputField(label, "", 5, tview.InputFieldInteger, nil)
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
	app.showModal(form, len(label)+10, 7)
}

func (app *Application) goToDefault() {
	app.UI.SetRoot(app.main, true)
}

func (app *Application) search() {
	text := tview.NewInputField().
		SetLabel("Text to search: ").
		SetText(app.searchTerm)

	text.SetDoneFunc(func(key tcell.Key) {
		app.searchTerm = text.GetText()
		line, err := app.srt.Search(app.searchTerm, 0)
		if err != nil {
			app.showMessage("Text not found!")
			return
		}
		app.table.Select(line+1, 0)
		app.goToDefault()
	}).SetBorder(true)

	app.showModal(text, 30, 3)
}

func (app *Application) searchNext() {
	row, _ := app.table.GetSelection()
	line, err := app.srt.Search(app.searchTerm, row)
	if err != nil {
		app.showMessage("Text not found!")
		return
	}
	app.table.Select(line+1, 0)
}

func (app *Application) editSub() {
	row, _ := app.table.GetSelection()
	text := &app.srt.Subtitles[row-1].Text
	form := tview.NewForm()

	form.AddInputField("Text", strings.ReplaceAll(*text, "\n", "\\n"), len(*text), nil, nil).
		AddButton("OK", func() {
			*text = form.GetFormItemByLabel("Text").(*tview.InputField).GetText()
			*text = strings.ReplaceAll(*text, "\\n", "\n")
			app.table.GetCell(row, 3).SetText(strings.ReplaceAll(*text, "\n", newLineSymbol))
			app.goToDefault()
		}).
		AddButton("cancel", app.goToDefault).
		AddTextView("Obs.:", "Use \"\\n\" to break lines", 0, 1, true, false).
		SetTitle(" Edit Subtitle ").
		SetBorder(true)

	app.showModal(form, len(*text)+10, 10)
}

func (app *Application) toogleSub() {
	row, _ := app.table.GetSelection()
	num := app.srt.Subtitles[row-1].Num
	color := tcell.ColorBlack
	app.srt.Subtitles[num-1].Delete = !app.srt.Subtitles[num-1].Delete
	if app.srt.Subtitles[num-1].Delete {
		color = tcell.ColorDarkRed
	}
	for c := 0; c < app.table.GetColumnCount(); c++ {
		app.table.GetCell(row, c).SetBackgroundColor(color)
	}
}

func (app *Application) showModal(p tview.Primitive, width, height int) {
	app.UI.SetRoot(
		tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, height, 1, true).
				AddItem(nil, 0, 1, false), width, 1, true).
			AddItem(nil, 0, 1, false), true)
}

func (app *Application) showMessage(message string) {
	app.UI.SetRoot(tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(btnIdx int, btnText string) {
			app.goToDefault()
		}), true)
}

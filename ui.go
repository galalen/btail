package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func appUI(tail Tail) {
	app := tview.NewApplication()

	header := tview.NewTextView().
		SetText("btail 🐝").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	table := tview.NewTable().
		SetBorders(false).
		SetFixed(1, 2).
		SetSelectable(true, false)

	table.SetCell(0, 0, tview.NewTableCell("Timestamp").
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetExpansion(1))
	table.SetCell(0, 1, tview.NewTableCell("Message").
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetExpansion(9))

	go func() {
		row := 1
		for line := range tail.Lines {
			table.SetCell(row, 0, tview.NewTableCell(line.Time.Format(time.RFC3339)).
				SetTextColor(tcell.ColorLightGoldenrodYellow).
				SetAlign(tview.AlignLeft).
				SetExpansion(1))

			table.SetCell(row, 1, tview.NewTableCell(line.Text).
				SetTextColor(tcell.ColorLimeGreen).
				SetAlign(tview.AlignLeft).
				SetExpansion(2))
			row++
		}
	}()

	mainContent := tview.NewFlex().
		AddItem(table, 0, 1, true)

	// Footer TextView
	footerInfo := tview.NewTextView()
	footerInfo.SetText("\tSearch (Ctrl+F) | Quit (Ctrl+Q) | Exit (ESC) | UP (↑) | Down (↓)")
	footerInfo.SetTextAlign(tview.AlignCenter)
	footerInfo.SetDynamicColors(true)
	footerInfo.SetBackgroundColor(tcell.ColorRebeccaPurple)

	footerInput := tview.NewInputField()
	footerInput.SetPlaceholder("(Ctrl+F) Search 🔍")
	footerInput.SetFieldTextColor(tcell.ColorWhite)
	footerInput.SetPlaceholderTextColor(tcell.ColorWhite)
	footerInput.SetFieldTextColor(tcell.ColorWhite)
	footerInput.SetBackgroundColor(tcell.ColorRebeccaPurple)

	footer := tview.NewFlex().
		AddItem(footerInput, 0, 5, true).
		AddItem(footerInfo, 0, 5, false)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(header, 1, 1, false).
		AddItem(mainContent, 0, 10, true).
		AddItem(footer, 1, 1, false)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlF:
			app.SetFocus(footerInput)
			return nil
		case tcell.KeyCtrlQ:
			app.Stop()
			return nil
		case tcell.KeyEsc:
			app.SetFocus(table)
			footerInput.SetText("")
			return nil
		default:
			// ignored
		}
		return event
	})

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}

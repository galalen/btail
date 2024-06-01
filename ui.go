package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func appUI(config Config) {
	app := tview.NewApplication()

	pattern := []string{"Timestamp", "Level", "Message"}
	expansions := []int{2, 2, 6}

	header := tview.NewTextView().
		SetText("btail 🐝").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	table := tview.NewTable().
		SetBorders(false).
		SetFixed(1, 2).
		SetSelectable(true, false)

	for i, text := range pattern {
		if i%2 == 0 {
			table.SetCell(0, i, tview.NewTableCell(text).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignLeft).
				SetExpansion(expansions[i]))
		} else {
			table.SetCell(0, i, tview.NewTableCell(text).
				SetTextColor(tcell.ColorRed).
				SetAlign(tview.AlignLeft).
				SetExpansion(expansions[i]))
		}
	}

	res, err := Tail(config)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	go func() {
		row := 1
		for line := range res.Lines {
			msg := strings.Split(line, "|")
			level := strings.TrimSpace(msg[0])
			message := strings.TrimSpace(msg[1])

			table.SetCell(row, 0, tview.NewTableCell(time.Now().Format(time.RFC3339)).
				SetTextColor(tcell.ColorGreenYellow).
				SetAlign(tview.AlignLeft).
				SetExpansion(2))

			table.SetCell(row, 1, tview.NewTableCell(level).
				SetTextColor(tcell.ColorRed).
				SetAlign(tview.AlignLeft).
				SetExpansion(2))

			table.SetCell(row, 2, tview.NewTableCell(message).
				SetTextColor(tcell.ColorGreenYellow).
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

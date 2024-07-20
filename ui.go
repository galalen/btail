package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func setupColumns(table *tview.Table) {
	table.SetCell(0, 0, tview.NewTableCell("No.").
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetStyle(tcell.StyleDefault.Bold(true)).
		SetExpansion(1))

	table.SetCell(0, 1, tview.NewTableCell("Timestamp").
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetStyle(tcell.StyleDefault.Bold(true)).
		SetExpansion(1))

	table.SetCell(0, 2, tview.NewTableCell("Message").
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetStyle(tcell.StyleDefault.Bold(true)).
		SetExpansion(9))
}

func clearTable(table *tview.Table) {
	table.Clear()
	setupColumns(table)
}

func renderRow(table *tview.Table, row int, line Line, newTextMsg string) {
	table.SetCell(row, 0, tview.NewTableCell(strconv.Itoa(row)).
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetExpansion(1))

	table.SetCell(row, 1, tview.NewTableCell(line.Time.Format(time.RFC3339)).
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetExpansion(1))

	if newTextMsg != "" {
		table.SetCell(row, 2, tview.NewTableCell(newTextMsg).
			SetTextColor(tcell.ColorLightGreen).
			SetAlign(tview.AlignLeft).
			SetExpansion(2))
	} else {
		table.SetCell(row, 2, tview.NewTableCell(highlightPatterns(line.Text)).
			SetTextColor(tcell.ColorLightGreen).
			SetAlign(tview.AlignLeft).
			SetExpansion(2))
	}
}

func appUI(tail Tail) {
	app := tview.NewApplication()

	var bufferedLines []Line
	var isSearchUsed bool

	table := tview.NewTable().
		SetBorders(false).
		SetFixed(1, 3).
		SetSelectable(true, false).
		ScrollToEnd()

	setupColumns(table)

	showBufferedLines := func(keyword string) {
		clearTable(table)
		row := table.GetRowCount()
		if len(keyword) > 0 {
			for _, line := range bufferedLines {
				if strings.Contains(line.Text, keyword) {
					highlightedText := highlightKeyword(line.Text, keyword)
					renderRow(table, row, line, highlightedText)
					row++
				}
			}
			isSearchUsed = true
		} else {
			// TODO: take the count when search used,
			// 	consider tail.Lines b/c data will be appended to both ch and bufferedLines
			for _, line := range bufferedLines {
				renderRow(table, row, line, "")
				row++
			}
			isSearchUsed = false
		}
		table.ScrollToEnd()
		table.Select(row, 1)
	}

	// main populate
	go func() {
		row := table.GetRowCount()
		for line := range tail.Lines {
			bufferedLines = append(bufferedLines, line)

			renderRow(table, row, line, "")
			row++

			app.QueueUpdateDraw(func() {
				_, _, _, height := table.GetInnerRect()
				lastVisibleRow := row - 1
				if lastVisibleRow > height {
					table.SetOffset(row, lastVisibleRow-height+1)
				}
				table.Select(lastVisibleRow, 1)
			})
		}
	}()

	header := tview.NewTextView().
		SetText("btail 🐝").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

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
	footerInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			keyword := footerInput.GetText()
			if len(keyword) > 0 {
				showBufferedLines(keyword)
			}
		}
	})

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
			if isSearchUsed {
				showBufferedLines("")
			}
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

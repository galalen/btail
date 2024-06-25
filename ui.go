package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func appUI(tail Tail) {
	app := tview.NewApplication()

	var bufferedLines []Line
	var isSearchUsed bool

	header := tview.NewTextView().
		SetText("btail 🐝").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	table := tview.NewTable().
		SetBorders(false).
		SetFixed(1, 3).
		SetSelectable(true, false).
		ScrollToEnd()

	table.SetCell(0, 0, tview.NewTableCell("No.").
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetExpansion(1))

	table.SetCell(0, 1, tview.NewTableCell("Timestamp").
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetExpansion(1))

	table.SetCell(0, 2, tview.NewTableCell("Message").
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetExpansion(9))

	clearTable := func() {
		table.Clear()

		table.SetCell(0, 0, tview.NewTableCell("No.").
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetExpansion(1))

		table.SetCell(0, 1, tview.NewTableCell("Timestamp").
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetExpansion(1))

		table.SetCell(0, 2, tview.NewTableCell("Message").
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetExpansion(9))

	}

	highlightKeyword := func(text, keyword string) string {
		return strings.ReplaceAll(text, keyword, fmt.Sprintf("[red]%s[#32CD32]", keyword))
	}

	showBufferedLines := func(keyword string) {
		clearTable()
		row := table.GetRowCount()
		if len(keyword) > 0 {
			for _, line := range bufferedLines {
				if strings.Contains(line.Text, keyword) {
					highlightedText := highlightKeyword(line.Text, keyword)

					table.SetCell(row, 0, tview.NewTableCell(strconv.Itoa(row)).
						SetTextColor(tcell.ColorWhite).
						SetAlign(tview.AlignLeft).
						SetExpansion(1))

					table.SetCell(row, 1, tview.NewTableCell(line.Time.Format(time.RFC3339)).
						SetTextColor(tcell.ColorLightGoldenrodYellow).
						SetAlign(tview.AlignLeft).
						SetExpansion(1))

					table.SetCell(row, 2, tview.NewTableCell(highlightedText).
						SetTextColor(tcell.ColorLimeGreen).
						SetAlign(tview.AlignLeft).
						SetExpansion(2))
					row++
				}
			}
			isSearchUsed = true
		} else {
			// TODO: take the count when search used,
			// 	consider tail.Lines b/c data will be appended to both ch and bufferedLines
			for _, line := range bufferedLines {
				table.SetCell(row, 0, tview.NewTableCell(strconv.Itoa(row)).
					SetTextColor(tcell.ColorWhite).
					SetAlign(tview.AlignLeft).
					SetExpansion(1))

				table.SetCell(row, 1, tview.NewTableCell(line.Time.Format(time.RFC3339)).
					SetTextColor(tcell.ColorLightGoldenrodYellow).
					SetAlign(tview.AlignLeft).
					SetExpansion(1))

				table.SetCell(row, 2, tview.NewTableCell(highlightPatterns(line.Text)).
					SetTextColor(tcell.ColorLimeGreen).
					SetAlign(tview.AlignLeft).
					SetExpansion(2))
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

			table.SetCell(row, 0, tview.NewTableCell(strconv.Itoa(row)).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft).
				SetExpansion(1))

			table.SetCell(row, 1, tview.NewTableCell(line.Time.Format(time.RFC3339)).
				SetTextColor(tcell.ColorLightGoldenrodYellow).
				SetAlign(tview.AlignLeft).
				SetExpansion(1))

			table.SetCell(row, 2, tview.NewTableCell(highlightPatterns(line.Text)).
				SetTextColor(tcell.ColorLimeGreen).
				SetAlign(tview.AlignLeft).
				SetExpansion(2))
			table.ScrollToEnd()
			table.Select(row, 1)
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

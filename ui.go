package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type BtailApp struct {
	app           *tview.Application
	table         *tview.Table
	searchInput   *tview.InputField
	tail          *Tail
	bufferedLines []Line
	isSearchUsed  bool
}

func NewBtailApp(tail *Tail) *BtailApp {
	table := tview.NewTable().
		SetBorders(false).
		SetFixed(1, 3).
		SetSelectable(true, false).
		ScrollToEnd()

	return &BtailApp{
		app:         tview.NewApplication(),
		table:       table,
		searchInput: tview.NewInputField(),
		tail:        tail,
	}
}

func (b *BtailApp) setupColumns() {
	b.table.SetCell(0, 0, tview.NewTableCell("No.").
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetStyle(tcell.StyleDefault.Bold(true)).
		SetExpansion(1))

	b.table.SetCell(0, 1, tview.NewTableCell("Timestamp").
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetStyle(tcell.StyleDefault.Bold(true)).
		SetExpansion(1))

	b.table.SetCell(0, 2, tview.NewTableCell("Message").
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetStyle(tcell.StyleDefault.Bold(true)).
		SetExpansion(9))
}

func (b *BtailApp) clearTable() {
	b.table.Clear()
	b.setupColumns()
}

func (b *BtailApp) renderRow(row int, line Line, newTextMsg string) {
	b.table.SetCell(row, 0, tview.NewTableCell(strconv.Itoa(row)).
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetExpansion(1))

	b.table.SetCell(row, 1, tview.NewTableCell(line.Time.Format(time.RFC3339)).
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetExpansion(1))

	if newTextMsg != "" {
		b.table.SetCell(row, 2, tview.NewTableCell(newTextMsg).
			SetTextColor(tcell.ColorLightGreen).
			SetAlign(tview.AlignLeft).
			SetExpansion(2))
	} else {
		b.table.SetCell(row, 2, tview.NewTableCell(highlightPatterns(line.Text)).
			SetTextColor(tcell.ColorLightGreen).
			SetAlign(tview.AlignLeft).
			SetExpansion(2))
	}
}

func (b *BtailApp) showBufferedLines(keyword string) {
	b.clearTable()
	row := b.table.GetRowCount()
	if len(keyword) > 0 {
		b.isSearchUsed = true

		for _, line := range b.bufferedLines {
			if strings.Contains(line.Text, keyword) {
				highlightedText := highlightKeyword(line.Text, keyword)
				b.renderRow(row, line, highlightedText)
				row++
			}
		}
	} else {
		b.isSearchUsed = false
		for _, line := range b.bufferedLines {
			b.renderRow(row, line, "")
			row++
		}
	}
	b.table.ScrollToEnd()
}

func (b *BtailApp) Run() {
	b.setupColumns()

	// main populate
	go func() {
		row := b.table.GetRowCount()
		for line := range b.tail.Lines {
			b.bufferedLines = append(b.bufferedLines, line)

			b.renderRow(row, line, "")
			row++

			b.app.QueueUpdateDraw(func() {
				_, _, _, height := b.table.GetInnerRect()
				lastVisibleRow := row - 1
				if lastVisibleRow > height {
					b.table.SetOffset(row, lastVisibleRow-height+1)
				}
			})
		}
	}()

	header := tview.NewTextView().
		SetText("btail 🐝").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	mainContent := tview.NewFlex().
		AddItem(b.table, 0, 1, true)

	// Footer TextView
	footerInfo := tview.NewTextView()
	footerInfo.SetText("Quit (Ctrl+Q)")
	footerInfo.SetTextAlign(tview.AlignCenter)
	footerInfo.SetDynamicColors(true)
	footerInfo.SetBackgroundColor(tcell.ColorGray)

	b.searchInput.SetPlaceholder("(Ctrl+F) Search 🔍")
	b.searchInput.SetFieldTextColor(tcell.ColorWhite)
	b.searchInput.SetPlaceholderTextColor(tcell.ColorWhite)
	b.searchInput.SetFieldTextColor(tcell.ColorWhite)
	b.searchInput.SetBackgroundColor(tcell.ColorDimGray)
	b.searchInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			keyword := b.searchInput.GetText()
			if len(keyword) > 0 {
				b.showBufferedLines(keyword)
			}
		}
	})

	footer := tview.NewFlex().
		AddItem(b.searchInput, 0, 8, true).
		AddItem(footerInfo, 0, 2, false)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(header, 1, 1, false).
		AddItem(mainContent, 0, 10, true).
		AddItem(footer, 1, 1, false)

	b.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlF:
			b.app.SetFocus(b.searchInput)
			b.searchInput.SetPlaceholder("...")
			return nil
		case tcell.KeyCtrlQ:
			b.app.Stop()
			return nil
		case tcell.KeyEsc:
			b.searchInput.SetPlaceholder("(Ctrl+F) Search 🔍")
			b.app.SetFocus(b.table)
			b.searchInput.SetText("")
			if b.isSearchUsed {
				b.showBufferedLines("")
			}
			return nil
		default:
			// ignored
		}
		return event
	})

	if err := b.app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}

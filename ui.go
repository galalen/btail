package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	tail          *Tail
	logsView      viewport.Model
	bufferedLines []Line
	width         int
	height        int
	searchInput   textinput.Model
	searching     bool
	searchTerm    string
	matchCount    int
	autoScroll    bool
	lastScrollPos int
}

const bufferedLinesCount = 500

func initialModel(tail *Tail) *model {
	lv := viewport.New(80, 20)
	lv.Style = baseStyle

	ti := textinput.New()
	ti.Placeholder = "Search..."

	return &model{
		tail:        tail,
		logsView:    lv,
		searchInput: ti,
		autoScroll:  true,
	}
}

func (m *model) Init() tea.Cmd {
	return m.tailFile()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			if !m.searching {
				return m, tea.Quit
			}
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+f":
			m.searching = !m.searching
			if m.searching {
				m.searchInput.Focus()
			} else {
				m.searchInput.Blur()
			}
			m.searchTerm = ""
			m.matchCount = 0
			m.updateLogsView()
		case "esc":
			if m.searching {
				m.searching = false
				m.searchInput.SetValue("")
				m.searchInput.Blur()
				m.searchTerm = ""
				m.matchCount = 0
				m.updateLogsView()
			}
		case "up":
			m.autoScroll = false
			m.logsView.LineUp(1)
		case "down":
			m.logsView.LineDown(1)
			if m.logsView.AtBottom() {
				m.autoScroll = true
			}
		case "home":
			m.autoScroll = false
			m.logsView.GotoTop()
		case "end":
			m.autoScroll = true
			m.logsView.GotoBottom()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.logsView.Width = m.width - 4
		m.logsView.Height = m.height - 6
		m.searchInput.Width = m.width / 3
		return m, nil
	case Line:
		m.bufferedLines = append(m.bufferedLines, msg)
		if len(m.bufferedLines) > bufferedLinesCount {
			m.bufferedLines = m.bufferedLines[1:]
		}

		m.updateLogsView()
		return m, m.tailFile()
	case nil:
		if m.tail.Config.Follow {
			return m, m.tailFile()
		}
		return m, nil
	}

	if m.searching {
		m.searchInput, cmd = m.searchInput.Update(msg)
		m.searchTerm = m.searchInput.Value()
		m.updateLogsView()
		return m, cmd
	}

	// Check if the viewport has been scrolled
	if m.logsView.YOffset != m.lastScrollPos {
		if m.logsView.AtBottom() {
			m.autoScroll = true
		} else {
			m.autoScroll = false
		}
		m.lastScrollPos = m.logsView.YOffset
	}

	m.logsView, cmd = m.logsView.Update(msg)
	return m, cmd
}

func (m *model) updateLogsView() {
	var content strings.Builder
	m.matchCount = 0

	for _, msg := range m.bufferedLines {
		highlightedContent := highlightPatterns(msg.Text)

		if m.searchTerm != "" {
			count := strings.Count(strings.ToLower(msg.Text), strings.ToLower(m.searchTerm))
			m.matchCount += count
			if count > 0 {
				highlightedContent = highlightSearch(highlightedContent, m.searchTerm)
			}
		}

		t := timeStyle.Render(msg.Time.Format("03:04:05 PM"))
		content.WriteString(fmt.Sprintf("[%s]\n%s\n\n", t, highlightedContent))
	}

	m.logsView.SetContent(content.String())
	if m.autoScroll {
		m.logsView.GotoBottom()
	}
}

func (m *model) View() string {
	title := titleStyle.Render("btail 🐝")

	var statusBar string
	if m.searching {
		searchInput := searchInputStyle.Render(m.searchInput.View())
		bufferInfo := fmt.Sprintf("buffer: %d/%d", len(m.bufferedLines), bufferedLinesCount)
		statusMessage := statusMessageStyle.Render(fmt.Sprintf("matches: %d | %s | esc: cancel", m.matchCount, bufferInfo))
		statusBar = lipgloss.JoinHorizontal(lipgloss.Left, searchInput, statusMessage)
	} else {
		statusBar = statusBarStyle.Render("\tq: quit | ctrl+f: search\t")
	}

	statusBar = lipgloss.NewStyle().Render(statusBar)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		m.logsView.View(),
		statusBar,
	)
}

func (m *model) tailFile() tea.Cmd {
	return func() tea.Msg {
		select {
		case line, ok := <-m.tail.Lines:
			if !ok {
				return nil
			}
			return line
		case <-time.After(100 * time.Millisecond):
			if m.tail.Config.Follow {
				return m.tailFile()()
			}
			return nil
		}
	}
}

func runBtailApp(tail *Tail) {
	p := tea.NewProgram(initialModel(tail), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

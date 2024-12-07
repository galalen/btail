package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/galalen/btail/pkg/tail"
)

type State int

const (
	StateNormal State = iota
	StateSearching
)

type Model struct {
	tailer        *tail.Tail
	logsView      viewport.Model
	searchInput   textinput.Model
	bufferedLines []tail.Line
	state         State
	width         int
	height        int
	searchTerm    string
	matchCount    int
	autoScroll    bool
	lastScrollPos int
}

func NewModel(tailer *tail.Tail) *Model {
	vp := viewport.New(80, 20)
	vp.Style = baseStyle

	ti := textinput.New()
	ti.Placeholder = "Search..."

	return &Model{
		tailer:        tailer,
		logsView:      vp,
		searchInput:   ti,
		state:         StateNormal,
		autoScroll:    true,
		bufferedLines: make([]tail.Line, 0, tailer.Config.UIBufferSize),
	}
}

func (m *Model) Init() tea.Cmd {
	return m.tailFile()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			if m.state == StateNormal {
				return m, tea.Quit
			}
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+f":
			return m.toggleSearch()
		case "esc":
			return m.exitSearch()
		case "up":
			m.scrollUp()
		case "down":
			m.scrollDown()
		case "home":
			m.scrollToTop()
		case "end":
			m.scrollToBottom()
		}
	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)
	case tail.Line:
		return m.handleNewLine(msg)
	case error:
		// TODO: show error in status bar
		return m, m.tailFile()
	}

	if m.state == StateSearching {
		m.searchInput, cmd = m.searchInput.Update(msg)
		m.searchTerm = m.searchInput.Value()
		m.updateContent()
		return m, cmd
	}

	m.updateScrollState()
	m.logsView, cmd = m.logsView.Update(msg)
	return m, cmd
}

func (m *Model) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.logsView.Width = msg.Width - 4
	m.logsView.Height = msg.Height - 6
	m.searchInput.Width = msg.Width / 3
	return m, nil
}

func (m *Model) handleNewLine(line tail.Line) (tea.Model, tea.Cmd) {
	m.bufferedLines = append(m.bufferedLines, line)
	if len(m.bufferedLines) > m.tailer.Config.UIBufferSize {
		m.bufferedLines = m.bufferedLines[1:]
	}
	m.updateContent()
	return m, m.tailFile()
}

func (m *Model) View() string {
	title := titleStyle.Render("btail 🐝")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		m.logsView.View(),
		m.renderStatusBar(),
	)
}

func (m *Model) renderStatusBar() string {
	if m.state == StateSearching {
		return m.renderSearchBar()
	}
	return statusBarStyle.Render("\tq: quit | ctrl+f: search\t")
}

func (m *Model) renderSearchBar() string {
	searchInput := searchInputStyle.Render(m.searchInput.View())
	bufferInfo := fmt.Sprintf("buffer: %d/%d", len(m.bufferedLines), m.tailer.Config.UIBufferSize)
	statusMessage := statusMessageStyle.Render(
		fmt.Sprintf("matches: %d | %s | esc: cancel", m.matchCount, bufferInfo),
	)
	return lipgloss.JoinHorizontal(lipgloss.Left, searchInput, statusMessage)
}

func (m *Model) updateContent() {
	var content strings.Builder
	m.matchCount = 0

	for _, line := range m.bufferedLines {
		highlightedContent := highlightPatterns(line.Text)

		if m.searchTerm != "" {
			count := strings.Count(
				strings.ToLower(line.Text),
				strings.ToLower(m.searchTerm),
			)
			m.matchCount += count
			if count > 0 {
				highlightedContent = highlightSearch(highlightedContent, m.searchTerm)
			}
		}

		timestamp := fmt.Sprintf("%s%s%s",
			bracketsStyle.Render("["),
			timeStyle.Render(line.Time.Format("03:04:05 PM")),
			bracketsStyle.Render("]"),
		)
		content.WriteString(fmt.Sprintf("%s %s\n", timestamp, highlightedContent))
	}

	m.logsView.SetContent(content.String())
	if m.autoScroll {
		m.logsView.GotoBottom()
	}
}

func (m *Model) scrollUp() {
	m.autoScroll = false
	m.logsView.LineUp(1)
}

func (m *Model) scrollDown() {
	m.logsView.LineDown(1)
	if m.logsView.AtBottom() {
		m.autoScroll = true
	}
}

func (m *Model) scrollToTop() {
	m.autoScroll = false
	m.logsView.GotoTop()
}

func (m *Model) scrollToBottom() {
	m.autoScroll = true
	m.logsView.GotoBottom()
}

func (m *Model) toggleSearch() (tea.Model, tea.Cmd) {
	if m.state == StateSearching {
		return m.exitSearch()
	}

	m.state = StateSearching
	m.searchInput.Focus()
	m.searchInput.SetValue("")
	m.searchTerm = ""
	m.matchCount = 0
	m.updateContent()
	return m, textinput.Blink
}

func (m *Model) exitSearch() (tea.Model, tea.Cmd) {
	if m.state == StateSearching {
		m.state = StateNormal
		m.searchInput.Blur()
		m.searchInput.SetValue("")
		m.searchTerm = ""
		m.matchCount = 0
		m.updateContent()
	}
	return m, nil
}

func (m *Model) tailFile() tea.Cmd {
	return func() tea.Msg {
		select {
		case line, ok := <-m.tailer.Lines:
			if !ok {
				return nil
			}
			return line
		case <-time.After(100 * time.Millisecond):
			if m.tailer.Config.Follow {
				return m.tailFile()()
			}
			return nil
		}
	}
}

func (m *Model) updateScrollState() {
	if m.logsView.YOffset != m.lastScrollPos {
		m.autoScroll = m.logsView.AtBottom()
		m.lastScrollPos = m.logsView.YOffset
	}
}

func Run(tailer *tail.Tail) error {
	p := tea.NewProgram(NewModel(tailer), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

package app

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/galalen/btail/pkg/tail"
)

type State int
type SearchMode string

const (
	StateNormal State = iota
	StateSearching
	SearchModeNormal SearchMode = "normal"
	SearchModeRegex  SearchMode = "regex"
	viewportWidth               = 80
	viewportHeight              = 20
)

type Model struct {
	tailer        *tail.Tail
	logsView      viewport.Model
	searchInput   textinput.Model
	bufferedLines []tail.Line
	state         State
	searchMode    SearchMode
	compiledRegex *regexp.Regexp
	searchErr     error
	searchTerm    string
	width         int
	height        int
	matchCount    int
	lastScrollPos int
	autoScroll    bool
	wrapLines     bool
}

func NewModel(tailer *tail.Tail) *Model {
	vp := viewport.New(viewportWidth, viewportHeight)
	vp.Style = baseStyle
	vp.SetHorizontalStep(4)

	ti := textinput.New()
	ti.Placeholder = "search..."

	return &Model{
		tailer:        tailer,
		logsView:      vp,
		searchInput:   ti,
		state:         StateNormal,
		searchMode:    SearchModeNormal,
		autoScroll:    true,
		wrapLines:     true,
		bufferedLines: make([]tail.Line, 0, tailer.Config.UIBufferSize),
	}
}

func (m *Model) Init() tea.Cmd {
	return m.tailFile()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

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
			m.searchInput.Placeholder = "search..."
			cmd := m.toggleSearch(SearchModeNormal)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case "ctrl+r":
			m.searchInput.Placeholder = "pattern..."
			cmd := m.toggleSearch(SearchModeRegex)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case "esc":
			cmd := m.exitSearch()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case "w":
			if m.state != StateSearching {
				m.wrapLines = !m.wrapLines
				if m.wrapLines {
					m.logsView.SetXOffset(0)
				}
				m.updateContent()
			}
		case "up":
			m.scrollUp()
		case "down":
			m.scrollDown()
		case "left":
			if !m.wrapLines {
				m.logsView.ScrollLeft(4)
			}
		case "right":
			if !m.wrapLines {
				m.logsView.ScrollRight(4)
			}
		case "home":
			m.scrollToTop()
		case "end":
			m.scrollToBottom()
		case "c":
			if m.state != StateSearching {
				m.clearBuffer()
			}
		}
	case tea.WindowSizeMsg:
		m.handleWindowResize(msg)
	case tail.Line:
		m.handleNewLine(msg)
		cmds = append(cmds, m.tailFile())
	case error:
		// TODO: show error in status bar
		cmds = append(cmds, m.tailFile())
	}

	if m.state == StateSearching {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.searchTerm = m.searchInput.Value()
		if m.searchMode == SearchModeRegex {
			if m.searchTerm != "" {
				var err error
				m.compiledRegex, err = regexp.Compile(m.searchTerm)
				m.searchErr = err
				if err != nil {
					m.compiledRegex = nil
				}
			} else {
				m.searchErr = nil
			}
		} else {
			m.compiledRegex = regexp.MustCompile(`(?i)` + regexp.QuoteMeta(m.searchTerm))
			m.searchErr = nil
		}
		m.updateContent()
	}

	m.updateScrollState()
	var cmd tea.Cmd
	m.logsView, cmd = m.logsView.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) clearBuffer() {
	m.bufferedLines = make([]tail.Line, 0, m.tailer.Config.UIBufferSize)
	m.updateContent()
}

func (m *Model) handleWindowResize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
	m.logsView.Width = msg.Width
	m.logsView.Height = msg.Height - 6
	m.searchInput.Width = msg.Width / 3
}

func (m *Model) handleNewLine(line tail.Line) {
	m.bufferedLines = append(m.bufferedLines, line)
	if len(m.bufferedLines) > m.tailer.Config.UIBufferSize {
		m.bufferedLines = m.bufferedLines[1:]
	}
	m.updateContent()
}

func (m *Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render("btail 🐝"),
		m.logsView.View(),
		m.renderStatusBar(),
	)
}

func (m *Model) renderBufferInfo() string {
	return fmt.Sprintf("buffer: %d/%d", len(m.bufferedLines), m.tailer.Config.UIBufferSize)
}

func (m *Model) renderStatusBar() string {
	if m.state == StateSearching {
		return m.renderSearchBar()
	}
	wrapMode := "wrap"
	if !m.wrapLines {
		wrapMode = "hscroll"
	}
	return statusBarStyle.Render(fmt.Sprintf("\t%s | %s | ctrl+f: search | ctrl+r: regex | w: wrap | left/right: scroll | c: clear | q: quit\t", m.renderBufferInfo(), wrapMode))
}

func (m *Model) renderSearchMode() string {
	return searchModeStyle.Render(fmt.Sprintf("mode: %s", m.searchMode))
}

func (m *Model) renderSearchBar() string {
	errText := ""
	if m.searchErr != nil {
		errText = fmt.Sprintf(" | %s", m.searchErr.Error())
	}
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		searchInputStyle.Render(m.searchInput.View()),
		pinkStyle.Render(fmt.Sprintf(" %d matches | ", m.matchCount)),
		m.renderSearchMode(),
		pinkStyle.Render(fmt.Sprintf(" | %s | esc: cancel%s", m.renderBufferInfo(), errText)),
	)
}

func (m *Model) highlightSearch(content string) string {
	if m.compiledRegex == nil {
		return content
	}

	return m.compiledRegex.ReplaceAllStringFunc(content, func(match string) string {
		return searchMatchStyle.Render(match)
	})
}

func (m *Model) updateContent() {
	var content strings.Builder
	m.matchCount = 0

	for _, line := range m.bufferedLines {
		highlightedContent := highlightPatterns(line.Text)

		if m.searchMode == SearchModeNormal {
			if m.searchTerm != "" {
				count := strings.Count(strings.ToLower(line.Text), strings.ToLower(m.searchTerm))
				if count > 0 {
					m.matchCount += count
					highlightedContent = m.highlightSearch(line.Text)
				}
			}
		} else {
			if m.compiledRegex != nil {
				count := len(m.compiledRegex.FindAllStringIndex(line.Text, -1))
				if count > 0 {
					m.matchCount += count
					highlightedContent = m.highlightSearch(line.Text)
				}
			}
		}

		timestamp := fmt.Sprintf("%s%s%s",
			bracketsStyle.Render("["),
			timeStyle.Render(line.Time.Format("03:04:05 PM")),
			bracketsStyle.Render("]"),
		)
		rawLine := fmt.Sprintf("%s %s", timestamp, highlightedContent)
		logLine := rawLine
		if m.wrapLines && m.logsView.Width > 0 {
			logLine = lipgloss.NewStyle().Width(m.logsView.Width).Render(rawLine)
		}
		content.WriteString(logLine + "\n")
	}

	m.logsView.SetContent(content.String())
	if m.autoScroll {
		m.logsView.GotoBottom()
	}
}

func (m *Model) scrollUp() {
	m.autoScroll = false
	m.logsView.ScrollUp(1)
}

func (m *Model) scrollDown() {
	m.logsView.ScrollDown(1)
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

func (m *Model) toggleSearch(searchMode SearchMode) tea.Cmd {
	if m.state == StateSearching {
		return m.exitSearch()
	}

	m.state = StateSearching
	m.searchMode = searchMode
	m.searchInput.Focus()
	m.searchInput.SetValue("")
	m.searchTerm = ""
	m.compiledRegex = nil
	m.searchErr = nil
	m.matchCount = 0
	m.updateContent()
	return textinput.Blink
}

func (m *Model) exitSearch() tea.Cmd {
	if m.state == StateSearching {
		m.state = StateNormal
		m.searchInput.Blur()
		m.searchInput.SetValue("")
		m.searchTerm = ""
		m.compiledRegex = nil
		m.searchErr = nil
		m.matchCount = 0
		m.updateContent()
	}
	return nil
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

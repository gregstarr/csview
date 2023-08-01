package table

import (
	"csview/utils"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

// types

type Column struct {
	name  string
	width int
	hide  bool
}

type KeyMap struct {
	Quit         key.Binding
	Up           key.Binding
	Down         key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	Left         key.Binding
	Right        key.Binding
	JumpLeft     key.Binding
	JumpRight    key.Binding
	Hide         key.Binding
	UnHideAll    key.Binding
	IncUnselFore key.Binding
	IncUnselBack key.Binding
	IncSelBack   key.Binding
	IncSelFore   key.Binding
	IncPadding   key.Binding
	DecPadding   key.Binding
}

type Cell struct {
	row, col int
}

type Styles struct {
	padding                   int
	unselectedForegroundColor uint
	unselectedBackgroundColor uint
	SelectedBackgroundColor   uint
	SelectedForegroundColor   uint
}

type Model struct {
	cell       Cell
	records    [][]string
	columns    []Column
	startRow   int
	stopRow    int
	startCol   int
	stopCol    int
	termWidth  int
	termHeight int
	KeyMap     KeyMap
	styles     Styles
}

func DefaultStyles() Styles {
	return makeStyles(1, 2, 0, 15, 0)
}

func makeStyles(padding int, unselFore, unselBack, selFore, selBack uint) Styles {
	return Styles{
		padding:                   padding,
		unselectedForegroundColor: unselFore,
		unselectedBackgroundColor: unselBack,
		SelectedForegroundColor:   selFore,
		SelectedBackgroundColor:   selBack,
	}
}

func (s Styles) Normal() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, s.padding).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		Foreground(lipgloss.ANSIColor(s.unselectedForegroundColor)).
		Background(lipgloss.ANSIColor(s.unselectedBackgroundColor)).
		BorderForeground(lipgloss.ANSIColor(s.unselectedForegroundColor)).
		BorderBackground(lipgloss.ANSIColor(s.unselectedBackgroundColor))
}

func (s Styles) Selected(style lipgloss.Style) lipgloss.Style {
	return style.
		Foreground(lipgloss.ANSIColor(s.SelectedForegroundColor)).
		Background(lipgloss.ANSIColor(s.SelectedBackgroundColor))
}

func (s Styles) Hidden(style lipgloss.Style) lipgloss.Style {
	return style.Padding(0, 0)
}

func (s Styles) Header(style lipgloss.Style) lipgloss.Style {
	return style.Border(lipgloss.NormalBorder(), false, true, true, false)
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl-c"),
			key.WithHelp("q/esc/ctrl-c", "quit"),
		),
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("left", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("right", "left"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "page down"),
		),
		JumpLeft: key.NewBinding(
			key.WithKeys("shift+left"),
			key.WithHelp("shift+left", "5 left"),
		),
		JumpRight: key.NewBinding(
			key.WithKeys("shift+right"),
			key.WithHelp("shift+right", "5 right"),
		),
		Hide: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "hide"),
		),
		UnHideAll: key.NewBinding(
			key.WithKeys("O"),
			key.WithHelp("O", "unhide all"),
		),
		IncUnselFore: key.NewBinding(key.WithKeys(",")),
		IncUnselBack: key.NewBinding(key.WithKeys("<")),
		IncSelBack:   key.NewBinding(key.WithKeys(">")),
		IncSelFore:   key.NewBinding(key.WithKeys(".")),
		IncPadding:   key.NewBinding(key.WithKeys("/")),
		DecPadding:   key.NewBinding(key.WithKeys("?")),
	}
}

func New(records [][]string) *Model {
	columns := make([]Column, len(records[0]))
	for i, col := range records[0] {
		columns[i] = Column{col, 1, false}
		for j := 0; j < len(records); j++ {
			if columns[i].width < len(records[j][i]) {
				columns[i].width = len(records[j][i])
			}
		}
	}

	m := Model{
		records:    records,
		startRow:   1,
		stopRow:    20,
		startCol:   1,
		stopCol:    8,
		columns:    columns,
		cell:       Cell{0, 0},
		termHeight: 99999,
		termWidth:  99999,
		KeyMap:     DefaultKeyMap(),
		styles:     DefaultStyles(),
	}
	return &m
}

// methods

func (m Model) getStyle(row, col int) lipgloss.Style {
	style := m.styles.Normal()
	if m.columns[col].hide {
		style = m.styles.Hidden(style)
	}
	if row == 0 {
		style = m.styles.Header(style)
	}
	if (row == 0 && m.cell.col == col) || (col == 0 && m.cell.row == row) || (m.cell.row == row && m.cell.col == col) {
		style = m.styles.Selected(style)
	}
	return style.Width(m.getColumnWidth(col) - 1)
}

func (m Model) renderCell(row, col int) string {
	if m.columns[col].hide {
		return m.getStyle(row, col).Render("-")
	} else {
		return m.getStyle(row, col).Render(m.records[row][col])
	}
}

func (m Model) renderHeader() string {
	headerStrings := make([]string, m.stopCol-m.startCol+1)
	headerStrings[0] = m.renderCell(0, 0)
	zz := 1
	for j := m.startCol; j < m.stopCol; j++ {
		headerStrings[zz] = m.renderCell(0, j)
		zz++
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, headerStrings...)
}

func (m Model) renderRow(row int) string {
	rowStrings := make([]string, m.stopCol-m.startCol+1)
	rowStrings[0] = m.renderCell(row, 0)
	zz := 1
	for j := m.startCol; j < m.stopCol; j++ {
		rowStrings[zz] = m.renderCell(row, j)
		zz++
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, rowStrings...)
}

func (m Model) renderAllRows() string {
	rowStrings := make([]string, m.stopRow-m.startRow+1)
	zz := 0
	for j := m.startRow; j < m.stopRow; j++ {
		rowStrings[zz] = m.renderRow(j)
		zz++
	}
	return lipgloss.JoinVertical(lipgloss.Left, rowStrings...)
}

func MoveUp(m *Model, n int) {
	m.cell.row = utils.Max(1, m.cell.row-n)
}

func MoveDown(m *Model, n int) {
	m.cell.row = utils.Min(m.NumRows()-1, m.cell.row+n)
}

func MoveLeft(m *Model, n int) {
	m.cell.col = utils.Max(1, m.cell.col-n)
}

func MoveRight(m *Model, n int) {
	m.cell.col = utils.Min(m.NumCols()-1, m.cell.col+n)
}

func Hide(m *Model) {
	m.columns[m.cell.col].hide = !m.columns[m.cell.col].hide
	m.stopCol = m.limitFromLeft(m.startCol)
}

func UnHideAll(m *Model) {
	for i := range m.columns {
		m.columns[i].hide = false
	}
	m.stopCol = m.limitFromLeft(m.startCol)
}

func fixView(m *Model) {
	if m.cell.row < m.startRow {
		m.startRow = utils.Max(1, utils.Min(m.startRow, m.cell.row))
		m.stopRow = m.startRow + m.termHeight - 3
	}
	if m.cell.row > m.stopRow-1 {
		m.stopRow = utils.Min(m.NumRows(), utils.Max(m.stopRow, m.cell.row+1))
		m.startRow = m.stopRow - m.termHeight + 3
	}
	if m.cell.col < m.startCol {
		m.startCol = utils.Max(1, utils.Min(m.startCol, m.cell.col))
		m.stopCol = m.limitFromLeft(m.cell.col)
	}
	if m.cell.col > m.stopCol-1 {
		m.stopCol = utils.Min(m.NumCols(), utils.Max(m.stopCol, m.cell.col+1))
		m.startCol = m.limitFromRight(m.cell.col)
	}
	m.cell.row = utils.Min(utils.Max(m.cell.row, m.startRow), m.stopRow-1)
	m.cell.col = utils.Min(utils.Max(m.cell.col, m.startCol), m.stopCol-1)
}

func (m Model) NumCols() int {
	return len(m.columns)
}

func (m Model) NumRows() int {
	return len(m.records)
}

func (m Model) getColumnWidth(col int) (w int) {
	if m.columns[col].hide {
		w = 2
	} else {
		w = m.columns[col].width + 2*m.styles.padding + 1
	}
	return w
}

func (m Model) limitFromLeft(col int) int {
	s := m.getColumnWidth(0)
	j := col
	for {
		if j == m.NumCols() {
			break
		}
		s += m.getColumnWidth(j)
		if s > m.termWidth {
			break
		}
		j++
	}
	return j
}

func (m Model) limitFromRight(col int) int {
	s := m.getColumnWidth(0)
	j := col
	for {
		if j == 0 {
			break
		}
		s += m.getColumnWidth(j)
		if s > m.termWidth {
			break
		}
		j--
	}
	return j + 1
}

func mod[T uint | int](x, m T) T {
	y := x
	for y >= m {
		y -= m
	}
	return y
}

// bubbletea Model interface

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termHeight = msg.Height
		m.termWidth = msg.Width
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.KeyMap.Up):
			MoveUp(&m, 1)
		case key.Matches(msg, m.KeyMap.Down):
			MoveDown(&m, 1)
		case key.Matches(msg, m.KeyMap.Left):
			MoveLeft(&m, 1)
		case key.Matches(msg, m.KeyMap.Right):
			MoveRight(&m, 1)
		case key.Matches(msg, m.KeyMap.PageUp):
			MoveUp(&m, m.termHeight/2)
		case key.Matches(msg, m.KeyMap.PageDown):
			MoveDown(&m, m.termHeight/2)
		case key.Matches(msg, m.KeyMap.JumpLeft):
			MoveLeft(&m, 5)
		case key.Matches(msg, m.KeyMap.JumpRight):
			MoveRight(&m, 5)
		case key.Matches(msg, m.KeyMap.Hide):
			Hide(&m)
		case key.Matches(msg, m.KeyMap.Hide):
			UnHideAll(&m)
		case key.Matches(msg, m.KeyMap.IncUnselFore):
			m.styles.unselectedForegroundColor = mod(m.styles.unselectedForegroundColor+1, 16)
		case key.Matches(msg, m.KeyMap.IncUnselBack):
			m.styles.unselectedBackgroundColor = mod(m.styles.unselectedBackgroundColor+1, 16)
		case key.Matches(msg, m.KeyMap.IncSelBack):
			m.styles.SelectedBackgroundColor = mod(m.styles.SelectedBackgroundColor+1, 16)
		case key.Matches(msg, m.KeyMap.IncSelFore):
			m.styles.SelectedForegroundColor = mod(m.styles.SelectedForegroundColor+1, 16)
		case key.Matches(msg, m.KeyMap.IncPadding):
			m.styles.padding++
			m.stopCol = m.limitFromLeft(m.startCol)
		case key.Matches(msg, m.KeyMap.DecPadding):
			m.styles.padding = utils.Max(0, m.styles.padding-1)
			m.stopCol = m.limitFromLeft(m.startCol)
		}
	}
	fixView(&m)
	return m, nil
}

func (m Model) View() string {
	header := m.renderHeader()
	rows := m.renderAllRows()
	s := lipgloss.JoinVertical(lipgloss.Left, header, rows)
	return strings.TrimRight(s, "\n ")
}

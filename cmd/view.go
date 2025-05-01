package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type popupModel struct {
	content      string
	viewport     viewport.Model
	width        int
	height       int
	windowWidth  int
	windowHeight int
	ready        bool
}

func (m popupModel) Init() tea.Cmd {
	return nil
}

func (m popupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "esc" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		m.width = min(100, msg.Width*3/4)
		m.height = min(40, msg.Height*3/4)

		if !m.ready {
			contentWidth := m.width - 4   // 2 for padding on each side
			contentHeight := m.height - 8 // space for header, footer, and padding

			m.viewport = viewport.New(contentWidth, contentHeight)
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			contentWidth := m.width - 4
			contentHeight := m.height - 8
			m.viewport.Width = contentWidth
			m.viewport.Height = contentHeight
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m popupModel) View() string {
	if !m.ready {
		return "Initializing..."
	}

	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.width - 4)

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Align(lipgloss.Center).
		Width(m.width - 4)

	header := headerStyle.Render("Query Results")
	footer := footerStyle.Render("↑/↓: Scroll • q/ESC: Close")

	separator := strings.Repeat("─", m.width-4)

	popupContent := lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		separator,
		m.viewport.View(),
		separator,
		footer,
	)

	boxContent := style.Width(m.width).Render(popupContent)

	boxWidth := lipgloss.Width(boxContent)
	boxHeight := lipgloss.Height(boxContent)

	horizontalOffset := (m.windowWidth - boxWidth) / 2
	verticalOffset := (m.windowHeight - boxHeight) / 2

	if horizontalOffset < 0 {
		horizontalOffset = 0
	}
	if verticalOffset < 0 {
		verticalOffset = 0
	}

	horizontalPadding := strings.Repeat(" ", horizontalOffset)
	verticalPadding := strings.Repeat("\n", verticalOffset)

	centeredContent := verticalPadding

	lines := strings.Split(boxContent, "\n")
	for i, line := range lines {
		centeredContent += horizontalPadding + line
		if i < len(lines)-1 {
			centeredContent += "\n"
		}
	}

	return centeredContent
}

func ShowPopup(content string) error {
	p := popupModel{content: content}
	prog := tea.NewProgram(p, tea.WithAltScreen())
	_, err := prog.Run()
	return err
}

func FormatQueryResult(result interface{}, stats map[string]interface{}, executionTime float64) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("📊 Results:\n\n%v\n\n", result))

	sb.WriteString("📈 Statistics:\n")
	sb.WriteString(fmt.Sprintf("⏱️  Execution time: %.2f ms\n", executionTime))

	if stats != nil {
		if scanned, ok := stats["scannedFull"].(int); ok {
			sb.WriteString(fmt.Sprintf("📄 Documents read: %d\n", scanned))
		}
		if written, ok := stats["writesExecuted"].(int); ok {
			sb.WriteString(fmt.Sprintf("✏️  Documents written: %d\n", written))
		}
		if memory, ok := stats["peakMemoryUsage"].(int); ok {
			sb.WriteString(fmt.Sprintf("💾 Memory usage: %s\n", formatBytes(memory)))
		}
		if serverTime, ok := stats["executionTime"].(float64); ok {
			sb.WriteString(fmt.Sprintf("⏱️  Server execution time: %.2f ms\n", serverTime))
		}
	}

	return sb.String()
}

func formatBytes(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := unit, 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

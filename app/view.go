// Package app defines tui
package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const maxWidth = 135

func (m Model) View() string {
	s := m.styles

	// 1. Left-side form (always visible)
	formView := strings.TrimSuffix(m.form.View(), "\n\n")
	form := m.lg.NewStyle().Margin(1, 0).Render(formView)

	// 2. Middle status (always visible)
	var selectedCluster string
	if raw := m.form.Get("cluster"); raw != nil {
		if v, ok := raw.([]string); ok {
			selectedCluster = strings.Join(v, "  \n")
		}
	}

	var selectedCommand string
	if raw := m.form.Get("command"); raw != nil {
		if v, ok := raw.([]string); ok {
			selectedCommand = strings.Join(v, "  \n")
		}
	}

	spinView := ""
	if m.spinning {
		spinView = m.spinner.View()
	}
	statusContent := s.StatusHeader.Render(
		"Build") + "  " + spinView + "\n" + "Cluster(s):" + "\n" + selectedCluster + "\n\n" + "Command(s):" + "\n" + selectedCommand

	// 3. Right-side completed output (only shown on output)

	// Compute heights
	formH := lipgloss.Height(form)
	statusH := lipgloss.Height(statusContent)
	outputH := lipgloss.Height(m.finalOutput)
	maxH := max(outputH, max(statusH, formH))

	// TODO: this is never displayed
	completedBox := ""
	if m.finalOutput != "" {
		completedContent := s.StatusHeader.Render("Output Content") + "\n" + m.finalOutput
		completedBox = s.Status.Width(maxWidth / 2).Render(completedContent)
	}

	statusBox := s.Status.Width(60).Height(maxH).Render(statusContent)
	if completedBox != "" {
		completedBox = s.Status.Width(60).Height(maxH).Render(
			s.StatusHeader.Render("Output") + "\n" + m.finalOutput,
		)
	}

	spacer := lipgloss.NewStyle().Width(2).Render(" ")
	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		form,
		spacer,
		statusBox,
		spacer,
		completedBox,
	)

	// header + footer same as before
	errors := m.form.Errors()
	header := m.appBoundaryView("BCP tui")
	if len(errors) > 0 {
		header = m.appErrorBoundaryView(m.errorView())
	}
	footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
	if len(errors) > 0 {
		footer = m.appErrorBoundaryView("")
	}

	return s.Base.Render(header + "\n" + body + "\n\n" + footer)
}

func (m Model) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}

func (m Model) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func (m Model) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(red),
	)
}

package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// update spinner
	if m.spinning {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {

	// Handle window resizing
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, maxWidth) - m.styles.Base.GetHorizontalFrameSize()

	case tea.KeyMsg:
		switch msg.String() {

		// Quit
		case "ctrl+c", "q":
			return m, tea.Quit

		// Move back one field
		case "esc":
			m.form.PrevField()
			return m, nil

		// Trigger action on Enter
		case "enter":
			// Capture selected commands
			if raw := m.form.Get("command"); raw != nil {
				if c, ok := raw.([]string); ok {
					m.commands = c
				}
				// Capture selected clusters
				if raw := m.form.Get("cluster"); raw != nil {
					if v, ok := raw.([]string); ok {
						m.clusters = v
					}
				}
				m.RunOnCluster() // populates m.finalOutput
				return m, nil
			}
		}
	}

	// Continue form updates
	form, formCmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	cmds = append(cmds, formCmd)
	return m, tea.Batch(cmds...)
}

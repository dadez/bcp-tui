package app

import (
	"github/dadez/bcp-tui/config"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type state int

type Model struct {
	state       state
	lg          *lipgloss.Renderer
	styles      *Styles
	clusters    []string
	commands    []string
	form        *huh.Form
	width       int
	finalOutput string
	spinner     spinner.Model
	spinning    bool
}

func customKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap() // gives you the default keymap struct

	// Override SelectAll for MultiSelect → Shift+A
	km.MultiSelect.SelectAll = key.NewBinding(
		key.WithKeys("A"), // Shift+A
		key.WithHelp("shift+a", "select all"),
	)

	// (Optional) change SelectNone too (example: Shift+N)
	km.MultiSelect.SelectNone = key.NewBinding(
		key.WithKeys("A"),
		key.WithHelp("shift+a", "select none"),
	)

	return km
}

func NewModel() Model {
	m := Model{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	clusters := config.AppConfig.Clusters
	clusterOpts := make([]huh.Option[string], 0, len(clusters))
	for _, cluster := range clusters {
		clusterOpts = append(clusterOpts, huh.NewOption(cluster, cluster))
	}

	commands := config.AppConfig.Commands
	commandsOpts := make([]huh.Option[string], 0, len(commands))
	for _, command := range commands {
		commandsOpts = append(commandsOpts, huh.NewOption(command.Name, command.Command))
	}

	// add a custom command
	commandsOpts = append(commandsOpts, huh.NewOption("Custom command", "custom"))

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Key("cluster").
				Options(clusterOpts...).
				Title("Choose your cluster").
				Description("Select cluster(s) bellow").
				Filterable(true).
				Value(&m.clusters),

			huh.NewMultiSelect[string]().
				Key("command").
				Value(&m.commands).
				Options(commandsOpts...).
				Title("Choose your action").
				Description("This will determine the action"),
		),
	).
		WithWidth(45).
		WithShowHelp(true).
		WithKeyMap(customKeyMap()).
		WithShowErrors(true)

	m.spinner = spinner.New()
	m.spinner.Spinner = spinner.Jump
	m.spinning = true

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.form.Init(),
		m.spinner.Tick,
	)
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

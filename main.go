package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github/dadez/bcp-tui/config"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const maxWidth = 135

var (
	red    = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	indigo = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green  = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
)

type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Highlight,
	ErrorHeaderText,
	Help lipgloss.Style
}

func NewStyles(lg *lipgloss.Renderer) *Styles {
	s := Styles{}
	s.Base = lg.NewStyle().
		Padding(1, 4, 0, 1)
	s.HeaderText = lg.NewStyle().
		Foreground(indigo).
		Bold(true).
		Padding(0, 1, 0, 2)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(indigo).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lg.NewStyle().
		Foreground(green).
		Bold(true)
	s.Highlight = lg.NewStyle().
		Foreground(lipgloss.Color("212"))
	s.ErrorHeaderText = s.HeaderText.
		Foreground(red)
	s.Help = lg.NewStyle().
		Foreground(lipgloss.Color("240"))
	return &s
}

type state int

const (
	statusNormal state = iota
	stateDone
)

type Model struct {
	state       state
	lg          *lipgloss.Renderer
	styles      *Styles
	clusters    []string
	commands    []string
	form        *huh.Form
	width       int
	finalOutput string
}

func customKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap() // gives you the default keymap struct

	// Override only SelectAll for MultiSelect → Shift+A
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

	return m
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				m.runOnCluster() // populates m.finalOutput
				return m, nil
			}
		}
	}

	// Continue form updates
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	return m, cmd
}

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

	statusContent := s.StatusHeader.Render(
		"Build") + "\n" + "Cluster(s):" + "\n" + selectedCluster + "\n\n" + "Command(s):" + "\n" + selectedCommand

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

func (m *Model) runOnCluster() {
	m.finalOutput = ""

	// Validate that every command contains at least one "%s"
	for _, cmdTemplate := range m.commands {
		placeholderIndex := strings.Index(cmdTemplate, "%s")

		// 1) Must contain %s
		if placeholderIndex == -1 {
			m.finalOutput += fmt.Sprintf(
				"Invalid command template: %q (missing %%s placeholder)\n",
				cmdTemplate,
			)
			return
		}

		// 2) Must NOT start with %s
		if placeholderIndex == 0 {
			m.finalOutput += fmt.Sprintf(
				"Invalid command template: %q (%%s cannot be the first token)\n",
				cmdTemplate,
			)
			return
		}
	}

	for _, cluster := range m.clusters {
		for _, cmdTemplate := range m.commands {
			// Safe because of validation above
			fullCmd := fmt.Sprintf(cmdTemplate, cluster)
			parts := strings.Fields(fullCmd)

			if len(parts) < 2 { // binary + cluster
				m.finalOutput += fmt.Sprintf(
					"Invalid expanded command for cluster %s: %q (insufficient args)\n",
					cluster,
					fullCmd,
				)
				continue
			}

			cmd := exec.Command(parts[0], parts[1:]...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				errorHeader := m.styles.ErrorHeaderText.Render("Command failed")
				m.finalOutput += fmt.Sprintf(
					"%s\nFailed to run command %q for cluster %s:\n%s\nError: %v\n\n",
					errorHeader,
					parts,
					cluster,
					string(out),
					err,
				)
				continue
			}

			m.finalOutput += fmt.Sprintf("Command OK for %s:\n%s\n", cluster, string(out))
		}
	}
}

func main() {
	// load configuration
	configPath := flag.String("c", "config.yaml", "Path to config file")
	flag.Parse()
	if err := config.Load(*configPath); err != nil {
		log.Fatal(err)
	}

	// build form
	_, err := tea.NewProgram(NewModel()).Run()
	if err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}
}

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github/dadez/bcp-tui/config"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const maxWidth = 80

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
	action      string
	command     string
	form        *huh.Form
	width       int
	finalOutput string
	done        bool
}

func NewModel() Model {
	m := Model{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	// clusters := []string{"dev-cluster-1", "prod-cluster-2", "staging-cluster-3", "test-cluster-4", "demo-cluster-5"}
	clusters := config.AppConfig.Clusters
	opts := make([]huh.Option[string], 0, len(clusters))
	for _, cluster := range clusters {
		opts = append(opts, huh.NewOption(cluster, cluster))
	}

	actions := config.AppConfig.Actions
	actionOpts := make([]huh.Option[string], 0, len(actions))
	for _, action := range actions {
		actionOpts = append(actionOpts, huh.NewOption(action, action))
	}

	commands := config.AppConfig.Commands
	commandsOpts := make([]huh.Option[string], 0, len(commands))
	for _, command := range commands {
		commandsOpts = append(commandsOpts, huh.NewOption(command.Name, command.Command))
	}

	urls := config.AppConfig.Urls
	urlOpts := make([]huh.Option[string], 0, len(urls))
	for _, url := range urls {
		urlOpts = append(urlOpts, huh.NewOption(url.Name, url.URL))
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Key("cluster").
				Options(opts...).
				Title("Choose your cluster").
				Description("Action will be performed on this cluster(s)").
				Filterable(true).
				Value(&m.clusters),

			huh.NewSelect[string]().
				Key("action").
				Value(&m.action).
				Options(actionOpts...).
				Title("Choose your action").
				Description("This will determine the action"),

			huh.NewSelect[string]().
				Key("command").
				Value(&m.command).
				OptionsFunc(func() []huh.Option[string] {
					switch m.action {
					case "command":
						return commandsOpts
					case "browse":
						return urlOpts
					default:
						return urlOpts
					}
				}, &m.action).
				Title("Choose your command").
				Description("This will determine the command or target"),
		),
	).
		WithWidth(45).
		WithShowHelp(false).
		WithShowErrors(false)

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
			// If we're on the action field, run the action
			if m.form.Get("action") != nil {
				if m.form.Get("command") != nil {
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
	var selected string
	if raw := m.form.Get("cluster"); raw != nil {
		if v, ok := raw.([]string); ok {
			selected = strings.Join(v, "  \n")
		}
	}
	action := m.form.GetString("action")
	command := m.form.GetString("command")

	statusContent := s.StatusHeader.Render(
		"Cluster(s)") + "\n" + selected + "\n" + "Action: " + action + "\n" + "Command: " + command

	// 3. Right-side completed output (only shown if needed)
	completedBox := ""
	if m.finalOutput != "" {
		completedContent := s.StatusHeader.Render("Output") + "\n" + m.finalOutput
		completedBox = s.Status.Width(maxWidth).Render(completedContent)
	}

	// Compute heights
	formH := lipgloss.Height(form)
	statusH := lipgloss.Height(statusContent)
	outputH := lipgloss.Height(m.finalOutput)
	maxH := formH
	if statusH > maxH {
		maxH = statusH
	}
	if outputH > maxH {
		maxH = outputH
	}

	statusBox := s.Status.Width(28).Height(maxH).Render(statusContent)
	if completedBox != "" {
		completedBox = s.Status.Width(maxWidth).Height(maxH).Render(
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
	command := m.form.GetString("command")
	m.finalOutput = ""

	for _, cluster := range m.clusters {
		c := fmt.Sprintf(command, cluster)
		parts := strings.Fields(c)
		cmd := exec.Command(parts[0], parts[1:]...)
		out, err := cmd.CombinedOutput() // <-- CRUCIAL
		if err != nil {
			errorHeader := m.styles.ErrorHeaderText.Render("Command failed")
			m.finalOutput += fmt.Sprintf(
				"%s\n Failed to run command %s with args %s for cluster %s:\n%s\nError: %v\n\n",
				errorHeader,
				cmd.Path,
				cmd.Args,
				cluster,
				string(out),
				err,
			)
			continue
		}
		// If it worked (or oc exists), you still show what's returned
		m.finalOutput += fmt.Sprintf("Command OK for %s:\n%s\n", cluster, string(out))
	}
}

func main() {
	// configuration
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

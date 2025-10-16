package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

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
	form        *huh.Form
	width       int
	finalOutput string
	done        bool
}

func NewModel() Model {
	m := Model{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	clusters := []string{"dev-cluster-1", "prod-cluster-2", "staging-cluster-3", "test-cluster-4", "demo-cluster-5"}
	opts := make([]huh.Option[string], 0, len(clusters))
	for _, cluster := range clusters {
		opts = append(opts, huh.NewOption(cluster, cluster))
	}

	// per default, activate yes box in confirmation
	done := true

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Key("cluster").
				// Options(huh.NewOptions(m.clusters...)...).
				Options(opts...).
				Title("Choose your cluster").
				Description("Action will be performed on this cluster(s)").
				Filterable(true).
				Value(&m.clusters),

			huh.NewSelect[string]().
				Key("action").
				// Options(huh.NewOptions("grafana", "argocd", "9999")...).
				Options(
					huh.NewOption("grafana", "grafana"),
					huh.NewOption("argocd", "argocd"),
					huh.NewOption("console", "console"),
					huh.NewOption("login", "login"),
				).
				Title("Choose your action").
				Description("This will determine the action"),

			huh.NewConfirm().
				Key("done").
				Title("All done?").
				Value(&done).
				Validate(func(v bool) error {
					if !v {
						return fmt.Errorf("finish before exiting")
					}
					return nil
				}).
				Affirmative("Yep").
				Negative("Wait, no"),
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
	if m.done {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "q" || msg.String() == "enter" {
				return m, tea.Quit
			}
		}
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, maxWidth) - m.styles.Base.GetHorizontalFrameSize()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Interrupt
		case "esc", "q":
			return m, tea.Quit
		}
	}

	// Process the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	// When done, manually commit the MultiSelect if it’s still active
	if m.form.State == huh.StateCompleted {
		if !m.done {
			// Force read from the form's MultiSelect value
			if raw := m.form.Get("cluster"); raw != nil {
				if v, ok := raw.([]string); ok {
					m.clusters = v // populate bound slice
				}
			}
			m.openCluster()
			m.done = true

		}
		// do no quit immediately for allow display of command output
		// return m, tea.Quit
		return m, nil
	}

	return m, cmd
}

func (m Model) View() string {
	s := m.styles

	switch m.form.State {
	case huh.StateCompleted:
		if m.finalOutput != "" {
			return m.styles.Status.Padding(1, 2).Width(60).Render(m.finalOutput)
		}

		var b strings.Builder

		action := m.form.GetString("action")
		if action == "" {
			fmt.Fprintf(&b, "No action was selected!\n")
			return b.String()
		}

		fmt.Fprintf(&b, "Action: %s\n", action)
		fmt.Fprintf(&b, "Selected actions:\n")

		for _, cluster := range m.clusters {
			switch action {
			case "login":
				fmt.Fprintf(&b, "login on %s\n", cluster)
			default:
				url := fmt.Sprintf("https://%s.%s.example.com", action, cluster)
				fmt.Fprintf(&b, "open %s\n", url)
			}
		}

		return s.Status.Padding(1, 2).Width(48).Render(b.String())

	default:

		// Form (left side)
		v := strings.TrimSuffix(m.form.View(), "\n\n")
		form := m.lg.NewStyle().Margin(1, 0).Render(v)

		// Status (right side)
		var status string
		{
			var action string
			if m.form.GetString("action") != "" {
				action = "action: " + m.form.GetString("action")
			}

			var selected string
			if raw := m.form.Get("cluster"); raw != nil {
				if v, ok := raw.([]string); ok {
					selected = strings.Join(v, "  \n")
				}
			}

			const statusWidth = 28
			statusMarginLeft := m.width - statusWidth - lipgloss.Width(form) - s.Status.GetMarginRight()
			status = s.Status.
				Height(lipgloss.Height(form)).
				Width(statusWidth).
				MarginLeft(statusMarginLeft).
				Render(
					s.StatusHeader.Render("Cluster(s)") + "\n" +
						selected + "\n" +
						action,
				)
		}

		errors := m.form.Errors()
		header := m.appBoundaryView("BCP tui")
		if len(errors) > 0 {
			header = m.appErrorBoundaryView(m.errorView())
		}
		body := lipgloss.JoinHorizontal(lipgloss.Left, form, status)

		footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
		if len(errors) > 0 {
			footer = m.appErrorBoundaryView("")
		}

		return s.Base.Render(header + "\n" + body + "\n\n" + footer)
	}
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

func openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	cmd.Start()
}

func (m *Model) runOcLogin(cluster string) {
	cmd := exec.Command("oc", "login", cluster)

	out, err := cmd.CombinedOutput() // <-- CRUCIAL
	if err != nil {
		// Store stderr/stdout (+ error) for display in the View
		m.finalOutput = fmt.Sprintf(
			"Failed to run oc login for %s:\n%s\nError: %v",
			cluster,
			string(out),
			err,
		)
		return
	}

	// If it worked (or oc exists), you still show what's returned
	m.finalOutput = fmt.Sprintf("Login OK for %s:\n%s", cluster, string(out))
}

func (m *Model) openCluster() {
	action := m.form.GetString("action")

	for _, cluster := range m.clusters {
		switch action {
		case "login":
			m.runOcLogin(cluster)
		default:
			url := fmt.Sprintf("https://%s.%s.example.com", action, cluster)
			openURL(url)
		}
	}
}

func main() {
	_, err := tea.NewProgram(NewModel()).Run()
	if err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}
}

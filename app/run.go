package app

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
)

func (m *Model) RunOnCluster() {
	m.finalOutput = ""

	// Check if user selected the custom option
	for i, cmdTemplate := range m.commands {
		if cmdTemplate == "custom" {
			var customInput string

			// Prompt user for the actual command
			prompt := huh.NewInput().
				Title("Enter your custom command").
				Description("You must use %s somewhere to insert the cluster name").
				Placeholder("e.g. kubectl get pods -n %s").
				Value(&customInput)

			// Run the prompt immediately
			if err := prompt.Run(); err != nil {
				m.finalOutput += fmt.Sprintf("Custom command input canceled: %v\n", err)
				return
			}

			// Replace the placeholder in the user’s selections
			m.commands[i] = customInput
		}
	}

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

			m.spinning = false

			m.finalOutput += fmt.Sprintf("Command OK for %s:\n%s\n", cluster, string(out))
		}
	}
}

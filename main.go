package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github/dadez/bcp-tui/app"
	"github/dadez/bcp-tui/config"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// load configuration
	configPath := flag.String("c", "", "Path to config file")
	flag.Parse()
	if err := config.Load(*configPath); err != nil {
		log.Fatal(err)
	}

	// build form
	_, err := tea.NewProgram(app.NewModel()).Run()
	if err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}
}

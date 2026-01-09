package main

import (
	"fmt"
	"io/ioutil"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
)

type Service struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`
}

type RenderConfig struct {
	Services []Service `yaml:"services"`
}

type model struct {
	cursor  int
	choices []string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			// Will add action later
		}
	}
	return m, nil
}

func (m model) View() string {
	s := "Render CLI TUI\n\n"

	if len(m.choices) == 0 {
		return s + "No services found in config.yaml\n"
	}

	s += "Services:\n"
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}
	s += "\nPress q to quit.\n"
	return s
}

func main() {
	yamlFile, err := ioutil.ReadFile("../../config/render.yaml")
	if err != nil {
		fmt.Printf("Error reading config.yaml: %v\n", err)
		os.Exit(1)
	}

	var config RenderConfig
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		fmt.Printf("Error parsing config.yaml: %v\n", err)
		os.Exit(1)
	}

	var services []string
	for _, service := range config.Services {
		services = append(services, service.Name)
	}

	p := tea.NewProgram(model{
		choices: services,
	})

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

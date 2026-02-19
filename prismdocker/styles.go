package main

import "github.com/charmbracelet/lipgloss"

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	ListHeaderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(subtle).
			MarginRight(2)

	ListItemStyle = lipgloss.NewStyle().PaddingLeft(1)

	checkMark = lipgloss.NewStyle().SetString("✓").
			Foreground(special).
			PaddingRight(1).
			String()

	listDone = func(s string) string {
		return checkMark + lipgloss.NewStyle().
			Strikethrough(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#969B86", Dark: "#696969"}).
			Render(s)
	}

	docStyle = lipgloss.NewStyle().Margin(1, 2)

	tableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Align(lipgloss.Center)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(true)

	statusUpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")) // Green

	statusExitedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")) // Red

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	// Port Colors
	portHostStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("45")) // Cyan for host

	portContainerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")) // Orange for container

	// Table Enhancements
	idStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("141")) // Purple (Dracula-ish)

	imageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")) // Dimmed gray

	rowEvenStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("235")) // Very dark gray for zebra stripe
)

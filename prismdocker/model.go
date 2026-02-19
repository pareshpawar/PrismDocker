package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/moby/moby/client"
)

type SortOrder int

type ActiveView int

const (
	viewContainers ActiveView = iota
	viewLogs
)

const (
	SortByID SortOrder = iota
	SortByName
	SortByImage
	SortByState
	SortByCPU
	SortByMem
)

func (s SortOrder) String() string {
	switch s {
	case SortByID:
		return "ID"
	case SortByName:
		return "Name"
	case SortByImage:
		return "Image"
	case SortByState:
		return "State"
	case SortByCPU:
		return "CPU%"
	case SortByMem:
		return "Mem"
	default:
		return "Unknown"
	}
}

type model struct {
	dockerClient       *client.Client
	allContainers      []Container
	filteredContainers []Container
	cursor             int
	err                error
	width              int
	height             int
	sortOrder          SortOrder
	showAll            bool
	tick               int
	tableOffset        int
	showStats          bool
	stats              map[string]Stats
	// Action confirm dialog
	confirmMode   bool
	confirmAction string // "remove"
	// Brief status message shown in footer
	statusMsg  string
	statusTick int // countdown to clear statusMsg
	// Log viewer
	activeView    ActiveView
	logLines      []string
	logFilter     string
	logFilterMode bool
	logOffset     int
	logContainer  string // ID of container being viewed
}

func initialModel() model {
	cli, err := NewDockerClient()
	if err != nil {
		return model{err: err}
	}

	return model{
		dockerClient:       cli,
		allContainers:      []Container{},
		filteredContainers: []Container{},
		cursor:             0,
		sortOrder:          SortByState,
		showAll:            false,
		tableOffset:        0,
		showStats:          false,
		stats:              make(map[string]Stats),
		activeView:         viewContainers,
	}
}

// Init starts the Bubble Tea program.
// It kicks off the tick loop and performs an initial container fetch.
func (m model) Init() tea.Cmd {
	return tea.Batch(
		waitForTick(),
		waitForAnimTick(),
		fetchContainers(m.dockerClient),
	)
}

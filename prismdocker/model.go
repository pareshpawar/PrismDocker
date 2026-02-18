package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/moby/moby/client"
)

type SortOrder int

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
	allContainers      []Container // Store all fetched containers
	filteredContainers []Container // Store containers to display
	cursor             int
	err                error
	width              int
	height             int
	sortOrder          SortOrder
	showAll            bool // True = show all, False = show running only
	tick               int
	tableOffset        int // Vertical scroll offset for the table
	showStats          bool
	stats              map[string]Stats // Cache stats by Container ID
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
		sortOrder:          SortByState, // Default sort by State (Running first)
		showAll:            false,       // Default show running only
		tableOffset:        0,
		showStats:          false,
		stats:              make(map[string]Stats),
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

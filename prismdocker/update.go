package main

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/moby/moby/client"
)

type tickMsg time.Time

type containersMsg []Container

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func waitForTick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchContainers(cli *client.Client) tea.Cmd {
	return func() tea.Msg {
		containers, err := ListContainers(cli)
		if err != nil {
			return errMsg{err}
		}
		return containersMsg(containers)
	}
}

// Update is the main update loop for the Bubble Tea program.
// It handles incoming messages and returns an updated model and a command.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Handle key presses
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Scroll up if cursor moves above offset
				if m.cursor < m.tableOffset {
					m.tableOffset = m.cursor
				}
			}
		case "down", "j":
			if m.cursor < len(m.filteredContainers)-1 {
				m.cursor++
				// Scroll down handled in logical update or view?
				// We need to know height to scroll down.
				// But height is dynamic.
				// Better to check bounds in View? No, view is read-only.
				// We can just keep cursor valid here.
				// The "keep cursor in view" logic for BOTTOM edge depends on window height.
				// We can store a maxVisibleRows in model updated during View/Resize?
				// Or approximate here?
				// Actually, we can just ensure offset <= cursor.
				// The logic "if cursor >= offset + height" needs height.
				// m.height is available.
				// Check approximate table height (height - 10 for header/footer)
				// exact calc in view is better, but state must be in model.
				// Let's assume a safe margin or handle it in View?
				// Bubble Tea best practice: Model should know everything.
				// We can roughly estimate table height here or just let the View adjust the offset?
				// View cannot adjust offset (it returns string).
				// So Update MUST know the table height.
				// The View calculates `headerHeight` etc.
				// Let's assume header ~10 lines, footer ~2 lines.
				// tableHeight ~= m.height - 12.
				//
				// Let's be reactive:
				// implementation detail:
				// If we implement the viewport in View, we can pass a message back? Too complex.
				// Simple way: Update uses m.height.

				// Conservative estimate:
				headerHeight := 10 // Approximate
				footerHeight := 2
				tableHeight := m.height - headerHeight - footerHeight
				if tableHeight < 1 {
					tableHeight = 1
				}

				if m.cursor >= m.tableOffset+tableHeight {
					m.tableOffset = m.cursor - tableHeight + 1
				}
			}
		case "r":
			// Manual refresh triggered by 'r'
			return m, fetchContainers(m.dockerClient)
		case "s":
			// Cycle sort order: base 4 always available; CPU/Mem only when stats on
			if m.showStats {
				m.sortOrder = (m.sortOrder + 1) % 6 // 0-5: ID,Name,Image,State,CPU,Mem
			} else {
				m.sortOrder = (m.sortOrder + 1) % 4 // 0-3: ID,Name,Image,State
			}
			m.filteredContainers = sortAndFilter(m.allContainers, m.sortOrder, m.showAll, m.stats)
			m.cursor = 0
			m.tableOffset = 0
		case "a":
			// Toggle show all
			m.showAll = !m.showAll
			m.filteredContainers = sortAndFilter(m.allContainers, m.sortOrder, m.showAll, m.stats)
			m.cursor = 0
			m.tableOffset = 0
		case "t":
			// Toggle stats
			m.showStats = !m.showStats
			if m.showStats {
				// Trigger immediate fetch
				return m, fetchAllStats(m.dockerClient, m.filteredContainers)
			} else {
				// If currently sorted by CPU/Mem, reset to State
				if m.sortOrder == SortByCPU || m.sortOrder == SortByMem {
					m.sortOrder = SortByState
					m.filteredContainers = sortAndFilter(m.allContainers, m.sortOrder, m.showAll, m.stats)
				}
			}
		}

	// Handle window resizing to ensure responsiveness
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	// Handle the tick message for periodic updates (Data Refresh)
	case tickMsg:
		cmds := []tea.Cmd{
			waitForTick(),                   // Schedule next data refresh
			fetchContainers(m.dockerClient), // Fetch latest container list
		}
		if m.showStats {
			cmds = append(cmds, fetchAllStats(m.dockerClient, m.filteredContainers))
		}
		return m, tea.Batch(cmds...)

	// Handle animation tick
	case animTickMsg:
		m.tick++
		return m, waitForAnimTick()

	// Handle the result of fetching containers
	// This message comes from the fetchContainers command.
	case containersMsg:
		m.allContainers = msg
		m.filteredContainers = sortAndFilter(m.allContainers, m.sortOrder, m.showAll, m.stats)

		// Adjust cursor if list shrank or is empty
		if m.cursor >= len(m.filteredContainers) && len(m.filteredContainers) > 0 {
			m.cursor = len(m.filteredContainers) - 1
		} else if len(m.filteredContainers) == 0 {
			m.cursor = 0
		}

	case statsMsg:
		m.stats = msg
		// Re-sort if sorted by CPU or Mem so order stays live
		if m.sortOrder == SortByCPU || m.sortOrder == SortByMem {
			m.filteredContainers = sortAndFilter(m.allContainers, m.sortOrder, m.showAll, m.stats)
		}

	case errMsg:
		m.err = msg.err
	}

	return m, nil
}

type statsMsg map[string]Stats

func fetchAllStats(cli *client.Client, containers []Container) tea.Cmd {
	return func() tea.Msg {
		results := make(map[string]Stats)
		// Parallel fetch could be better but sticking to serial for simplicity first or simple parallelism
		// For TUI responsiveness, we should do it in parallel
		// But let's keep it simple: fetch only for *Running* containers in view?
		// User sees `filteredContainers`. If we filter by running, we fetch those.
		// If we show all, we only fetch stats for running ones anyway (stats fail for stopped).

		for _, c := range containers {
			if strings.HasPrefix(strings.ToLower(c.Status), "up") { // Simple check if running
				stats, err := GetContainerStats(cli, c.ID)
				if err == nil {
					results[c.ID] = stats
				}
			}
		}
		return statsMsg(results)
	}
}

// Modify waitForTick to be faster for animation.
// However, we don't want to fetch containers every 200ms.
// We need two ticks.

type animTickMsg time.Time

func waitForAnimTick() tea.Cmd {
	return tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
		return animTickMsg(t)
	})
}

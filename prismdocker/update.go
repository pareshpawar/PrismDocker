package main

import (
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/moby/moby/client"
)

type tickMsg time.Time
type containersMsg []Container
type errMsg struct{ err error }
type actionMsg struct{ err error }
type logLineMsg string
type execDoneMsg struct{ err error }
type openBrowserMsg struct{}
type inspectMsg struct {
	data ContainerInspect
	err  error
}

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

func doAction(fn func() error) tea.Cmd {
	return func() tea.Msg {
		return actionMsg{fn()}
	}
}

func fetchInspect(cli *client.Client, containerID string) tea.Cmd {
	return func() tea.Msg {
		data, err := InspectContainer(cli, containerID)
		return inspectMsg{data, err}
	}
}

// applyFilter applies current sort/filter/search/compose settings.
func (m *model) applyFilter() {
	if m.groupByCompose {
		m.filteredContainers = sortAndFilterWithCompose(m.allContainers, m.sortOrder, m.showAll, m.stats, m.searchQuery)
	} else {
		m.filteredContainers = sortAndFilter(m.allContainers, m.sortOrder, m.showAll, m.stats, m.searchQuery)
	}
}

// Update is the main update loop for the Bubble Tea program.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// ── Inspect view mode ──────────────────────────────────────────
		if m.activeView == viewInspect {
			switch msg.String() {
			case "esc", "q":
				m.activeView = viewContainers
				m.inspectOffset = 0
			case "up", "k":
				if m.inspectOffset > 0 {
					m.inspectOffset--
				}
			case "down", "j":
				m.inspectOffset++
			}
			return m, nil
		}

		// ── Log view mode ──────────────────────────────────────────────
		if m.activeView == viewLogs {
			switch msg.String() {
			case "esc", "q":
				m.activeView = viewContainers
				m.logLines = nil
				m.logFilter = ""
				m.logFilterMode = false
				m.logOffset = 0
			case "/":
				m.logFilterMode = !m.logFilterMode
			case "up", "k":
				if m.logOffset > 0 {
					m.logOffset--
				}
			case "down", "j":
				m.logOffset++
			case "backspace":
				if m.logFilterMode && len(m.logFilter) > 0 {
					m.logFilter = m.logFilter[:len(m.logFilter)-1]
				}
			default:
				if m.logFilterMode && len(msg.String()) == 1 {
					m.logFilter += msg.String()
				}
			}
			return m, nil
		}

		// ── Help overlay ───────────────────────────────────────────────
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		// ── Confirm dialog mode ────────────────────────────────────────
		if m.confirmMode {
			switch msg.String() {
			case "y", "Y":
				m.confirmMode = false
				if m.confirmAction == "remove" && m.cursor < len(m.filteredContainers) {
					c := m.filteredContainers[m.cursor]
					return m, doAction(func() error {
						return RemoveContainer(m.dockerClient, c.ID)
					})
				}
			case "n", "N", "esc":
				m.confirmMode = false
			}
			return m, nil
		}

		// ── Search mode ────────────────────────────────────────────────
		if m.searchMode {
			switch msg.String() {
			case "esc":
				m.searchMode = false
				m.searchQuery = ""
				m.applyFilter()
				m.cursor = 0
				m.tableOffset = 0
			case "enter":
				m.searchMode = false
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.applyFilter()
					m.cursor = 0
					m.tableOffset = 0
				}
			default:
				if len(msg.String()) == 1 {
					m.searchQuery += msg.String()
					m.applyFilter()
					m.cursor = 0
					m.tableOffset = 0
				}
			}
			return m, nil
		}

		// ── Normal container view ──────────────────────────────────────
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.tableOffset {
					m.tableOffset = m.cursor
				}
			}

		case "down", "j":
			if m.cursor < len(m.filteredContainers)-1 {
				m.cursor++
				headerHeight := 10
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
			return m, fetchContainers(m.dockerClient)

		case "s":
			if m.showStats {
				m.sortOrder = (m.sortOrder + 1) % 6
			} else {
				m.sortOrder = (m.sortOrder + 1) % 4
			}
			m.applyFilter()
			m.cursor = 0
			m.tableOffset = 0

		case "a":
			m.showAll = !m.showAll
			m.applyFilter()
			m.cursor = 0
			m.tableOffset = 0

		case "t":
			m.showStats = !m.showStats
			if m.showStats {
				return m, fetchAllStats(m.dockerClient, m.filteredContainers)
			} else {
				if m.sortOrder == SortByCPU || m.sortOrder == SortByMem {
					m.sortOrder = SortByState
					m.applyFilter()
				}
			}

		case "/":
			m.searchMode = true

		case "?":
			m.showHelp = !m.showHelp

		case "g":
			m.groupByCompose = !m.groupByCompose
			m.applyFilter()
			m.cursor = 0
			m.tableOffset = 0

		// ── Container actions ──────────────────────────────────────────
		case "S": // Stop
			if m.cursor < len(m.filteredContainers) {
				c := m.filteredContainers[m.cursor]
				if c.State == "running" {
					m.statusMsg = "Stopping " + c.Names + "..."
					return m, doAction(func() error {
						return StopContainer(m.dockerClient, c.ID)
					})
				}
			}

		case "u": // Start (Up)
			if m.cursor < len(m.filteredContainers) {
				c := m.filteredContainers[m.cursor]
				if c.State != "running" {
					m.statusMsg = "Starting " + c.Names + "..."
					return m, doAction(func() error {
						return StartContainer(m.dockerClient, c.ID)
					})
				}
			}

		case "R": // Restart
			if m.cursor < len(m.filteredContainers) {
				c := m.filteredContainers[m.cursor]
				m.statusMsg = "Restarting " + c.Names + "..."
				return m, doAction(func() error {
					return RestartContainer(m.dockerClient, c.ID)
				})
			}

		case "x": // Remove (with confirm)
			if m.cursor < len(m.filteredContainers) {
				m.confirmMode = true
				m.confirmAction = "remove"
			}

		case "l": // Logs
			if m.cursor < len(m.filteredContainers) {
				c := m.filteredContainers[m.cursor]
				m.activeView = viewLogs
				m.logLines = nil
				m.logFilter = ""
				m.logFilterMode = false
				m.logOffset = 0
				m.logContainer = c.ID
				return m, fetchLogs(m.dockerClient, c.ID)
			}

		case "d": // Inspect/Details
			if m.cursor < len(m.filteredContainers) {
				c := m.filteredContainers[m.cursor]
				m.activeView = viewInspect
				m.inspectOffset = 0
				return m, fetchInspect(m.dockerClient, c.ID)
			}

		case "enter", "i": // Shell exec
			if m.cursor < len(m.filteredContainers) {
				c := m.filteredContainers[m.cursor]
				if c.State == "running" {
					cmd := exec.Command("docker", "exec", "-it", c.ID, "/bin/sh")
					return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
						return execDoneMsg{err}
					})
				}
			}

		case "o": // Open port in browser
			if m.cursor < len(m.filteredContainers) {
				c := m.filteredContainers[m.cursor]
				port := firstPublicPort(c.Ports)
				if port != "" {
					url := "http://localhost:" + port
					m.statusMsg = "Opening " + url
					return m, func() tea.Msg {
						exec.Command("xdg-open", url).Start()
						return openBrowserMsg{}
					}
				} else {
					m.statusMsg = "No public port found"
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		cmds := []tea.Cmd{
			waitForTick(),
			fetchContainers(m.dockerClient),
		}
		if m.showStats {
			cmds = append(cmds, fetchAllStats(m.dockerClient, m.filteredContainers))
		}
		// Decrement status message countdown
		if m.statusMsg != "" {
			m.statusTick--
			if m.statusTick <= 0 {
				m.statusMsg = ""
			}
		}
		return m, tea.Batch(cmds...)

	case animTickMsg:
		m.tick++
		return m, waitForAnimTick()

	case containersMsg:
		m.allContainers = msg
		m.applyFilter()
		if m.cursor >= len(m.filteredContainers) && len(m.filteredContainers) > 0 {
			m.cursor = len(m.filteredContainers) - 1
		} else if len(m.filteredContainers) == 0 {
			m.cursor = 0
		}

	case statsMsg:
		m.stats = msg
		if m.sortOrder == SortByCPU || m.sortOrder == SortByMem {
			m.applyFilter()
		}

	case actionMsg:
		if msg.err != nil {
			m.statusMsg = "Error: " + msg.err.Error()
		} else {
			m.statusMsg = "Done."
		}
		m.statusTick = 3
		return m, fetchContainers(m.dockerClient)

	case logLinesMsg:
		m.logLines = msg
		// Auto-scroll to bottom
		m.logOffset = len(m.logLines)

	case inspectMsg:
		if msg.err != nil {
			m.statusMsg = "Inspect error: " + msg.err.Error()
			m.activeView = viewContainers
		} else {
			m.inspectData = msg.data
		}

	case execDoneMsg:
		return m, fetchContainers(m.dockerClient)

	case openBrowserMsg:
		// nothing to do

	case errMsg:
		m.err = msg.err
	}

	return m, nil
}

type statsMsg map[string]Stats

func fetchAllStats(cli *client.Client, containers []Container) tea.Cmd {
	return func() tea.Msg {
		results := make(map[string]Stats)
		for _, c := range containers {
			if strings.HasPrefix(strings.ToLower(c.Status), "up") {
				stats, err := GetContainerStats(cli, c.ID)
				if err == nil {
					results[c.ID] = stats
				}
			}
		}
		return statsMsg(results)
	}
}

type animTickMsg time.Time

func waitForAnimTick() tea.Cmd {
	return tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
		return animTickMsg(t)
	})
}

// firstPublicPort extracts the first public (host) port from a ports string.
func firstPublicPort(ports string) string {
	if ports == "" {
		return ""
	}
	for _, part := range strings.Split(ports, ", ") {
		if strings.Contains(part, "->") {
			host := strings.Split(part, "->")[0]
			// strip protocol e.g. "0.0.0.0:8080" -> "8080"
			if idx := strings.LastIndex(host, ":"); idx >= 0 {
				return host[idx+1:]
			}
			return host
		}
	}
	return ""
}

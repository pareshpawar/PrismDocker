package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	widthID     = 14
	widthName   = 30
	widthImage  = 30
	widthStatus = 20
	widthPorts  = 40
)

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	// Dispatch to log viewer
	if m.activeView == viewLogs {
		return m.renderLogsView()
	}
	// Calculate dynamic widths based on terminal width
	// Total available width roughly: m.width - 4 (borders/padding)
	// We want to ensure at least some view.
	availableWidth := m.width - 10 // Safety margin for borders and padding
	if availableWidth < 40 {
		availableWidth = 40 // Minimum reasonable width
	}

	// Dynamic Widths
	var wID, wName, wImage, wStatus, wPorts, wCPU, wMem, wNet int

	if m.showStats {
		wID = 15
		wStatus = 12
		wCPU = 20
		wMem = 20
		wNet = 20

		available := availableWidth - wID - wStatus - wCPU - wMem - wNet
		if available < 20 {
			available = 20
		}

		wName = int(float64(available) * 0.5)
		wImage = available - wName
		wPorts = 0 // Hidden
	} else {
		// Standard View
		wID = 15
		wStatus = 20
		remaining := availableWidth - wID - wStatus
		if remaining < 0 {
			remaining = 0
		}
		wName = int(float64(remaining) * 0.35)
		wImage = int(float64(remaining) * 0.35)
		wPorts = remaining - wName - wImage

		if wName < 15 {
			wName = 15
		}
		if wImage < 15 {
			wImage = 15
		}
		if wPorts < 20 {
			wPorts = 20
		}
	}
	// Header logic
	runningCount := 0
	for _, c := range m.allContainers {
		if c.State == "running" {
			runningCount++
		}
	}

	// Prism ASCII Art Construction
	// We build it line by line to control colors
	// Idea:
	//      / \
	// --- /   \ === [ ]
	//     \   / === [ ]
	//      \ /  === [ ]

	lightColor := lipgloss.Color("255") // White
	prismColor := lipgloss.Color("63")  // Blue-ish/Cyan

	// Beams (Rainbow)
	beam1 := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("~") // Red
	beam2 := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render("~") // Orange
	beam3 := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("~") // Yellow
	beam4 := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render("~")  // Green
	beam5 := lipgloss.NewStyle().Foreground(lipgloss.Color("21")).Render("~")  // Blue
	beam6 := lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Render("~")  // Cyan
	beam7 := lipgloss.NewStyle().Foreground(lipgloss.Color("129")).Render("~") // Violet

	// Blocks (Containers)
	box1 := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("[]")
	box2 := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render("[]")
	box3 := lipgloss.NewStyle().Foreground(lipgloss.Color("21")).Render("[]")

	// Prism Shape
	pTop := lipgloss.NewStyle().Foreground(prismColor).Render("      / \\")
	pMid1 := lipgloss.NewStyle().Foreground(prismColor).Render("     /   \\")
	pMid2 := lipgloss.NewStyle().Foreground(prismColor).Render("    /     \\")
	pBot := lipgloss.NewStyle().Foreground(prismColor).Render("   /_______\\")

	// Incoming Light
	inRay := lipgloss.NewStyle().Foreground(lightColor).Render("---")

	// Composition
	// Line 1:       / \
	// Line 2:      /   \    ~~ []
	// Line 3: --- /     \ ~~~~ []
	// Line 4:    /_______\~~~~ []

	line1 := pTop

	// Ray out 1
	out1 := fmt.Sprintf("   %s%s%s %s", beam1, beam2, beam3, box1)
	line2 := pMid1 + out1

	// Ray out 2
	out2 := fmt.Sprintf(" %s%s%s %s", beam4, beam5, beam6, box2)
	line3 := inRay + pMid2 + out2

	// Ray out 3
	out3 := fmt.Sprintf("%s%s%s %s", beam7, beam1, beam2, box3)
	line4 := pBot + out3

	prismLogo := lipgloss.JoinVertical(lipgloss.Left, line1, line2, line3, line4)

	// Title next to it
	titleText := `    ____       _                 
   / __ \_____(_)________ ___    
  / /_/ / ___/ / ___/ __  __ \   
 / ____/ /  / (__  ) / / / / /   
/_/   /_/  /_/____/_/ /_/ /_/    `

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Render(titleText)

	// Combine Prism + Title
	fullLogo := lipgloss.JoinHorizontal(lipgloss.Bottom, prismLogo, "   ", title)

	stats := fmt.Sprintf("Running: %d | Total: %d", runningCount, len(m.allContainers))

	showStatus := "All"
	if !m.showAll {
		showStatus = "Running"
	}

	statsLabel := "OFF"
	if m.showStats {
		statsLabel = "ON"
	}
	statusInfo := fmt.Sprintf("Sort: %s | Show: %s | Stats: %s", m.sortOrder, showStatus, statsLabel)

	metaInfo := lipgloss.JoinVertical(lipgloss.Right,
		lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(stats),
		lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Render(statusInfo),
	)

	headerContent := lipgloss.JoinHorizontal(lipgloss.Bottom,
		fullLogo,
		lipgloss.NewStyle().Width(availableWidth-lipgloss.Width(fullLogo)).Align(lipgloss.Right).Render(metaInfo),
	)

	header := ListHeaderStyle.Render(headerContent)

	// Table Header
	var tHeaderCols string
	if m.showStats {
		tHeaderCols = lipgloss.JoinHorizontal(lipgloss.Top,
			ListItemStyle.Width(wID).Render("ID"),
			ListItemStyle.Width(wName).Render("Name"),
			ListItemStyle.Width(wImage).Render("Image"),
			ListItemStyle.Width(wStatus).Render("Status"),
			ListItemStyle.Width(wCPU).Render("CPU%"),
			ListItemStyle.Width(wMem).Render("MEM"),
			ListItemStyle.Width(wNet).Render("NET I/O"),
		)
	} else {
		tHeaderCols = lipgloss.JoinHorizontal(lipgloss.Top,
			ListItemStyle.Width(wID).Render("ID"),
			ListItemStyle.Width(wName).Render("Name"),
			ListItemStyle.Width(wImage).Render("Image"),
			ListItemStyle.Width(wStatus).Render("Status"),
			ListItemStyle.Width(wPorts).Render("Ports"),
		)
	}
	tHeader := tableStyle.Render(tHeaderCols)

	// Footer definition (moved up for height interp)
	// Footer
	footerText := "↑/k↓/j: Nav • r: Refresh • s: Sort • a: All/Running • t: Stats • S: Stop • u: Start • R: Restart • x: Remove • l: Logs • i: Shell • o: Open • q: Quit"
	if m.statusMsg != "" {
		footerText = m.statusMsg
	}
	footer := helpStyle.Render(footerText)

	// Calculate Heights
	headerH := lipgloss.Height(header)
	tHeaderH := lipgloss.Height(tHeader)
	footerH := lipgloss.Height(footer)

	bodyHeight := m.height - headerH - tHeaderH - footerH
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	// Viewport Slicing
	start := m.tableOffset
	end := start + bodyHeight
	if end > len(m.filteredContainers) {
		end = len(m.filteredContainers)
	}

	// Table Rows
	var rows []string

	if len(m.filteredContainers) == 0 {
		rows = append(rows, "No containers found.")
	} else {
		for i := start; i < end; i++ {
			c := m.filteredContainers[i]
			cursor := "  "
			if m.cursor == i {
				cursor = "> "
			}

			// Style (Zebra + Selection)
			style := ListItemStyle
			if m.cursor == i {
				style = selectedStyle
			} else if i%2 == 0 {
				style = style.Copy().Background(lipgloss.Color("235")) // Zebra stripe
			}

			status := minifyStatus(c.Status)
			if c.State == "running" {
				status = statusUpStyle.Render(status)
			} else {
				status = statusExitedStyle.Render(status)
			}

			// Truncate/Scroll logic
			padding := 2

			isSelected := m.cursor == i

			// ID Style
			idVal := c.ID
			if !isSelected {
				idVal = idStyle.Render(idVal)
			}
			id := style.Width(wID).Render(idVal)

			// Name: Scroll effectively
			nameRaw := c.Names
			if isSelected {
				nameRaw = scrollText(nameRaw, wName-padding, m.tick)
			} else {
				nameRaw = truncate(nameRaw, wName-padding)
			}
			name := style.Width(wName).Render(nameRaw)

			// Image: Scroll effectively
			imageRaw := c.Image
			if isSelected {
				imageRaw = scrollText(imageRaw, wImage-padding, m.tick)
			} else {
				imageRaw = truncate(imageRaw, wImage-padding)
				if !isSelected {
					imageRaw = imageStyle.Render(imageRaw)
				}
			}
			image := style.Width(wImage).Render(imageRaw)

			stat := style.Width(wStatus).Render(status)

			var row string
			if m.showStats {
				// Stats columns with progress bars
				s := m.stats[c.ID]
				var cpuStr, memStr, netStr string
				if c.State != "running" {
					cpuStr = "-"
					memStr = "-"
					netStr = "-"
				} else {
					cpuStr = renderBar(s.CPUPercent, 8) + fmt.Sprintf(" %.1f%%", s.CPUPercent)
					memPct := 0.0
					if s.MemLimit > 0 {
						memPct = s.MemUsage / s.MemLimit * 100
					}
					// Alert: yellow >80%, red >95%
					if memPct > 95 {
						style = style.Copy().Background(lipgloss.Color("196"))
					} else if memPct > 80 {
						style = style.Copy().Background(lipgloss.Color("214"))
					}
					memStr = renderBar(memPct, 8) + fmt.Sprintf(" %s/%s", formatBytesShort(s.MemUsage), formatBytesShort(s.MemLimit))
					netStr = fmt.Sprintf("%s↑%s↓", formatBytesShort(s.NetTx), formatBytesShort(s.NetRx))
				}
				cpuCol := style.Width(wCPU).Render(cpuStr)
				memCol := style.Width(wMem).Render(memStr)
				netCol := style.Width(wNet).Render(netStr)
				row = lipgloss.JoinHorizontal(lipgloss.Top,
					style.Render(cursor),
					id,
					name,
					image,
					stat,
					cpuCol,
					memCol,
					netCol,
				)
			} else {
				// Ports: Discrete Scroll
				rawPorts := c.Ports
				visibleWidth := wPorts - padding
				var finalPort string
				if isSelected {
					scrolled := scrollPorts(rawPorts, visibleWidth, m.tick)
					finalPort = style.Width(wPorts).Render(scrolled)
				} else {
					truncated := truncatePorts(rawPorts, visibleWidth)
					finalPort = style.Width(wPorts).Render(truncated)
				}
				row = lipgloss.JoinHorizontal(lipgloss.Top,
					style.Render(cursor),
					id,
					name,
					image,
					stat,
					finalPort,
				)
			}
			rows = append(rows, row)
		}
	}

	// Fill empty space if rows < bodyHeight
	if len(rows) < bodyHeight {
		fill := bodyHeight - len(rows)
		for j := 0; j < fill; j++ {
			rows = append(rows, "")
		}
	}

	body := lipgloss.JoinVertical(lipgloss.Left, rows...)

	// Base view
	base := lipgloss.JoinVertical(lipgloss.Left,
		header,
		tHeader,
		body,
		footer,
	)

	// Overlay confirm popup if needed
	if m.confirmMode {
		return renderConfirmPopup(base, m.width, m.height)
	}
	return base
}

func formatPorts(p string) string {
	if p == "" {
		return ""
	}
	parts := strings.Split(p, ", ")
	var styledParts []string
	for _, part := range parts {
		if strings.Contains(part, "->") {
			sub := strings.Split(part, "->")
			if len(sub) == 2 {
				// Colorize left (Host) and right (Container)
				host := portHostStyle.Render(sub[0])
				container := portContainerStyle.Render(sub[1])
				styledParts = append(styledParts, fmt.Sprintf("%s->%s", host, container))
			} else {
				styledParts = append(styledParts, part)
			}
		} else {
			styledParts = append(styledParts, part)
		}
	}
	return strings.Join(styledParts, ", ")
}

func truncate(s string, max int) string {
	if len(s) > max {
		if max > 3 {
			return s[:max-3] + "..."
		}
		return s[:max]
	}
	return s
}

func scrollText(text string, width int, tick int) string {
	if len(text) <= width {
		return text
	}
	gap := 5
	full := text + strings.Repeat(" ", gap)
	offset := (tick / 2) % len(full) // Speed control: 1 char per 2 ticks (400ms)
	s := full[offset:] + full[:offset]
	if len(s) > width {
		s = s[:width]
	}
	return s
}

func scrollPorts(ports string, width int, tick int) string {
	parts := strings.Split(ports, ", ")
	if len(parts) == 0 {
		return ""
	}

	// Rotate
	n := len(parts)
	offset := (tick / 10) % n // Switch every 2s (10 ticks)
	rotated := make([]string, n)
	copy(rotated, parts[offset:])
	copy(rotated[n-offset:], parts[:offset])

	// Assemble fitting parts
	var displayParts []string
	currentLen := 0
	for _, p := range rotated {
		// +2 for comma space, roughly
		// Naive length check, doesn't account for ANSI, but inputs are raw here.
		flen := len(p)
		if currentLen == 0 {
			// First item
			displayParts = append(displayParts, p)
			currentLen += flen
		} else {
			if currentLen+2+flen > width {
				break
			}
			displayParts = append(displayParts, p)
			currentLen += 2 + flen
		}
	}

	raw := strings.Join(displayParts, ", ")
	return formatPorts(raw)
}

func truncatePorts(ports string, width int) string {
	if len(ports) <= width {
		return formatPorts(ports)
	}

	parts := strings.Split(ports, ", ")
	var displayParts []string
	currentLen := 0

	for _, p := range parts {
		partLen := len(p)
		sepLen := 2 // ", "
		if currentLen == 0 {
			sepLen = 0
		}

		// We need room for this part + separator + "..." (3)
		if currentLen+sepLen+partLen+3 > width {
			// Can't fit
			return formatPorts(strings.Join(displayParts, ", ")) + "..."
		}

		displayParts = append(displayParts, p)
		currentLen += sepLen + partLen
	}

	return formatPorts(strings.Join(displayParts, ", "))
}

func formatBytes(b float64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%.1f B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", b/float64(div), "kMGTPE"[exp])
}

// formatBytesShort is a compact version: 128M, 1.2G, 512k
func formatBytesShort(b float64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%.0fB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	val := b / float64(div)
	if val >= 10 {
		return fmt.Sprintf("%.0f%c", val, "KMGTP"[exp])
	}
	return fmt.Sprintf("%.1f%c", val, "KMGTP"[exp])
}

// minifyStatus shortens Docker status strings for compact display.
// e.g. "Up 3 hours" -> "Up 3h", "Exited (0) 2 days ago" -> "Exit 2d"
func minifyStatus(s string) string {
	replacer := strings.NewReplacer(
		" seconds", "s",
		" second", "s",
		" minutes", "m",
		" minute", "m",
		" hours", "h",
		" hour", "h",
		" days", "d",
		" day", "d",
		" weeks", "w",
		" week", "w",
		" months", "mo",
		" month", "mo",
		" ago", "",
		"Exited", "Exit",
		"Created", "New",
		"Restarting", "Restart",
		"Paused", "Paused",
	)
	return replacer.Replace(s)
}

// renderBar renders an ASCII progress bar like [■■■□□□□□] for a given percentage.
func renderBar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int(pct / 100 * float64(width))
	bar := strings.Repeat("■", filled) + strings.Repeat("□", width-filled)

	var color lipgloss.Color
	switch {
	case pct > 95:
		color = lipgloss.Color("196") // Red
	case pct > 80:
		color = lipgloss.Color("214") // Orange
	case pct > 60:
		color = lipgloss.Color("226") // Yellow
	default:
		color = lipgloss.Color("46") // Green
	}
	return lipgloss.NewStyle().Foreground(color).Render("[" + bar + "]")
}

// renderLogsView renders the full-screen log viewer.
func (m model) renderLogsView() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	title := titleStyle.Render(fmt.Sprintf("Logs: %s", m.logContainer))

	filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	filterBar := ""
	if m.logFilterMode {
		filterBar = filterStyle.Render(fmt.Sprintf("Filter: %s█", m.logFilter))
	} else if m.logFilter != "" {
		filterBar = filterStyle.Render(fmt.Sprintf("Filter: %s  (/ to edit)", m.logFilter))
	}

	footerStr := "Esc/q: Back • /: Filter • ↑/k↓/j: Scroll"
	footer := helpStyle.Render(footerStr)

	headerH := lipgloss.Height(title) + 1
	footerH := lipgloss.Height(footer)
	filterH := 0
	if filterBar != "" {
		filterH = 1
	}
	bodyH := m.height - headerH - footerH - filterH - 1
	if bodyH < 1 {
		bodyH = 1
	}

	// Filter lines
	lines := m.logLines
	if m.logFilter != "" {
		var filtered []string
		for _, l := range lines {
			if strings.Contains(strings.ToLower(l), strings.ToLower(m.logFilter)) {
				filtered = append(filtered, l)
			}
		}
		lines = filtered
	}

	// Scroll
	offset := m.logOffset
	if offset > len(lines)-bodyH {
		offset = len(lines) - bodyH
	}
	if offset < 0 {
		offset = 0
	}
	end := offset + bodyH
	if end > len(lines) {
		end = len(lines)
	}
	visible := lines[offset:end]

	// Pad
	for len(visible) < bodyH {
		visible = append(visible, "")
	}

	body := strings.Join(visible, "\n")

	parts := []string{title, body}
	if filterBar != "" {
		parts = append(parts, filterBar)
	}
	parts = append(parts, footer)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderConfirmPopup overlays a centered confirmation dialog on top of the base view.
func renderConfirmPopup(base string, width, height int) string {
	popupStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 4).
		Bold(true)

	msg := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("⚠  Remove this container?") +
		"\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Render("  [y] Yes    [n] No / Esc")

	popup := popupStyle.Render(msg)
	pw := lipgloss.Width(popup)
	ph := lipgloss.Height(popup)

	left := (width - pw) / 2
	top := (height - ph) / 2
	if left < 0 {
		left = 0
	}
	if top < 0 {
		top = 0
	}

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, popup,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.AdaptiveColor{Light: "0", Dark: "0"}),
	)
}

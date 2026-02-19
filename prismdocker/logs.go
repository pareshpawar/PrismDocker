package main

import (
	"bufio"
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/moby/moby/client"
)

type logLinesMsg []string

func fetchLogs(cli *client.Client, containerID string) tea.Cmd {
	return func() tea.Msg {
		rc, err := cli.ContainerLogs(context.Background(), containerID, client.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Tail:       "500",
		})
		if err != nil {
			return logLinesMsg([]string{"Error fetching logs: " + err.Error()})
		}
		defer rc.Close()

		var lines []string
		scanner := bufio.NewScanner(rc)
		for scanner.Scan() {
			line := scanner.Text()
			// Docker multiplexes stdout/stderr with an 8-byte header; strip it
			if len(line) > 8 {
				line = line[8:]
			}
			lines = append(lines, line)
		}
		if len(lines) == 0 {
			lines = []string{"(no logs)"}
		}
		return logLinesMsg(lines)
	}
}

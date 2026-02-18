# 🌈 PrismDocker

> Taking the "white light" (plain text) of Docker and breaking it into a spectrum.

A beautiful, fast terminal UI for Docker container management — built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

```
      / \                  ____       _
     /   \   ~~~ []       / __ \_____(_)________ ___
---    /     \ ~~~ []    / /_/ / ___/ / ___/ __  __ \
   /_______\~~~ []      / ____/ /  / (__  ) / / / / /
                       /_/   /_/  /_/____/_/ /_/ /_/
```

## Features

- 📋 **Live container list** — auto-refreshes every 2 seconds
- 📊 **Stats mode** — real-time CPU%, memory, and network I/O per container
- 🎨 **Color-coded status** — running containers in green, exited in red
- 🔀 **Multi-column sorting** — sort by ID, Name, Image, State, CPU%, or Memory
- 🔍 **Filter view** — toggle between running-only and all containers
- 📜 **Scrolling text** — long names/images/ports scroll when selected
- 🦓 **Zebra striping** — alternating row backgrounds for readability
- ⌨️ **Vim-style navigation** — `j`/`k` or arrow keys

## Installation

### From source

```bash
git clone https://github.com/pareshpawar/PrismDocker.git
cd PrismDocker/prismdocker
go build -o prismdocker .
./prismdocker
```

### With `go install`

```bash
go install github.com/pareshpawar/PrismDocker/prismdocker@latest
```

> Requires Go 1.24+ and a running Docker daemon.

## Keybindings

| Key       | Action                                      |
|-----------|---------------------------------------------|
| `↑` / `k` | Move cursor up                              |
| `↓` / `j` | Move cursor down                            |
| `r`       | Manual refresh                              |
| `s`       | Cycle sort order (ID → Name → Image → State → CPU% → Mem) |
| `a`       | Toggle All / Running-only view              |
| `t`       | Toggle stats mode (CPU%, Mem, Net I/O)      |
| `q`       | Quit                                        |

> **Note:** CPU% and Mem sort options are only available when stats mode is on (`t`).

## Stats Mode

Press `t` to enable live stats. The Ports column is replaced with:

| Column  | Description                        |
|---------|------------------------------------|
| `CPU%`  | CPU usage percentage                |
| `MEM`   | Memory usage / limit (e.g. `128M/512M`) |
| `NET`   | Network Tx↑ / Rx↓ (e.g. `1.2M↑3.4M↓`) |

Stopped containers show `-` for all stats columns.

## Requirements

- Go 1.24+
- Docker daemon running locally (or accessible via `DOCKER_HOST`)

## License

MIT

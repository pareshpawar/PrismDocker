# 🌈 PrismDocker

> A beautiful, fast terminal UI for Docker container management — built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

```
      / \                  ____       _
     /   \   ~~~ []       / __ \_____(_)________ ___
---    /     \ ~~~ []    / /_/ / ___/ / ___/ __  __ \
   /_______\~~~ []      / ____/ /  / (__  ) / / / / /
                       /_/   /_/  /_/____/_/ /_/ /_/
```

## Features

- 📋 **Live container list** — auto-refreshes every 2 seconds
- 📊 **Stats mode** — real-time CPU%, memory usage, and network I/O per container with color-coded progress bars
- 🚦 **Row alerting** — rows turn yellow when memory > 80%, red when > 95%
- 🎨 **Color-coded status** — running containers in green, stopped in red
- 🔀 **Multi-column sorting** — sort by ID, Name, Image, State, CPU%, or Memory
- 🔍 **All / Running toggle** — show all containers or only running ones
- 📜 **Log viewer** — press `l` to view live container logs with a built-in search/filter bar
- 🐚 **Shell exec** — press `i` to drop straight into a shell inside a container; TUI resumes on exit
- ⚡ **Quick actions** — stop, start, restart, and remove containers with single keystrokes
- 🌐 **Open in browser** — press `o` to open a container's exposed port in your default browser
- 🦓 **Zebra striping** — alternating row backgrounds for readability
- ⌨️ **Vim-style navigation** — `j`/`k` or arrow keys

## Installation

### Homebrew (macOS & Linux)

```bash
brew tap pareshpawar/tap
brew install pareshpawar/tap/prism
```

### With `go install`

```bash
go install github.com/pareshpawar/PrismDocker/prismdocker@latest
```

### From source

```bash
git clone https://github.com/pareshpawar/PrismDocker.git
cd PrismDocker/prismdocker
go build -o prism .
./prism
```

> Requires Go 1.24+ and a running Docker daemon.

## Keybindings

### Navigation

| Key       | Action                       |
|-----------|------------------------------|
| `↑` / `k` | Move cursor up               |
| `↓` / `j` | Move cursor down             |
| `r`       | Manual refresh               |
| `q`       | Quit                         |

### View

| Key | Action                                                       |
|-----|--------------------------------------------------------------|
| `s` | Cycle sort order: ID → Name → Image → State → CPU% → Mem     |
| `a` | Toggle All / Running-only view                               |
| `t` | Toggle stats mode (CPU%, Mem, Net I/O)                       |

> **Note:** CPU% and Mem sort options are only available when stats mode is on (`t`).

### Container Actions

| Key   | Action                                              |
|-------|-----------------------------------------------------|
| `S`   | Stop the highlighted container                      |
| `u`   | Start (up) the highlighted container                |
| `R`   | Restart the highlighted container                   |
| `x`   | Remove container — shows a confirmation popup first |
| `l`   | Open log viewer (last 500 lines)                    |
| `i` / `Enter` | Drop into a shell inside the container (`/bin/sh`) |
| `o`   | Open the container's first public port in browser   |

### Log Viewer

| Key        | Action               |
|------------|----------------------|
| `Esc` / `q` | Return to container list |
| `/`        | Toggle filter mode   |
| `↑` / `k`  | Scroll up            |
| `↓` / `j`  | Scroll down          |

## Stats Mode

Press `t` to enable live stats. The Ports column is replaced with:

| Column  | Description                                      |
|---------|--------------------------------------------------|
| `CPU%`  | CPU usage % with color-coded progress bar         |
| `MEM`   | Memory usage / limit (e.g. `128M/512M`)           |
| `NET`   | Network Tx↑ / Rx↓ (e.g. `1.2M↑3.4M↓`)           |

Stopped containers show `-` for all stats columns.  
Rows turn **yellow** when memory > 80%, **red** when memory > 95%.

## Requirements

- Go 1.24+ (for building from source)
- Docker daemon running locally (or accessible via `DOCKER_HOST`)

## Author

Built by [Paresh Pawar](https://github.com/pareshpawar).

## License

[MIT](./LICENSE)

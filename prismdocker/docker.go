package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/moby/moby/client"
)

type Container struct {
	ID             string
	Names          string
	Image          string
	Status         string
	State          string // "running", "exited", etc.
	Ports          string
	ComposeProject string // from com.docker.compose.project label
	ComposeService string // from com.docker.compose.service label
}

func NewDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

func ListContainers(cli *client.Client) ([]Container, error) {
	// Use client.ContainerListOptions as indicated by go doc.
	// If this fails, we will try types.ContainerListOptions.
	containers, err := cli.ContainerList(context.Background(), client.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}

	// Range over containers.Items
	var result []Container
	for _, c := range containers.Items {
		// Names are returned as "/name", so we strip the leading slash
		names := strings.Join(c.Names, ", ")
		names = strings.ReplaceAll(names, "/", "")

		// Format ports with deduplication
		portSet := make(map[string]struct{})
		var ports []string
		for _, p := range c.Ports {
			// Ignore IP in the key to deduplicate 0.0.0.0 vs :::
			fmtStr := fmt.Sprintf("%d->%d/%s", p.PublicPort, p.PrivatePort, p.Type)
			if p.PublicPort == 0 {
				fmtStr = fmt.Sprintf("%d/%s", p.PrivatePort, p.Type)
			}

			if _, exists := portSet[fmtStr]; !exists {
				portSet[fmtStr] = struct{}{}
				ports = append(ports, fmtStr)
			}
		}

		result = append(result, Container{
			ID:             c.ID[:12],
			Names:          names,
			Image:          c.Image,
			Status:         c.Status,
			State:          string(c.State),
			Ports:          strings.Join(ports, ", "),
			ComposeProject: c.Labels["com.docker.compose.project"],
			ComposeService: c.Labels["com.docker.compose.service"],
		})
	}
	return result, nil
}

// Stats represents minimal container statistics
type Stats struct {
	CPUPercent float64
	MemUsage   float64 // bytes
	MemLimit   float64 // bytes
	NetRx      float64 // bytes
	NetTx      float64 // bytes
}

// Docker stats JSON structure (simplified)
type statsJSON struct {
	Read    string `json:"read"`
	Preread string `json:"preread"`
	Pids    struct {
		Current int `json:"current"`
	} `json:"pids_stats"`
	Blkio struct {
		IoServiceBytesRecursive []struct {
			Major int    `json:"major"`
			Minor int    `json:"minor"`
			Op    string `json:"op"`
			Value int    `json:"value"`
		} `json:"io_service_bytes_recursive"`
	} `json:"blkio_stats"`
	CPU struct {
		CPUUsage struct {
			TotalUsage        uint64 `json:"total_usage"`
			UsageInKernelmode uint64 `json:"usage_in_kernelmode"`
			UsageInUsermode   uint64 `json:"usage_in_usermode"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs     uint32 `json:"online_cpus"`
	} `json:"cpu_stats"`
	PreCPU struct {
		CPUUsage struct {
			TotalUsage        uint64 `json:"total_usage"`
			UsageInKernelmode uint64 `json:"usage_in_kernelmode"`
			UsageInUsermode   uint64 `json:"usage_in_usermode"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs     uint32 `json:"online_cpus"`
	} `json:"precpu_stats"`
	Memory struct {
		Usage    uint64 `json:"usage"`
		MaxUsage uint64 `json:"max_usage"`
		Stats    struct {
			ActiveAnon              uint64 `json:"active_anon"`
			ActiveFile              uint64 `json:"active_file"`
			Cache                   uint64 `json:"cache"`
			Dirty                   uint64 `json:"dirty"`
			HierarchicalMemoryLimit uint64 `json:"hierarchical_memory_limit"`
			InactiveAnon            uint64 `json:"inactive_anon"`
			InactiveFile            uint64 `json:"inactive_file"`
			MappedFile              uint64 `json:"mapped_file"`
			Pgfault                 uint64 `json:"pgfault"`
			Pgmajfault              uint64 `json:"pgmajfault"`
			Pgpgin                  uint64 `json:"pgpgin"`
			Pgpgout                 uint64 `json:"pgpgout"`
			Rss                     uint64 `json:"rss"`
			RssHuge                 uint64 `json:"rss_huge"`
			TotalActiveAnon         uint64 `json:"total_active_anon"`
			TotalActiveFile         uint64 `json:"total_active_file"`
			TotalCache              uint64 `json:"total_cache"`
			TotalDirty              uint64 `json:"total_dirty"`
			TotalInactiveAnon       uint64 `json:"total_inactive_anon"`
			TotalInactiveFile       uint64 `json:"total_inactive_file"`
			TotalMappedFile         uint64 `json:"total_mapped_file"`
			TotalPgfault            uint64 `json:"total_pgfault"`
			TotalPgmajfault         uint64 `json:"total_pgmajfault"`
			TotalPgpgin             uint64 `json:"total_pgpgin"`
			TotalPgpgout            uint64 `json:"total_pgpgout"`
			TotalRss                uint64 `json:"total_rss"`
			TotalRssHuge            uint64 `json:"total_rss_huge"`
			TotalUnevictable        uint64 `json:"total_unevictable"`
			TotalWriteback          uint64 `json:"total_writeback"`
			Unevictable             uint64 `json:"unevictable"`
			Writeback               uint64 `json:"writeback"`
		} `json:"stats"`
		Limit uint64 `json:"limit"`
	} `json:"memory_stats"`
	Networks map[string]struct {
		RxBytes   uint64 `json:"rx_bytes"`
		RxPackets uint64 `json:"rx_packets"`
		TxBytes   uint64 `json:"tx_bytes"`
		TxPackets uint64 `json:"tx_packets"`
	} `json:"networks"`
}

func calculateCPUPercent(v statsJSON) float64 {
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(v.CPU.CPUUsage.TotalUsage) - float64(v.PreCPU.CPUUsage.TotalUsage)
		// calculate the change for the entire system between readings
		systemDelta = float64(v.CPU.SystemCPUUsage) - float64(v.PreCPU.SystemCPUUsage)
		onlineCPUs  = float64(v.CPU.OnlineCPUs)
	)

	if onlineCPUs == 0.0 {
		// Fallback if online_cpus is missing (unlikely in newer docker)
		// UsageInKernelmode is typically uint64. If we don't have online CPUs, assume 1.
		onlineCPUs = 1.0
	}

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0
	}
	return cpuPercent
}

func calculateMemUsage(v statsJSON) float64 {
	// Mem usage is slightly different in cgroup v1 vs v2.
	// Common docker stats: Usage - Cache (v1)
	// For simplicity, using Usage.
	// But `docker stats` often does usage - cache for some versions.
	// Let's use Usage for now.
	return float64(v.Memory.Usage)
}

func calculateNetIO(v statsJSON) (float64, float64) {
	var rx, tx float64
	for _, n := range v.Networks {
		rx += float64(n.RxBytes)
		tx += float64(n.TxBytes)
	}
	return rx, tx
}

// GetContainerStats fetches stats for a single container.
// It uses stream=false to get a one-shot snapshot.
func GetContainerStats(cli *client.Client, containerID string) (Stats, error) {
	resp, err := cli.ContainerStats(context.Background(), containerID, client.ContainerStatsOptions{Stream: false})
	if err != nil {
		return Stats{}, err
	}
	defer resp.Body.Close()

	var v statsJSON
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return Stats{}, err
	}

	return Stats{
		CPUPercent: calculateCPUPercent(v),
		MemUsage:   calculateMemUsage(v),
		MemLimit:   float64(v.Memory.Limit),
		NetRx:      func() float64 { r, _ := calculateNetIO(v); return r }(),
		NetTx:      func() float64 { _, t := calculateNetIO(v); return t }(),
	}, nil
}

func StopContainer(cli *client.Client, containerID string) error {
	_, err := cli.ContainerStop(context.Background(), containerID, client.ContainerStopOptions{})
	return err
}

func StartContainer(cli *client.Client, containerID string) error {
	_, err := cli.ContainerStart(context.Background(), containerID, client.ContainerStartOptions{})
	return err
}

func RestartContainer(cli *client.Client, containerID string) error {
	_, err := cli.ContainerRestart(context.Background(), containerID, client.ContainerRestartOptions{})
	return err
}

func RemoveContainer(cli *client.Client, containerID string) error {
	_, err := cli.ContainerRemove(context.Background(), containerID, client.ContainerRemoveOptions{Force: true})
	return err
}

// ContainerInspect holds detailed information about a container.
type ContainerInspect struct {
	ID            string
	Name          string
	Image         string
	Created       string
	State         string
	RestartPolicy string
	Env           []string
	Mounts        []MountInfo
	Networks      []NetworkInfo
	Labels        map[string]string
	Ports         string
	Cmd           []string
	Entrypoint    []string
}

type MountInfo struct {
	Source      string
	Destination string
	Mode        string
	Type        string
}

type NetworkInfo struct {
	Name      string
	IPAddress string
	Gateway   string
}

func InspectContainer(cli *client.Client, containerID string) (ContainerInspect, error) {
	result, err := cli.ContainerInspect(context.Background(), containerID, client.ContainerInspectOptions{})
	if err != nil {
		return ContainerInspect{}, err
	}
	info := result.Container

	var mounts []MountInfo
	for _, m := range info.Mounts {
		mounts = append(mounts, MountInfo{
			Source:      m.Source,
			Destination: m.Destination,
			Mode:        m.Mode,
			Type:        string(m.Type),
		})
	}

	var networks []NetworkInfo
	if info.NetworkSettings != nil {
		for name, net := range info.NetworkSettings.Networks {
			networks = append(networks, NetworkInfo{
				Name:      name,
				IPAddress: net.IPAddress.String(),
				Gateway:   net.Gateway.String(),
			})
		}
	}

	// Format ports
	var portStrs []string
	if info.NetworkSettings != nil {
		for port, bindings := range info.NetworkSettings.Ports {
			for _, b := range bindings {
				portStrs = append(portStrs, fmt.Sprintf("%s:%s->%s", b.HostIP, b.HostPort, port.String()))
			}
			if len(bindings) == 0 {
				portStrs = append(portStrs, port.String())
			}
		}
	}

	restartPolicy := ""
	if info.HostConfig != nil {
		restartPolicy = string(info.HostConfig.RestartPolicy.Name)
	}

	created := info.Created
	if len(created) > 19 {
		created = created[:19] // Trim to "2024-01-01T00:00:00"
	}

	stateStr := ""
	if info.State != nil {
		stateStr = string(info.State.Status)
	}

	var env, cmd, entrypoint []string
	var labels map[string]string
	var image string
	if info.Config != nil {
		env = info.Config.Env
		labels = info.Config.Labels
		image = info.Config.Image
		cmd = info.Config.Cmd
		entrypoint = info.Config.Entrypoint
	}

	id := info.ID
	if len(id) > 12 {
		id = id[:12]
	}

	return ContainerInspect{
		ID:            id,
		Name:          strings.TrimPrefix(info.Name, "/"),
		Image:         image,
		Created:       created,
		State:         stateStr,
		RestartPolicy: restartPolicy,
		Env:           env,
		Mounts:        mounts,
		Networks:      networks,
		Labels:        labels,
		Ports:         strings.Join(portStrs, ", "),
		Cmd:           cmd,
		Entrypoint:    entrypoint,
	}, nil
}

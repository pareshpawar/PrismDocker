package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/moby/moby/client"
)

type Container struct {
	ID     string
	Names  string
	Image  string
	Status string
	State  string // "running", "exited", etc.
	Ports  string
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
			ID:     c.ID[:12],
			Names:  names,
			Image:  c.Image,
			Status: c.Status,
			State:  string(c.State),
			Ports:  strings.Join(ports, ", "),
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
	// Use explicit struct as requested by compiler
	// Assuming Stream: false gives us one shot?
	// If Stream is false, it might return empty body?
	// ContainerStats doc: "Get container stats. One-shot if stream is false."

	// We need to verify if `client.ContainerStatsOptions` has a `Stream` field.
	// Most likely yes.

	// Note: `client` is `github.com/moby/moby/client`.
	// If types are missing we might need `github.com/docker/docker/api/types/container`.
	// But the error explicitly said `client.ContainerStatsOptions`.

	// Let's assume it works.
	// But we can't use `client.ContainerStatsOptions` literal if we aren't sure of fields.
	// Assuming it's `Stream bool`.

	// Wait, I cannot construct `client.ContainerStatsOptions` if I don't know fields.
	// But the error output implies that type EXISTS in `client`.

	// Let's try to verify via `go doc` equivalent? I can't.
	// I'll take a leap of faith that it has a Stream field.

	// Wait! `ContainerStats` in `moby/client` typically takes `Stream bool` directly in v1.xx
	// BUT `container` (api/types) has `StatsOptions`.
	// Why does `client` package have `ContainerStatsOptions`?
	// Maybe it's an alias?

	// Let's just try to compile with `Stream: false`.

	// But wait! If I just pass `true` or `false` I get the error.
	// So I MUST use the struct.

	// IMPORTANT: I need to handle `view.go` redeclares too.

	// Let's do `docker.go` first.

	// Correction: I replaced the whole function body with comments in previous step.
	// I need to put back the logic!

	// Oh, I see that I replaced the func with a placeholder `return Stats{}, fmt.Errorf...`
	// I need to restore the full body.

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
		NetRx:      func() float64 { r, _ := calculateNetIO(v); return r }(), // simple wrapper
		NetTx:      func() float64 { _, t := calculateNetIO(v); return t }(),
	}, nil
}

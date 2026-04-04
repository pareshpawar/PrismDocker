package main

import (
	"sort"
	"strings"
)

func sortAndFilter(containers []Container, order SortOrder, showAll bool, stats map[string]Stats, searchQuery ...string) []Container {
	query := ""
	if len(searchQuery) > 0 {
		query = strings.ToLower(searchQuery[0])
	}

	var filtered []Container
	for _, c := range containers {
		if !showAll && c.State != "running" {
			continue
		}
		if query != "" {
			lower := strings.ToLower
			if !strings.Contains(lower(c.Names), query) &&
				!strings.Contains(lower(c.Image), query) &&
				!strings.Contains(lower(c.ID), query) {
				continue
			}
		}
		filtered = append(filtered, c)
	}

	sort.Slice(filtered, func(i, j int) bool {
		switch order {
		case SortByID:
			return filtered[i].ID < filtered[j].ID
		case SortByName:
			return filtered[i].Names < filtered[j].Names
		case SortByImage:
			return filtered[i].Image < filtered[j].Image
		case SortByState:
			// Sort by state (running first), then by name
			if filtered[i].State != filtered[j].State {
				if filtered[i].State == "running" {
					return true
				}
				if filtered[j].State == "running" {
					return false
				}
				return filtered[i].State < filtered[j].State
			}
			return filtered[i].Names < filtered[j].Names
		case SortByCPU:
			si := stats[filtered[i].ID]
			sj := stats[filtered[j].ID]
			return si.CPUPercent > sj.CPUPercent // descending
		case SortByMem:
			si := stats[filtered[i].ID]
			sj := stats[filtered[j].ID]
			return si.MemUsage > sj.MemUsage // descending
		default:
			return filtered[i].ID < filtered[j].ID
		}
	})

	return filtered
}

// sortAndFilterWithCompose sorts containers with compose project grouping.
// Containers are first grouped by compose project, then sorted within each group.
func sortAndFilterWithCompose(containers []Container, order SortOrder, showAll bool, stats map[string]Stats, searchQuery string) []Container {
	filtered := sortAndFilter(containers, order, showAll, stats, searchQuery)

	// Stable sort by compose project (preserving inner sort order)
	sort.SliceStable(filtered, func(i, j int) bool {
		pi := filtered[i].ComposeProject
		pj := filtered[j].ComposeProject
		if pi == "" {
			pi = "~standalone" // Sort standalone last
		}
		if pj == "" {
			pj = "~standalone"
		}
		return pi < pj
	})

	return filtered
}

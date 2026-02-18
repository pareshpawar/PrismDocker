package main

import (
	"sort"
)

func sortAndFilter(containers []Container, order SortOrder, showAll bool, stats map[string]Stats) []Container {
	var filtered []Container
	for _, c := range containers {
		if !showAll && c.State != "running" {
			continue
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

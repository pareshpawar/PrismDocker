package main

import (
	"testing"
)

func TestSortAndFilter_EmptyList(t *testing.T) {
	result := sortAndFilter(nil, SortByName, true, nil)
	if len(result) != 0 {
		t.Errorf("expected empty, got %d items", len(result))
	}
}

func TestSortAndFilter_RunningOnly(t *testing.T) {
	containers := []Container{
		{ID: "aaa", Names: "web", State: "running"},
		{ID: "bbb", Names: "db", State: "exited"},
		{ID: "ccc", Names: "cache", State: "running"},
	}
	result := sortAndFilter(containers, SortByName, false, nil)
	if len(result) != 2 {
		t.Fatalf("expected 2 running, got %d", len(result))
	}
	for _, c := range result {
		if c.State != "running" {
			t.Errorf("expected running, got %s for %s", c.State, c.Names)
		}
	}
}

func TestSortAndFilter_ShowAll(t *testing.T) {
	containers := []Container{
		{ID: "aaa", Names: "web", State: "running"},
		{ID: "bbb", Names: "db", State: "exited"},
	}
	result := sortAndFilter(containers, SortByName, true, nil)
	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
}

func TestSortAndFilter_SortByName(t *testing.T) {
	containers := []Container{
		{ID: "ccc", Names: "zeta", State: "running"},
		{ID: "aaa", Names: "alpha", State: "running"},
		{ID: "bbb", Names: "beta", State: "running"},
	}
	result := sortAndFilter(containers, SortByName, true, nil)
	if result[0].Names != "alpha" || result[1].Names != "beta" || result[2].Names != "zeta" {
		t.Errorf("unexpected sort order: %s, %s, %s", result[0].Names, result[1].Names, result[2].Names)
	}
}

func TestSortAndFilter_SortByState(t *testing.T) {
	containers := []Container{
		{ID: "aaa", Names: "db", State: "exited"},
		{ID: "bbb", Names: "web", State: "running"},
	}
	result := sortAndFilter(containers, SortByState, true, nil)
	if result[0].State != "running" {
		t.Errorf("expected running first, got %s", result[0].State)
	}
}

func TestSortAndFilter_SearchQuery(t *testing.T) {
	containers := []Container{
		{ID: "aaa", Names: "web-server", Image: "nginx", State: "running"},
		{ID: "bbb", Names: "database", Image: "postgres", State: "running"},
		{ID: "ccc", Names: "cache", Image: "redis", State: "running"},
	}

	// Search by name
	result := sortAndFilter(containers, SortByName, true, nil, "web")
	if len(result) != 1 || result[0].Names != "web-server" {
		t.Errorf("search by name: expected web-server, got %v", result)
	}

	// Search by image
	result = sortAndFilter(containers, SortByName, true, nil, "redis")
	if len(result) != 1 || result[0].Names != "cache" {
		t.Errorf("search by image: expected cache, got %v", result)
	}

	// Search case-insensitive
	result = sortAndFilter(containers, SortByName, true, nil, "NGINX")
	if len(result) != 1 {
		t.Errorf("case-insensitive search: expected 1, got %d", len(result))
	}

	// No match
	result = sortAndFilter(containers, SortByName, true, nil, "nonexistent")
	if len(result) != 0 {
		t.Errorf("no match: expected 0, got %d", len(result))
	}
}

func TestSortAndFilter_SortByCPU(t *testing.T) {
	containers := []Container{
		{ID: "aaa", Names: "low", State: "running"},
		{ID: "bbb", Names: "high", State: "running"},
	}
	stats := map[string]Stats{
		"aaa": {CPUPercent: 10},
		"bbb": {CPUPercent: 90},
	}
	result := sortAndFilter(containers, SortByCPU, true, stats)
	if result[0].Names != "high" {
		t.Errorf("expected high CPU first, got %s", result[0].Names)
	}
}

func TestSortAndFilter_SortByMem(t *testing.T) {
	containers := []Container{
		{ID: "aaa", Names: "low", State: "running"},
		{ID: "bbb", Names: "high", State: "running"},
	}
	stats := map[string]Stats{
		"aaa": {MemUsage: 100},
		"bbb": {MemUsage: 9000},
	}
	result := sortAndFilter(containers, SortByMem, true, stats)
	if result[0].Names != "high" {
		t.Errorf("expected high memory first, got %s", result[0].Names)
	}
}

func TestSortAndFilterWithCompose(t *testing.T) {
	containers := []Container{
		{ID: "aaa", Names: "web", State: "running", ComposeProject: "myapp"},
		{ID: "bbb", Names: "standalone", State: "running", ComposeProject: ""},
		{ID: "ccc", Names: "db", State: "running", ComposeProject: "myapp"},
		{ID: "ddd", Names: "other-web", State: "running", ComposeProject: "other"},
	}
	result := sortAndFilterWithCompose(containers, SortByName, true, nil, "")

	// myapp containers should be grouped together, other next, standalone last
	if result[0].ComposeProject != "myapp" || result[1].ComposeProject != "myapp" {
		t.Errorf("expected myapp containers first, got %s, %s", result[0].ComposeProject, result[1].ComposeProject)
	}
	if result[2].ComposeProject != "other" {
		t.Errorf("expected 'other' project third, got %s", result[2].ComposeProject)
	}
	if result[3].ComposeProject != "" {
		t.Errorf("expected standalone last, got %s", result[3].ComposeProject)
	}
}

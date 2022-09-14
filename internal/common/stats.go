package common

import (
	"fmt"
	"sync"
	"time"
)

type RunStats struct {
	mu        sync.Mutex
	Date      time.Time
	RepoStats map[string]*RepoStats
}

func NewRunStats() *RunStats {
	r := RunStats{}
	r.Date = time.Now()
	r.RepoStats = make(map[string]*RepoStats)
	return &r

}
func (c *RunStats) Add(repo string) *RepoStats {
	c.mu.Lock()
	r := NewRepoStats(repo)
	r.Name = repo
	// Lock so only one goroutine at a time can access the map c.v.
	c.RepoStats[repo] = r
	c.mu.Unlock()
	return r
}

type RepoStats struct {
	mu   sync.Mutex
	Name string
	//SitemapCount     int
	//SitemapHttpError int
	//SitemapIssues    int
	//SitemapSummoned  int
	counts map[string]int
}

func NewRepoStats(name string) *RepoStats {
	r := RepoStats{Name: name}
	r.counts = make(map[string]int)
	return &r
}

const Count string = "SitemapCount"
const HttpError string = "SitemapHttpError"
const Issues string = "SitemapIssues"
const Summoned string = "SitemapSummoned"
const EmptyDoc string = "SitemapEmptyDoc"
const Stored string = "SitemapStored"
const StoreError string = "SitemapStoredError"
const HeadlessError string = "HeadlessServerError"

// Inc increments the counter for the given key.
func (c *RepoStats) Inc(key string) {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	//_, ok := c.counts[key]
	c.counts[key]++
	c.mu.Unlock()
}

// Inc sets a value for the given key.
func (c *RepoStats) Set(key string, value int) {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	c.counts[key] = value
	c.mu.Unlock()
}

// Value returns the current value of the counter for the given key.
func (c *RepoStats) Value(key string) int {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	defer c.mu.Unlock()
	return c.counts[key]
}
func (c *RunStats) Output() string {
	out := fmt.Sprintln("-------RUN STATS --------")
	out += fmt.Sprintf("Start %s\n", c.Date)
	for name, repo := range c.RepoStats {
		out += fmt.Sprintf("---%s----\n", name)
		for r, count := range repo.counts {
			out += fmt.Sprintf("   %s: %d \n", r, count)
		}
	}
	return out
}

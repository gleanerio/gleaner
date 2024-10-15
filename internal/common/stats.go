package common

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

type RunStats struct {
	mu         sync.Mutex
	Date       time.Time
	StopReason string
	RepoStats  map[string]*RepoStats
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
	r.Start = time.Now()
	// Lock so only one goroutine at a time can access the map c.v.
	c.RepoStats[repo] = r
	c.mu.Unlock()
	return r
}

type RepoStats struct {
	mu    sync.Mutex
	Name  string
	Start time.Time
	End   time.Time
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
func (c *RepoStats) setEndTime() {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	defer c.mu.Unlock()
	c.End = time.Now()
}

const Count string = "SitemapCount"
const HttpError string = "SitemapHttpError"
const Issues string = "SitemapIssues"
const Summoned string = "SitemapSummoned"
const EmptyDoc string = "SitemapEmptyDoc"
const Stored string = "SitemapStored"
const StoreError string = "SitemapStoredError"
const HeadlessError string = "HeadlessServerError"
const NotAuthorized string = "NotAuthorized"
const BadUrl string = "BadURL404"
const RepoServerError string = "RepoServerError"
const GenericIssue = "GenericUrlIssue"

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
	out := fmt.Sprintln("RunStats:")
	out += fmt.Sprintf("  Start: %s\n", c.Date)
	out += fmt.Sprintf("  Reason: %s\n", c.StopReason)
	out += fmt.Sprintf("  Soruce:\n")
	for name, repo := range c.RepoStats {

		out += fmt.Sprintf("    - name: %s\n", name)
		out += fmt.Sprintf("      Start: %s\n", repo.Start)
		out += fmt.Sprintf("      End: %s\n", repo.End)
		for r, count := range repo.counts {
			out += fmt.Sprintf("      %s: %d \n", r, count)
		}
	}
	return out
}

func (c *RepoStats) Output() string {
	c.setEndTime()
	out := fmt.Sprintln("SourceStats:")
	out += fmt.Sprintf("  Start: %s\n", c.Start)
	out += fmt.Sprintf("  End: %s\n", c.End)
	out += fmt.Sprintf("  Soruce:\n")

	out += fmt.Sprintf("    - name: %s\n", c.Name)
	for r, count := range c.counts {
		out += fmt.Sprintf("      %s: %d \n", r, count)
	}

	return out
}

func RunRepoStatsOutput(repoStats *RepoStats, source string) {
	fmt.Print(repoStats.Output())
	const layout = "2006-01-02-15-04-05"
	t := time.Now()
	lf := fmt.Sprintf("%s/repo-%s-stats-%s.log", Logpath, source, t.Format(layout))

	LogFile := lf // log to custom file
	logFile, err := os.OpenFile(LogFile, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	logFile.WriteString(repoStats.Output())
	logFile.Close()
}

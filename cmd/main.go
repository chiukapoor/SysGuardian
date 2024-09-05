package main

import (
	"os"
	"sync"

	"github.com/chiukapoor/SysGuardian/pkg/healthcheck"
	"github.com/jedib0t/go-pretty/table"
)

// collectResults gathers results from all checks
func collectResults() <-chan healthcheck.Result {
	var wg sync.WaitGroup
	results := make(chan healthcheck.Result, 4)

	checks := []healthcheck.Check{
		healthcheck.CheckCPU(),
		healthcheck.CheckDisk(),
		healthcheck.CheckMemory(),
		healthcheck.CheckOS(),
	}

	for _, check := range checks {
		wg.Add(1)
		go func(c healthcheck.Check) {
			defer wg.Done()
			for _, result := range c.GetAll() {
				results <- result
			}
		}(check)
	}

	// Close the results channel when all checks are done
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

// renderResults renders the collected results into a table format
func renderResults(results <-chan healthcheck.Result) {
	// Prepare the table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleColoredDark)
	t.AppendHeader(table.Row{"Resource", "Check", "Status", "Info"})

	// Append results to the table
	for result := range results {
		t.AppendRow(table.Row{result.Component, result.Name, result.Status, result.Info})
	}

	// Render the table
	t.Render()
}

func main() {
	results := collectResults()
	renderResults(results)
}

package healthcheck

import (
	"os/exec"
	"strings"
	"sync"

	"github.com/fatih/color"
)

// Define status levels
const (
	OK      = "OK"
	Warning = "Warning"
	Error   = "Error"
)

// Result is a struct containing Component, Name, Status and Info
type Result struct {
	Component string
	Name      string
	Status    string
	Info      string
}

// Check interface requires the GetAll() function which return a list of Result struct to be implemented
type Check interface {
	GetAll() []Result
}

// Define function signature for checks
type checkFunc func(string, string) Result

type listOfChecks struct {
	component string
	name      string
	fn        checkFunc
}

// runCommand executes the list of string and return a space trimmed output
func runCommand(cmdList ...string) (string, error) {
	cmd := exec.Command(cmdList[0], cmdList[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	usage := strings.TrimSpace(string(output))
	return usage, nil
}

// getStatus returns the Status based on the status name
func getStatus(statusName string) string {
	switch statusName {
	case OK:
		return color.GreenString("OK")
	case Warning:
		return color.YellowString("Warning")
	case Error:
		return color.RedString("Error")
	default:
		return color.HiRedString("Unknown")
	}
}

// runChecks runs a list of checks concurrently and returns the collected results
func runChecks(checks []listOfChecks) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, len(checks))

	for _, check := range checks {
		wg.Add(1)
		go func(component string, name string, fn checkFunc) {
			defer wg.Done()
			result := fn(component, name)
			results <- result
		}(check.component, check.name, check.fn)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	getAllResult := []Result{}
	for result := range results {
		getAllResult = append(getAllResult, result)
	}
	return getAllResult
}

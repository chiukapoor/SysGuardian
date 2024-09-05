package healthcheck

import (
	"fmt"
	"strconv"
	"strings"
)

type CPUCheck struct{}

func CheckCPU() Check {
	return &CPUCheck{}
}

func (c *CPUCheck) GetAll() []Result {
	// Slice of functions to execute
	checks := []listOfChecks{
		{"CPU", "Usage", c.getCPUUsage},
	}
	return runChecks(checks)
}

func (c *CPUCheck) getCPUUsage(componentName string, checkName string) Result {
	// Execute the 'uptime' command
	output, err := runCommand("uptime")
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: err.Error()}
	}

	// Parse the output to extract the 15-minute load average

	fields := strings.Fields(string(output))
	loadAvg15mStr := fields[len(fields)-1] // The 15-minute load average is last

	// Convert the load average to a float
	loadAvg15m, err := strconv.ParseFloat(strings.Trim(loadAvg15mStr, ","), 64)
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: err.Error()}
	}

	// Get the number of CPU cores to calculate the CPU usage percentage
	numCoresOutput, err := runCommand("nproc")
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: err.Error()}
	}
	numCores, err := strconv.Atoi(strings.TrimSpace(string(numCoresOutput)))
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: err.Error()}
	}

	// Calculate CPU usage as a percentage of the load average
	usagePercentage := (loadAvg15m / float64(numCores)) * 100

	// Check if the CPU usage is high
	status := GetStatus("OK")
	if usagePercentage > 95 {
		status = GetStatus("Error")
	}
	if usagePercentage > 80 {
		status = GetStatus("Warning")
	}

	return Result{Component: componentName, Name: checkName, Status: status, Info: fmt.Sprintf("15-minute load average: %.2f, Usage: %.2f%%", loadAvg15m, usagePercentage)}
}

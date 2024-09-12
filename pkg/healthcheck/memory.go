package healthcheck

import (
	"fmt"
	"strconv"
	"strings"
)

type memoryCheck struct{}

// CheckMemory initializes and returns a new instance of the memoryCheck struct which implements the Check interface.
// It provides a standardized way to run these checks through the health check framework.
func CheckMemory() Check {
	return &memoryCheck{}
}

func (m *memoryCheck) GetAll() []Result {
	checks := []listOfChecks{
		{"Memory", "RAM Usage", m.getMemoryUsage},
		{"Memory", "Swap Usage", m.getSwapUsage},
		{"Memory", "Pressure", m.getMemoryPressure},
		{"Memory", "OOM Events", m.getOOMEvents},
	}
	return runChecks(checks)
}

func (m *memoryCheck) getMemoryUsage(componentName string, checkName string) Result {
	output, err := runCommand("free")
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: err.Error()}
	}

	// Parse the output of the 'free' command
	lines := strings.Split(string(output), "\n")

	// Find the line that contains the word "Mem"
	var memLine string
	for _, line := range lines {
		if strings.HasPrefix(line, "Mem:") {
			memLine = line
			break
		}
	}

	// If we didn't find the "Mem:" line, return an error
	if memLine == "" {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: "Mem line not found"}
	}

	// Extract the total and available memory from the second line (the 'Mem:' line)
	fields := strings.Fields(memLine)
	if len(fields) < 7 {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: "Unexpected output format"}
	}

	totalMemory, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: "Error parsing total memory"}
	}

	availableMemory, err := strconv.ParseFloat(fields[6], 64)
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: "Error parsing available memory"}
	}

	// Calculate memory usage percentage
	usedMemory := totalMemory - availableMemory
	usagePercentage := (usedMemory / totalMemory) * 100

	// Determine the status based on the usage percentage
	status := getStatus("OK")
	if usagePercentage > 90 {
		status = getStatus("Error")
	} else if usagePercentage > 75 {
		status = getStatus("Warning")
	}

	return Result{Component: componentName, Name: checkName, Status: status, Info: fmt.Sprintf("Usage: %.2f%%", usagePercentage)}
}

func (m *memoryCheck) getSwapUsage(componentName string, checkName string) Result {
	output, err := runCommand("free")
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: err.Error()}
	}

	// Parse the output of the 'free' command
	// Parse the output of the 'free' command
	lines := strings.Split(string(output), "\n")

	// Find the line that contains the word "Mem"
	var swapLine string
	for _, line := range lines {
		if strings.HasPrefix(line, "Swap:") {
			swapLine = line
			break
		}
	}

	// If we didn't find the "Swap:" line, return an error
	if swapLine == "" {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: "Mem line not found"}
	}

	// Extract the total and available memory from the second line (the 'Swap:' line)
	fields := strings.Fields(swapLine)
	if len(fields) < 3 {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: "Unexpected output format"}
	}

	totalMemory, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: "Error parsing total memory"}
	}

	availableMemory, err := strconv.ParseFloat(fields[3], 64)
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: "Error parsing available memory"}
	}

	// Calculate memory usage percentage
	usedMemory := totalMemory - availableMemory
	usagePercentage := (usedMemory / totalMemory) * 100

	// Determine the status based on the usage percentage
	status := getStatus("OK")
	if usagePercentage > 90 {
		status = getStatus("Error")
	} else if usagePercentage > 75 {
		status = getStatus("Warning")
	}

	return Result{Component: componentName, Name: checkName, Status: status, Info: fmt.Sprintf("Usage: %.2f%%", usagePercentage)}
}

func (m *memoryCheck) getMemoryPressure(componentName string, checkName string) Result {
	output, err := runCommand("cat", "/proc/pressure/memory")
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: err.Error()}
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 1 {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: "Unexpected output format"}
	}

	// The "some" line provides information about memory pressure
	someLine := strings.Fields(lines[0])
	if len(someLine) < 2 {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: "Error parsing memory pressure"}
	}

	pressureValue, err := strconv.ParseFloat(strings.Split(someLine[1], "=")[1], 64)

	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: fmt.Sprintf("Error parsing memory pressure: %s", err.Error())}
	}

	// Determine the status based on the pressure value
	status := getStatus("OK")
	if pressureValue > 20 {
		status = getStatus("Warning")
	}
	if pressureValue > 40 {
		status = getStatus("Error")
	}

	return Result{Component: componentName, Name: checkName, Status: status, Info: fmt.Sprintf("Memory Pressure: %.2f%%", pressureValue)}
}

func (m *memoryCheck) getOOMEvents(componentName string, checkName string) Result {
	output, err := runCommand("dmesg", "-T")
	if err != nil && err.Error() != "exit status 1" {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: err.Error()}
	}

	lines := strings.Split(string(output), "\n")
	var filteredLines []string
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "killed process") {
			filteredLines = append(filteredLines, line)
		}
	}

	// Count the number of OOM events
	oomCount := len(filteredLines)

	status := getStatus("OK")
	if oomCount > 0 {
		status = getStatus("Error")
	}

	return Result{Component: componentName, Name: checkName, Status: status, Info: fmt.Sprintf("OOM Events: %d", oomCount)}
}

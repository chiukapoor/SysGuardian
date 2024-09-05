package healthcheck

import (
	"fmt"
	"strconv"
	"strings"
)

type MemoryCheck struct{}

func CheckMemory() Check {
	return &MemoryCheck{}
}

func (m *MemoryCheck) GetAll() []Result {
	checks := []listOfChecks{
		{"Memory", "RAM Usage", m.getMemoryUsage},
		{"Memory", "Swap Usage", m.getSwapUsage},
		{"Memory", "Pressure", m.getMemoryPressure},
		{"Memory", "OOM Events", m.getOOMEvents},
	}
	return runChecks(checks)
}

func (m *MemoryCheck) getMemoryUsage(componentName string, checkName string) Result {
	output, err := runCommand("free")
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: err.Error()}
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
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: "Mem line not found"}
	}

	// Extract the total and available memory from the second line (the 'Mem:' line)
	fields := strings.Fields(memLine)
	if len(fields) < 7 {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: "Unexpected output format"}
	}

	totalMemory, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: "Error parsing total memory"}
	}

	availableMemory, err := strconv.ParseFloat(fields[6], 64)
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: "Error parsing available memory"}
	}

	// Calculate memory usage percentage
	usedMemory := totalMemory - availableMemory
	usagePercentage := (usedMemory / totalMemory) * 100

	// Determine the status based on the usage percentage
	status := GetStatus("OK")
	if usagePercentage > 90 {
		status = GetStatus("Error")
	} else if usagePercentage > 75 {
		status = GetStatus("Warning")
	}

	return Result{Component: componentName, Name: checkName, Status: status, Info: fmt.Sprintf("Usage: %.2f%%", usagePercentage)}
}

func (m *MemoryCheck) getSwapUsage(componentName string, checkName string) Result {
	output, err := runCommand("free")
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: err.Error()}
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
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: "Mem line not found"}
	}

	// Extract the total and available memory from the second line (the 'Swap:' line)
	fields := strings.Fields(swapLine)
	if len(fields) < 3 {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: "Unexpected output format"}
	}

	totalMemory, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: "Error parsing total memory"}
	}

	availableMemory, err := strconv.ParseFloat(fields[3], 64)
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: "Error parsing available memory"}
	}

	// Calculate memory usage percentage
	usedMemory := totalMemory - availableMemory
	usagePercentage := (usedMemory / totalMemory) * 100

	// Determine the status based on the usage percentage
	status := GetStatus("OK")
	if usagePercentage > 90 {
		status = GetStatus("Error")
	} else if usagePercentage > 75 {
		status = GetStatus("Warning")
	}

	return Result{Component: componentName, Name: checkName, Status: status, Info: fmt.Sprintf("Usage: %.2f%%", usagePercentage)}
}

func (m *MemoryCheck) getMemoryPressure(componentName string, checkName string) Result {
	output, err := runCommand("cat", "/proc/pressure/memory")
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: err.Error()}
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 1 {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: "Unexpected output format"}
	}

	// The "some" line provides information about memory pressure
	someLine := strings.Fields(lines[0])
	if len(someLine) < 2 {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: "Error parsing memory pressure"}
	}

	pressureValue, err := strconv.ParseFloat(strings.Split(someLine[1], "=")[1], 64)

	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: fmt.Sprintf("Error parsing memory pressure: %s", err.Error())}
	}

	// Determine the status based on the pressure value
	status := GetStatus("OK")
	if pressureValue > 20 {
		status = GetStatus("Warning")
	}
	if pressureValue > 40 {
		status = GetStatus("Error")
	}

	return Result{Component: componentName, Name: checkName, Status: status, Info: fmt.Sprintf("Memory Pressure: %.2f%%", pressureValue)}
}

func (m *MemoryCheck) getOOMEvents(componentName string, checkName string) Result {
	output, err := runCommand("dmesg", "-T")
	if err != nil && err.Error() != "exit status 1" {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: err.Error()}
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

	status := GetStatus("OK")
	if oomCount > 0 {
		status = GetStatus("Error")
	}

	return Result{Component: componentName, Name: checkName, Status: status, Info: fmt.Sprintf("OOM Events: %d", oomCount)}
}

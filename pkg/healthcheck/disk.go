package healthcheck

import (
	"fmt"
	"strconv"
	"strings"
)

type diskCheck struct{}

// CheckDisk initializes and returns a new instance of the diskCheck struct which implements the Check interface.
// It provides a standardized way to run these checks through the health check framework.
func CheckDisk() Check {
	return &diskCheck{}
}

// GetAll retrieves a list of disk-related checks that need to be performed and executes them.
func (d *diskCheck) GetAll() []Result {
	checks := []listOfChecks{
		{"Disk", "Root Usage", d.getRootPartitionUsage},
		{"Disk", "Root Inode Usage", d.getFilesystemInodeUsage},
	}
	return runChecks(checks)
}

func (d *diskCheck) getRootPartitionUsage(componentName string, checkName string) Result {
	// Get disk usage information from the df command
	output, err := runCommand("df", "--output=pcent", "/")
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: err.Error()}
	}

	// Parse the output to get used space and available space
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: "Unexpected df output format"}
	}

	usageStr := strings.TrimSpace(lines[1])
	usageStr = strings.Trim(usageStr, "%")
	usagePercentage, err := strconv.Atoi(usageStr)
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: err.Error()}
	}

	// Check if the Disk usage is higher than 90%
	status := getStatus("OK")
	if usagePercentage > 95 {
		status = getStatus("Error")
	}
	if usagePercentage > 75 {
		status = getStatus("Warning")
	}

	return Result{Component: componentName, Name: checkName, Status: status, Info: fmt.Sprintf("Disk Usage: %d%%", usagePercentage)}
}

func (d *diskCheck) getFilesystemInodeUsage(componentName string, checkName string) Result {
	output, err := runCommand("df", "-i")
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: err.Error()}
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return Result{Component: componentName, Name: checkName, Status: getStatus("Error"), Info: "Unexpected output format"}
	}

	var inodeUsagePercentage float64
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		if fields[5] == "/" {
			usage, err := strconv.ParseFloat(strings.TrimSuffix(fields[4], "%"), 64)
			if err != nil {
				continue
			}
			if usage > inodeUsagePercentage {
				inodeUsagePercentage = usage
				break
			}
		}
	}

	status := getStatus("OK")
	if inodeUsagePercentage > 75 {
		status = getStatus("Warning")
	}
	if inodeUsagePercentage > 90 {
		status = getStatus("Error")
	}

	return Result{Component: componentName, Name: checkName, Status: status, Info: fmt.Sprintf("Inode Usage: %.2f%%", inodeUsagePercentage)}
}

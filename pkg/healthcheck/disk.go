package healthcheck

import (
	"fmt"
	"strconv"
	"strings"
)

type DiskCheck struct{}

func CheckDisk() Check {
	return &DiskCheck{}
}

func (d *DiskCheck) GetAll() []Result {
	checks := []listOfChecks{
		{"Disk", "Root Usage", d.getRootPartitionUsage},
		{"Disk", "Root Inode Usage", d.getFilesystemInodeUsage},
	}
	return runChecks(checks)
}

func (d *DiskCheck) getRootPartitionUsage(componentName string, checkName string) Result {
	// Get disk usage information from the df command
	output, err := runCommand("df", "--output=pcent", "/")
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: err.Error()}
	}

	// Parse the output to get used space and available space
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: "Unexpected df output format"}
	}

	usageStr := strings.TrimSpace(lines[1])
	usageStr = strings.Trim(usageStr, "%")
	usagePercentage, err := strconv.Atoi(usageStr)
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: err.Error()}
	}

	// Check if the Disk usage is higher than 90%
	status := GetStatus("OK")
	if usagePercentage > 95 {
		status = GetStatus("Error")
	}
	if usagePercentage > 75 {
		status = GetStatus("Warning")
	}

	return Result{Component: componentName, Name: checkName, Status: status, Info: fmt.Sprintf("Disk Usage: %d%%", usagePercentage)}
}

func (d *DiskCheck) getFilesystemInodeUsage(componentName string, checkName string) Result {
	output, err := runCommand("df", "-i")
	if err != nil {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: err.Error()}
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return Result{Component: componentName, Name: checkName, Status: GetStatus("Error"), Info: "Unexpected output format"}
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

	status := GetStatus("OK")
	if inodeUsagePercentage > 75 {
		status = GetStatus("Warning")
	}
	if inodeUsagePercentage > 90 {
		status = GetStatus("Error")
	}

	return Result{Component: componentName, Name: checkName, Status: status, Info: fmt.Sprintf("Inode Usage: %.2f%%", inodeUsagePercentage)}
}

package healthcheck

import (
	"fmt"
	"time"

	"github.com/beevik/ntp"
)

type osCheck struct{}

// CheckOS initializes and returns a new instance of the osCheck struct which implements the Check interface.
// It provides a standardized way to run these checks through the health check framework.
func CheckOS() Check {
	return &osCheck{}
}

// GetAll retrieves a list of os-related checks that need to be performed and executes them.
func (o *osCheck) GetAll() []Result {
	checks := []listOfChecks{
		{"OS", "Time Synchronization", o.getTimeSynchronization},
	}
	return runChecks(checks)
}

func (o *osCheck) getTimeSynchronization(componentName string, checkName string) Result {
	server := "pool.ntp.org" // Replace with your preferred NTP server

	// Get time from the NTP server
	ntpTime, err := ntp.Time(server)
	if err != nil {
		return Result{Component: "OS", Name: "NTP Time Synchronization", Status: getStatus("Error"), Info: fmt.Sprintf("Error querying NTP server: %s", err)}
	}

	// Get local system time
	localTime := time.Now()

	// Calculate the difference
	timeDiff := ntpTime.Sub(localTime)
	const tolerance = 2 * time.Second

	status := getStatus("OK")
	if timeDiff > tolerance || timeDiff < -tolerance {
		status = getStatus("Warning")
	}

	return Result{Component: "OS", Name: "NTP Time Synchronization", Status: status, Info: fmt.Sprintf("NTP Time: %s, Local Time: %s, Difference: %v", ntpTime.Format(time.RFC3339), localTime.Format(time.RFC3339), timeDiff)}
}

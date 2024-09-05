package healthcheck

import (
	"fmt"
	"time"

	"github.com/beevik/ntp"
)

type OSCheck struct{}

func CheckOS() Check {
	return &OSCheck{}
}

func (o *OSCheck) GetAll() []Result {
	checks := []listOfChecks{
		{"OS", "Time Synchronization", o.getTimeSynchronization},
	}
	return runChecks(checks)
}

func (o *OSCheck) getTimeSynchronization(componentName string, checkName string) Result {
	server := "pool.ntp.org" // Replace with your preferred NTP server

	// Get time from the NTP server
	ntpTime, err := ntp.Time(server)
	if err != nil {
		return Result{Component: "OS", Name: "NTP Time Synchronization", Status: GetStatus("Error"), Info: fmt.Sprintf("Error querying NTP server: %s", err)}
	}

	// Get local system time
	localTime := time.Now()

	// Calculate the difference
	timeDiff := ntpTime.Sub(localTime)
	const tolerance = 2 * time.Second

	status := GetStatus("OK")
	if timeDiff > tolerance || timeDiff < -tolerance {
		status = GetStatus("Warning")
	}

	return Result{Component: "OS", Name: "NTP Time Synchronization", Status: status, Info: fmt.Sprintf("NTP Time: %s, Local Time: %s, Difference: %v", ntpTime.Format(time.RFC3339), localTime.Format(time.RFC3339), timeDiff)}
}

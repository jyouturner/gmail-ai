package main

import (
	"fmt"
	"os/exec"
)

func LaunchCtl() {
	// Create a new launchd job with a start interval of 60 seconds
	cmd := exec.Command("launchctl", "submit", "-l", "com.example.myprogram", "-i", "60", "/path/to/my/program")
	if err := cmd.Run(); err != nil {
		fmt.Println("Error creating job:", err)
		return
	}

	fmt.Println("Job created successfully!")
}

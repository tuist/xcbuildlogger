package main

import (
	"io"
	"log"
	"os"
	"os/exec"
)

func main() {
	// Path to XCBBuildService
	xcbuildServicePath := "/Applications/Xcode.app/Contents/SharedFrameworks/XCBuild.framework/PlugIns/XCBBuildService.bundle/Contents/MacOS/XCBBuildService"

	// Path to log file (modify if needed)
	logFilePath := "/tmp/xcodebuildlogging.txt"

	// Open the log file for writing
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Create a new command for XCBBuildService
	cmd := exec.Command(xcbuildServicePath)
	cmd.Stdin = os.Stdin                            // Forward stdin
	cmd.Stdout = io.MultiWriter(os.Stdout, logFile) // Forward and log stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, logFile) // Forward and log stderr

	// Start the XCBBuildService process
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start XCBBuildService: %v", err)
	}

	// Wait for the process to finish
	if err := cmd.Wait(); err != nil {
		log.Printf("XCBBuildService exited with error: %v", err)
	}
}

package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// RPCPacket represents a decoded message
type RPCPacket struct {
	Channel uint64 `json:"channel"`
	Payload string `json:"payload"`
}

func main() {
	xcbuildServicePath := "/Applications/Xcode.app/Contents/SharedFrameworks/XCBuild.framework/PlugIns/XCBBuildService.bundle/Contents/MacOS/XCBBuildService"

	// Open log file
	logFilePath := "/tmp/xcode_xcbbuildservice.log"
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	cmd := exec.Command(xcbuildServicePath)

	// Create pipes
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("Failed to create stdin pipe: %v", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to create stdout pipe: %v", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("Failed to create stderr pipe: %v", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start XCBBuildService: %v", err)
	}

	// Handle signals and forward them to the child process
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range signalChan {
			cmd.Process.Signal(sig)
		}
	}()

	// Handle Xcode -> XCBBuildService messages (stdin)
	go func() {
		decodeAndForwardStdin(os.Stdin, stdinPipe, logFile)
	}()

	// Handle XCBBuildService -> Xcode messages (stdout & stderr)
	go decodeAndForwardStdout(stdoutPipe, os.Stdout, logFile)
	go decodeAndForwardStdout(stderrPipe, os.Stderr, logFile)

	// Wait for the process to finish and handle termination properly
	go func() {
		// Wait for the underlying process to exit
		err := cmd.Wait()
		if err != nil {
			logToFile(logFile, fmt.Sprintf("XCBBuildService exited with error: %v", err))
		}
		close(signalChan) // Close the signal channel to prevent further signals after process exits
	}()

	// Block main goroutine until the process finishes
	<-signalChan
	log.Println("Main process is terminating")
}

func decodeAndForwardStdout(input io.Reader, output io.Writer, logFile *os.File) {
	for {
		packet, rawData, err := readRPCPacket(input)
		if err == io.EOF {
			break
		}
		if err != nil {
			logToFile(logFile, fmt.Sprintf("Error reading RPCPacket from stdout: %v", err))
			continue
		}

		// Log decoded packet with full metadata
		logDecodedPacket(logFile, packet)

		// Forward raw data to Xcode (or wherever necessary)
		_, err = output.Write(rawData)
		if err != nil {
			logToFile(logFile, fmt.Sprintf("Error forwarding data from stdout: %v", err))
			break
		}
	}
}

// decodeAndForwardStdin intercepts, decodes, logs, and forwards stdin messages
func decodeAndForwardStdin(input io.Reader, output io.Writer, logFile *os.File) {
	for {
		packet, rawData, err := readRPCPacket(input)
		if err == io.EOF {
			break
		}
		if err != nil {
			logToFile(logFile, fmt.Sprintf("Error reading RPCPacket: %v", err))
			continue
		}

		// Log decoded packet with full metadata
		logDecodedPacket(logFile, packet)

		// Forward raw data to XCBBuildService
		_, err = output.Write(rawData)
		if err != nil {
			logToFile(logFile, fmt.Sprintf("Error forwarding data to XCBBuildService: %v", err))
			break
		}
	}
}
func readRPCPacket(reader io.Reader) (*RPCPacket, []byte, error) {
	// Read the 12-byte header (8-byte channel + 4-byte payload size)
	header := make([]byte, 12)
	_, err := io.ReadFull(reader, header)
	if err != nil {
		return nil, nil, err
	}

	// Parse header fields
	channel := binary.LittleEndian.Uint64(header[:8])
	payloadSize := binary.LittleEndian.Uint32(header[8:12])

	// Read the payload
	payload := make([]byte, payloadSize)
	_, err = io.ReadFull(reader, payload)
	if err != nil {
		return nil, nil, err
	}

	// Return the fully decoded structure
	return &RPCPacket{
		Channel: channel,
		Payload: string(payload),
	}, append(header, payload...), nil
}

// logDecodedPacket writes a decoded RPCPacket with metadata
func logDecodedPacket(logFile *os.File, packet *RPCPacket) {
	packetJSON, _ := json.MarshalIndent(packet, "", "  ") // Pretty-print JSON
	logToFile(logFile, string(packetJSON))                // Convert byte slice to string directly
}

// logToFile writes log messages safely
func logToFile(logFile *os.File, message string) {
	logFile.WriteString(message + "\n")
	logFile.Sync() // Ensure it's written to disk
}
